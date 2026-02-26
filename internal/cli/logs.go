package cli

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(logsCmd)
	logsCmd.Flags().BoolP("follow", "f", false, "Follow log output until deployment completes")
}

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Show logs for the linked service's latest deployment",
	Long: `Show deployment logs for the currently linked service.

Requires a fully linked directory (workspace/project/env/service). Fetches
the latest deployment and displays its log output. Use --follow to stream
updates.`,
	Example: "  ancla logs\n  ancla logs -f",
	GroupID: "workflow",
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfg.Workspace == "" || cfg.Project == "" || cfg.Env == "" || cfg.Service == "" {
			return fmt.Errorf("not fully linked — run `ancla link <ws>/<proj>/<env>/<svc>` first")
		}

		// Get service pipeline status to find latest deployment
		svcPath := "/workspaces/" + cfg.Workspace + "/projects/" + cfg.Project + "/envs/" + cfg.Env + "/services/" + cfg.Service
		req, _ := http.NewRequest("GET", apiURL(svcPath+"/pipeline-status"), nil)
		body, err := doRequest(req)
		if err != nil {
			return err
		}

		var status struct {
			Deploy *struct {
				ID     string `json:"id"`
				Status string `json:"status"`
			} `json:"deploy"`
		}
		if err := json.Unmarshal(body, &status); err != nil {
			return fmt.Errorf("parsing pipeline status: %w", err)
		}
		if status.Deploy == nil || status.Deploy.ID == "" {
			fmt.Println("No deployments found.")
			return nil
		}

		deployID := status.Deploy.ID

		// Fetch deployment logs
		logReq, _ := http.NewRequest("GET", apiURL("/deploys/"+deployID+"/log"), nil)
		logBody, err := doRequest(logReq)
		if err != nil {
			return err
		}

		var result struct {
			Status  string `json:"status"`
			LogText string `json:"log_text"`
		}
		json.Unmarshal(logBody, &result)

		if isJSON() {
			return printJSON(result)
		}

		fmt.Printf("Deployment %s — %s\n\n", deployID[:8], colorStatus(result.Status))
		if result.LogText != "" {
			fmt.Print(result.LogText)
		} else {
			fmt.Println("(no log output yet)")
		}

		follow, _ := cmd.Flags().GetBool("follow")
		if follow {
			return followDeployLog(deployID)
		}
		return nil
	},
}
