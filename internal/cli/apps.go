package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(appsCmd)
	appsCmd.AddCommand(appsListCmd)
	appsCmd.AddCommand(appsGetCmd)
	appsCmd.AddCommand(appsDeployCmd)
	appsCmd.AddCommand(appsScaleCmd)
	appsCmd.AddCommand(appsStatusCmd)
}

var appsCmd = &cobra.Command{
	Use:     "apps",
	Aliases: []string{"app", "a"},
	Short:   "Manage applications",
	Long: `Manage applications within a project.

Applications are the deployable units in Ancla. Each application belongs to an
org/project and has its own images, releases, deployments, and configuration.
Use sub-commands to list, inspect, deploy, and scale your applications.`,
	Example: "  ancla apps list my-org/my-project\n  ancla apps get my-org/my-project/my-app\n  ancla apps deploy <app-id>",
	GroupID: "resources",
}

var appsListCmd = &cobra.Command{
	Use:               "list <org>/<project>",
	Short:             "List applications in a project",
	Example:           "  ancla apps list my-org/my-project",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeProjects,
	RunE: func(cmd *cobra.Command, args []string) error {
		req, _ := http.NewRequest("GET", apiURL("/applications/"+args[0]), nil)
		body, err := doRequest(req)
		if err != nil {
			return err
		}

		var apps []struct {
			Name     string `json:"name"`
			Slug     string `json:"slug"`
			Platform string `json:"platform"`
		}
		if err := json.Unmarshal(body, &apps); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		if isJSON() {
			return printJSON(apps)
		}

		var rows [][]string
		for _, a := range apps {
			rows = append(rows, []string{a.Slug, a.Name, a.Platform})
		}
		table([]string{"SLUG", "NAME", "PLATFORM"}, rows)
		return nil
	},
}

var appsGetCmd = &cobra.Command{
	Use:     "get <org>/<project>/<app>",
	Short:   "Get application details",
	Example: "  ancla apps get my-org/my-project/my-app",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		req, _ := http.NewRequest("GET", apiURL("/applications/"+args[0]), nil)
		body, err := doRequest(req)
		if err != nil {
			return err
		}

		var app struct {
			Name             string         `json:"name"`
			Slug             string         `json:"slug"`
			Platform         string         `json:"platform"`
			GithubRepository string         `json:"github_repository"`
			AutoDeployBranch string         `json:"auto_deploy_branch"`
			ProcessCounts    map[string]int `json:"process_counts"`
		}
		if err := json.Unmarshal(body, &app); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		if isJSON() {
			return printJSON(app)
		}

		fmt.Printf("Application: %s (%s)\n", app.Name, app.Slug)
		fmt.Printf("Platform: %s\n", app.Platform)
		if app.GithubRepository != "" {
			fmt.Printf("Repository: %s\n", app.GithubRepository)
		}
		if app.AutoDeployBranch != "" {
			fmt.Printf("Auto-deploy branch: %s\n", app.AutoDeployBranch)
		}
		if len(app.ProcessCounts) > 0 {
			fmt.Println("Processes:")
			for proc, count := range app.ProcessCounts {
				fmt.Printf("  %s: %d\n", proc, count)
			}
		}
		return nil
	},
}

var appsDeployCmd = &cobra.Command{
	Use:     "deploy <app-id>",
	Short:   "Trigger a full deploy for an application",
	Example: "  ancla apps deploy abc12345",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		req, _ := http.NewRequest("POST", apiURL("/applications/"+args[0]+"/deploy"), nil)
		body, err := doRequest(req)
		if err != nil {
			return err
		}

		var result struct {
			ImageID string `json:"image_id"`
		}
		json.Unmarshal(body, &result)
		fmt.Printf("Deploy triggered. Image ID: %s\n", result.ImageID)
		return nil
	},
}

var appsScaleCmd = &cobra.Command{
	Use:     "scale <app-id> <process>=<count> ...",
	Short:   "Scale application processes",
	Example: "  ancla apps scale abc12345 web=2 worker=1",
	Args:    cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		counts := make(map[string]int)
		for _, arg := range args[1:] {
			var proc string
			var count int
			if _, err := fmt.Sscanf(arg, "%[^=]=%d", &proc, &count); err != nil {
				return fmt.Errorf("invalid scale argument %q (expected process=count)", arg)
			}
			counts[proc] = count
		}

		payload, _ := json.Marshal(map[string]any{"process_counts": counts})
		req, _ := http.NewRequest("POST", apiURL("/applications/"+args[0]+"/scale"), bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		if _, err := doRequest(req); err != nil {
			return err
		}

		fmt.Println("Scaled successfully.")
		return nil
	},
}

var appsStatusCmd = &cobra.Command{
	Use:     "status <app-id>",
	Short:   "Show pipeline status for an application",
	Example: "  ancla apps status abc12345",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		req, _ := http.NewRequest("GET", apiURL("/applications/"+args[0]+"/pipeline-status"), nil)
		body, err := doRequest(req)
		if err != nil {
			return err
		}

		var status struct {
			Build   *struct{ Status string } `json:"build"`
			Release *struct{ Status string } `json:"release"`
			Deploy  *struct{ Status string } `json:"deploy"`
		}
		json.Unmarshal(body, &status)

		if isJSON() {
			return printJSON(status)
		}

		var rows [][]string
		buildS, relS, depS := "-", "-", "-"
		if status.Build != nil {
			buildS = colorStatus(status.Build.Status)
		}
		if status.Release != nil {
			relS = colorStatus(status.Release.Status)
		}
		if status.Deploy != nil {
			depS = colorStatus(status.Deploy.Status)
		}
		rows = append(rows, []string{"Build", buildS})
		rows = append(rows, []string{"Release", relS})
		rows = append(rows, []string{"Deploy", depS})
		table([]string{"STAGE", "STATUS"}, rows)
		return nil
	},
}
