package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login via browser and save credentials",
	Long: `Start a device-code login flow, open browser for web authorization,
poll for API key token, and save credentials to config file.`,
	RunE: runLogin,
}

var (
	loginTimeoutSec int
	loginNoBrowser  bool
	deviceStartPath string
	devicePollPath  string
)

type loginResponse struct {
	BaseURL    string `json:"base_url"`
	ConfigFile string `json:"config_file"`
}

type deviceStartResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

type devicePollResponse struct {
	AccessToken       string `json:"access_token"`
	TokenType         string `json:"token_type"`
	RetryAfterSeconds int    `json:"retry_after_seconds"`
	Error             string `json:"error"`
}

type devicePollError struct {
	Code       string
	Message    string
	RetryAfter time.Duration
	Fatal      bool
}

func (e *devicePollError) Error() string {
	if e == nil {
		return ""
	}
	if strings.TrimSpace(e.Message) != "" {
		return e.Message
	}
	if strings.TrimSpace(e.Code) != "" {
		return e.Code
	}
	return "device poll failed"
}

func init() {
	rootCmd.AddCommand(loginCmd)

	loginCmd.Flags().IntVar(&loginTimeoutSec, "timeout", 180, "Login timeout in seconds")
	loginCmd.Flags().BoolVar(&loginNoBrowser, "no-browser", false, "Do not auto-open browser; only print verification URL")
	loginCmd.Flags().StringVar(&deviceStartPath, "device-start-path", "/api/auth/device/start", "Device login start API path or full URL")
	loginCmd.Flags().StringVar(&devicePollPath, "device-poll-path", "/api/auth/device/poll", "Device login poll API path or full URL")
}

func runLogin(cmd *cobra.Command, args []string) error {
	if loginTimeoutSec <= 0 {
		return newCLIError("invalid_argument", "--timeout must be greater than 0", 1, nil)
	}

	base := strings.TrimSpace(viper.GetString("base_url"))
	if base == "" {
		return newCLIError("missing_base_url", "base URL is required (use --base-url or set ROBOTX_BASE_URL)", 1, nil)
	}
	base = strings.TrimRight(base, "/")

	startURL, err := resolveEndpoint(base, strings.TrimSpace(deviceStartPath))
	if err != nil {
		return newCLIError("invalid_argument", "invalid --device-start-path", 1, err)
	}
	pollURL, err := resolveEndpoint(base, strings.TrimSpace(devicePollPath))
	if err != nil {
		return newCLIError("invalid_argument", "invalid --device-poll-path", 1, err)
	}

	logf("🔐 Starting RobotX device login flow...\n")
	startResp, err := startDeviceLogin(startURL)
	if err != nil {
		return newCLIError("login_start_failed", "failed to start device login", 2, err)
	}
	if strings.TrimSpace(startResp.DeviceCode) == "" {
		return newCLIError("login_start_failed", "device login response missing device_code", 2, nil)
	}

	verificationURL := buildVerificationURL(base, startResp)
	if verificationURL == "" {
		return newCLIError("login_start_failed", "device login response missing verification URL", 2, nil)
	}

	logf("🧾 User Code: %s\n", valueOrDash(startResp.UserCode))
	logf("🌐 Verification URL: %s\n", verificationURL)
	if loginNoBrowser {
		logf("🧭 Open the URL above in your browser and complete login.\n")
	} else if err := openBrowser(verificationURL); err != nil {
		logf("⚠️  Failed to open browser automatically: %v\n", err)
		logf("🧭 Open the URL above in your browser and complete login.\n")
	} else {
		logf("🧭 Browser opened. Complete login to continue...\n")
	}

	interval := time.Duration(startResp.Interval) * time.Second
	if interval <= 0 {
		interval = 5 * time.Second
	}

	logf("⏳ Waiting for authorization...\n")
	apiKey, err := pollForDeviceToken(pollURL, startResp.DeviceCode, interval, time.Duration(loginTimeoutSec)*time.Second)
	if err != nil {
		return newCLIError("login_failed", "device login failed", 2, err)
	}

	configPath, err := resolveConfigWritePath()
	if err != nil {
		return newCLIError("config_error", "failed to resolve config path", 1, err)
	}
	if err := writeCredentialsToConfig(configPath, base, apiKey); err != nil {
		return newCLIError("config_write_failed", "failed to write credentials to config", 1, err)
	}

	logf("✅ Login successful. Credentials saved to: %s\n", configPath)
	if err := emitSuccess(cmd.Name(), loginResponse{
		BaseURL:    base,
		ConfigFile: configPath,
	}); err != nil {
		return newCLIError("output_error", "failed to render JSON output", 1, err)
	}
	return nil
}

