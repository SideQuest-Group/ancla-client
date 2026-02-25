package cli

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(releasesCmd)
	releasesCmd.AddCommand(releasesListCmd)
	releasesCmd.AddCommand(releasesCreateCmd)
	releasesCmd.AddCommand(releasesDeployCmd)
}

var releasesCmd = &cobra.Command{
	Use:     "releases",
	Aliases: []string{"release", "rel"},
	Short:   "Manage releases",
	Long: `Manage releases for an application.

A release combines a built image with configuration to create a deployable
artifact. Releases are versioned and can be deployed to your infrastructure.
Use sub-commands to list releases, create new ones, or deploy them.`,
	Example: "  ancla releases list <app-id>\n  ancla releases create <app-id>",
	GroupID: "resources",
}

var releasesListCmd = &cobra.Command{
	Use:     "list <app-id>",
	Short:   "List releases for an application",
	Example: "  ancla releases list abc12345",
	Args:    cobra.ExactArgs(1),
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

		if isJSON() {
			return printJSON(result)
		}

		var rows [][]string
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
			rows = append(rows, []string{fmt.Sprintf("v%d", r.Version), id, r.Platform, colorStatus(status), r.Created})
		}
		table([]string{"VERSION", "ID", "PLATFORM", "STATUS", "CREATED"}, rows)
		return nil
	},
}

var releasesCreateCmd = &cobra.Command{
	Use:     "create <app-id>",
	Short:   "Create a new release",
	Example: "  ancla releases create abc12345",
	Args:    cobra.ExactArgs(1),
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
	Use:     "deploy <release-id>",
	Short:   "Deploy a release",
	Example: "  ancla releases deploy <release-id>",
	Args:    cobra.ExactArgs(1),
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
