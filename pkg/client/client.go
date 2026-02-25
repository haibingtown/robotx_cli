package client

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Project represents a RobotX project
type Project struct {
	ProjectID  string    `json:"project_id"`
	Name       string    `json:"name"`
	Visibility string    `json:"visibility"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// BuildPlan describes detected build instructions from server-side scanning.
type BuildPlan struct {
	Strategy       string   `json:"strategy,omitempty"`
	NeedsBuild     bool     `json:"needs_build"`
	ProjectType    string   `json:"project_type,omitempty"`
	PackageManager string   `json:"package_manager,omitempty"`
	InstallCommand string   `json:"install_command,omitempty"`
	BuildCommand   string   `json:"build_command,omitempty"`
	OutputDir      string   `json:"output_dir,omitempty"`
	NodeVersion    string   `json:"node_version,omitempty"`
	RuntimeImage   string   `json:"runtime_image,omitempty"`
	Notes          []string `json:"notes,omitempty"`
}

// ScannerResult mirrors server-side scanning results attached to commits.
type ScannerResult struct {
	BuildPlan *BuildPlan `json:"build_plan,omitempty"`
}

// SourceCommit represents an uploaded source bundle.
type SourceCommit struct {
	CommitID      string         `json:"commit_id"`
	ProjectID     string         `json:"project_id"`
	ScannerResult *ScannerResult `json:"scanner_result,omitempty"`
}

// Build represents a build task
type Build struct {
	BuildID           string     `json:"build_id"`
	ProjectID         string     `json:"project_id"`
	CommitID          string     `json:"commit_id"`
	Status            string     `json:"status"`
	RuntimeArtifactID string     `json:"runtime_artifact_id,omitempty"`
	ErrorMsg          string     `json:"error_msg,omitempty"`
	PreviewPath       string     `json:"preview_path,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	FinishedAt        *time.Time `json:"finished_at,omitempty"`
}

// CreateProjectRequest represents project creation request
type CreateProjectRequest struct {
	Name       string `json:"name"`
	Visibility string `json:"visibility,omitempty"`
}

// CreateProject creates a new project
func (c *Client) CreateProject(req CreateProjectRequest) (*Project, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.doRequest("POST", "/api/projects", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, c.parseError(resp)
	}

	var project Project
	if err := json.NewDecoder(resp.Body).Decode(&project); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &project, nil
}

// GetProject retrieves project information
func (c *Client) GetProject(projectID string) (*Project, error) {
	resp, err := c.doRequest("GET", fmt.Sprintf("/api/projects/%s", projectID), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var project Project
	if err := json.NewDecoder(resp.Body).Decode(&project); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &project, nil
}

// UploadSource uploads source code and creates a commit/build.
func (c *Client) UploadSource(projectID, sourcePath string) (*SourceCommit, *Build, error) {
	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add file
	file, err := os.Open(sourcePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open source file: %w", err)
	}
	defer file.Close()

	part, err := writer.CreateFormFile("file", filepath.Base(sourcePath))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return nil, nil, fmt.Errorf("failed to copy file: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, nil, fmt.Errorf("failed to close writer: %w", err)
	}

	// Create request
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/projects/%s/commits", c.baseURL, projectID), body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to upload source: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		if resp.StatusCode == http.StatusAccepted {
			// Continue parsing for APIs that accept upload asynchronously.
		} else {
			return nil, nil, c.parseError(resp)
		}
	}

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result struct {
		Commit   *SourceCommit `json:"commit"`
		Build    *Build        `json:"build"`
		CommitID string        `json:"commit_id"`
	}
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &result); err != nil {
			// Try common wrapped payload structure: {"data": {...}}
			var wrapped struct {
				Data struct {
					Commit   *SourceCommit `json:"commit"`
					Build    *Build        `json:"build"`
					CommitID string        `json:"commit_id"`
				} `json:"data"`
			}
			if err2 := json.Unmarshal(rawBody, &wrapped); err2 == nil {
				result.Commit = wrapped.Data.Commit
				result.Build = wrapped.Data.Build
				result.CommitID = wrapped.Data.CommitID
			} else {
				return nil, nil, fmt.Errorf("failed to decode response: %w", err)
			}
		}
	}

	if result.Commit == nil && result.CommitID != "" {
		result.Commit = &SourceCommit{CommitID: result.CommitID, ProjectID: projectID}
	}

	// Some APIs return a top-level build_id without build object.
	if result.Build == nil && len(rawBody) > 0 {
		var fallback struct {
			BuildID string `json:"build_id"`
			Data    struct {
				BuildID string `json:"build_id"`
			} `json:"data"`
		}
		if err := json.Unmarshal(rawBody, &fallback); err == nil {
			buildID := strings.TrimSpace(fallback.BuildID)
			if buildID == "" {
				buildID = strings.TrimSpace(fallback.Data.BuildID)
			}
			if buildID != "" {
				result.Build = &Build{BuildID: buildID, ProjectID: projectID}
			}
		}
	}

	return result.Commit, result.Build, nil
}

