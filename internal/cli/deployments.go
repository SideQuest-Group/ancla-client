package cli

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(deploymentsCmd)
	deploymentsCmd.AddCommand(deploymentsGetCmd)
	deploymentsCmd.AddCommand(deploymentsLogCmd)
}

var deploymentsCmd = &cobra.Command{
	Use:   "deployments",
	Short: "Manage deployments",
}

var deploymentsGetCmd = &cobra.Command{
	Use:   "get <deployment-id>",
	Short: "Get deployment details",
	Args:  cobra.ExactArgs(1),
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

		status := "in progress"
		if dpl.Error {
			status = "error"
		} else if dpl.Complete {
			status = "complete"
		}

		fmt.Printf("Deployment: %s\n", dpl.ID)
		fmt.Printf("Status: %s\n", status)
		if dpl.ErrorDtl != "" {
			fmt.Printf("Error: %s\n", dpl.ErrorDtl)
		}
		if dpl.Created != "" {
			fmt.Printf("Created: %s\n", dpl.Created)
		}
		if dpl.Updated != "" {
			fmt.Printf("Updated: %s\n", dpl.Updated)
		}
		return nil
	},
}

var deploymentsLogCmd = &cobra.Command{
	Use:   "log <deployment-id>",
	Short: "Show deployment log",
	Args:  cobra.ExactArgs(1),
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
		return nil
	},
}
