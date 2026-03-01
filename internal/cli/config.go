package cli

import (
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
	rootCmd.AddCommand(configCmd)
	configCmd.PersistentFlags().String("scope", "service", "Config scope: workspace, project, env, or service")
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configDeleteCmd)
	configCmd.AddCommand(configImportCmd)
	configImportCmd.Flags().StringP("file", "f", "", "Path to .env file to import")
	configImportCmd.Flags().Bool("restart", false, "Trigger a config-only deploy after import")
	configListCmd.Flags().Bool("show-secrets", false, "Show secret values instead of masking them")
	configSetCmd.Flags().Bool("restart", false, "Trigger a config-only deploy after setting the variable")
	configDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	configCmd.AddCommand(configApplyCmd)
	configApplyCmd.Flags().StringP("file", "f", "", "Path to .env file to import")
	configApplyCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
}

var configCmd = &cobra.Command{
	Use:     "config",
	Aliases: []string{"cfg", "env"},
	Short:   "Manage configuration variables",
	Long: `Manage configuration variables at different scopes.

Configuration variables are key-value pairs injected into your service at
runtime. Use the --scope flag to target a specific level (workspace, project,
env, or service). Variables can be marked as secrets (values hidden by default)
or as build-time variables available during image builds. Use sub-commands to
list, set, delete, or bulk-import configuration from .env files.`,
	Example: `  ancla config list my-ws/my-proj/staging/my-svc
  ancla config set my-ws/my-proj/staging/my-svc KEY=value
  ancla config list --scope workspace my-ws`,
	GroupID: "config",
	RunE: func(cmd *cobra.Command, args []string) error {
		return configListCmd.RunE(cmd, args)
	},
}

// configAPIPath resolcts the API path for configuration based on the --scope
// flag and positional argument. Returns the full API config path.
func configAPIPath(cmd *cobra.Command, arg string) (string, error) {
	scope, _ := cmd.Flags().GetString("scope")
	if scope == "" {
		scope = "service"
	}

	ws, proj, env, svc, err := config.ResolveServicePath(arg, cfg)
	if err != nil {
		return "", err
	}

	switch scope {
	case "workspace":
		if ws == "" {
			return "", fmt.Errorf("workspace is required for --scope workspace")
		}
		return "/workspaces/" + ws + "/config/", nil
	case "project":
		if ws == "" || proj == "" {
			return "", fmt.Errorf("workspace and project are required for --scope project")
		}
		return "/workspaces/" + ws + "/projects/" + proj + "/config/", nil
	case "env":
		if ws == "" || proj == "" || env == "" {
			return "", fmt.Errorf("workspace, project, and env are required for --scope env")
		}
		return "/workspaces/" + ws + "/projects/" + proj + "/envs/" + env + "/config/", nil
	case "service":
		if ws == "" || proj == "" || env == "" || svc == "" {
			return "", fmt.Errorf("workspace, project, env, and service are required for --scope service")
		}
		return "/workspaces/" + ws + "/projects/" + proj + "/envs/" + env + "/services/" + svc + "/config/", nil
	default:
		return "", fmt.Errorf("invalid scope %q — use workspace, project, env, or service", scope)
	}
}

var configListCmd = &cobra.Command{
	Use:     "list [ws/proj/env/svc]",
	Short:   "List configuration variables",
	Example: "  ancla config list my-ws/my-proj/staging/my-svc\n  ancla config list --scope workspace my-ws",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var arg string
		if len(args) == 1 {
			arg = args[0]
		}
		cfgPath, err := configAPIPath(cmd, arg)
		if err != nil {
			return err
		}

		req, _ := http.NewRequest("GET", apiURL(cfgPath), nil)
		body, err := doRequest(req)
		if err != nil {
			return err
		}

		var configs []struct {
			ID        string `json:"id"`
			Name      string `json:"name"`
			Value     string `json:"value"`
			Secret    bool   `json:"secret"`
			Buildtime bool   `json:"buildtime"`
		}
		if err := json.Unmarshal(body, &configs); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		showSecrets, _ := cmd.Flags().GetBool("show-secrets")

		if !showSecrets {
			for i := range configs {
				if configs[i].Secret {
					configs[i].Value = "********"
				}
			}
		}

		if isJSON() {
			return printJSON(configs)
		}

		var rows [][]string
		for _, c := range configs {
			rows = append(rows, []string{c.Name, c.Value, fmt.Sprintf("%v", c.Secret), fmt.Sprintf("%v", c.Buildtime)})
		}
		table([]string{"NAME", "VALUE", "SECRET", "BUILDTIME"}, rows)
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:     "set [ws/proj/env/svc] KEY=value",
	Short:   "Set a configuration variable",
	Example: "  ancla config set my-ws/my-proj/staging/my-svc DATABASE_URL=postgres://localhost/mydb",
	Args:    cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		var arg, kvPair string
		if len(args) == 2 {
			arg = args[0]
			kvPair = args[1]
		} else {
			kvPair = args[0]
		}

		cfgPath, err := configAPIPath(cmd, arg)
		if err != nil {
			return err
		}

		parts := strings.SplitN(kvPair, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("expected KEY=value format")
		}

		payload, _ := json.Marshal(map[string]any{
			"name":  parts[0],
			"value": parts[1],
		})
		req, _ := http.NewRequest("POST", apiURL(cfgPath), bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		if _, err := doRequest(req); err != nil {
			return err
		}
		fmt.Printf("Set %s\n", parts[0])

		restart, _ := cmd.Flags().GetBool("restart")
		if restart {
			return triggerConfigOnlyDeploy(cmd, arg)
		}
		return nil
	},
}