// TriggerBuild triggers a build for a commit (legacy API).
func (c *Client) TriggerBuild(projectID, commitID string) (*Build, error) {
	body, err := json.Marshal(map[string]string{
		"commit_id": commitID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.doRequest("POST", fmt.Sprintf("/api/projects/%s/builds", projectID), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, c.parseError(resp)
	}

	var build Build
	if err := json.NewDecoder(resp.Body).Decode(&build); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &build, nil
}

// StartBuild requests the server to start a build by build ID.
func (c *Client) StartBuild(projectID, buildID string) error {
	body, err := json.Marshal(map[string]string{
		"project_id": projectID,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}
	resp, err := c.doRequest("POST", fmt.Sprintf("/api/builds/%s/start", buildID), bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return c.parseError(resp)
	}
	return nil
}

// GetBuild retrieves build information.
func (c *Client) GetBuild(projectID, buildID string) (*Build, error) {
	resp, err := c.doRequest("GET", fmt.Sprintf("/api/builds/%s", buildID), nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusNotFound && projectID != "" {
		resp.Body.Close()
		resp, err = c.doRequest("GET", fmt.Sprintf("/api/projects/%s/builds/%s", projectID, buildID), nil)
		if err != nil {
			return nil, err
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var build Build
	if err := json.NewDecoder(resp.Body).Decode(&build); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &build, nil
}

// PublishBuild publishes a build to production
func (c *Client) PublishBuild(projectID, buildID string) (string, error) {
	body, err := json.Marshal(map[string]string{
		"build_id": buildID,
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.doRequest("POST", fmt.Sprintf("/api/projects/%s/publish", projectID), bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", c.parseError(resp)
	}

	var result struct {
		PublicPath string `json:"public_path"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err == nil {
		return result.PublicPath, nil
	}
	return "", nil
}

// GetBuildLogs retrieves build logs
func (c *Client) GetBuildLogs(projectID, buildID string) (string, error) {
	resp, err := c.doRequest("GET", fmt.Sprintf("/api/builds/%s/logs/stream", buildID), nil)
	if err != nil {
		return "", err
	}
	if resp.StatusCode == http.StatusNotFound && projectID != "" {
		resp.Body.Close()
		resp, err = c.doRequest("GET", fmt.Sprintf("/api/projects/%s/builds/%s/logs", projectID, buildID), nil)
		if err != nil {
			return "", err
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", c.parseError(resp)
	}

	logs, err := readSSE(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read logs: %w", err)
	}
	return logs, nil
}

// UploadBuildArtifacts uploads a zip of build outputs for a given build.
func (c *Client) UploadBuildArtifacts(buildID, zipPath string) (*Build, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	file, err := os.Open(zipPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open artifact file: %w", err)
	}
	defer file.Close()

	part, err := writer.CreateFormFile("file", filepath.Base(zipPath))
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("failed to copy file: %w", err)
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/builds/%s/artifacts", c.baseURL, buildID), body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to upload artifacts: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		return nil, c.parseError(resp)
	}

	var build Build
	if err := json.NewDecoder(resp.Body).Decode(&build); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &build, nil
}

func readSSE(r io.Reader) (string, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	var lines []string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			lines = append(lines, strings.TrimPrefix(line, "data: "))
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return strings.Join(lines, "\n"), nil
}

func (c *Client) doRequest(method, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, c.baseURL+path, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

func (c *Client) parseError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)
	var errResp struct {
		Error   interface{} `json:"error"`
		Message string      `json:"message"`
		Detail  string      `json:"detail"`
		Code    string      `json:"code"`
	}
	if err := json.Unmarshal(body, &errResp); err == nil {
		msg := strings.TrimSpace(errResp.Message)
		if msg == "" {
			msg = strings.TrimSpace(errResp.Detail)
		}
		if msg == "" {
			switch v := errResp.Error.(type) {
			case string:
				msg = strings.TrimSpace(v)
			case map[string]interface{}:
				for _, key := range []string{"message", "detail", "error", "msg"} {
					if raw, ok := v[key]; ok {
						if s, ok := raw.(string); ok && strings.TrimSpace(s) != "" {
							msg = strings.TrimSpace(s)
							break
						}
					}
				}
			}
		}

		if msg != "" {
			if strings.TrimSpace(errResp.Code) != "" {
				return fmt.Errorf("API error (status %d, code %s): %s", resp.StatusCode, strings.TrimSpace(errResp.Code), msg)
			}
			return fmt.Errorf("API error (status %d): %s", resp.StatusCode, msg)
		}
	}

	trimmedBody := strings.TrimSpace(string(body))
	if trimmedBody == "" {
		return fmt.Errorf("API error: status %d", resp.StatusCode)
	}
	return fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, trimmedBody)
}
