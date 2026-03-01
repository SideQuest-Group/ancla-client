package cli

import (
	"bytes"
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
	buildsCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	buildsCmd.Flags().BoolP("follow", "f", false, "Follow build progress until complete")
	buildsCmd.Flags().String("strategy", "", "Build strategy: dockerfile or buildpack")
	buildsTriggerCmd.Flags().BoolP("follow", "f", false, "Follow build progress until complete")
	buildsTriggerCmd.Flags().String("strategy", "", "Build strategy: dockerfile or buildpack")
	buildsLogCmd.Flags().BoolP("follow", "f", false, "Poll for log updates until build completes")
}

var buildsCmd = &cobra.Command{
	Use:     "builds",
	Aliases: []string{"build", "b"},
	Short:   "Manage builds",
	Long: `Manage builds for a service.

Builds are created from your service source code. Each build produces a new
versioned artifact that can be deployed. Use sub-commands to list builds,
trigger a new build, or view build logs.

When a service is linked (via ancla link), running "ancla build" with no
subcommand will prompt to trigger a new build. Use --yes to skip the prompt.`,
	Example: "  ancla build\n  ancla build --yes --follow\n  ancla builds list my-ws/my-proj/staging/my-svc",
	GroupID: "resources",
	RunE: func(cmd *cobra.Command, args []string) error {
		// If a service is linked, prompt to trigger a build.
		ws, proj, env, svc, err := resolveServicePath(args)
		if err == nil && ws != "" && proj != "" && env != "" && svc != "" {
			path := ws + "/" + proj + "/" + env + "/" + svc
			if !confirmAction(cmd, fmt.Sprintf("Build %s?", stAccent.Render(path))) {
				return nil
			}
			return buildsTriggerCmd.RunE(cmd, args)
		}
		// Fall back to listing builds.
		return buildsListCmd.RunE(cmd, args)
	},
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
				ID       string  `json:"id"`
				Version  int     `json:"version"`
				Built    bool    `json:"built"`
				Error    bool    `json:"error"`
				Created  string  `json:"created"`
				Strategy *string `json:"strategy"`
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
			strategy := "dockerfile"
			if b.Strategy != nil && *b.Strategy != "" {
				strategy = *b.Strategy
			}
			rows = append(rows, []string{fmt.Sprintf("v%d", b.Version), id, colorStatus(status), strategy, b.Created})
		}
		table([]string{"VERSION", "ID", "STATUS", "STRATEGY", "CREATED"}, rows)
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
		var reqBody *bytes.Reader
		strategy, _ := cmd.Flags().GetString("strategy")
		if strategy != "" {
			payload, _ := json.Marshal(map[string]any{"strategy": strategy})
			reqBody = bytes.NewReader(payload)
		}
		var req *http.Request
		if reqBody != nil {
			req, _ = http.NewRequest("POST", apiURL(servicePath(ws, proj, env, svc)+"/builds/trigger"), reqBody)
			req.Header.Set("Content-Type", "application/json")
		} else {
			req, _ = http.NewRequest("POST", apiURL(servicePath(ws, proj, env, svc)+"/builds/trigger"), nil)
		}
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
		if follow && result.Version > 0 {
			return followBuildLog(servicePath(ws, proj, env, svc), fmt.Sprintf("%d", result.Version))
		}
		return nil
	},
}

var buildsLogCmd = &cobra.Command{
	Use:     "log [<ws>/<proj>/<env>/<svc>] <version>",
	Short:   "Show build log",
	Example: "  ancla builds log 3\n  ancla builds log my-ws/my-proj/staging/my-svc 2",
	Args:    cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		sp, version, err := resolveBuildArgs(args)
		if err != nil {
			return err
		}

		req, _ := http.NewRequest("GET", apiURL(sp+"/builds/"+version+"/log"), nil)
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
			return followBuildLog(sp, version)
		}
		return nil
	},
}

// resolveBuildArgs handles two calling conventions:
//
//	builds log <version>                              — uses linked service context
//	builds log <ws>/<proj>/<env>/<svc> <version>      — explicit path
//
// Returns the service path prefix and build version string.
func resolveBuildArgs(args []string) (sp, version string, err error) {
	if len(args) == 2 {
		ws, proj, env, svc, e := resolveServicePath(args[:1])
		if e != nil {
			return "", "", e
		}
		if proj == "" || env == "" || svc == "" {
			return "", "", fmt.Errorf("all four segments required: <ws>/<proj>/<env>/<svc>")
		}
		return servicePath(ws, proj, env, svc), args[1], nil
	}
	// Single arg — version, resolve service from linked config.
	ws, proj, env, svc, e := resolveServicePath(nil)
	if e != nil || ws == "" || proj == "" || env == "" || svc == "" {
		return "", "", fmt.Errorf("no linked service — provide <ws>/<proj>/<env>/<svc> before the version, or run `ancla link`")
	}
	return servicePath(ws, proj, env, svc), args[0], nil
}

// followBuildLog polls the build log endpoint until the build completes or errors.
func followBuildLog(sp, version string) error {
	var lastLen int
	stop := spin("Building...")
	defer stop()

	for {
		time.Sleep(3 * time.Second)
		req, _ := http.NewRequest("GET", apiURL(sp+"/builds/"+version+"/log"), nil)
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
		case "success":
			stop()
			fmt.Println("\n" + stSuccess.Render(symCheck+" Build complete."))
			return nil
		case "error":
			stop()
			return fmt.Errorf("%s", stError.Render(symCross+" Build failed"))
		}
	}
}
