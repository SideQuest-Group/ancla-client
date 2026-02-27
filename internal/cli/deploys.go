package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(deploysCmd)
	deploysCmd.AddCommand(deploysListCmd)
	deploysCmd.AddCommand(deploysGetCmd)
	deploysCmd.AddCommand(deploysLogCmd)
	deploysGetCmd.Flags().BoolP("follow", "f", false, "Follow deployment progress until complete")
	deploysLogCmd.Flags().BoolP("follow", "f", false, "Poll for log updates until deployment completes")
}

var deploysCmd = &cobra.Command{
	Use:     "deploys",
	Aliases: []string{"d"},
	Short:   "Manage deploys",
	Long: `Manage deploys for your services.

Deploys represent the rollout of a build to your infrastructure. Each deploy
tracks its progress and can be inspected for status, errors, and logs.
Use sub-commands to list deploys, view details, or stream deploy logs.`,
	Example: "  ancla deploys list my-ws/my-proj/staging/my-svc\n  ancla deploys get <deploy-id>\n  ancla deploys log <deploy-id>",
	GroupID: "resources",
}

var deploysListCmd = &cobra.Command{
	Use:     "list <ws>/<proj>/<env>/<svc>",
	Short:   "List deploys for a service",
	Example: "  ancla deploys list my-ws/my-proj/staging/my-svc",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, proj, env, svc, err := resolveServicePath(args)
		if err != nil {
			return err
		}
		if proj == "" || env == "" || svc == "" {
			return fmt.Errorf("usage: deploys list <ws>/<proj>/<env>/<svc>")
		}

		req, _ := http.NewRequest("GET", apiURL(servicePath(ws, proj, env, svc)+"/deploys/"), nil)
		body, err := doRequest(req)
		if err != nil {
			return err
		}

		var result struct {
			Items []struct {
				ID       string `json:"id"`
				Complete bool   `json:"complete"`
				Error    bool   `json:"error"`
				Created  string `json:"created"`
			} `json:"items"`
		}
		if err := json.Unmarshal(body, &result); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		if isJSON() {
			return printJSON(result)
		}

		var rows [][]string
		for _, d := range result.Items {
			status := "in progress"
			if d.Error {
				status = "error"
			} else if d.Complete {
				status = "complete"
			}
			id := d.ID
			if len(id) > 8 {
				id = id[:8]
			}
			rows = append(rows, []string{id, colorStatus(status), d.Created})
		}
		table([]string{"ID", "STATUS", "CREATED"}, rows)
		return nil
	},
}

var deploysGetCmd = &cobra.Command{
	Use:     "get <deploy-id>",
	Short:   "Get deploy details",
	Example: "  ancla deploys get abc12345",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		req, _ := http.NewRequest("GET", apiURL("/deploys/"+args[0]+"/detail"), nil)
		body, err := doRequest(req)
		if err != nil {
			return err
		}

		var dpl struct {
			ID       string `json:"id"`
			Complete bool   `json:"complete"`
			Error    bool   `json:"error"`
			ErrorDtl string `json:"error_detail"`
			JobID    string `json:"job_id"`
			Created  string `json:"created"`
			Updated  string `json:"updated"`
		}
		if err := json.Unmarshal(body, &dpl); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		if isJSON() {
			return printJSON(dpl)
		}

		status := "in progress"
		if dpl.Error {
			status = "error"
		} else if dpl.Complete {
			status = "complete"
		}

		fmt.Printf("Deploy: %s\n", dpl.ID)
		fmt.Printf("Status: %s\n", colorStatus(status))
		if dpl.ErrorDtl != "" {
			fmt.Printf("Error: %s\n", dpl.ErrorDtl)
		}
		if dpl.Created != "" {
			fmt.Printf("Created: %s\n", dpl.Created)
		}
		if dpl.Updated != "" {
			fmt.Printf("Updated: %s\n", dpl.Updated)
		}

		follow, _ := cmd.Flags().GetBool("follow")
		if follow && !dpl.Complete && !dpl.Error {
			return followDeploy(args[0])
		}
		return nil
	},
}

var deploysLogCmd = &cobra.Command{
	Use:     "log <deploy-id>",
	Short:   "Show deploy log",
	Example: "  ancla deploys log abc12345",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		req, _ := http.NewRequest("GET", apiURL("/deploys/"+args[0]+"/log"), nil)
		body, err := doRequest(req)
		if err != nil {
			return err
		}

		var result struct {
			Status  string `json:"status"`
			LogText string `json:"log_text"`
		}
		json.Unmarshal(body, &result)

		fmt.Printf("Deploy — %s\n\n", result.Status)
		if result.LogText != "" {
			fmt.Println(result.LogText)
		} else {
			fmt.Println("(no log output yet)")
		}

		follow, _ := cmd.Flags().GetBool("follow")
		if follow {
			return followDeployLog(args[0])
		}
		return nil
	},
}

// followDeploy polls deploy status until complete or error.
func followDeploy(deployID string) error {
	stop := spin("Deploying...")
	defer stop()

	for {
		time.Sleep(3 * time.Second)
		req, _ := http.NewRequest("GET", apiURL("/deploys/"+deployID+"/detail"), nil)
		body, err := doRequest(req)
		if err != nil {
			return err
		}
		var dpl struct {
			Complete bool   `json:"complete"`
			Error    bool   `json:"error"`
			ErrorDtl string `json:"error_detail"`
		}
		json.Unmarshal(body, &dpl)

		if dpl.Error {
			stop()
			if dpl.ErrorDtl != "" {
				return fmt.Errorf("deploy failed: %s", dpl.ErrorDtl)
			}
			return fmt.Errorf("deploy failed")
		}
		if dpl.Complete {
			stop()
			fmt.Println("\nDeploy complete.")
			return nil
		}
	}
}

// followDeployLog polls deploy logs until complete or error.
func followDeployLog(deployID string) error {
	var lastLen int
	stop := spin("Deploying...")
	defer stop()

	for {
		time.Sleep(3 * time.Second)
		req, _ := http.NewRequest("GET", apiURL("/deploys/"+deployID+"/log"), nil)
		body, err := doRequest(req)
		if err != nil {
			return err
		}
		var result struct {
			Status  string `json:"status"`
			LogText string `json:"log_text"`
		}
		json.Unmarshal(body, &result)

		if len(result.LogText) > lastLen {
			stop()
			fmt.Print(result.LogText[lastLen:])
			lastLen = len(result.LogText)
			stop = spin("Deploying...")
		}

		switch result.Status {
		case "complete", "success":
			stop()
			fmt.Println("\nDeploy complete.")
			return nil
		case "error", "failed":
			stop()
			return fmt.Errorf("deploy failed")
		}
	}
}
