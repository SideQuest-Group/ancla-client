package cli

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(deployActionCmd)
	deployActionCmd.Flags().Bool("no-follow", false, "Fire and forget — don't stream build logs")
}

var deployActionCmd = &cobra.Command{
	Use:   "deploy [<ws>/<proj>/<env>/<svc>]",
	Short: "Deploy your service",
	Long: `Trigger a full deploy for a service.

This is the primary way to ship code. It triggers a build and deploy pipeline
for the specified service. By default it streams the build log until the
pipeline completes. Use --no-follow to fire and forget.

The service can be specified as a slash-separated path or resolved from
the linked context (see ancla link).`,
	Example: "  ancla deploy my-ws/my-proj/staging/my-svc\n  ancla deploy  # uses linked context",
	GroupID: "workflow",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, proj, env, svc, err := resolveServicePath(args)
		if err != nil {
			return err
		}
		if proj == "" || env == "" || svc == "" {
			return fmt.Errorf("all four segments required: <ws>/<proj>/<env>/<svc>")
		}

		path := fmt.Sprintf("%s/%s/%s/%s", ws, proj, env, svc)
		if !isQuiet() {
			fmt.Printf("Deploying %s...\n", path)
		}

		stop := spin("Triggering deploy...")
		req, _ := http.NewRequest("POST", apiURL(servicePath(ws, proj, env, svc)+"/deploy"), nil)
		body, err := doRequest(req)
		stop()
		if err != nil {
			return err
		}

		var result struct {
			BuildID string `json:"build_id"`
		}
		if err := json.Unmarshal(body, &result); err != nil {
			fmt.Println("Deploy triggered, but the response could not be parsed.")
			return nil
		}

		if isJSON() {
			return printJSON(result)
		}

		noFollow, _ := cmd.Flags().GetBool("no-follow")
		if noFollow || result.BuildID == "" {
			fmt.Printf("Deploy triggered. Build ID: %s\n", result.BuildID)
			return nil
		}

		fmt.Printf("Build ID: %s\n", result.BuildID)
		if err := followBuild(result.BuildID); err != nil {
			return err
		}

		fmt.Println("Deploy pipeline complete. Run `ancla deploys list` to inspect deploy status.")
		return nil
	},
}
