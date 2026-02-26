package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/SideQuest-Group/ancla-client/internal/config"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Link this directory to an Ancla workspace/project/env/service",
	Long: `Initialize (link) the current directory to an Ancla service.

This interactive command walks you through selecting a workspace, project,
environment, and service, then saves the link context to a local
.ancla/config.yaml file. Subsequent commands run from this directory will
automatically use the linked workspace, project, env, and service without
requiring explicit arguments.`,
	Example: `  ancla init
  ancla init   # re-link to a different service`,
	GroupID: "workflow",
	RunE:    runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	// Step 1: Check if already linked
	if cfg.IsLinked() {
		fmt.Printf("This directory is already linked to: %s\n", cfg.ServicePath())
		fmt.Print("Continue and re-link? [y/N] ")
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" && answer != "yes" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	// Step 2: Select workspace
	wsSlug, err := selectWorkspace(reader)
	if err != nil {
		return err
	}

	// Step 3: Select project
	projectSlug, err := selectProject(reader, wsSlug)
	if err != nil {
		return err
	}

	// Step 4: Select environment
	envSlug, err := selectEnv(reader, wsSlug, projectSlug)
	if err != nil {
		return err
	}

	// Step 5: Select service
	svcSlug, err := selectService(reader, wsSlug, projectSlug, envSlug)
	if err != nil {
		return err
	}

	// Step 6: Save link context
	cfg.Workspace = wsSlug
	cfg.Project = projectSlug
	cfg.Env = envSlug
	cfg.Service = svcSlug

	if err := config.SaveLocal(cfg); err != nil {
		return fmt.Errorf("saving local config: %w", err)
	}

	// Step 7: Print summary
	fmt.Println()
	fmt.Println("Linked successfully!")
	fmt.Printf("  Workspace:   %s\n", wsSlug)
	fmt.Printf("  Project:     %s\n", projectSlug)
	fmt.Printf("  Environment: %s\n", envSlug)
	fmt.Printf("  Service:     %s\n", svcSlug)
	fmt.Println()
	fmt.Println("Saved to .ancla/config.yaml")
	return nil
}

// selectWorkspace fetches the user's workspaces and prompts for a selection.
func selectWorkspace(reader *bufio.Reader) (string, error) {
	req, _ := http.NewRequest("GET", apiURL("/workspaces/"), nil)
	body, err := doRequest(req)
	if err != nil {
		return "", fmt.Errorf("fetching workspaces: %w", err)
	}

	var workspaces []struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	if err := json.Unmarshal(body, &workspaces); err != nil {
		return "", fmt.Errorf("parsing workspaces: %w", err)
	}

	if len(workspaces) == 0 {
		return "", fmt.Errorf("no workspaces found — create one at %s first", cfg.Server)
	}

	fmt.Println()
	fmt.Println("Select a workspace:")
	for i, w := range workspaces {
		fmt.Printf("  [%d] %s (%s)\n", i+1, w.Name, w.Slug)
	}
	fmt.Print("Enter number or slug: ")

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	// Try as number first
	if n, err := strconv.Atoi(input); err == nil {
		if n < 1 || n > len(workspaces) {
			return "", fmt.Errorf("invalid selection: %d (must be 1-%d)", n, len(workspaces))
		}
		return workspaces[n-1].Slug, nil
	}

	// Otherwise treat as slug — verify it exists
	for _, w := range workspaces {
		if w.Slug == input {
			return w.Slug, nil
		}
	}
	return "", fmt.Errorf("workspace %q not found", input)
}

// selectProject fetches projects for the given workspace and prompts for a selection.
func selectProject(reader *bufio.Reader, wsSlug string) (string, error) {
	req, _ := http.NewRequest("GET", apiURL("/workspaces/"+wsSlug+"/projects/"), nil)
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

	if len(projects) == 0 {
		return "", fmt.Errorf("no projects found in workspace %q — create one at %s first", wsSlug, cfg.Server)
	}

	fmt.Println()
	fmt.Println("Select a project:")
	for i, p := range projects {
		fmt.Printf("  [%d] %s (%s)\n", i+1, p.Name, p.Slug)
	}
	fmt.Print("Enter number or slug: ")

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if n, err := strconv.Atoi(input); err == nil {
		if n < 1 || n > len(projects) {
			return "", fmt.Errorf("invalid selection: %d (must be 1-%d)", n, len(projects))
		}
		return projects[n-1].Slug, nil
	}

	for _, p := range projects {
		if p.Slug == input {
			return p.Slug, nil
		}
	}
	return "", fmt.Errorf("project %q not found in workspace %q", input, wsSlug)
}

// selectEnv fetches environments for the given workspace/project and prompts for a selection.
func selectEnv(reader *bufio.Reader, wsSlug, projectSlug string) (string, error) {
	req, _ := http.NewRequest("GET", apiURL("/workspaces/"+wsSlug+"/projects/"+projectSlug+"/envs/"), nil)
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
		return "", fmt.Errorf("no environments found in %s/%s — create one at %s first", wsSlug, projectSlug, cfg.Server)
	}

	fmt.Println()
	fmt.Println("Select an environment:")
	for i, e := range envs {
		fmt.Printf("  [%d] %s (%s)\n", i+1, e.Name, e.Slug)
	}
	fmt.Print("Enter number or slug: ")

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if n, err := strconv.Atoi(input); err == nil {
		if n < 1 || n > len(envs) {
			return "", fmt.Errorf("invalid selection: %d (must be 1-%d)", n, len(envs))
		}
		return envs[n-1].Slug, nil
	}

	for _, e := range envs {
		if e.Slug == input {
			return e.Slug, nil
		}
	}
	return "", fmt.Errorf("environment %q not found in %s/%s", input, wsSlug, projectSlug)
}

// selectService fetches services for the given workspace/project/env and prompts for a selection.
func selectService(reader *bufio.Reader, wsSlug, projectSlug, envSlug string) (string, error) {
	req, _ := http.NewRequest("GET", apiURL("/workspaces/"+wsSlug+"/projects/"+projectSlug+"/envs/"+envSlug+"/services/"), nil)
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

	if len(services) == 0 {
		return "", fmt.Errorf("no services found in %s/%s/%s — create one at %s first", wsSlug, projectSlug, envSlug, cfg.Server)
	}

	fmt.Println()
	fmt.Println("Select a service:")
	for i, s := range services {
		fmt.Printf("  [%d] %s (%s)\n", i+1, s.Name, s.Slug)
	}
	fmt.Print("Enter number or slug: ")

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if n, err := strconv.Atoi(input); err == nil {
		if n < 1 || n > len(services) {
			return "", fmt.Errorf("invalid selection: %d (must be 1-%d)", n, len(services))
		}
		return services[n-1].Slug, nil
	}

	for _, s := range services {
		if s.Slug == input {
			return s.Slug, nil
		}
	}
	return "", fmt.Errorf("service %q not found in %s/%s/%s", input, wsSlug, projectSlug, envSlug)
}
