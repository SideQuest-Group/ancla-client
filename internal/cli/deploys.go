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
	Example: "  ancla deploys list my-ws/my-proj/staging/my-svc\n  ancla deploys get <ws>/<proj>/<env>/<svc> <deploy-id>\n  ancla deploys log <ws>/<proj>/<env>/<svc> <deploy-id>",
	GroupID: "resources",
	RunE: func(cmd *cobra.Command, args []string) error {
		return deploysListCmd.RunE(cmd, args)
	},
}

var deploysListCmd = &cobra.Command{
	Use:     "list [<ws>/<proj>/<env>/<svc>]",
	Short:   "List deploys for a service",
	Example: "  ancla deploys list\n  ancla deploys list my-ws/my-proj/staging/my-svc",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, proj, env, svc, err := resolveServicePath(args)
		if err != nil {
			return err
		}
		if proj == "" || env == "" || svc == "" {
			return fmt.Errorf("no linked service — provide <ws>/<proj>/<env>/<svc>, or run `ancla link`")
		}

		req, _ := http.NewRequest("GET", apiURL(servicePath(ws, proj, env, svc)+"/deploys/"), nil)
		body, err := doRequest(req)
		if err != nil {
			return err
		}

		var items []struct {
			ID       string `json:"id"`
			Complete bool   `json:"complete"`
			Error    bool   `json:"error"`
			Created  string `json:"created"`
		}
		if err := json.Unmarshal(body, &items); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		if isJSON() {
			return printJSON(items)
		}

		var rows [][]string
		for _, d := range items {
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
	Use:     "get [<ws>/<proj>/<env>/<svc>] <deploy-id>",
	Short:   "Get deploy details",
	Example: "  ancla deploys get abc12345\n  ancla deploys get my-ws/my-proj/staging/my-svc abc12345",
	Args:    cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ep, deployID, err := resolveDeployArgs(args)
		if err != nil {
			return err
		}

		req, _ := http.NewRequest("GET", apiURL(ep+"/deploys/"+deployID), nil)
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
			return followDeploy(ep, deployID)
		}
		return nil
	},
}

var deploysLogCmd = &cobra.Command{
	Use:     "log [<ws>/<proj>/<env>/<svc>] <deploy-id>",
	Short:   "Show deploy log",
	Example: "  ancla deploys log abc12345\n  ancla deploys log my-ws/my-proj/staging/my-svc abc12345",
	Args:    cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ep, deployID, err := resolveDeployArgs(args)
		if err != nil {
			return err
		}

		req, _ := http.NewRequest("GET", apiURL(ep+"/deploys/"+deployID+"/log"), nil)
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
			return followDeployLog(ep, deployID)
		}
		return nil
	},
}

// resolveDeployArgs handles two calling conventions:
//
//	deploys get <deploy-id>                         — uses linked service context
//	deploys get <ws>/<proj>/<env>/<svc> <deploy-id> — explicit path
//
// Returns the env-level path prefix and deploy ID.
func resolveDeployArgs(args []string) (ep, deployID string, err error) {
	if len(args) == 2 {
		ws, proj, env, _, e := resolveServicePath(args[:1])
		if e != nil {
			return "", "", e
		}
		if proj == "" || env == "" {
			return "", "", fmt.Errorf("at least <ws>/<proj>/<env> required")
		}
		return envPath(ws, proj, env), args[1], nil
	}
	// Single arg — deploy ID, resolve from linked config.
	ws, proj, env, svc, e := resolveServicePath(nil)
	if e != nil || ws == "" || proj == "" || env == "" || svc == "" {
		return "", "", fmt.Errorf("no linked service — provide <ws>/<proj>/<env>/<svc> before the deploy ID, or run `ancla link`")
	}
	return envPath(ws, proj, env), args[0], nil
}

// followDeploy polls deploy status until complete or error.
func followDeploy(ep, deployID string) error {
	stop := spin("Deploying...")
	defer stop()

	for {
		time.Sleep(3 * time.Second)
		req, _ := http.NewRequest("GET", apiURL(ep+"/deploys/"+deployID), nil)
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
				return fmt.Errorf("%s %s", stError.Render(symCross+" Deploy failed:"), dpl.ErrorDtl)
			}
			return fmt.Errorf("%s", stError.Render(symCross+" Deploy failed"))
		}
		if dpl.Complete {
			stop()
			fmt.Println("\n" + stSuccess.Render(symCheck+" Deploy complete."))
			return nil
		}
	}
}

// followDeployLog polls deploy logs until complete or error.
func followDeployLog(ep, deployID string) error {
	var lastLen int
	stop := spin("Deploying...")
	defer stop()

	for {
		time.Sleep(3 * time.Second)
		req, _ := http.NewRequest("GET", apiURL(ep+"/deploys/"+deployID+"/log"), nil)
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
			fmt.Println("\n" + stSuccess.Render(symCheck+" Deploy complete."))
			return nil
		case "error", "failed":
			stop()
			return fmt.Errorf("%s", stError.Render(symCross+" Deploy failed"))
		}
	}
}
