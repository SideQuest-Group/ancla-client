package cli

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all your projects grouped by workspace",
	Example: "  ancla list",
	GroupID: "resources",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Fetch all workspaces
		wsReq, _ := http.NewRequest("GET", apiURL("/workspaces/"), nil)
		wsBody, err := doRequest(wsReq)
		if err != nil {
			return err
		}

		var workspaces []struct {
			Name string `json:"name"`
			Slug string `json:"slug"`
		}
		if err := json.Unmarshal(wsBody, &workspaces); err != nil {
			return fmt.Errorf("parsing workspaces: %w", err)
		}

		// Fetch projects for each workspace
		type projectInfo struct {
			Name          string `json:"name"`
			Slug          string `json:"slug"`
			WorkspaceSlug string `json:"workspace_slug"`
		}

		allProjects := make(map[string][]projectInfo)
		for _, ws := range workspaces {
			projReq, _ := http.NewRequest("GET", apiURL("/workspaces/"+ws.Slug+"/projects/"), nil)
			projBody, err := doRequest(projReq)
			if err != nil {
				continue
			}
			var projects []projectInfo
			if json.Unmarshal(projBody, &projects) == nil {
				allProjects[ws.Slug] = projects
			}
		}

		if isJSON() {
			grouped := make(map[string][]string)
			for wsSlug, projects := range allProjects {
				for _, p := range projects {
					grouped[wsSlug] = append(grouped[wsSlug], p.Name)
				}
			}
			return printJSON(grouped)
		}

		// Display projects grouped by workspace
		for _, ws := range workspaces {
			bold.Println(ws.Name)
			projs := allProjects[ws.Slug]
			if len(projs) == 0 {
				fmt.Println(color.HiBlackString("  (no projects)"))
			}
			for _, p := range projs {
				fmt.Printf("  %s\n", p.Name)
			}
			fmt.Println()
		}

		return nil
	},
}
