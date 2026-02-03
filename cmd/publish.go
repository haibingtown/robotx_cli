package cmd

import (
	"fmt"

	"haibingtown/robotx_cli/pkg/client"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish a build to production",
	Long:  `Publish a specific build to the production environment.`,
	RunE:  runPublish,
}

var (
	publishProjectID string
	publishBuildID   string
)

func init() {
	rootCmd.AddCommand(publishCmd)

	publishCmd.Flags().StringVarP(&publishProjectID, "project-id", "p", "", "Project ID (required)")
	publishCmd.Flags().StringVarP(&publishBuildID, "build-id", "b", "", "Build ID (required)")
	publishCmd.MarkFlagRequired("project-id")
	publishCmd.MarkFlagRequired("build-id")
}

func runPublish(cmd *cobra.Command, args []string) error {
	baseURL := viper.GetString("base_url")
	apiKey := viper.GetString("api_key")

	if baseURL == "" {
		return fmt.Errorf("base URL is required")
	}
	if apiKey == "" {
		return fmt.Errorf("API key is required")
	}

	c := client.NewClient(baseURL, apiKey)

	fmt.Printf("🚀 Publishing build %s to production...\n", publishBuildID)
	if err := c.PublishBuild(publishProjectID, publishBuildID); err != nil {
		return fmt.Errorf("failed to publish: %w", err)
	}

	fmt.Printf("✅ Published successfully!\n")
	prodURL := fmt.Sprintf("%s/%s", baseURL, publishProjectID)
	fmt.Printf("🌐 Production URL: %s\n", prodURL)

	return nil
}