func startDeviceLogin(startURL string) (*deviceStartResponse, error) {
	req, err := http.NewRequest(http.MethodPost, startURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create device-start request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	httpClient := &http.Client{Timeout: 20 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("device-start request failed: %w", err)
	}
	defer resp.Body.Close()

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read device-start response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("device-start API error (status %d): %s", resp.StatusCode, compactForError(rawBody))
	}

	var out deviceStartResponse
	if err := json.Unmarshal(rawBody, &out); err != nil {
		return nil, fmt.Errorf("failed to decode device-start response: %w", err)
	}
	return &out, nil
}

func pollForDeviceToken(pollURL, deviceCode string, interval, timeout time.Duration) (string, error) {
	deadline := time.Now().Add(timeout)
	for {
		if time.Now().After(deadline) {
			return "", fmt.Errorf("login timed out after %d seconds", int(timeout.Seconds()))
		}

		token, err := pollDeviceToken(pollURL, deviceCode)
		if err == nil {
			if strings.TrimSpace(token) == "" {
				return "", fmt.Errorf("device poll succeeded but no access token found")
			}
			return strings.TrimSpace(token), nil
		}

		var pollErr *devicePollError
		if !errors.As(err, &pollErr) {
			return "", err
		}
		code := strings.TrimSpace(pollErr.Code)
		switch code {
		case "authorization_pending":
			if !sleepUntilDeadline(deadline, interval) {
				return "", fmt.Errorf("login timed out after %d seconds", int(timeout.Seconds()))
			}
			continue
		case "slow_down":
			waitFor := pollErr.RetryAfter
			if waitFor <= 0 {
				waitFor = interval + 2*time.Second
			}
			if !sleepUntilDeadline(deadline, waitFor) {
				return "", fmt.Errorf("login timed out after %d seconds", int(timeout.Seconds()))
			}
			continue
		default:
			if pollErr.Fatal {
				return "", err
			}
			if !sleepUntilDeadline(deadline, interval) {
				return "", fmt.Errorf("login timed out after %d seconds", int(timeout.Seconds()))
			}
		}
	}
}

func pollDeviceToken(pollURL, deviceCode string) (string, error) {
	payload := map[string]string{
		"device_code": strings.TrimSpace(deviceCode),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to encode device poll payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, pollURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create device poll request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	httpClient := &http.Client{Timeout: 20 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("device poll request failed: %w", err)
	}
	defer resp.Body.Close()

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read device poll response: %w", err)
	}

	var parsed devicePollResponse
	_ = json.Unmarshal(rawBody, &parsed)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		token := strings.TrimSpace(firstNonEmpty(
			strings.TrimSpace(parsed.AccessToken),
			extractAPIKey(rawBody),
		))
		if token == "" {
			return "", fmt.Errorf("device poll response missing access token: %s", compactForError(rawBody))
		}
		return token, nil
	}

	code := strings.TrimSpace(parsed.Error)
	if code == "" {
		code = fmt.Sprintf("http_%d", resp.StatusCode)
	}
	retryAfter := time.Duration(parsed.RetryAfterSeconds) * time.Second
	if retryAfter < 0 {
		retryAfter = 0
	}

	fatal := true
	if code == "authorization_pending" || code == "slow_down" {
		fatal = false
	}
	return "", &devicePollError{
		Code:       code,
		Message:    fmt.Sprintf("device poll failed (%s): %s", code, compactForError(rawBody)),
		RetryAfter: retryAfter,
		Fatal:      fatal,
	}
}

