package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/SideQuest-Group/ancla-client/internal/config"
)

func init() {
	rootCmd.AddCommand(deployActionCmd)
	deployActionCmd.Flags().Bool("no-follow", false, "Fire and forget — don't stream build logs")
	// Suppress cobra usage dump on RunE errors — deploy errors are handled
	// with styled error cards, not usage text.
	deployActionCmd.SilenceUsage = true
	deployActionCmd.SilenceErrors = true
}

var deployActionCmd = &cobra.Command{
	Use:   "deploy [<ws>/<proj>/<env>/<svc>]",
	Short: "Deploy your service",
	Long: `Deploy your service in one command.

If the current directory is not linked to a service, deploy walks you through
selecting (or creating) a workspace, project, environment, and service
interactively. For Python projects it can also scaffold a Dockerfile.

Once linked, subsequent runs skip straight to the deploy.

Use --no-follow to trigger the deploy without streaming build logs.`,
	Example: "  ancla deploy\n  ancla deploy my-ws/my-proj/staging/my-svc\n  ancla deploy --no-follow",
	GroupID: "workflow",
	Args:    cobra.MaximumNArgs(1),
	RunE:    runDeploy,
}

func runDeploy(cmd *cobra.Command, args []string) error {
	// If an explicit path was given, skip the wizard entirely.
	if len(args) > 0 {
		return deployDirect(cmd, args)
	}

	// --- Preflight ensure chain ---
	changed := false

	ws, proj, env, svc := cfg.Workspace, cfg.Project, cfg.Env, cfg.Service

	var err error

	// 1. Ensure logged in
	if err = ensureLoggedIn(); err != nil {
		return err
	}

	// 2. Ensure workspace
	ws, err = ensureWorkspace(ws)
	if err != nil {
		return err
	}
	if ws != cfg.Workspace {
		cfg.Workspace = ws
		changed = true
	}

	// 3. Ensure project
	proj, err = ensureProject(ws, proj)
	if err != nil {
		return err
	}
	if proj != cfg.Project {
		cfg.Project = proj
		changed = true
	}

	// 4. Ensure environment
	env, err = ensureEnv(ws, proj, env)
	if err != nil {
		return err
	}
	if env != cfg.Env {
		cfg.Env = env
		changed = true
	}

	// 5. Ensure service
	svc, err = ensureService(ws, proj, env, svc)
	if err != nil {
		return err
	}
	if svc != cfg.Service {
		cfg.Service = svc
		changed = true
	}

	// 6. Ensure Dockerfile (skip for buildpack services)
	strategy := fetchServiceBuildStrategy(ws, proj, env, svc)
	if strategy != "buildpack" {
		if err = ensureDockerfile(); err != nil {
			return err
		}
	} else if !isQuiet() {
		fmt.Println("\n" + stepActive("Buildpack service — skipping Dockerfile check."))
	}

	// 7. Save link context if anything changed
	if changed {
		cfg.Workspace = ws
		cfg.Project = proj
		cfg.Env = env
		cfg.Service = svc
		if err := config.SaveLocal(cfg); err != nil {
			return fmt.Errorf("saving link context: %w", err)
		}
	}

	path := fmt.Sprintf("%s/%s/%s/%s", ws, proj, env, svc)
	if !isQuiet() {
		if changed {
			fmt.Println("\n" + stepDone("Linked: "+stAccent.Render(path)))
			fmt.Println(stDim.Render("  Saved to .ancla/config.yaml"))
		}
		fmt.Printf("\n%s %s\n", stBold.Render("Deploying"), stAccent.Render(path))
	}

	// --- Existing deploy logic ---
	return triggerAndFollow(cmd, ws, proj, env, svc)
}

// deployDirect handles the case where the user gave an explicit ws/proj/env/svc argument.
func deployDirect(cmd *cobra.Command, args []string) error {
	ws, proj, env, svc, err := resolveServicePath(args)
	if err != nil {
		return err
	}
	if proj == "" || env == "" || svc == "" {
		return fmt.Errorf("all four segments required: <ws>/<proj>/<env>/<svc>")
	}

	path := fmt.Sprintf("%s/%s/%s/%s", ws, proj, env, svc)
	if !isQuiet() {
		fmt.Printf("%s %s\n", stBold.Render("Deploying"), stAccent.Render(path))
	}

	return triggerAndFollow(cmd, ws, proj, env, svc)
}

