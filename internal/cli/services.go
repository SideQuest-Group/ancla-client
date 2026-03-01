package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/SideQuest-Group/ancla-client/internal/config"
)

func init() {
	rootCmd.AddCommand(servicesCmd)
	servicesCmd.AddCommand(servicesListCmd)
	servicesCmd.AddCommand(servicesGetCmd)
	servicesCmd.AddCommand(servicesDeployCmd)
	servicesCmd.AddCommand(servicesScaleCmd)
	servicesCmd.AddCommand(servicesStatusCmd)
	servicesScaleCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
}

var servicesCmd = &cobra.Command{
	Use:     "services",
	Aliases: []string{"svc", "s"},
	Short:   "Manage services",
	Long: `Manage services within a project environment.

Services are the deployable units in Ancla. Each service belongs to a
workspace/project/environment and has its own builds, deploys, and configuration.
Use sub-commands to list, inspect, deploy, and scale your services.`,
	Example: "  ancla services list my-ws/my-proj/staging\n  ancla services get my-ws/my-proj/staging/my-svc\n  ancla services deploy my-ws/my-proj/staging/my-svc",
	GroupID: "resources",
	RunE: func(cmd *cobra.Command, args []string) error {
		return servicesListCmd.RunE(cmd, args)
	},
}

// resolveServicePath extracts ws/proj/env/svc from a slash-separated argument,
// falling back to cfg fields for missing segments. Returns an error if the
// workspace segment is empty (minimum required context).
func resolveServicePath(args []string) (ws, proj, env, svc string, err error) {
	var arg string
	if len(args) > 0 {
		arg = args[0]
	}
	ws, proj, env, svc, err = config.ResolveServicePath(arg, cfg)
	if err != nil {
		return
	}
	if ws == "" {
		err = fmt.Errorf("workspace is required — provide <ws>/... or run `ancla link`")
	}
	return
}

// envPath builds the nested API path prefix up to the environment level.
func envPath(ws, proj, env string) string {
	return fmt.Sprintf("/workspaces/%s/projects/%s/envs/%s", ws, proj, env)
}

// servicePath builds the nested API path prefix for a service resource.
func servicePath(ws, proj, env, svc string) string {
	return envPath(ws, proj, env) + "/services/" + svc
}

// serviceBasePath builds the nested API path prefix up to the environment level.
func serviceBasePath(ws, proj, env string) string {
	return envPath(ws, proj, env) + "/services/"
}

var servicesListCmd = &cobra.Command{
	Use:               "list <ws>/<proj>/<env>",
	Short:             "List services in an environment",
	Example:           "  ancla services list my-ws/my-proj/staging",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeProjects,
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, proj, env, _, err := resolveServicePath(args)
		if err != nil {
			return err
		}
		if proj == "" || env == "" {
			return fmt.Errorf("usage: services list <ws>/<proj>/<env>")
		}

		req, _ := http.NewRequest("GET", apiURL(serviceBasePath(ws, proj, env)), nil)
		body, err := doRequest(req)
		if err != nil {
			return err
		}

		var services []struct {
			Name     string `json:"name"`
			Slug     string `json:"slug"`
			Platform string `json:"platform"`
		}
		if err := json.Unmarshal(body, &services); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		if isJSON() {
			return printJSON(services)
		}

		var rows [][]string
		for _, s := range services {
			rows = append(rows, []string{s.Slug, s.Name, s.Platform})
		}
		table([]string{"SLUG", "NAME", "PLATFORM"}, rows)
		return nil
	},
}

var servicesGetCmd = &cobra.Command{
	Use:     "get <ws>/<proj>/<env>/<svc>",
	Short:   "Get service details",
	Example: "  ancla services get my-ws/my-proj/staging/my-svc",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, proj, env, svc, err := resolveServicePath(args)
		if err != nil {
			return err
		}
		if proj == "" || env == "" || svc == "" {
			return fmt.Errorf("usage: services get <ws>/<proj>/<env>/<svc>")
		}

		req, _ := http.NewRequest("GET", apiURL(servicePath(ws, proj, env, svc)), nil)
		body, err := doRequest(req)
		if err != nil {
			return err
		}

		var service struct {
			Name             string         `json:"name"`
			Slug             string         `json:"slug"`
			Platform         string         `json:"platform"`
			GithubRepository string         `json:"github_repository"`
			AutoDeployBranch string         `json:"auto_deploy_branch"`
			ProcessCounts    map[string]int `json:"process_counts"`
		}
		if err := json.Unmarshal(body, &service); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		if isJSON() {
			return printJSON(service)
		}

		fmt.Printf("Service: %s (%s)\n", service.Name, service.Slug)
		fmt.Printf("Platform: %s\n", service.Platform)
		if service.GithubRepository != "" {
			fmt.Printf("Repository: %s\n", service.GithubRepository)
		}
		if service.AutoDeployBranch != "" {
			fmt.Printf("Auto-deploy branch: %s\n", service.AutoDeployBranch)
		}
		if len(service.ProcessCounts) > 0 {
			fmt.Println("Processes:")
			for proc, count := range service.ProcessCounts {
				fmt.Printf("  %s: %d\n", proc, count)
			}
		}
		return nil
	},
}

