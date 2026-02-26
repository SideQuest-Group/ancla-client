package cli

import (
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
	shellCmd.Flags().StringP("process", "p", "web", "Process type to connect to")
	shellCmd.Flags().StringP("command", "c", "/bin/sh", "Command to execute in the container")
	rootCmd.AddCommand(shellCmd)
}

var shellCmd = &cobra.Command{
	Use:   "shell [ws/proj/env/svc]",
	Short: "Open an interactive shell in a running container",
	Long: `Open an interactive shell session in a running service container.

Uses the linked context or an explicit ws/proj/env/svc path. Unlike ssh,
this command uses the platform exec API directly and does not require SSH keys.`,
	Example: `  ancla shell
  ancla shell my-ws/my-proj/staging/my-svc
  ancla shell -p worker
  ancla shell -c /bin/bash`,
	GroupID: "workflow",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var arg string
		if len(args) == 1 {
			arg = args[0]
		}
		ws, proj, env, svc, err := config.ResolveServicePath(arg, cfg)
		if err != nil {
			return err
		}
		if ws == "" || proj == "" || env == "" || svc == "" {
			return fmt.Errorf("no service specified — provide an argument or run `ancla link` first")
		}

		process, _ := cmd.Flags().GetString("process")
		command, _ := cmd.Flags().GetString("command")

		// Request an exec session from the API
		svcPath := "/workspaces/" + ws + "/projects/" + proj + "/envs/" + env + "/services/" + svc
		payload := fmt.Sprintf(`{"process":"%s","command":"%s"}`, process, command)
		req, _ := http.NewRequest("POST", apiURL(svcPath+"/exec"), strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")

		stop := spin("Connecting...")
		body, err := doRequest(req)
		stop()
		if err != nil {
			return fmt.Errorf("exec not available: %w", err)
		}

		var session struct {
			WebSocketURL string `json:"websocket_url"`
			Host         string `json:"host"`
			Port         int    `json:"port"`
			Token        string `json:"token"`
		}
		if err := json.Unmarshal(body, &session); err != nil {
			return fmt.Errorf("parsing exec response: %w", err)
		}

		// Fall back to SSH if we get host/port/token
		if session.Host != "" && session.Token != "" {
			sshCmd := exec.Command("ssh",
				"-o", "StrictHostKeyChecking=no",
				"-p", fmt.Sprintf("%d", session.Port),
				fmt.Sprintf("token:%s@%s", session.Token, session.Host),
				command,
			)
			sshCmd.Stdin = os.Stdin
			sshCmd.Stdout = os.Stdout
			sshCmd.Stderr = os.Stderr
			return sshCmd.Run()
		}

		return fmt.Errorf("exec session did not return connection details")
	},
}
