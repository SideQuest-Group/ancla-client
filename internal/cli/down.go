package cli

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	downCmd.Flags().BoolVarP(&downYes, "yes", "y", false, "Skip confirmation prompt")
	rootCmd.AddCommand(downCmd)
}

var downYes bool

var downCmd = &cobra.Command{
	Use:   "down [org/project/app]",
	Short: "Scale all processes to 0 for an application",
	Long: `Tear down an application by scaling all of its processes to zero.

If no app path is provided, the linked app context from the local
.ancla/config.yaml is used (set via "ancla link"). You can also pass
the full org/project/app path as an argument.

Use --yes to skip the confirmation prompt (useful in scripts and CI).`,
	Example: `  ancla down
  ancla down my-org/my-project/my-app
  ancla down my-org/my-project/my-app --yes`,
	GroupID: "workflow",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Resolve app path from argument or link context.
		var appPath string
		if len(args) == 1 {
			appPath = args[0]
		} else if cfg.IsLinked() && cfg.Org != "" && cfg.Project != "" && cfg.App != "" {
			appPath = cfg.Org + "/" + cfg.Project + "/" + cfg.App
		} else {
			return fmt.Errorf("no app specified — pass org/project/app or link a directory first")
		}

		// Fetch the app to discover current process types.
		req, err := http.NewRequest("GET", apiURL("/applications/"+appPath), nil)
		if err != nil {
			return fmt.Errorf("building request: %w", err)
		}
		body, err := doRequest(req)
		if err != nil {
			return err
		}

		var app struct {
			ProcessCounts map[string]int `json:"process_counts"`
		}
		if err := json.Unmarshal(body, &app); err != nil {
			return fmt.Errorf("parsing application response: %w", err)
		}

		if len(app.ProcessCounts) == 0 {
			fmt.Println("No processes found for this application.")
			return nil
		}

		// Warn and confirm.
		fmt.Printf("This will scale all processes to 0 for %s\n", appPath)
		if !downYes {
			fmt.Print("Continue? [y/N] ")
			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			answer = strings.TrimSpace(strings.ToLower(answer))
			if answer != "y" && answer != "yes" {
				fmt.Println("Aborted.")
				return nil
			}
		}

		// Build zero-scaled process counts.
		zeroed := make(map[string]int, len(app.ProcessCounts))
		for proc := range app.ProcessCounts {
			zeroed[proc] = 0
		}

		payload, _ := json.Marshal(map[string]any{"process_counts": zeroed})

		stop := spin("Scaling down...")
		scaleReq, err := http.NewRequest("POST", apiURL("/applications/"+appPath+"/scale"), bytes.NewReader(payload))
		if err != nil {
			stop()
			return fmt.Errorf("building request: %w", err)
		}
		scaleReq.Header.Set("Content-Type", "application/json")

		scaleBody, err := doRequest(scaleReq)
		stop()
		if err != nil {
			return err
		}

		if isJSON() {
			var result any
			if json.Unmarshal(scaleBody, &result) == nil {
				return printJSON(result)
			}
		}

		fmt.Println("All processes scaled to 0.")
		return nil
	},
}
