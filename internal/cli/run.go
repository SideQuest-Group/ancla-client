package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run -- <command> [args...]",
	Short: "Run a local command with the app's config vars injected",
	Long: `Execute a command locally with the linked application's configuration
variables injected as environment variables.

Requires a fully linked directory (org/project/app). Fetches all non-secret
configuration variables from the API and passes them as environment variables
to the specified command.`,
	Example: "  ancla run -- python manage.py migrate\n  ancla run -- env | grep DATABASE",
	GroupID: "workflow",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfg.Org == "" || cfg.Project == "" || cfg.App == "" {
			return fmt.Errorf("not fully linked — run `ancla link <org>/<project>/<app>` first")
		}

		// Fetch app config
		appPath := cfg.Org + "/" + cfg.Project + "/" + cfg.App
		req, _ := http.NewRequest("GET", apiURL("/configurations/"+appPath), nil)
		body, err := doRequest(req)
		if err != nil {
			return fmt.Errorf("fetching config: %w", err)
		}

		var configs []struct {
			Name   string `json:"name"`
			Value  string `json:"value"`
			Secret bool   `json:"secret"`
		}
		if err := json.Unmarshal(body, &configs); err != nil {
			return fmt.Errorf("parsing config: %w", err)
		}

		// Build environment: inherit current env + overlay app config
		env := os.Environ()
		for _, c := range configs {
			if !c.Secret {
				env = append(env, c.Name+"="+c.Value)
			}
		}

		// Execute the command
		c := exec.Command(args[0], args[1:]...)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		c.Env = env

		return c.Run()
	},
}