var servicesDeployCmd = &cobra.Command{
	Use:     "deploy <ws>/<proj>/<env>/<svc>",
	Short:   "Trigger a full deploy for a service",
	Example: "  ancla services deploy my-ws/my-proj/staging/my-svc",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, proj, env, svc, err := resolveServicePath(args)
		if err != nil {
			return err
		}
		if proj == "" || env == "" || svc == "" {
			return fmt.Errorf("usage: services deploy <ws>/<proj>/<env>/<svc>")
		}

		stop := spin("Deploying...")
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
			fmt.Println("Deploy likely succeeded, but the response could not be parsed (unexpected format).")
			return nil
		}
		fmt.Printf("Deploy triggered. Build ID: %s\n", result.BuildID)
		return nil
	},
}

var servicesScaleCmd = &cobra.Command{
	Use:     "scale <ws>/<proj>/<env>/<svc> <process>=<count> ...",
	Short:   "Scale service processes",
	Example: "  ancla services scale my-ws/my-proj/staging/my-svc web=2 worker=1",
	Args:    cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, proj, env, svc, err := resolveServicePath(args)
		if err != nil {
			return err
		}
		if proj == "" || env == "" || svc == "" {
			return fmt.Errorf("usage: services scale <ws>/<proj>/<env>/<svc> <process>=<count> ...")
		}

		counts := make(map[string]int)
		for _, arg := range args[1:] {
			parts := strings.SplitN(arg, "=", 2)
			if len(parts) != 2 || parts[0] == "" {
				return fmt.Errorf("invalid scale argument %q (expected process=count)", arg)
			}
			count, err := strconv.Atoi(parts[1])
			if err != nil {
				return fmt.Errorf("invalid scale argument %q: count must be an integer", arg)
			}
			counts[parts[0]] = count
		}

		// Warn when scaling any process to 0 — this effectively stops it.
		for proc, count := range counts {
			if count == 0 {
				msg := fmt.Sprintf("Scaling %q to 0 will stop the process.", proc)
				if !confirmAction(cmd, msg) {
					fmt.Println("Aborted.")
					return nil
				}
				break // only need to confirm once
			}
		}

		stop := spin("Scaling...")
		payload, _ := json.Marshal(map[string]any{"process_counts": counts})
		req, _ := http.NewRequest("POST", apiURL(servicePath(ws, proj, env, svc)+"/scale"), bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		if _, err := doRequest(req); err != nil {
			stop()
			return err
		}
		stop()

		fmt.Println("Scaled successfully.")
		return nil
	},
}

var servicesStatusCmd = &cobra.Command{
	Use:     "status <ws>/<proj>/<env>/<svc>",
	Short:   "Show pipeline status for a service",
	Example: "  ancla services status my-ws/my-proj/staging/my-svc",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, proj, env, svc, err := resolveServicePath(args)
		if err != nil {
			return err
		}
		if proj == "" || env == "" || svc == "" {
			return fmt.Errorf("usage: services status <ws>/<proj>/<env>/<svc>")
		}

		req, _ := http.NewRequest("GET", apiURL(pipelineStatusPath(ws, proj, env, svc)), nil)
		body, err := doRequest(req)
		if err != nil {
			return err
		}

		var status struct {
			Build  *struct{ Status string } `json:"build"`
			Deploy *struct{ Status string } `json:"deploy"`
		}
		if err := json.Unmarshal(body, &status); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		if isJSON() {
			return printJSON(status)
		}

		var rows [][]string
		buildS, depS := "-", "-"
		if status.Build != nil {
			buildS = colorStatus(status.Build.Status)
		}
		if status.Deploy != nil {
			depS = colorStatus(status.Deploy.Status)
		}
		rows = append(rows, []string{"Build", buildS})
		rows = append(rows, []string{"Deploy", depS})
		table([]string{"STAGE", "STATUS"}, rows)
		return nil
	},
}
