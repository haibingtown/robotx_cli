package cmd

import (
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update [project-path]",
	Short: "Update an existing project",
	Long: `Update an existing project with new code. This command will:
1. Package and upload new source code
2. Trigger a new build
3. Wait for build completion
4. Update preview environment`,
	Args: cobra.MaximumNArgs(1),
	RunE: runUpdate,
}

var (
	updateProjectID string
	updatePublish   bool
	updateWait      bool
)

func init() {
	rootCmd.AddCommand(updateCmd)

	updateCmd.Flags().StringVarP(&updateProjectID, "project-id", "p", "", "Project ID (required)")
	updateCmd.Flags().BoolVar(&updatePublish, "publish", false, "Publish to production after successful build")
	updateCmd.Flags().BoolVar(&updateWait, "wait", true, "Wait for build completion")
	updateCmd.MarkFlagRequired("project-id")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	// Reuse deploy logic with existing project ID
	projectID = updateProjectID
	publish = updatePublish
	wait = updateWait

	return runDeploy(cmd, args)
}
