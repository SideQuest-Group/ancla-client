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

	"github.com/SideQuest-Group/ancla-client/internal/config"
)

func init() {
	downCmd.Flags().BoolVarP(&downYes, "yes", "y", false, "Skip confirmation prompt")
	rootCmd.AddCommand(downCmd)
}

var downYes bool

var downCmd = &cobra.Command{
	Use:   "down [ws/proj/env/svc]",
	Short: "Scale all processes to 0 for a service",
	Long: `Tear down a service by scaling all of its processes to zero.

If no service path is provided, the linked context from the local
.ancla/config.yaml is used (set via "ancla link"). You can also pass
the full ws/proj/env/svc path as an argument.

Use --yes to skip the confirmation prompt (useful in scripts and CI).`,
	Example: `  ancla down
  ancla down my-ws/my-proj/staging/my-svc
  ancla down my-ws/my-proj/staging/my-svc --yes`,
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
			return fmt.Errorf("no service specified — pass ws/proj/env/svc or link a directory first")
		}

		svcPath := "/workspaces/" + ws + "/projects/" + proj + "/envs/" + env + "/services/" + svc
		displayPath := ws + "/" + proj + "/" + env + "/" + svc

		// Fetch the service to discover current process types.
		req, err := http.NewRequest("GET", apiURL(svcPath), nil)
		if err != nil {
			return fmt.Errorf("building request: %w", err)
		}
		body, err := doRequest(req)
		if err != nil {
			return err
		}

		var service struct {
			ProcessCounts map[string]int `json:"process_counts"`
		}
		if err := json.Unmarshal(body, &service); err != nil {
			return fmt.Errorf("parsing service response: %w", err)
		}

		if len(service.ProcessCounts) == 0 {
			fmt.Println("No processes found for this service.")
			return nil
		}

		// Warn and confirm.
		fmt.Printf("This will scale all processes to 0 for %s\n", displayPath)
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
		zeroed := make(map[string]int, len(service.ProcessCounts))
		for proc := range service.ProcessCounts {
			zeroed[proc] = 0
		}

		payload, _ := json.Marshal(map[string]any{"process_counts": zeroed})

		stop := spin("Scaling down...")
		scaleReq, err := http.NewRequest("POST", apiURL(svcPath+"/scale"), bytes.NewReader(payload))
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
