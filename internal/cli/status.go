package cli

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of the linked workspace/project/env/service",
	Long: `Show a unified status view for the currently linked resource.

Requires a linked directory (see ancla link). Displays the workspace, project,
environment, service details, and current pipeline status in a single view.`,
	Example: "  ancla status",
	GroupID: "workflow",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !cfg.IsLinked() {
			return fmt.Errorf("not linked — run `ancla link <ws>/<proj>/<env>/<svc>` first")
		}

		type statusOutput struct {
			Workspace string `json:"workspace"`
			Project   string `json:"project,omitempty"`
			Env       string `json:"env,omitempty"`
			Service   string `json:"service,omitempty"`
			Build     string `json:"build,omitempty"`
			Deploy    string `json:"deploy,omitempty"`
		}
		out := statusOutput{
			Workspace: cfg.Workspace,
			Project:   cfg.Project,
			Env:       cfg.Env,
			Service:   cfg.Service,
		}

		// If we have a full service path, fetch pipeline status
		if cfg.Workspace != "" && cfg.Project != "" && cfg.Env != "" && cfg.Service != "" {
			svcPath := "/workspaces/" + cfg.Workspace + "/projects/" + cfg.Project + "/envs/" + cfg.Env + "/services/" + cfg.Service
			req, _ := http.NewRequest("GET", apiURL(svcPath+"/pipeline/status"), nil)
			body, err := doRequest(req)
			if err == nil {
				var status struct {
					Build  *struct{ Status string } `json:"build"`
					Deploy *struct{ Status string } `json:"deploy"`
				}
				json.Unmarshal(body, &status)
				if status.Build != nil {
					out.Build = status.Build.Status
				}
				if status.Deploy != nil {
					out.Deploy = status.Deploy.Status
				}
			}
		}

		if isJSON() {
			return printJSON(out)
		}

		fmt.Println(stHeading.Render(symAnchor + " Status"))
		fmt.Println()
		fmt.Println(kv("Workspace", out.Workspace))
		if out.Project != "" {
			fmt.Println(kv("Project", out.Project))
		}
		if out.Env != "" {
			fmt.Println(kv("Environment", out.Env))
		}
		if out.Service != "" {
			fmt.Println(kv("Service", out.Service))
		}

		if out.Build != "" || out.Deploy != "" {
			fmt.Println()
			if out.Build != "" {
				fmt.Println(kv("Build", colorStatus(out.Build)))
			}
			if out.Deploy != "" {
				fmt.Println(kv("Deploy", colorStatus(out.Deploy)))
			}
		}

		return nil
	},
}