// triggerAndFollow POSTs the deploy and polls builds/deploys until complete.
func triggerAndFollow(cmd *cobra.Command, ws, proj, env, svc string) error {
	stop := spin("Triggering deploy...")
	req, _ := http.NewRequest("POST", apiURL(servicePath(ws, proj, env, svc)+"/deploy"), nil)
	body, err := doRequest(req)
	stop()
	if err != nil {
		return err
	}

	// Parse whatever the server returns — field names vary.
	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Println("Deploy triggered, but the response could not be parsed.")
		return nil
	}

	if isJSON() {
		return printJSON(result)
	}

	noFollow, _ := cmd.Flags().GetBool("no-follow")
	if noFollow {
		fmt.Println(stepDone("Deploy triggered."))
		return nil
	}

	// Poll builds list + deploys list to track the pipeline.
	return followPipeline(ws, proj, env, svc)
}

// pipelineStatusPath returns the project-level pipeline status URL with
// service and env as query params.
func pipelineStatusPath(ws, proj, env, svc string) string {
	return fmt.Sprintf("/workspaces/%s/projects/%s/pipeline/status?service=%s&env=%s", ws, proj, svc, env)
}

// followPipeline polls the pipeline status endpoint until both the build
// and deploy phases complete (or one errors).
//
// Important: the deploy stage is only evaluated AFTER the build completes,
// because until a new deploy record is created (which happens post-build),
// the pipeline returns the previous deploy's status — which may be "success".
func followPipeline(ws, proj, env, svc string) error {
	type stageStatus struct {
		Status      string  `json:"status"`
		ErrorDetail *string `json:"error_detail"`
	}

	buildDone := false
	prevBuildStatus := ""
	prevDeployStatus := ""
	stop := spin("Building...")
	defer stop()

	for first := true; ; first = false {
		if !first {
			time.Sleep(3 * time.Second)
		}

		req, _ := http.NewRequest("GET", apiURL(pipelineStatusPath(ws, proj, env, svc)), nil)
		body, err := doRequest(req)
		if err != nil {
			return err
		}

		var status struct {
			Build  *stageStatus `json:"build"`
			Deploy *stageStatus `json:"deploy"`
		}
		if err := json.Unmarshal(body, &status); err != nil {
			return fmt.Errorf("parsing pipeline status: %w", err)
		}

		// Track build phase.
		if !buildDone && status.Build != nil && status.Build.Status != prevBuildStatus {
			prevBuildStatus = status.Build.Status
			switch status.Build.Status {
			case "success":
				stop()
				fmt.Println(stepDone("Build complete"))
				buildDone = true
				// Reset deploy tracking — ignore any stale deploy status
				// from before this build. The new deploy will appear shortly.
				prevDeployStatus = ""
				stop = spin("Deploying...")
			case "error":
				stop()
				pe := &pipelineError{
					Kind:      errBuild,
					Workspace: ws, Project: proj, Env: env, Service: svc,
				}
				if status.Build.ErrorDetail != nil {
					pe.Detail = *status.Build.ErrorDetail
				}
				renderErrorCard(pe)
				return fmt.Errorf("build failed")
			}
		}

		// Track deploy phase — only after build is done.
		if buildDone && status.Deploy != nil && status.Deploy.Status != prevDeployStatus {
			prevDeployStatus = status.Deploy.Status
			switch status.Deploy.Status {
			case "success":
				stop()
				fmt.Println(stepDone("Deploy complete"))
				fmt.Println("\n" + stSuccess.Render(symCheck+" Deploy pipeline complete."))
				return nil
			case "error":
				stop()
				pe := &pipelineError{
					Kind:      errDeploy,
					Workspace: ws, Project: proj, Env: env, Service: svc,
				}
				if status.Deploy.ErrorDetail != nil {
					pe.Detail = *status.Deploy.ErrorDetail
				}
				renderErrorCard(pe)
				return fmt.Errorf("deploy failed")
			}
		}
	}
}

// --- Preflight ensure steps ---

// ensureLoggedIn checks that we have a valid API key. If not, triggers the
// browser login flow.
func ensureLoggedIn() error {
	if cfg.APIKey != "" {
		req, _ := http.NewRequest("GET", apiURL("/workspaces/"), nil)
		if _, err := doRequest(req); err == nil {
			return nil
		}
		if !isQuiet() {
			fmt.Println(stepActive("API key is invalid or expired."))
		}
	} else if !isQuiet() {
		fmt.Println(stepActive("Not logged in."))
	}

	if !isQuiet() {
		fmt.Println(stDim.Render("  Opening browser to log in..."))
	}
	if err := loginBrowser(); err != nil {
		return fmt.Errorf("login failed: %w", err)
	}
	fmt.Println(stepDone("Logged in"))
	return nil
}

