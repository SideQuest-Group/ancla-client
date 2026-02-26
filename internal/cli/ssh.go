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

	"github.com/SideQuest-Group/ancla-client/internal/config"
)

func init() {
	sshCmd.Flags().String("process", "web", "Process type to connect to")
	rootCmd.AddCommand(sshCmd)
}

var sshCmd = &cobra.Command{
	Use:   "ssh [ws/proj/env/svc]",
	Short: "Open an SSH session to a running container",
	Long: `Open an interactive SSH session to a running service container.

If no service path is provided, the linked context from .ancla/config.yaml
is used (set via 'ancla link'). Otherwise, specify the full service path as
ws/proj/env/svc.

The command requests ephemeral connection credentials from the Ancla API
and launches an SSH session to the container running the specified process type.`,
	Example: `  ancla ssh my-ws/my-proj/staging/my-svc
  ancla ssh --process worker
  ancla ssh my-ws/my-proj/staging/my-svc --process worker`,
	GroupID: "workflow",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Resolve service path from argument or link context.
		var arg string
		if len(args) == 1 {
			arg = args[0]
		}
		ws, proj, env, svc, err := config.ResolveServicePath(arg, cfg)
		if err != nil {
			return err
		}
		if ws == "" || proj == "" || env == "" || svc == "" {
			return fmt.Errorf("no service specified — provide a service path or link a project first with `ancla link`")
		}

		// Validate the path has four segments.
		displayPath := ws + "/" + proj + "/" + env + "/" + svc
		parts := strings.Split(displayPath, "/")
		if len(parts) != 4 || parts[0] == "" || parts[1] == "" || parts[2] == "" || parts[3] == "" {
			return fmt.Errorf("invalid service path %q — expected ws/proj/env/svc", displayPath)
		}

		processType, _ := cmd.Flags().GetString("process")

		// Request exec credentials from the API.
		svcPath := "/workspaces/" + ws + "/projects/" + proj + "/envs/" + env + "/services/" + svc
		payload, _ := json.Marshal(map[string]string{
			"process": processType,
		})
		req, err := http.NewRequest("POST", apiURL(svcPath+"/exec"), bytes.NewReader(payload))
		if err != nil {
			return fmt.Errorf("building request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		body, err := doRequest(req)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				return fmt.Errorf("exec is not available for %s — the service may not be running or exec is not supported", displayPath)
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

		fmt.Fprintf(os.Stderr, "Connecting to %s (%s process)...\n", displayPath, processType)
		if err := c.Run(); err != nil {
			return fmt.Errorf("ssh session failed: %w", err)
		}
		return nil
	},
}
