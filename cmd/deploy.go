package cmd

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"os/exec"
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
	localBuild  bool
	installCmd  string
	buildCmd    string
	outputDir   string
)

func init() {
	rootCmd.AddCommand(deployCmd)

	deployCmd.Flags().StringVarP(&projectName, "name", "n", "", "Project name (required for new projects)")
	deployCmd.Flags().StringVarP(&projectID, "project-id", "p", "", "Existing project ID (skip creation)")
	deployCmd.Flags().StringVarP(&visibility, "visibility", "v", "private", "Project visibility (public/private)")
	deployCmd.Flags().BoolVar(&publish, "publish", false, "Publish to production after successful build")
	deployCmd.Flags().BoolVar(&wait, "wait", true, "Wait for build completion")
	deployCmd.Flags().IntVar(&timeout, "timeout", 600, "Build timeout in seconds")
	deployCmd.Flags().BoolVar(&localBuild, "local-build", false, "Build locally and upload artifacts instead of using RobotX cloud build")
	deployCmd.Flags().StringVar(&installCmd, "install-command", "", "Override install command for local build")
	deployCmd.Flags().StringVar(&buildCmd, "build-command", "", "Override build command for local build")
	deployCmd.Flags().StringVar(&outputDir, "output-dir", "", "Override output directory for local build")
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
	commit, build, err := c.UploadSource(proj.ProjectID, zipPath)
	if err != nil {
		return fmt.Errorf("failed to upload source: %w", err)
	}
	if commit != nil && commit.CommitID != "" {
		fmt.Printf("✅ Source uploaded: %s\n", commit.CommitID)
	}
	if build != nil && build.BuildID != "" {
		fmt.Printf("✅ Build created: %s\n", build.BuildID)
	}

	// Step 4: Trigger build (remote) or run locally
	if localBuild {
		if build == nil || build.BuildID == "" {
			return fmt.Errorf("server did not return a build ID; local build upload is not supported by this server")
		}
		plan := (*client.BuildPlan)(nil)
		if commit != nil && commit.ScannerResult != nil {
			plan = commit.ScannerResult.BuildPlan
		}
		if err := runLocalBuild(absPath, plan); err != nil {
			return err
		}
		artifactDir := outputDir
		if artifactDir == "" && plan != nil && strings.TrimSpace(plan.OutputDir) != "" {
			artifactDir = strings.TrimSpace(plan.OutputDir)
		}
		if artifactDir == "" {
			artifactDir = "dist"
		}
		artifactPath := filepath.Join(absPath, artifactDir)
		if stat, err := os.Stat(artifactPath); err != nil || !stat.IsDir() {
			return fmt.Errorf("output directory missing: %s", artifactPath)
		}
		fmt.Printf("📦 Packaging build output from: %s\n", artifactPath)
		artifactZip, err := packageDirectory(artifactPath)
		if err != nil {
			return fmt.Errorf("failed to package build output: %w", err)
		}
		defer os.Remove(artifactZip)
		fmt.Printf("✅ Build output packaged: %s\n", artifactZip)

		fmt.Printf("⬆️  Uploading build artifacts...\n")
		build, err = c.UploadBuildArtifacts(build.BuildID, artifactZip)
		if err != nil {
			return fmt.Errorf("failed to upload build artifacts: %w", err)
		}
		fmt.Printf("✅ Build artifacts uploaded\n")
	} else {
		fmt.Printf("🔨 Triggering build...\n")
		if build == nil || build.BuildID == "" {
			if commit == nil || commit.CommitID == "" {
				return fmt.Errorf("no commit ID available to trigger build")
			}
			build, err = c.TriggerBuild(proj.ProjectID, commit.CommitID)
			if err != nil {
				return fmt.Errorf("failed to trigger build: %w", err)
			}
			fmt.Printf("✅ Build started: %s\n", build.BuildID)
		} else {
			if err := c.StartBuild(proj.ProjectID, build.BuildID); err != nil {
				return fmt.Errorf("failed to start build: %w", err)
			}
			fmt.Printf("✅ Build started: %s\n", build.BuildID)
		}
	}

	// Step 5: Wait for build completion
	if wait && !localBuild {
		fmt.Printf("⏳ Waiting for build to complete (timeout: %ds)...\n", timeout)
		build, err = waitForBuild(c, proj.ProjectID, build.BuildID, timeout)
		if err != nil {
			return fmt.Errorf("build failed: %w", err)
		}

		if build.Status == "success" {
			fmt.Printf("✅ Build completed successfully!\n")

			// Get preview URL
			previewURL := build.PreviewPath
			if previewURL == "" {
				previewURL = fmt.Sprintf("%s/preview/%s", baseURL, proj.ProjectID)
			}
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
	} else if localBuild && build != nil && build.Status == "success" {
		fmt.Printf("✅ Local build completed successfully!\n")
		previewURL := build.PreviewPath
		if previewURL == "" {
			previewURL = fmt.Sprintf("%s/preview/%s", baseURL, proj.ProjectID)
		}
		fmt.Printf("🌐 Preview URL: %s\n", previewURL)
	}

	// Step 6: Publish to production
	if publish && build != nil && build.Status == "success" {
		fmt.Printf("🚀 Publishing to production...\n")
		publicPath, err := c.PublishBuild(proj.ProjectID, build.BuildID)
		if err != nil {
			return fmt.Errorf("failed to publish: %w", err)
		}
		fmt.Printf("✅ Published successfully!\n")

		// Get production URL
		prodURL := publicPath
		if prodURL == "" {
			prodURL = fmt.Sprintf("%s/%s", baseURL, proj.ProjectID)
		}
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

func packageDirectory(root string) (string, error) {
	tmpFile, err := os.CreateTemp("", "robotx-artifacts-*.zip")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	zipWriter := zip.NewWriter(tmpFile)
	defer zipWriter.Close()

	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
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

func runLocalBuild(projectPath string, plan *client.BuildPlan) error {
	install := strings.TrimSpace(installCmd)
	build := strings.TrimSpace(buildCmd)

	if install == "" && plan != nil && strings.TrimSpace(plan.InstallCommand) != "" {
		install = strings.TrimSpace(plan.InstallCommand)
	}
	if build == "" && plan != nil && strings.TrimSpace(plan.BuildCommand) != "" {
		build = strings.TrimSpace(plan.BuildCommand)
	}

	if install == "" && fileExists(filepath.Join(projectPath, "package.json")) {
		install = "npm install"
	}
	if build == "" && fileExists(filepath.Join(projectPath, "package.json")) {
		build = "npm run build"
	}

	if plan != nil && !plan.NeedsBuild && installCmd == "" && buildCmd == "" {
		install = ""
		build = ""
	}

	if install != "" {
		fmt.Printf("🛠️  Running %s\n", install)
		if err := runShell(projectPath, install); err != nil {
			return fmt.Errorf("install failed: %w", err)
		}
	}
	if build != "" {
		fmt.Printf("🛠️  Running %s\n", build)
		if err := runShell(projectPath, build); err != nil {
			return fmt.Errorf("build failed: %w", err)
		}
	}
	return nil
}

func runShell(dir, command string) error {
	cmd := exec.Command("sh", "-lc", command)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
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