func buildVerificationURL(baseURL string, startResp *deviceStartResponse) string {
	if startResp == nil {
		return ""
	}
	verifyComplete := strings.TrimSpace(startResp.VerificationURIComplete)
	if verifyComplete != "" {
		return resolveURLAgainstBase(baseURL, verifyComplete)
	}

	verify := strings.TrimSpace(startResp.VerificationURI)
	if verify == "" {
		return ""
	}
	full := resolveURLAgainstBase(baseURL, verify)
	userCode := strings.TrimSpace(startResp.UserCode)
	if userCode == "" {
		return full
	}
	u, err := url.Parse(full)
	if err != nil {
		return full
	}
	q := u.Query()
	if strings.TrimSpace(q.Get("user_code")) == "" {
		q.Set("user_code", userCode)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func resolveEndpoint(baseURL, pathOrURL string) (string, error) {
	candidate := strings.TrimSpace(pathOrURL)
	if candidate == "" {
		return "", fmt.Errorf("empty path")
	}
	if strings.HasPrefix(candidate, "http://") || strings.HasPrefix(candidate, "https://") {
		u, err := url.Parse(candidate)
		if err != nil || u.Scheme == "" || u.Host == "" {
			return "", fmt.Errorf("invalid URL: %s", candidate)
		}
		return u.String(), nil
	}
	return strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(candidate, "/"), nil
}

func resolveURLAgainstBase(baseURL, maybeRelative string) string {
	target := strings.TrimSpace(maybeRelative)
	if target == "" {
		return ""
	}
	if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
		return target
	}
	base, err := url.Parse(strings.TrimRight(baseURL, "/"))
	if err != nil || base.Scheme == "" || base.Host == "" {
		return target
	}
	ref, err := url.Parse(target)
	if err != nil {
		return target
	}
	return base.ResolveReference(ref).String()
}

func sleepUntilDeadline(deadline time.Time, waitFor time.Duration) bool {
	if waitFor <= 0 {
		waitFor = time.Second
	}
	remaining := time.Until(deadline)
	if remaining <= 0 {
		return false
	}
	if waitFor > remaining {
		waitFor = remaining
	}
	time.Sleep(waitFor)
	return true
}

func extractAPIKey(raw []byte) string {
	var decoded interface{}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return ""
	}
	return findCredential(decoded)
}

func findCredential(node interface{}) string {
	switch v := node.(type) {
	case map[string]interface{}:
		for _, key := range []string{"api_key", "apiKey", "token", "access_token", "accessToken"} {
			if raw, ok := v[key]; ok {
				if s, ok := raw.(string); ok && strings.TrimSpace(s) != "" {
					return strings.TrimSpace(s)
				}
			}
		}
		for _, key := range []string{"data", "result", "credential", "credentials"} {
			if child, ok := v[key]; ok {
				if extracted := findCredential(child); extracted != "" {
					return extracted
				}
			}
		}
		for _, child := range v {
			if extracted := findCredential(child); extracted != "" {
				return extracted
			}
		}
	case []interface{}:
		for _, child := range v {
			if extracted := findCredential(child); extracted != "" {
				return extracted
			}
		}
	}
	return ""
}

func compactForError(raw []byte) string {
	s := strings.TrimSpace(string(raw))
	if s == "" {
		return "(empty response body)"
	}
	if len(s) > 240 {
		return s[:240] + "..."
	}
	return s
}

func resolveConfigWritePath() (string, error) {
	if strings.TrimSpace(cfgFile) != "" {
		return strings.TrimSpace(cfgFile), nil
	}
	return resolveDefaultConfigPath()
}

func writeCredentialsToConfig(path, baseURL, apiKey string) error {
	cfg := map[string]interface{}{}
	existing, err := os.ReadFile(path)
	if err == nil {
		if len(bytes.TrimSpace(existing)) > 0 {
			if unmarshalErr := yaml.Unmarshal(existing, &cfg); unmarshalErr != nil {
				return fmt.Errorf("failed to parse existing config: %w", unmarshalErr)
			}
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to read existing config: %w", err)
	}
	if cfg == nil {
		cfg = map[string]interface{}{}
	}
	cfg["base_url"] = strings.TrimSpace(baseURL)
	cfg["api_key"] = strings.TrimSpace(apiKey)

	out, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to encode config YAML: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("failed to ensure config directory: %w", err)
	}
	if err := os.WriteFile(path, out, 0o600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	return nil
}

func openBrowser(target string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", target)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", target)
	default:
		cmd = exec.Command("xdg-open", target)
	}
	return cmd.Start()
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
