package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(deploymentsCmd)
	deploymentsCmd.AddCommand(deploymentsGetCmd)
	deploymentsCmd.AddCommand(deploymentsLogCmd)
	deploymentsGetCmd.Flags().BoolP("follow", "f", false, "Follow deployment progress until complete")
	deploymentsLogCmd.Flags().BoolP("follow", "f", false, "Poll for log updates until deployment completes")
}

var deploymentsCmd = &cobra.Command{
	Use:     "deployments",
	Aliases: []string{"deployment", "dep"},
	Short:   "Manage deployments",
	Long: `Manage deployments for your applications.

Deployments represent the rollout of a release to your infrastructure. Each
deployment tracks its progress and can be inspected for status, errors, and
logs. Use sub-commands to view deployment details or stream deployment logs.`,
	Example: "  ancla deployments get <deployment-id>\n  ancla deployments log <deployment-id>",
	GroupID: "resources",
}

var deploymentsGetCmd = &cobra.Command{
	Use:     "get <deployment-id>",
	Short:   "Get deployment details",
	Example: "  ancla deployments get abc12345",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		req, _ := http.NewRequest("GET", apiURL("/deployments/"+args[0]+"/detail"), nil)
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

		fmt.Printf("Deployment: %s\n", dpl.ID)
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
			return followDeployment(args[0])
		}
		return nil
	},
}

var deploymentsLogCmd = &cobra.Command{
	Use:     "log <deployment-id>",
	Short:   "Show deployment log",
	Example: "  ancla deployments log abc12345",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		req, _ := http.NewRequest("GET", apiURL("/deployments/"+args[0]+"/log"), nil)
		body, err := doRequest(req)
		if err != nil {
			return err
		}

		var result struct {
			Status  string `json:"status"`
			LogText string `json:"log_text"`
		}
		json.Unmarshal(body, &result)

		fmt.Printf("Deployment — %s\n\n", result.Status)
		if result.LogText != "" {
			fmt.Println(result.LogText)
		} else {
			fmt.Println("(no log output yet)")
		}

		follow, _ := cmd.Flags().GetBool("follow")
		if follow {
			return followDeploymentLog(args[0])
		}
		return nil
	},
}

// followDeployment polls deployment status until complete or error.
func followDeployment(deploymentID string) error {
	stop := spin("Deploying...")
	defer stop()

	for {
		time.Sleep(3 * time.Second)
		req, _ := http.NewRequest("GET", apiURL("/deployments/"+deploymentID+"/detail"), nil)
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
				return fmt.Errorf("deployment failed: %s", dpl.ErrorDtl)
			}
			return fmt.Errorf("deployment failed")
		}
		if dpl.Complete {
			stop()
			fmt.Println("\nDeployment complete.")
			return nil
		}
	}
}

// followDeploymentLog polls deployment logs until complete or error.
func followDeploymentLog(deploymentID string) error {
	var lastLen int
	stop := spin("Deploying...")
	defer stop()

	for {
		time.Sleep(3 * time.Second)
		req, _ := http.NewRequest("GET", apiURL("/deployments/"+deploymentID+"/log"), nil)
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
			fmt.Println("\nDeployment complete.")
			return nil
		case "error", "failed":
			stop()
			return fmt.Errorf("deployment failed")
		}
	}
}
