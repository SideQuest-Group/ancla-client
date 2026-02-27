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

	"github.com/spf13/cobra"

	"github.com/SideQuest-Group/ancla-client/internal/config"
)

func init() {
	rootCmd.AddCommand(deployActionCmd)
	deployActionCmd.Flags().Bool("no-follow", false, "Fire and forget — don't stream build logs")
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

	// 6. Ensure Dockerfile
	if err = ensureDockerfile(); err != nil {
		return err
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
			fmt.Printf("\n→ Linked: %s\n", path)
			fmt.Println("  Saved to .ancla/config.yaml")
		}
		fmt.Printf("\nDeploying %s...\n", path)
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
		fmt.Printf("Deploying %s...\n", path)
	}

	return triggerAndFollow(cmd, ws, proj, env, svc)
}

// triggerAndFollow POSTs the deploy and optionally follows the build log.
func triggerAndFollow(cmd *cobra.Command, ws, proj, env, svc string) error {
	stop := spin("Triggering deploy...")
	req, _ := http.NewRequest("POST", apiURL(servicePath(ws, proj, env, svc)+"/deploy"), nil)
	body, err := doRequest(req)
	stop()
	if err != nil {
		return err
	}

	var result struct {
		BuildID string `json:"build_id"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Println("Deploy triggered, but the response could not be parsed.")
		return nil
	}

	if isJSON() {
		return printJSON(result)
	}

	noFollow, _ := cmd.Flags().GetBool("no-follow")
	if noFollow || result.BuildID == "" {
		fmt.Printf("Deploy triggered. Build ID: %s\n", result.BuildID)
		return nil
	}

	fmt.Printf("Build ID: %s\n", result.BuildID)
	if err := followBuild(result.BuildID); err != nil {
		return err
	}

	fmt.Println("Deploy pipeline complete.")
	return nil
}

// --- Preflight ensure steps ---

// ensureLoggedIn checks that we have a valid API key. If not, triggers the
// browser login flow.
func ensureLoggedIn() error {
	if cfg.APIKey != "" {
		// Validate the key with a lightweight request
		req, _ := http.NewRequest("GET", apiURL("/workspaces/"), nil)
		if _, err := doRequest(req); err == nil {
			return nil // key is valid
		}
		if !isQuiet() {
			fmt.Println("→ API key is invalid or expired.")
		}
	} else if !isQuiet() {
		fmt.Println("→ Not logged in.")
	}

	if !isQuiet() {
		fmt.Println("  Opening browser to log in...")
	}
	if err := loginBrowser(); err != nil {
		return fmt.Errorf("login failed: %w", err)
	}
	fmt.Println("  ✓ Logged in")
	return nil
}

// ensureWorkspace ensures a workspace is selected. Returns the workspace slug.
func ensureWorkspace(current string) (string, error) {
	if current != "" {
		// Validate it exists
		req, _ := http.NewRequest("GET", apiURL("/workspaces/"+current+"/"), nil)
		if _, err := doRequest(req); err == nil {
			if !isQuiet() {
				fmt.Printf("\n→ Workspace: %s\n", current)
			}
			return current, nil
		}
		if !isQuiet() {
			fmt.Printf("\n→ Workspace %q not found, re-selecting...\n", current)
		}
	}

	// Fetch workspaces
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
		// Create a personal workspace
		if !isQuiet() {
			fmt.Println("\n→ No workspaces found. Creating a personal workspace...")
		}
		name := cfg.Username
		if name == "" {
			name = "personal"
		}
		return createWorkspace(name+"'s workspace", true)

	case 1:
		ws := workspaces[0]
		if !isQuiet() {
			fmt.Printf("\n→ Using workspace: %s (%s)\n", ws.Name, ws.Slug)
		}
		return ws.Slug, nil

	default:
		if !isQuiet() {
			fmt.Println("\n→ No workspace linked.")
		}
		items := make([]promptItem, len(workspaces))
		for i, w := range workspaces {
			items[i] = promptItem{Slug: w.Slug, Name: w.Name}
		}
		slug, err := promptSelect("Select a workspace:", items, "")
		if err != nil {
			return "", err
		}
		fmt.Printf("  ✓ Workspace: %s\n", slug)
		return slug, nil
	}
}

// createWorkspace creates a new workspace via the API and returns its slug.
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
	fmt.Printf("  ✓ Workspace: %s\n", ws.Slug)
	return ws.Slug, nil
}

// ensureProject ensures a project is selected within the workspace.
func ensureProject(ws, current string) (string, error) {
	if current != "" {
		req, _ := http.NewRequest("GET", apiURL("/workspaces/"+ws+"/projects/"+current+"/"), nil)
		if _, err := doRequest(req); err == nil {
			if !isQuiet() {
				fmt.Printf("\n→ Project: %s\n", current)
			}
			return current, nil
		}
		if !isQuiet() {
			fmt.Printf("\n→ Project %q not found, re-selecting...\n", current)
		}
	}

	// Fetch projects
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

	if !isQuiet() {
		fmt.Println("\n→ No project linked.")
	}

	if len(projects) > 0 {
		items := make([]promptItem, len(projects))
		for i, p := range projects {
			items[i] = promptItem{Slug: p.Slug, Name: p.Name}
		}
		slug, existing, err := promptSelectOrCreate("Select a project:", items, "Create new project")
		if err != nil {
			return "", err
		}
		if existing {
			fmt.Printf("  ✓ Project: %s\n", slug)
			return slug, nil
		}
	}

	// Create new project
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

// createProject creates a new project via the API and returns its slug.
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
	fmt.Printf("  ✓ Created project %q (environments: production, staging, development)\n", proj.Name)
	return proj.Slug, nil
}

// ensureEnv ensures an environment is selected within the project.
func ensureEnv(ws, proj, current string) (string, error) {
	if current != "" {
		req, _ := http.NewRequest("GET", apiURL("/workspaces/"+ws+"/projects/"+proj+"/envs/"+current+"/"), nil)
		if _, err := doRequest(req); err == nil {
			if !isQuiet() {
				fmt.Printf("\n→ Environment: %s\n", current)
			}
			return current, nil
		}
		if !isQuiet() {
			fmt.Printf("\n→ Environment %q not found, re-selecting...\n", current)
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

	if len(envs) == 0 {
		return "", fmt.Errorf("no environments found — this shouldn't happen if the project was just created")
	}

	if !isQuiet() {
		fmt.Println("\n→ No environment linked.")
	}
	items := make([]promptItem, len(envs))
	for i, e := range envs {
		items[i] = promptItem{Slug: e.Slug, Name: e.Name}
	}
	slug, err := promptSelect("Select an environment:", items, "production")
	if err != nil {
		return "", err
	}
	fmt.Printf("  ✓ Environment: %s\n", slug)
	return slug, nil
}

// ensureService ensures a service is selected within the environment.
func ensureService(ws, proj, env, current string) (string, error) {
	if current != "" {
		req, _ := http.NewRequest("GET", apiURL(servicePath(ws, proj, env, current)), nil)
		if _, err := doRequest(req); err == nil {
			if !isQuiet() {
				fmt.Printf("\n→ Service: %s\n", current)
			}
			return current, nil
		}
		if !isQuiet() {
			fmt.Printf("\n→ Service %q not found, re-selecting...\n", current)
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

	if !isQuiet() {
		fmt.Println("\n→ No service linked.")
	}

	if len(services) > 0 {
		items := make([]promptItem, len(services))
		for i, s := range services {
			items[i] = promptItem{Slug: s.Slug, Name: s.Name}
		}
		slug, existing, err := promptSelectOrCreate("Select a service:", items, "Create new service")
		if err != nil {
			return "", err
		}
		if existing {
			fmt.Printf("  ✓ Service: %s\n", slug)
			return slug, nil
		}
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
	fmt.Printf("  ✓ Created service %q\n", svc.Name)
	return svc.Slug, nil
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
