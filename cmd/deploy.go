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

type deployResponse struct {
	ProjectID     string `json:"project_id"`
	ProjectName   string `json:"project_name,omitempty"`
	CommitID      string `json:"commit_id,omitempty"`
	BuildID       string `json:"build_id,omitempty"`
	BuildStatus   string `json:"build_status,omitempty"`
	PreviewURL    string `json:"preview_url,omitempty"`
	ProductionURL string `json:"production_url,omitempty"`
	Published     bool   `json:"published"`
	Waited        bool   `json:"waited"`
	LocalBuild    bool   `json:"local_build"`
}

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
	projectPath := "."
	if len(args) > 0 {
		projectPath = args[0]
	}

	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return newCLIError("invalid_project_path", "invalid project path", 1, err)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return newCLIError("invalid_project_path", fmt.Sprintf("project path does not exist: %s", absPath), 1, nil)
	}

	baseURL := viper.GetString("base_url")
	apiKey := viper.GetString("api_key")

	if baseURL == "" {
		return newCLIError("missing_base_url", "base URL is required (use --base-url or set ROBOTX_BASE_URL)", 1, nil)
	}
	if apiKey == "" {
		return newCLIError("missing_api_key", "API key is required (use --api-key or set ROBOTX_API_KEY)", 1, nil)
	}

	c := client.NewClient(baseURL, apiKey)
	usedProjectName := strings.TrimSpace(projectName)
	var previewURL string
	var productionURL string

	var proj *client.Project
	if projectID != "" {
		logf("📦 Using existing project: %s\n", projectID)
		proj, err = c.GetProject(projectID)
		if err != nil {
			return newCLIError("api_error", "failed to get project", 2, err)
		}
		if usedProjectName == "" {
			usedProjectName = proj.Name
		}
	} else {
		if usedProjectName == "" {
			usedProjectName = filepath.Base(absPath)
		}
		logf("📦 Creating project: %s\n", usedProjectName)
		proj, err = c.CreateProject(client.CreateProjectRequest{
			Name:       usedProjectName,
			Visibility: visibility,
		})
		if err != nil {
			return newCLIError("api_error", "failed to create project", 2, err)
		}
		usedProjectName = proj.Name
		logf("✅ Project created: %s\n", proj.ProjectID)
	}

	logf("📦 Packaging source code from: %s\n", absPath)
	zipPath, err := packageSource(absPath)
	if err != nil {
		return newCLIError("package_failed", "failed to package source", 1, err)
	}
	defer os.Remove(zipPath)

	if stat, statErr := os.Stat(zipPath); statErr == nil {
		sizeMB := float64(stat.Size()) / (1024.0 * 1024.0)
		logf("📏 Source archive size: %.2f MB\n", sizeMB)
	}
	logf("✅ Source packaged: %s\n", zipPath)

	logf("⬆️  Uploading source code...\n")
	commit, build, err := c.UploadSource(proj.ProjectID, zipPath)
	if err != nil {
		return newCLIError("api_error", "failed to upload source", 2, err)
	}
	if commit != nil && commit.CommitID != "" {
		logf("✅ Source uploaded: %s\n", commit.CommitID)
	}
	if build != nil && build.BuildID != "" {
		logf("✅ Build created: %s\n", build.BuildID)
	}

	if localBuild {
		if build == nil || build.BuildID == "" {
			return newCLIError("local_build_unsupported", "server did not return a build ID; local build upload is not supported by this server", 2, nil)
		}
		plan := (*client.BuildPlan)(nil)
		if commit != nil && commit.ScannerResult != nil {
			plan = commit.ScannerResult.BuildPlan
		}
		if err := runLocalBuild(absPath, plan); err != nil {
			return newCLIError("build_failed", "local build failed", 3, err)
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
			return newCLIError("build_failed", fmt.Sprintf("output directory missing: %s", artifactPath), 3, nil)
		}
		logf("📦 Packaging build output from: %s\n", artifactPath)
		artifactZip, err := packageDirectory(artifactPath)
		if err != nil {
			return newCLIError("build_failed", "failed to package build output", 3, err)
		}
		defer os.Remove(artifactZip)
		logf("✅ Build output packaged: %s\n", artifactZip)

		logf("⬆️  Uploading build artifacts...\n")
		build, err = c.UploadBuildArtifacts(build.BuildID, artifactZip)
		if err != nil {
			return newCLIError("api_error", "failed to upload build artifacts", 2, err)
		}
		logf("✅ Build artifacts uploaded\n")
	} else {
		logf("🔨 Triggering build...\n")
		if build == nil || build.BuildID == "" {
			if commit == nil || commit.CommitID == "" {
				return newCLIError("build_failed", "no commit ID available to trigger build", 3, nil)
			}
			build, err = c.TriggerBuild(proj.ProjectID, commit.CommitID)
			if err != nil {
				return newCLIError("api_error", "failed to trigger build", 2, err)
			}
			logf("✅ Build started: %s\n", build.BuildID)
		} else {
			if err := c.StartBuild(proj.ProjectID, build.BuildID); err != nil {
				return newCLIError("api_error", "failed to start build", 2, err)
			}
			logf("✅ Build started: %s\n", build.BuildID)
		}
	}

	if wait && !localBuild {
		logf("⏳ Waiting for build to complete (timeout: %ds)...\n", timeout)
		build, err = waitForBuild(c, proj.ProjectID, build.BuildID, timeout)
		if err != nil {
			return newCLIError("build_failed", "build failed", 3, err)
		}

		if build.Status == "success" {
			logf("✅ Build completed successfully!\n")
			previewURL = build.PreviewPath
			if previewURL == "" {
				previewURL = fmt.Sprintf("%s/preview/%s", baseURL, proj.ProjectID)
			}
			logf("🌐 Preview URL: %s\n", previewURL)
		} else {
			logf("❌ Build failed with status: %s\n", build.Status)

			logs, err := c.GetBuildLogs(proj.ProjectID, build.BuildID)
			if err == nil && logs != "" {
				logf("\n📋 Build logs:\n%s\n", logs)
			}
			return newCLIError("build_failed", fmt.Sprintf("build failed with status: %s", build.Status), 3, nil)
		}
	} else if localBuild && build != nil && build.Status == "success" {
		logf("✅ Local build completed successfully!\n")
		previewURL = build.PreviewPath
		if previewURL == "" {
			previewURL = fmt.Sprintf("%s/preview/%s", baseURL, proj.ProjectID)
		}
		logf("🌐 Preview URL: %s\n", previewURL)
	}

	if publish && build != nil && build.Status == "success" {
		logf("🚀 Publishing to production...\n")
		publicPath, err := c.PublishBuild(proj.ProjectID, build.BuildID)
		if err != nil {
			return newCLIError("publish_failed", "failed to publish", 4, err)
		}
		logf("✅ Published successfully!\n")

		productionURL = publicPath
		if productionURL == "" {
			productionURL = fmt.Sprintf("%s/%s", baseURL, proj.ProjectID)
		}
		logf("🌐 Production URL: %s\n", productionURL)
	}

	if previewURL == "" && build != nil && build.Status == "success" {
		previewURL = build.PreviewPath
		if previewURL == "" {
			previewURL = fmt.Sprintf("%s/preview/%s", baseURL, proj.ProjectID)
		}
	}
	if productionURL == "" && publish && build != nil && build.Status == "success" {
		productionURL = fmt.Sprintf("%s/%s", baseURL, proj.ProjectID)
	}

	if err := emitSuccess(cmd.Name(), deployResponse{
		ProjectID:     proj.ProjectID,
		ProjectName:   usedProjectName,
		CommitID:      safeCommitID(commit),
		BuildID:       safeBuildID(build),
		BuildStatus:   safeBuildStatus(build),
		PreviewURL:    previewURL,
		ProductionURL: productionURL,
		Published:     publish && productionURL != "",
		Waited:        wait,
		LocalBuild:    localBuild,
	}); err != nil {
		return newCLIError("output_error", "failed to render JSON output", 1, err)
	}

	return nil
}

func safeCommitID(commit *client.SourceCommit) string {
	if commit == nil {
		return ""
	}
	return commit.CommitID
}

func safeBuildID(build *client.Build) string {
	if build == nil {
		return ""
	}
	return build.BuildID
}

func safeBuildStatus(build *client.Build) string {
	if build == nil {
		return ""
	}
	return build.Status
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
		logf("🛠️  Running %s\n", install)
		if err := runShell(projectPath, install); err != nil {
			return fmt.Errorf("install failed: %w", err)
		}
	}
	if build != "" {
		logf("🛠️  Running %s\n", build)
		if err := runShell(projectPath, build); err != nil {
			return fmt.Errorf("build failed: %w", err)
		}
	}
	return nil
}

func runShell(dir, command string) error {
	cmd := exec.Command("sh", "-lc", command)
	cmd.Dir = dir
	if isJSONOutput() {
		cmd.Stdout = os.Stderr
	} else {
		cmd.Stdout = os.Stdout
	}
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
			logf("⏳ Build status: %s (elapsed: %ds)\n", build.Status, int(time.Since(start).Seconds()))
			time.Sleep(5 * time.Second)
		default:
			return nil, fmt.Errorf("unknown build status: %s", build.Status)
		}
	}
}
