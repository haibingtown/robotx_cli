package cmd

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/haibingtown/robotx_cli/pkg/client"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var deployCmd = &cobra.Command{
	Use:   "deploy [project-path]",
	Short: "Deploy a project to RobotX",
	Long: `Deploy a project to RobotX platform. This command will:
1. Create a project (if not exists)
2. Package and upload source code
3. Trigger a build
4. Wait for build completion
5. Optionally publish to production`,
	Args: cobra.MaximumNArgs(1),
	RunE: runDeploy,
}

var (
	projectName string
	projectID   string
	visibility  string
	publish     bool
	wait        bool
	timeout     int
)

func init() {
	rootCmd.AddCommand(deployCmd)

	deployCmd.Flags().StringVarP(&projectName, "name", "n", "", "Project name (required for new projects)")
	deployCmd.Flags().StringVarP(&projectID, "project-id", "p", "", "Existing project ID (skip creation)")
	deployCmd.Flags().StringVarP(&visibility, "visibility", "v", "private", "Project visibility (public/private)")
	deployCmd.Flags().BoolVar(&publish, "publish", false, "Publish to production after successful build")
	deployCmd.Flags().BoolVar(&wait, "wait", true, "Wait for build completion")
	deployCmd.Flags().IntVar(&timeout, "timeout", 600, "Build timeout in seconds")
}

func runDeploy(cmd *cobra.Command, args []string) error {
	// Get project path
	projectPath := "."
	if len(args) > 0 {
		projectPath = args[0]
	}

	// Validate project path
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return fmt.Errorf("invalid project path: %w", err)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("project path does not exist: %s", absPath)
	}

	// Get configuration
	baseURL := viper.GetString("base_url")
	apiKey := viper.GetString("api_key")

	if baseURL == "" {
		return fmt.Errorf("base URL is required (use --base-url or set ROBOTX_BASE_URL)")
	}
	if apiKey == "" {
		return fmt.Errorf("API key is required (use --api-key or set ROBOTX_API_KEY)")
	}

	// Create client
	c := client.NewClient(baseURL, apiKey)

	// Step 1: Get or create project
	var proj *client.Project
	if projectID != "" {
		fmt.Printf("📦 Using existing project: %s\n", projectID)
		proj, err = c.GetProject(projectID)
		if err != nil {
			return fmt.Errorf("failed to get project: %w", err)
		}
	} else {
		if projectName == "" {
			projectName = filepath.Base(absPath)
		}
		fmt.Printf("📦 Creating project: %s\n", projectName)
		proj, err = c.CreateProject(client.CreateProjectRequest{
			Name:       projectName,
			Visibility: visibility,
		})
		if err != nil {
			return fmt.Errorf("failed to create project: %w", err)
		}
		fmt.Printf("✅ Project created: %s\n", proj.ProjectID)
	}

	// Step 2: Package source code
	fmt.Printf("📦 Packaging source code from: %s\n", absPath)
	zipPath, err := packageSource(absPath)
	if err != nil {
		return fmt.Errorf("failed to package source: %w", err)
	}
	defer os.Remove(zipPath)
	fmt.Printf("✅ Source packaged: %s\n", zipPath)

	// Step 3: Upload source
	fmt.Printf("⬆️  Uploading source code...\n")
	commitID, err := c.UploadSource(proj.ProjectID, zipPath)
	if err != nil {
		return fmt.Errorf("failed to upload source: %w", err)
	}
	fmt.Printf("✅ Source uploaded: %s\n", commitID)

	// Step 4: Trigger build
	fmt.Printf("🔨 Triggering build...\n")
	build, err := c.TriggerBuild(proj.ProjectID, commitID)
	if err != nil {
		return fmt.Errorf("failed to trigger build: %w", err)
	}
	fmt.Printf("✅ Build started: %s\n", build.BuildID)

	// Step 5: Wait for build completion
	if wait {
		fmt.Printf("⏳ Waiting for build to complete (timeout: %ds)...\n", timeout)
		build, err = waitForBuild(c, proj.ProjectID, build.BuildID, timeout)
		if err != nil {
			return fmt.Errorf("build failed: %w", err)
		}

		if build.Status == "success" {
			fmt.Printf("✅ Build completed successfully!\n")

			// Get preview URL
			previewURL := fmt.Sprintf("%s/preview/%s", baseURL, proj.ProjectID)
			fmt.Printf("🌐 Preview URL: %s\n", previewURL)
		} else {
			fmt.Printf("❌ Build failed with status: %s\n", build.Status)

			// Try to get logs
			logs, err := c.GetBuildLogs(proj.ProjectID, build.BuildID)
			if err == nil && logs != "" {
				fmt.Printf("\n📋 Build logs:\n%s\n", logs)
			}
			return fmt.Errorf("build failed")
		}
	}

	// Step 6: Publish to production
	if publish && build.Status == "success" {
		fmt.Printf("🚀 Publishing to production...\n")
		if err := c.PublishBuild(proj.ProjectID, build.BuildID); err != nil {
			return fmt.Errorf("failed to publish: %w", err)
		}
		fmt.Printf("✅ Published successfully!\n")

		// Get production URL
		prodURL := fmt.Sprintf("%s/%s", baseURL, proj.ProjectID)
		fmt.Printf("🌐 Production URL: %s\n", prodURL)
	}

	return nil
}

func packageSource(projectPath string) (string, error) {
	// Create temporary zip file
	tmpFile, err := os.CreateTemp("", "robotx-source-*.zip")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	zipWriter := zip.NewWriter(tmpFile)
	defer zipWriter.Close()

	// Walk through project directory
	err = filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip certain directories
		relPath, err := filepath.Rel(projectPath, path)
		if err != nil {
			return err
		}

		if shouldSkip(relPath) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip directories themselves
		if info.IsDir() {
			return nil
		}

		// Add file to zip
		zipFile, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(zipFile, file)
		return err
	})

	if err != nil {
		os.Remove(tmpFile.Name())
		return "", err
	}

	return tmpFile.Name(), nil
}

func shouldSkip(path string) bool {
	skipDirs := []string{
		"node_modules",
		".git",
		".next",
		"dist",
		"build",
		".DS_Store",
		"__pycache__",
		".venv",
		"venv",
	}

	for _, skip := range skipDirs {
		if strings.HasPrefix(path, skip) || strings.Contains(path, string(filepath.Separator)+skip) {
			return true
		}
	}

	return false
}

func waitForBuild(c *client.Client, projectID, buildID string, timeoutSec int) (*client.Build, error) {
	start := time.Now()
	timeout := time.Duration(timeoutSec) * time.Second

	for {
		if time.Since(start) > timeout {
			return nil, fmt.Errorf("build timeout after %d seconds", timeoutSec)
		}

		build, err := c.GetBuild(projectID, buildID)
		if err != nil {
			return nil, err
		}

		switch build.Status {
		case "success", "failed":
			return build, nil
		case "queued", "running":
			fmt.Printf("⏳ Build status: %s (elapsed: %ds)\n", build.Status, int(time.Since(start).Seconds()))
			time.Sleep(5 * time.Second)
		default:
			return nil, fmt.Errorf("unknown build status: %s", build.Status)
		}
	}
}
