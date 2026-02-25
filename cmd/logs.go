package cmd

import (
	"fmt"
	"strings"

	"github.com/haibingtown/robotx_cli/pkg/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var logsCmd = &cobra.Command{
	Use:   "logs [build-id]",
	Short: "Get build logs",
	Long:  "Get logs for a specific build.",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runLogs,
}

var (
	logsProjectID string
	logsBuildID   string
	logsFollow    bool
)

type logsResponse struct {
	ProjectID string `json:"project_id,omitempty"`
	BuildID   string `json:"build_id"`
	Logs      string `json:"logs"`
}

func init() {
	rootCmd.AddCommand(logsCmd)

	logsCmd.Flags().StringVarP(&logsProjectID, "project-id", "p", "", "Project ID (optional)")
	logsCmd.Flags().StringVarP(&logsBuildID, "build-id", "b", "", "Build ID")
	logsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "Follow logs in realtime (not implemented yet)")
}

func runLogs(cmd *cobra.Command, args []string) error {
	buildID := strings.TrimSpace(logsBuildID)
	if len(args) > 0 {
		if buildID != "" {
			return newCLIError("invalid_argument", "build ID provided both as arg and --build-id", 1, nil)
		}
		buildID = strings.TrimSpace(args[0])
	}
	if buildID == "" {
		return newCLIError("missing_argument", "build ID is required (use positional arg or --build-id)", 1, nil)
	}
	if logsFollow {
		return newCLIError("not_implemented", "--follow is not implemented yet", 1, nil)
	}

	baseURL := viper.GetString("base_url")
	apiKey := viper.GetString("api_key")
	if baseURL == "" {
		return newCLIError("missing_base_url", "base URL is required", 1, nil)
	}
	if apiKey == "" {
		return newCLIError("missing_api_key", "API key is required", 1, nil)
	}

	c := client.NewClient(baseURL, apiKey)
	logs, err := c.GetBuildLogs(logsProjectID, buildID)
	if err != nil {
		return newCLIError("api_error", "failed to get logs", 2, err)
	}

	if err := emitSuccess(cmd.Name(), logsResponse{
		ProjectID: logsProjectID,
		BuildID:   buildID,
		Logs:      logs,
	}); err != nil {
		return newCLIError("output_error", "failed to render JSON output", 1, err)
	}
	if isJSONOutput() {
		return nil
	}

	if logs == "" {
		fmt.Println("(no logs)")
		return nil
	}
	fmt.Println(logs)
	return nil
}
