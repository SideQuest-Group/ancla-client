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
	Short: "Show status of the linked org/project/app",
	Long: `Show a unified status view for the currently linked resource.

Requires a linked directory (see ancla link). Displays the org, project,
application details, and current pipeline status in a single view.`,
	Example: "  ancla status",
	GroupID: "workflow",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !cfg.IsLinked() {
			return fmt.Errorf("not linked — run `ancla link <org>/<project>/<app>` first")
		}

		type statusOutput struct {
			Org     string `json:"org"`
			Project string `json:"project,omitempty"`
			App     string `json:"app,omitempty"`
			Build   string `json:"build,omitempty"`
			Release string `json:"release,omitempty"`
			Deploy  string `json:"deploy,omitempty"`
		}
		out := statusOutput{
			Org:     cfg.Org,
			Project: cfg.Project,
			App:     cfg.App,
		}

		// If we have a full app path, fetch pipeline status
		if cfg.Org != "" && cfg.Project != "" && cfg.App != "" {
			appPath := cfg.Org + "/" + cfg.Project + "/" + cfg.App
			req, _ := http.NewRequest("GET", apiURL("/applications/"+appPath+"/pipeline-status"), nil)
			body, err := doRequest(req)
			if err == nil {
				var status struct {
					Build   *struct{ Status string } `json:"build"`
					Release *struct{ Status string } `json:"release"`
					Deploy  *struct{ Status string } `json:"deploy"`
				}
				json.Unmarshal(body, &status)
				if status.Build != nil {
					out.Build = status.Build.Status
				}
				if status.Release != nil {
					out.Release = status.Release.Status
				}
				if status.Deploy != nil {
					out.Deploy = status.Deploy.Status
				}
			}
		}

		if isJSON() {
			return printJSON(out)
		}

		fmt.Printf("Org:     %s\n", out.Org)
		if out.Project != "" {
			fmt.Printf("Project: %s\n", out.Project)
		}
		if out.App != "" {
			fmt.Printf("App:     %s\n", out.App)
		}

		if out.Build != "" || out.Release != "" || out.Deploy != "" {
			fmt.Println()
			var rows [][]string
			if out.Build != "" {
				rows = append(rows, []string{"Build", colorStatus(out.Build)})
			}
			if out.Release != "" {
				rows = append(rows, []string{"Release", colorStatus(out.Release)})
			}
			if out.Deploy != "" {
				rows = append(rows, []string{"Deploy", colorStatus(out.Deploy)})
			}
			table([]string{"STAGE", "STATUS"}, rows)
		}

		return nil
	},
}
