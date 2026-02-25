package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(releasesCmd)
	releasesCmd.AddCommand(releasesListCmd)
	releasesCmd.AddCommand(releasesCreateCmd)
	releasesCmd.AddCommand(releasesDeployCmd)
}

var releasesCmd = &cobra.Command{
	Use:   "releases",
	Short: "Manage releases",
}

var releasesListCmd = &cobra.Command{
	Use:   "list <app-id>",
	Short: "List releases for an application",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		req, _ := http.NewRequest("GET", apiURL("/releases/"+args[0]), nil)
		body, err := doRequest(req)
		if err != nil {
			return err
		}

		var result struct {
			Items []struct {
				ID       string `json:"id"`
				Version  int    `json:"version"`
				Platform string `json:"platform"`
				Built    bool   `json:"built"`
				Error    bool   `json:"error"`
				Created  string `json:"created"`
			} `json:"items"`
		}
		if err := json.Unmarshal(body, &result); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "VERSION\tID\tPLATFORM\tSTATUS\tCREATED")
		for _, r := range result.Items {
			status := "building"
			if r.Error {
				status = "error"
			} else if r.Built {
				status = "built"
			}
			id := r.ID
			if len(id) > 8 {
				id = id[:8]
			}
			fmt.Fprintf(w, "v%d\t%s\t%s\t%s\t%s\n", r.Version, id, r.Platform, status, r.Created)
		}
		return w.Flush()
	},
}

var releasesCreateCmd = &cobra.Command{
	Use:   "create <app-id>",
	Short: "Create a new release",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		req, _ := http.NewRequest("POST", apiURL("/releases/"+args[0]+"/create"), nil)
		body, err := doRequest(req)
		if err != nil {
			return err
		}

		var result struct {
			ReleaseID string `json:"release_id"`
			Version   int    `json:"version"`
		}
		json.Unmarshal(body, &result)
		fmt.Printf("Release created: %s (v%d)\n", result.ReleaseID, result.Version)
		return nil
	},
}

var releasesDeployCmd = &cobra.Command{
	Use:   "deploy <release-id>",
	Short: "Deploy a release",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		req, _ := http.NewRequest("POST", apiURL("/releases/"+args[0]+"/deploy"), nil)
		body, err := doRequest(req)
		if err != nil {
			return err
		}

		var result struct {
			DeploymentID string `json:"deployment_id"`
		}
		json.Unmarshal(body, &result)
		fmt.Printf("Deployment created: %s\n", result.DeploymentID)
		return nil
	},
}
