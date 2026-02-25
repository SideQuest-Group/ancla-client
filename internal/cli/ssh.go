package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	sshCmd.Flags().String("process", "web", "Process type to connect to")
	rootCmd.AddCommand(sshCmd)
}

var sshCmd = &cobra.Command{
	Use:   "ssh [org/project/app]",
	Short: "Open an SSH session to a running container",
	Long: `Open an interactive SSH session to a running service container.

If no app path is provided, the linked app context from .ancla/config.yaml
is used (set via 'ancla link'). Otherwise, specify the full app path as
org/project/app.

The command requests ephemeral connection credentials from the Ancla API
and launches an SSH session to the container running the specified process type.`,
	Example: `  ancla ssh my-org/my-project/my-app
  ancla ssh --process worker
  ancla ssh my-org/my-project/my-app --process worker`,
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
			return fmt.Errorf("no app specified — provide an app path or link a project first with `ancla link`")
		}

		// Validate the path has three segments.
		parts := strings.Split(appPath, "/")
		if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
			return fmt.Errorf("invalid app path %q — expected org/project/app", appPath)
		}

		processType, _ := cmd.Flags().GetString("process")

		// Request exec credentials from the API.
		payload, _ := json.Marshal(map[string]string{
			"process": processType,
		})
		req, err := http.NewRequest("POST", apiURL("/applications/"+appPath+"/exec"), bytes.NewReader(payload))
		if err != nil {
			return fmt.Errorf("building request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		body, err := doRequest(req)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				return fmt.Errorf("exec is not available for %s — the app may not be running or exec is not supported", appPath)
			}
			return err
		}

		var connInfo struct {
			Host  string `json:"host"`
			Port  int    `json:"port"`
			Token string `json:"token"`
		}
		if err := json.Unmarshal(body, &connInfo); err != nil {
			return fmt.Errorf("parsing exec response: %w", err)
		}

		if connInfo.Host == "" || connInfo.Port == 0 || connInfo.Token == "" {
			return fmt.Errorf("incomplete connection details received from API")
		}

		// Build and execute the SSH command.
		sshArgs := []string{
			"-o", "StrictHostKeyChecking=no",
			"-p", fmt.Sprintf("%d", connInfo.Port),
			fmt.Sprintf("token:%s@%s", connInfo.Token, connInfo.Host),
		}

		sshBin, err := exec.LookPath("ssh")
		if err != nil {
			return fmt.Errorf("ssh not found in PATH — install OpenSSH to use this command")
		}

		c := exec.Command(sshBin, sshArgs...)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr

		fmt.Fprintf(os.Stderr, "Connecting to %s (%s process)...\n", appPath, processType)
		if err := c.Run(); err != nil {
			return fmt.Errorf("ssh session failed: %w", err)
		}
		return nil
	},
}
