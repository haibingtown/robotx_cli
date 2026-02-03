package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"haibingtown/robotx_cli/pkg/client"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get project or build status",
	Long:  `Get the status of a project or specific build.`,
	RunE:  runStatus,
}

var (
	statusProjectID string
	statusBuildID   string
	showLogs        bool
)

func init() {
	rootCmd.AddCommand(statusCmd)

	statusCmd.Flags().StringVarP(&statusProjectID, "project-id", "p", "", "Project ID (required)")
	statusCmd.Flags().StringVarP(&statusBuildID, "build-id", "b", "", "Build ID (optional)")
	statusCmd.Flags().BoolVarP(&showLogs, "logs", "l", false, "Show build logs")
	statusCmd.MarkFlagRequired("project-id")
}

func runStatus(cmd *cobra.Command, args []string) error {
	baseURL := viper.GetString("base_url")
	apiKey := viper.GetString("api_key")

	if baseURL == "" {
		return fmt.Errorf("base URL is required")
	}
	if apiKey == "" {
		return fmt.Errorf("API key is required")
	}

	c := client.NewClient(baseURL, apiKey)

	// Get project info
	fmt.Printf("📦 Fetching project information...\n")
	project, err := c.GetProject(statusProjectID)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	// Display project info
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "\n📋 Project Information:\n")
	fmt.Fprintf(w, "ID:\t%s\n", project.ProjectID)
	fmt.Fprintf(w, "Name:\t%s\n", project.Name)
	fmt.Fprintf(w, "Visibility:\t%s\n", project.Visibility)
	fmt.Fprintf(w, "Created:\t%s\n", project.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(w, "Updated:\t%s\n", project.UpdatedAt.Format("2006-01-02 15:04:05"))
	w.Flush()

	// Get build info if specified
	if statusBuildID != "" {
		fmt.Printf("\n🔨 Fetching build information...\n")
		build, err := c.GetBuild(statusProjectID, statusBuildID)
		if err != nil {
			return fmt.Errorf("failed to get build: %w", err)
		}

		fmt.Fprintf(w, "\n📋 Build Information:\n")
		fmt.Fprintf(w, "ID:\t%s\n", build.BuildID)
		fmt.Fprintf(w, "Status:\t%s\n", build.Status)
		fmt.Fprintf(w, "Commit:\t%s\n", build.CommitID)
		fmt.Fprintf(w, "Created:\t%s\n", build.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Fprintf(w, "Updated:\t%s\n", build.UpdatedAt.Format("2006-01-02 15:04:05"))
		w.Flush()

		// Show logs if requested
		if showLogs {
			fmt.Printf("\n📋 Build Logs:\n")
			logs, err := c.GetBuildLogs(statusProjectID, statusBuildID)
			if err != nil {
				fmt.Printf("⚠️  Failed to get logs: %v\n", err)
			} else {
				fmt.Println(logs)
			}
		}
	}

	// Show URLs
	fmt.Printf("\n🌐 URLs:\n")
	fmt.Printf("Preview: %s/preview/%s\n", baseURL, project.ProjectID)
	fmt.Printf("Production: %s/%s\n", baseURL, project.ProjectID)

	return nil
}
