package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/haibingtown/robotx_cli/pkg/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "List projects",
	Long:  `List projects for the current account.`,
	RunE:  runProjects,
}

var (
	projectsLimit int
)

type projectsResponse struct {
	Limit    int               `json:"limit,omitempty"`
	Projects []*client.Project `json:"projects"`
}

func init() {
	rootCmd.AddCommand(projectsCmd)

	projectsCmd.Flags().IntVar(&projectsLimit, "limit", 50, "Number of projects to list (max enforced by server)")
}

func runProjects(cmd *cobra.Command, args []string) error {
	baseURL := viper.GetString("base_url")
	apiKey := viper.GetString("api_key")

	if baseURL == "" {
		return newCLIError("missing_base_url", "base URL is required", 1, nil)
	}
	if apiKey == "" {
		return newCLIError("missing_api_key", "API key is required", 1, nil)
	}

	c := client.NewClient(baseURL, apiKey)
	logf("📋 Listing projects...\n")
	projects, err := c.ListProjects(projectsLimit)
	if err != nil {
		return newCLIError("api_error", "failed to list projects", 2, err)
	}

	resp := projectsResponse{
		Limit:    projectsLimit,
		Projects: projects,
	}
	if err := emitSuccess(cmd.Name(), resp); err != nil {
		return newCLIError("output_error", "failed to render JSON output", 1, err)
	}
	if isJSONOutput() {
		return nil
	}

	if len(projects) == 0 {
		fmt.Fprintln(os.Stdout, "No projects found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "PROJECT_ID\tNAME\tVISIBILITY\tCREATED_AT\tUPDATED_AT\tPREVIEW_URL\tPRODUCTION_URL")
	for _, project := range projects {
		fmt.Fprintf(
			w,
			"%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			project.ProjectID,
			valueOrDash(project.Name),
			valueOrDash(project.Visibility),
			formatBuildTime(project.CreatedAt),
			formatBuildTime(project.UpdatedAt),
			valueOrDash(projectPreviewURL(project, baseURL)),
			valueOrDash(resolvePublishURL(baseURL, project)),
		)
	}
	_ = w.Flush()

	return nil
}
