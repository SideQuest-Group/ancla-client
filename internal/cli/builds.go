package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(buildsCmd)
	buildsCmd.AddCommand(buildsListCmd)
	buildsCmd.AddCommand(buildsTriggerCmd)
	buildsCmd.AddCommand(buildsLogCmd)
	buildsTriggerCmd.Flags().BoolP("follow", "f", false, "Follow build progress until complete")
	buildsLogCmd.Flags().BoolP("follow", "f", false, "Poll for log updates until build completes")
}

var buildsCmd = &cobra.Command{
	Use:     "builds",
	Aliases: []string{"build", "b"},
	Short:   "Manage builds",
	Long: `Manage builds for a service.

Builds are created from your service source code. Each build produces a new
versioned artifact that can be deployed. Use sub-commands to list builds,
trigger a new build, or view build logs.`,
	Example: "  ancla builds list my-ws/my-proj/staging/my-svc\n  ancla builds trigger my-ws/my-proj/staging/my-svc",
	GroupID: "resources",
}

var buildsListCmd = &cobra.Command{
	Use:     "list <ws>/<proj>/<env>/<svc>",
	Short:   "List builds for a service",
	Example: "  ancla builds list my-ws/my-proj/staging/my-svc",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, proj, env, svc, err := resolveServicePath(args)
		if err != nil {
			return err
		}
		if proj == "" || env == "" || svc == "" {
			return fmt.Errorf("usage: builds list <ws>/<proj>/<env>/<svc>")
		}

		req, _ := http.NewRequest("GET", apiURL(servicePath(ws, proj, env, svc)+"/builds/"), nil)
		body, err := doRequest(req)
		if err != nil {
			return err
		}

		var result struct {
			Items []struct {
				ID      string `json:"id"`
				Version int    `json:"version"`
				Built   bool   `json:"built"`
				Error   bool   `json:"error"`
				Created string `json:"created"`
			} `json:"items"`
		}
		if err := json.Unmarshal(body, &result); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		if isJSON() {
			return printJSON(result)
		}

		var rows [][]string
		for _, b := range result.Items {
			status := "building"
			if b.Error {
				status = "error"
			} else if b.Built {
				status = "built"
			}
			id := b.ID
			if len(id) > 8 {
				id = id[:8]
			}
			rows = append(rows, []string{fmt.Sprintf("v%d", b.Version), id, colorStatus(status), b.Created})
		}
		table([]string{"VERSION", "ID", "STATUS", "CREATED"}, rows)
		return nil
	},
}

var buildsTriggerCmd = &cobra.Command{
	Use:     "trigger <ws>/<proj>/<env>/<svc>",
	Short:   "Trigger a build for a service",
	Example: "  ancla builds trigger my-ws/my-proj/staging/my-svc",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, proj, env, svc, err := resolveServicePath(args)
		if err != nil {
			return err
		}
		if proj == "" || env == "" || svc == "" {
			return fmt.Errorf("usage: builds trigger <ws>/<proj>/<env>/<svc>")
		}

		stop := spin("Triggering build...")
		req, _ := http.NewRequest("POST", apiURL(servicePath(ws, proj, env, svc)+"/builds/trigger"), nil)
		body, err := doRequest(req)
		stop()
		if err != nil {
			return err
		}

		var result struct {
			BuildID string `json:"build_id"`
			Version int    `json:"version"`
		}
		if err := json.Unmarshal(body, &result); err != nil {
			fmt.Println("Build likely triggered, but the response could not be parsed (unexpected format).")
			return nil
		}
		fmt.Printf("Build triggered. Build: %s (v%d)\n", result.BuildID, result.Version)

		follow, _ := cmd.Flags().GetBool("follow")
		if follow && result.BuildID != "" {
			return followBuild(result.BuildID)
		}
		return nil
	},
}

var buildsLogCmd = &cobra.Command{
	Use:     "log <build-id>",
	Short:   "Show build log",
	Example: "  ancla builds log <build-id>",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		req, _ := http.NewRequest("GET", apiURL("/builds/"+args[0]+"/log"), nil)
		body, err := doRequest(req)
		if err != nil {
			return err
		}

		var result struct {
			Status  string `json:"status"`
			Version int    `json:"version"`
			LogText string `json:"log_text"`
		}
		if err := json.Unmarshal(body, &result); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		fmt.Printf("Build v%d — %s\n\n", result.Version, result.Status)
		if result.LogText != "" {
			fmt.Println(result.LogText)
		} else {
			fmt.Println("(no log output yet)")
		}

		follow, _ := cmd.Flags().GetBool("follow")
		if follow {
			return followBuild(args[0])
		}
		return nil
	},
}

// followBuild polls the build log endpoint until the build completes or errors.
func followBuild(buildID string) error {
	var lastLen int
	stop := spin("Building...")
	defer stop()

	for {
		time.Sleep(3 * time.Second)
		req, _ := http.NewRequest("GET", apiURL("/builds/"+buildID+"/log"), nil)
		body, err := doRequest(req)
		if err != nil {
			return err
		}
		var result struct {
			Status  string `json:"status"`
			LogText string `json:"log_text"`
		}
		if err := json.Unmarshal(body, &result); err != nil {
			return fmt.Errorf("parsing poll response: %w", err)
		}

		// Print new log lines
		if len(result.LogText) > lastLen {
			stop()
			fmt.Print(result.LogText[lastLen:])
			lastLen = len(result.LogText)
			stop = spin("Building...")
		}

		switch result.Status {
		case "built", "success", "complete":
			stop()
			fmt.Println("\nBuild complete.")
			return nil
		case "error", "failed":
			stop()
			return fmt.Errorf("build failed")
		}
	}
}
