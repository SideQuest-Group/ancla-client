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

		// Get latest deploy from the deploys list.
		svcPath := servicePath(cfg.Workspace, cfg.Project, cfg.Env, cfg.Service)
		req, _ := http.NewRequest("GET", apiURL(svcPath+"/deploys/"), nil)
		body, err := doRequest(req)
		if err != nil {
			return err
		}

		var deploys struct {
			Items []struct {
				ID string `json:"id"`
			} `json:"items"`
		}
		if err := json.Unmarshal(body, &deploys); err != nil {
			return fmt.Errorf("parsing deploys: %w", err)
		}
		if len(deploys.Items) == 0 || deploys.Items[0].ID == "" {
			fmt.Println("No deployments found.")
			return nil
		}

		deployID := deploys.Items[0].ID
		ep := envPath(cfg.Workspace, cfg.Project, cfg.Env)

		// Fetch deployment logs (env-level endpoint).
		logReq, _ := http.NewRequest("GET", apiURL(ep+"/deploys/"+deployID+"/log"), nil)
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

		shortID := deployID
		if len(shortID) > 8 {
			shortID = shortID[:8]
		}
		fmt.Printf("Deployment %s — %s\n\n", shortID, colorStatus(result.Status))
		if result.LogText != "" {
			fmt.Print(result.LogText)
		} else {
			fmt.Println("(no log output yet)")
		}

		follow, _ := cmd.Flags().GetBool("follow")
		if follow {
			return followDeployLog(ep, deployID)
		}
		return nil
	},
}