// ensureWorkspace ensures a workspace is selected. Returns the workspace slug.
func ensureWorkspace(current string) (string, error) {
	if current != "" {
		req, _ := http.NewRequest("GET", apiURL("/workspaces/"+current+"/"), nil)
		if _, err := doRequest(req); err == nil {
			if !isQuiet() {
				fmt.Println(stepDone("Workspace: " + stAccent.Render(current)))
			}
			return current, nil
		}
		if !isQuiet() {
			fmt.Println(stepActive(fmt.Sprintf("Workspace %q not found, re-selecting...", current)))
		}
	}

	req, _ := http.NewRequest("GET", apiURL("/workspaces/"), nil)
	body, err := doRequest(req)
	if err != nil {
		return "", fmt.Errorf("fetching workspaces: %w", err)
	}

	var workspaces []struct {
		Name     string `json:"name"`
		Slug     string `json:"slug"`
		Personal bool   `json:"personal"`
	}
	if err := json.Unmarshal(body, &workspaces); err != nil {
		return "", fmt.Errorf("parsing workspaces: %w", err)
	}

	switch len(workspaces) {
	case 0:
		if !isQuiet() {
			fmt.Println(stepActive("No workspaces found. Creating a personal workspace..."))
		}
		name := cfg.Username
		if name == "" {
			name = "personal"
		}
		return createWorkspace(name+"'s workspace", true)

	case 1:
		ws := workspaces[0]
		if !isQuiet() {
			fmt.Println(stepDone("Workspace: " + stAccent.Render(ws.Slug)))
		}
		return ws.Slug, nil

	default:
		items := make([]promptItem, len(workspaces))
		for i, w := range workspaces {
			items[i] = promptItem{Slug: w.Slug, Name: w.Name}
		}
		slug, existing, err := promptSelectOrCreate("Select a workspace:", items, "Create new workspace")
		if err != nil {
			return "", err
		}
		if existing {
			fmt.Println(stepDone("Workspace: " + stAccent.Render(slug)))
			return slug, nil
		}
		name, err := promptInput("  Workspace name", "")
		if err != nil {
			return "", err
		}
		if name == "" {
			return "", fmt.Errorf("workspace name is required")
		}
		return createWorkspace(name, false)
	}
}