var configDeleteCmd = &cobra.Command{
	Use:     "delete [ws/proj/env/svc] <config-id>",
	Short:   "Delete a configuration variable",
	Example: "  ancla config delete my-ws/my-proj/staging/my-svc <config-id>",
	Args:    cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		var arg, configID string
		if len(args) == 2 {
			arg = args[0]
			configID = args[1]
		} else {
			configID = args[0]
		}

		cfgPath, err := configAPIPath(cmd, arg)
		if err != nil {
			return err
		}

		if !confirmAction(cmd, "This will delete the configuration variable.") {
			fmt.Println("Aborted.")
			return nil
		}
		req, _ := http.NewRequest("DELETE", apiURL(cfgPath+configID), nil)
		if _, err := doRequest(req); err != nil {
			return err
		}
		fmt.Println("Deleted.")
		return nil
	},
}

var configImportCmd = &cobra.Command{
	Use:     "import [ws/proj/env/svc]",
	Short:   "Bulk import configuration from a .env file",
	Example: "  ancla config import my-ws/my-proj/staging/my-svc --file .env",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var arg string
		if len(args) == 1 {
			arg = args[0]
		}

		cfgPath, err := configAPIPath(cmd, arg)
		if err != nil {
			return err
		}

		filePath, _ := cmd.Flags().GetString("file")
		if filePath == "" {
			return fmt.Errorf("--file flag is required")
		}

		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("reading file: %w", err)
		}

		payload, _ := json.Marshal(map[string]any{
			"raw": string(data),
		})
		req, _ := http.NewRequest("POST", apiURL(cfgPath+"bulk"), bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		body, err := doRequest(req)
		if err != nil {
			return err
		}

		var result struct {
			Created []string `json:"created"`
			Skipped []string `json:"skipped"`
			Errors  []struct {
				Name  string `json:"name"`
				Error string `json:"error"`
			} `json:"errors"`
		}
		json.Unmarshal(body, &result)

		fmt.Printf("Created: %d variables\n", len(result.Created))
		if len(result.Skipped) > 0 {
			fmt.Printf("Skipped (secret): %s\n", strings.Join(result.Skipped, ", "))
		}
		if len(result.Errors) > 0 {
			fmt.Println("Errors:")
			for _, e := range result.Errors {
				fmt.Printf("  %s: %s\n", e.Name, e.Error)
			}
		}

		restart, _ := cmd.Flags().GetBool("restart")
		if restart {
			return triggerConfigOnlyDeploy(cmd, arg)
		}
		return nil
	},
}

var configApplyCmd = &cobra.Command{
	Use:   "apply [ws/proj/env/svc]",
	Short: "Bulk import .env + trigger config-only deploy",
	Long: `Import variables from a .env file and immediately trigger a config-only deploy.

This is a convenience command that combines 'config import' with an automatic
config-only redeploy so your changes take effect immediately.`,
	Example: "  ancla config apply my-ws/my-proj/staging/my-svc --file .env",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var arg string
		if len(args) == 1 {
			arg = args[0]
		}

		cfgPath, err := configAPIPath(cmd, arg)
		if err != nil {
			return err
		}

		filePath, _ := cmd.Flags().GetString("file")
		if filePath == "" {
			return fmt.Errorf("--file flag is required")
		}

		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("reading file: %w", err)
		}

		payload, _ := json.Marshal(map[string]any{
			"raw": string(data),
		})
		req, _ := http.NewRequest("POST", apiURL(cfgPath+"bulk"), bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		body, err := doRequest(req)
		if err != nil {
			return err
		}

		var result struct {
			Created []string `json:"created"`
			Skipped []string `json:"skipped"`
			Errors  []struct {
				Name  string `json:"name"`
				Error string `json:"error"`
			} `json:"errors"`
		}
		json.Unmarshal(body, &result)

		fmt.Printf("Created: %d variables\n", len(result.Created))
		if len(result.Skipped) > 0 {
			fmt.Printf("Skipped (secret): %s\n", strings.Join(result.Skipped, ", "))
		}
		if len(result.Errors) > 0 {
			fmt.Println("Errors:")
			for _, e := range result.Errors {
				fmt.Printf("  %s: %s\n", e.Name, e.Error)
			}
		}

		// Trigger config-only deploy
		return triggerConfigOnlyDeploy(cmd, arg)
	},
}

// triggerConfigOnlyDeploy triggers a config-only deploy for the service
// identified by the positional argument (or linked context).
func triggerConfigOnlyDeploy(cmd *cobra.Command, arg string) error {
	ws, proj, env, svc, err := config.ResolveServicePath(arg, cfg)
	if err != nil {
		return err
	}
	if ws == "" || proj == "" || env == "" || svc == "" {
		return fmt.Errorf("full service path required for config-only deploy")
	}

	stop := spin("Triggering config-only deploy...")
	payload, _ := json.Marshal(map[string]any{"config_only": true})
	req, _ := http.NewRequest("POST", apiURL(servicePath(ws, proj, env, svc)+"/deploy"), bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	body, err := doRequest(req)
	stop()
	if err != nil {
		return err
	}

	var result struct {
		DeployID string `json:"deploy_id"`
	}
	json.Unmarshal(body, &result)
	fmt.Printf("Config-only deploy triggered: %s\n", result.DeployID)
	return nil
}
