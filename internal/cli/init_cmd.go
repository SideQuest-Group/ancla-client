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
	Short: "Link this directory to an Ancla org/project/app",
	Long: `Initialize (link) the current directory to an Ancla application.

This interactive command walks you through selecting an organization, project,
and application, then saves the link context to a local .ancla/config.yaml file.
Subsequent commands run from this directory will automatically use the linked
org, project, and app without requiring explicit arguments.`,
	Example: `  ancla init
  ancla init   # re-link to a different app`,
	GroupID: "workflow",
	RunE:    runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	// Step 1: Check if already linked
	if cfg.IsLinked() {
		fmt.Printf("This directory is already linked to: %s/%s/%s\n", cfg.Org, cfg.Project, cfg.App)
		fmt.Print("Continue and re-link? [y/N] ")
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" && answer != "yes" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	// Step 2: Fetch and select organization
	orgSlug, err := selectOrg(reader)
	if err != nil {
		return err
	}

	// Step 3: Fetch and select project
	projectSlug, err := selectProject(reader, orgSlug)
	if err != nil {
		return err
	}

	// Step 4: Fetch and select application
	appSlug, err := selectApp(reader, orgSlug, projectSlug)
	if err != nil {
		return err
	}

	// Step 5: Save link context
	cfg.Org = orgSlug
	cfg.Project = projectSlug
	cfg.App = appSlug

	if err := config.SaveLocal(cfg); err != nil {
		return fmt.Errorf("saving local config: %w", err)
	}

	// Step 6: Print summary
	fmt.Println()
	fmt.Println("Linked successfully!")
	fmt.Printf("  Org:     %s\n", orgSlug)
	fmt.Printf("  Project: %s\n", projectSlug)
	fmt.Printf("  App:     %s\n", appSlug)
	fmt.Println()
	fmt.Println("Saved to .ancla/config.yaml")
	return nil
}

// selectOrg fetches the user's organizations and prompts for a selection.
func selectOrg(reader *bufio.Reader) (string, error) {
	req, _ := http.NewRequest("GET", apiURL("/organizations/"), nil)
	body, err := doRequest(req)
	if err != nil {
		return "", fmt.Errorf("fetching organizations: %w", err)
	}

	var orgs []struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	if err := json.Unmarshal(body, &orgs); err != nil {
		return "", fmt.Errorf("parsing organizations: %w", err)
	}

	if len(orgs) == 0 {
		return "", fmt.Errorf("no organizations found — create one at %s first", cfg.Server)
	}

	fmt.Println()
	fmt.Println("Select an organization:")
	for i, o := range orgs {
		fmt.Printf("  [%d] %s (%s)\n", i+1, o.Name, o.Slug)
	}
	fmt.Print("Enter number or slug: ")

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	// Try as number first
	if n, err := strconv.Atoi(input); err == nil {
		if n < 1 || n > len(orgs) {
			return "", fmt.Errorf("invalid selection: %d (must be 1-%d)", n, len(orgs))
		}
		return orgs[n-1].Slug, nil
	}

	// Otherwise treat as slug — verify it exists
	for _, o := range orgs {
		if o.Slug == input {
			return o.Slug, nil
		}
	}
	return "", fmt.Errorf("organization %q not found", input)
}

// selectProject fetches projects for the given org and prompts for a selection.
func selectProject(reader *bufio.Reader, orgSlug string) (string, error) {
	req, _ := http.NewRequest("GET", apiURL("/projects/"), nil)
	body, err := doRequest(req)
	if err != nil {
		return "", fmt.Errorf("fetching projects: %w", err)
	}

	var allProjects []struct {
		Name             string `json:"name"`
		Slug             string `json:"slug"`
		OrganizationSlug string `json:"organization_slug"`
	}
	if err := json.Unmarshal(body, &allProjects); err != nil {
		return "", fmt.Errorf("parsing projects: %w", err)
	}

	// Filter to the selected org
	var projects []struct {
		Name             string `json:"name"`
		Slug             string `json:"slug"`
		OrganizationSlug string `json:"organization_slug"`
	}
	for _, p := range allProjects {
		if p.OrganizationSlug == orgSlug {
			projects = append(projects, p)
		}
	}

	if len(projects) == 0 {
		return "", fmt.Errorf("no projects found in org %q — create one at %s first", orgSlug, cfg.Server)
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
	return "", fmt.Errorf("project %q not found in org %q", input, orgSlug)
}

// selectApp fetches applications for the given org/project and prompts for a selection.
func selectApp(reader *bufio.Reader, orgSlug, projectSlug string) (string, error) {
	req, _ := http.NewRequest("GET", apiURL("/applications/"+orgSlug+"/"+projectSlug), nil)
	body, err := doRequest(req)
	if err != nil {
		return "", fmt.Errorf("fetching applications: %w", err)
	}

	var apps []struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	if err := json.Unmarshal(body, &apps); err != nil {
		return "", fmt.Errorf("parsing applications: %w", err)
	}

	if len(apps) == 0 {
		return "", fmt.Errorf("no applications found in %s/%s — create one at %s first", orgSlug, projectSlug, cfg.Server)
	}

	fmt.Println()
	fmt.Println("Select an application:")
	for i, a := range apps {
		fmt.Printf("  [%d] %s (%s)\n", i+1, a.Name, a.Slug)
	}
	fmt.Print("Enter number or slug: ")

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if n, err := strconv.Atoi(input); err == nil {
		if n < 1 || n > len(apps) {
			return "", fmt.Errorf("invalid selection: %d (must be 1-%d)", n, len(apps))
		}
		return apps[n-1].Slug, nil
	}

	for _, a := range apps {
		if a.Slug == input {
			return a.Slug, nil
		}
	}
	return "", fmt.Errorf("application %q not found in %s/%s", input, orgSlug, projectSlug)
}