func createWorkspace(name string, personal bool) (string, error) {
	payload, _ := json.Marshal(map[string]any{
		"name":     name,
		"personal": personal,
	})
	req, _ := http.NewRequest("POST", apiURL("/workspaces/"), bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	body, err := doRequest(req)
	if err != nil {
		return "", fmt.Errorf("creating workspace: %w", err)
	}
	var ws struct {
		Slug string `json:"slug"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(body, &ws); err != nil {
		return "", fmt.Errorf("parsing workspace response: %w", err)
	}
	fmt.Println(stepDone("Workspace: " + stAccent.Render(ws.Slug)))
	return ws.Slug, nil
}

// ensureProject ensures a project is selected within the workspace.
func ensureProject(ws, current string) (string, error) {
	if current != "" {
		req, _ := http.NewRequest("GET", apiURL("/workspaces/"+ws+"/projects/"+current+"/"), nil)
		if _, err := doRequest(req); err == nil {
			if !isQuiet() {
				fmt.Println(stepDone("Project: " + stAccent.Render(current)))
			}
			return current, nil
		}
		if !isQuiet() {
			fmt.Println(stepActive(fmt.Sprintf("Project %q not found, re-selecting...", current)))
		}
	}

	req, _ := http.NewRequest("GET", apiURL("/workspaces/"+ws+"/projects/"), nil)
	body, err := doRequest(req)
	if err != nil {
		return "", fmt.Errorf("fetching projects: %w", err)
	}

	var projects []struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	if err := json.Unmarshal(body, &projects); err != nil {
		return "", fmt.Errorf("parsing projects: %w", err)
	}

	items := make([]promptItem, len(projects))
	for i, p := range projects {
		items[i] = promptItem{Slug: p.Slug, Name: p.Name}
	}
	slug, action, err := promptSelectCreateSkip("Select a project:", items, "Create new project", "Link to workspace only")
	if err != nil {
		return "", err
	}
	switch action {
	case "existing":
		fmt.Println(stepDone("Project: " + stAccent.Render(slug)))
		return slug, nil
	case "skip":
		return "", nil
	}

	defaultName := currentDirName()
	name, err := promptInput("  Project name", defaultName)
	if err != nil {
		return "", err
	}
	if name == "" {
		return "", fmt.Errorf("project name is required")
	}

	return createProject(ws, name)
}

func createProject(ws, name string) (string, error) {
	slug := slugify(name)
	payload, _ := json.Marshal(map[string]any{
		"name": name,
		"slug": slug,
	})
	req, _ := http.NewRequest("POST", apiURL("/workspaces/"+ws+"/projects/"), bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	body, err := doRequest(req)
	if err != nil {
		return "", fmt.Errorf("creating project: %w", err)
	}
	var proj struct {
		Slug string `json:"slug"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(body, &proj); err != nil {
		return "", fmt.Errorf("parsing project response: %w", err)
	}
	fmt.Println(stepDone("Created project " + stAccent.Render(proj.Name) + stDim.Render(" (environments: production, staging, development)")))
	return proj.Slug, nil
}

// ensureEnv ensures an environment is selected within the project.
func ensureEnv(ws, proj, current string) (string, error) {
	if current != "" {
		req, _ := http.NewRequest("GET", apiURL("/workspaces/"+ws+"/projects/"+proj+"/envs/"+current+"/"), nil)
		if _, err := doRequest(req); err == nil {
			if !isQuiet() {
				fmt.Println(stepDone("Environment: " + stAccent.Render(current)))
			}
			return current, nil
		}
		if !isQuiet() {
			fmt.Println(stepActive(fmt.Sprintf("Environment %q not found, re-selecting...", current)))
		}
	}

	req, _ := http.NewRequest("GET", apiURL("/workspaces/"+ws+"/projects/"+proj+"/envs/"), nil)
	body, err := doRequest(req)
	if err != nil {
		return "", fmt.Errorf("fetching environments: %w", err)
	}

	var envs []struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	if err := json.Unmarshal(body, &envs); err != nil {
		return "", fmt.Errorf("parsing environments: %w", err)
	}

	items := make([]promptItem, len(envs))
	for i, e := range envs {
		items[i] = promptItem{Slug: e.Slug, Name: e.Name}
	}
	slug, action, err := promptSelectCreateSkip("Select an environment:", items, "Create new environment", "Link to project only")
	if err != nil {
		return "", err
	}
	switch action {
	case "existing":
		fmt.Println(stepDone("Environment: " + stAccent.Render(slug)))
		return slug, nil
	case "skip":
		return "", nil
	}

	// Create new environment
	name, err := promptInput("  Environment name", "production")
	if err != nil {
		return "", err
	}
	if name == "" {
		return "", fmt.Errorf("environment name is required")
	}

	return createEnv(ws, proj, name)
}

// createEnv creates a new environment via the API and returns its slug.
func createEnv(ws, proj, name string) (string, error) {
	payload, _ := json.Marshal(map[string]string{"name": name})
	req, _ := http.NewRequest("POST", apiURL("/workspaces/"+ws+"/projects/"+proj+"/envs/"), bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	body, err := doRequest(req)
	if err != nil {
		return "", fmt.Errorf("creating environment: %w", err)
	}
	var e struct {
		Slug string `json:"slug"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(body, &e); err != nil {
		return "", fmt.Errorf("parsing environment response: %w", err)
	}
	fmt.Println(stepDone("Created environment " + stAccent.Render(e.Name)))
	return e.Slug, nil
}

// ensureService ensures a service is selected within the environment.
func ensureService(ws, proj, env, current string) (string, error) {
	if current != "" {
		req, _ := http.NewRequest("GET", apiURL(servicePath(ws, proj, env, current)), nil)
		if _, err := doRequest(req); err == nil {
			if !isQuiet() {
				fmt.Println(stepDone("Service: " + stAccent.Render(current)))
			}
			return current, nil
		}
		if !isQuiet() {
			fmt.Println(stepActive(fmt.Sprintf("Service %q not found, re-selecting...", current)))
		}
	}

	basePath := serviceBasePath(ws, proj, env)
	req, _ := http.NewRequest("GET", apiURL(basePath), nil)
	body, err := doRequest(req)
	if err != nil {
		return "", fmt.Errorf("fetching services: %w", err)
	}

	var services []struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	if err := json.Unmarshal(body, &services); err != nil {
		return "", fmt.Errorf("parsing services: %w", err)
	}

	items := make([]promptItem, len(services))
	for i, s := range services {
		items[i] = promptItem{Slug: s.Slug, Name: s.Name}
	}
	slug, action, err := promptSelectCreateSkip("Select a service:", items, "Create new service", "Link to environment only")
	if err != nil {
		return "", err
	}
	switch action {
	case "existing":
		fmt.Println(stepDone("Service: " + stAccent.Render(slug)))
		return slug, nil
	case "skip":
		return "", nil
	}

	// Create new service
	defaultName := cfg.Project
	if defaultName == "" {
		defaultName = currentDirName()
	}
	name, err := promptInput("  Service name", defaultName)
	if err != nil {
		return "", err
	}
	if name == "" {
		return "", fmt.Errorf("service name is required")
	}

	return createService(ws, proj, env, name)
}

// createService creates a new service via the API and returns its slug.
func createService(ws, proj, env, name string) (string, error) {
	slug := slugify(name)
	payload := map[string]any{
		"name":     name,
		"slug":     slug,
		"platform": "wind",
	}

	// Try to detect GitHub repo
	if repo := detectGitHubRepo(); repo != "" {
		payload["github_repository"] = repo
	}

	// Ask for build strategy
	strategyItems := []promptItem{
		{Slug: "dockerfile", Name: "Dockerfile — build from your Dockerfile"},
		{Slug: "buildpack", Name: "Buildpack — automatic detection, no Dockerfile required"},
	}
	strategy, err := promptSelect("  Build strategy:", strategyItems, "dockerfile")
	if err == nil && strategy != "" {
		payload["build_strategy"] = strategy
	}

	data, _ := json.Marshal(payload)
	basePath := serviceBasePath(ws, proj, env)
	req, _ := http.NewRequest("POST", apiURL(basePath), bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	body, err := doRequest(req)
	if err != nil {
		return "", fmt.Errorf("creating service: %w", err)
	}
	var svc struct {
		Slug string `json:"slug"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(body, &svc); err != nil {
		return "", fmt.Errorf("parsing service response: %w", err)
	}
	fmt.Println(stepDone("Created service " + stAccent.Render(svc.Name)))
	return svc.Slug, nil
}

// fetchServiceBuildStrategy fetches the build_strategy for a service.
// Returns "dockerfile" (default), "buildpack", or "" on error.
func fetchServiceBuildStrategy(ws, proj, env, svc string) string {
	req, _ := http.NewRequest("GET", apiURL(servicePath(ws, proj, env, svc)), nil)
	body, err := doRequest(req)
	if err != nil {
		return ""
	}
	var detail struct {
		BuildStrategy *string `json:"build_strategy"`
	}
	if err := json.Unmarshal(body, &detail); err != nil {
		return ""
	}
	if detail.BuildStrategy == nil || *detail.BuildStrategy == "" {
		return "dockerfile"
	}
	return *detail.BuildStrategy
}

// --- Helpers ---

// currentDirName returns the base name of the current working directory.
func currentDirName() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return filepath.Base(cwd)
}

// slugify converts a name to a URL-safe slug.
func slugify(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		if r == ' ' || r == '_' {
			return '-'
		}
		return -1
	}, s)
	// Collapse multiple dashes
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	return strings.Trim(s, "-")
}

// detectGitHubRepo tries to extract owner/repo from the git remote origin URL.
func detectGitHubRepo() string {
	out, err := exec.Command("git", "remote", "get-url", "origin").Output()
	if err != nil {
		return ""
	}
	url := strings.TrimSpace(string(out))

	// SSH: git@github.com:owner/repo.git
	if strings.HasPrefix(url, "git@github.com:") {
		repo := strings.TrimPrefix(url, "git@github.com:")
		repo = strings.TrimSuffix(repo, ".git")
		return repo
	}

	// HTTPS: https://github.com/owner/repo.git
	if strings.Contains(url, "github.com/") {
		idx := strings.Index(url, "github.com/")
		repo := url[idx+len("github.com/"):]
		repo = strings.TrimSuffix(repo, ".git")
		return repo
	}

	return ""
}
