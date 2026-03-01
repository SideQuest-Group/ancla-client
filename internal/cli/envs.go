package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/cobra"

	"github.com/SideQuest-Group/ancla-client/internal/config"
)

func init() {
	rootCmd.AddCommand(envsCmd)
	envsCmd.AddCommand(envsListCmd)
	envsCmd.AddCommand(envsGetCmd)
	envsCmd.AddCommand(envsCreateCmd)
}

var envsCmd = &cobra.Command{
	Use:     "envs",
	Aliases: []string{"env", "e"},
	Short:   "Manage environments",
	Long: `Manage environments within a project.

Environments represent deployment targets (e.g. staging, production) for
the services in a project. Each environment can have its own configuration,
releases, and scaling settings.
Use sub-commands to list, inspect, or create environments.`,
	Example: "  ancla envs list my-ws/my-proj\n  ancla envs get my-ws/my-proj/staging\n  ancla envs create my-ws/my-proj production",
	GroupID: "resources",
	RunE: func(cmd *cobra.Command, args []string) error {
		return envsListCmd.RunE(cmd, args)
	},
}

var envsListCmd = &cobra.Command{
	Use:     "list [workspace/project]",
	Short:   "List environments in a project",
	Example: "  ancla envs list my-ws/my-proj",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var arg string
		if len(args) == 1 {
			arg = args[0]
		}
		ws, proj, _, _, err := config.ResolveServicePath(arg, cfg)
		if err != nil {
			return err
		}
		if ws == "" || proj == "" {
			return fmt.Errorf("workspace and project are required\n\n  ancla envs <workspace>/<project>\n\n  Hierarchy: workspace → project → env → service\n  Hint: run `ancla link` to set defaults")
		}

		req, _ := http.NewRequest("GET", apiURL("/workspaces/"+ws+"/projects/"+proj+"/envs/"), nil)
		body, err := doRequest(req)
		if err != nil {
			return err
		}

		var envs []struct {
			ID           string `json:"id"`
			Name         string `json:"name"`
			Slug         string `json:"slug"`
			ServiceCount int    `json:"service_count"`
			Created      string `json:"created"`
		}
		if err := json.Unmarshal(body, &envs); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		if isJSON() {
			return printJSON(envs)
		}

		var rows [][]string
		for _, e := range envs {
			rows = append(rows, []string{e.Slug, e.Name, fmt.Sprintf("%d", e.ServiceCount), e.Created})
		}
		table([]string{"SLUG", "NAME", "SERVICES", "CREATED"}, rows)
		return nil
	},
}

var envsGetCmd = &cobra.Command{
	Use:     "get <workspace>/<project>/<env>",
	Short:   "Get environment details",
	Example: "  ancla envs get my-ws/my-proj/staging",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		parts := strings.SplitN(args[0], "/", 3)
		if len(parts) != 3 {
			return fmt.Errorf("argument must be in the form <workspace>/<project>/<env>")
		}
		ws, proj, env := parts[0], parts[1], parts[2]

		req, _ := http.NewRequest("GET", apiURL("/workspaces/"+ws+"/projects/"+proj+"/envs/"+env), nil)
		body, err := doRequest(req)
		if err != nil {
			return err
		}

		var e struct {
			ID           string `json:"id"`
			Name         string `json:"name"`
			Slug         string `json:"slug"`
			ServiceCount int    `json:"service_count"`
			Created      string `json:"created"`
			Updated      string `json:"updated"`
		}
		if err := json.Unmarshal(body, &e); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		if isJSON() {
			return printJSON(e)
		}

		fmt.Printf("Environment: %s (%s)\n", e.Name, e.Slug)
		fmt.Printf("Services: %d\n", e.ServiceCount)
		if e.Created != "" {
			fmt.Printf("Created: %s\n", e.Created)
		}
		if e.Updated != "" {
			fmt.Printf("Updated: %s\n", e.Updated)
		}
		return nil
	},
}

var envsCreateCmd = &cobra.Command{
	Use:     "create <workspace>/<project> <name>",
	Short:   "Create a new environment",
	Example: "  ancla envs create my-ws/my-proj production",
	Args:    cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		parts := strings.SplitN(args[0], "/", 2)
		if len(parts) != 2 {
			return fmt.Errorf("first argument must be in the form <workspace>/<project>")
		}
		ws, proj := parts[0], parts[1]
		name := args[1]

		payload, _ := json.Marshal(map[string]string{"name": name})
		req, _ := http.NewRequest("POST", apiURL("/workspaces/"+ws+"/projects/"+proj+"/envs/"), bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")

		stop := spin("Creating environment...")
		body, err := doRequest(req)
		stop()
		if err != nil {
			return err
		}

		var e struct {
			ID   string `json:"id"`
			Name string `json:"name"`
			Slug string `json:"slug"`
		}
		if err := json.Unmarshal(body, &e); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		if isJSON() {
			return printJSON(e)
		}

		fmt.Printf("Created environment: %s (%s)\n", e.Name, e.Slug)
		return nil
	},
}
