package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/SideQuest-Group/ancla-client/internal/config"
)

func init() {
	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run [ws/proj/env/svc] -- <command> [args...]",
	Short: "Run a local command with the service's config vars injected",
	Long: `Execute a command locally with the linked service's configuration
variables injected as environment variables.

Requires a fully linked directory (workspace/project/env/service) or an
explicit service path argument. Fetches all non-secret configuration
variables from the API and passes them as environment variables to the
specified command.`,
	Example: "  ancla run -- python manage.py migrate\n  ancla run my-ws/my-proj/staging/my-svc -- env | grep DATABASE",
	GroupID: "workflow",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Determine if the first arg is a service path or the command.
		// If the first arg contains "/" it is treated as an explicit path.
		var cmdArgs []string
		var argPath string
		if len(args) > 1 && !isDashDash(args) {
			argPath = args[0]
			cmdArgs = args[1:]
		} else {
			cmdArgs = args
		}

		ws, proj, env, svc, err := config.ResolveServicePath(argPath, cfg)
		if err != nil {
			return err
		}
		if ws == "" || proj == "" || env == "" || svc == "" {
			return fmt.Errorf("not fully linked — run `ancla link <ws>/<proj>/<env>/<svc>` first")
		}

		// Fetch service config
		svcPath := "/workspaces/" + ws + "/projects/" + proj + "/envs/" + env + "/services/" + svc + "/config/"
		req, _ := http.NewRequest("GET", apiURL(svcPath), nil)
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

		// Build environment: inherit current env + overlay service config
		environ := os.Environ()
		for _, c := range configs {
			if !c.Secret {
				environ = append(environ, c.Name+"="+c.Value)
			}
		}

		// Execute the command
		c := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		c.Env = environ

		return c.Run()
	},
}

// isDashDash returns true if args starts with a non-path argument (no slash).
func isDashDash(args []string) bool {
	return len(args) > 0 && args[0] != "" && args[0][0] != '/'
}
