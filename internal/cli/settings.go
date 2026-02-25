package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/SideQuest-Group/ancla-client/internal/config"
)

func init() {
	rootCmd.AddCommand(settingsCmd)
	settingsCmd.AddCommand(settingsShowCmd)
	settingsCmd.AddCommand(settingsSetCmd)
	settingsCmd.AddCommand(settingsEditCmd)
	settingsCmd.AddCommand(settingsPathCmd)
}

var settingsCmd = &cobra.Command{
	Use:     "settings",
	Aliases: []string{"setting"},
	Short:   "Manage CLI settings (~/.ancla/config.yaml)",
	Example: "  ancla settings show\n  ancla settings set server https://ancla.dev",
	GroupID: "config",
}

// maskSecret masks a secret value, showing only the last 4 characters.
// For short values (4 chars or fewer), the entire value is masked.
func maskSecret(val string) string {
	if len(val) > 4 {
		return strings.Repeat("*", len(val)-4) + val[len(val)-4:]
	}
	return strings.Repeat("*", len(val))
}

var settingsShowCmd = &cobra.Command{
	Use:     "show",
	Short:   "Show current CLI settings",
	Example: "  ancla settings show",
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfg.APIKey != "" {
			fmt.Printf("api_key: %s\n", maskSecret(cfg.APIKey))
		} else {
			fmt.Printf("api_key: (not set)\n")
		}
		return nil
	},
}

var settingsSetCmd = &cobra.Command{
	Use:     "set <key> <value>",
	Short:   "Set a CLI setting (api_key)",
	Example: "  ancla settings set server https://ancla.dev\n  ancla settings set api_key mykey123",
	Args:    cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key, value := args[0], args[1]
		switch key {
		case "server":
			cfg.Server = value
		case "api_key":
			cfg.APIKey = value
		default:
			return fmt.Errorf("unknown setting %q (valid: server, api_key)", key)
		}
		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}
		displayValue := value
		if key == "api_key" {
			displayValue = maskSecret(value)
		}
		fmt.Printf("Set %s = %s\n", key, displayValue)
		return nil
	},
}

var settingsEditCmd = &cobra.Command{
	Use:     "edit",
	Short:   "Open config in $EDITOR",
	Example: "  ancla settings edit",
	RunE: func(cmd *cobra.Command, args []string) error {
		path := config.FilePath()
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vi"
		}
		c := exec.Command(editor, path)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	},
}

var settingsPathCmd = &cobra.Command{
	Use:     "path",
	Short:   "Show config file locations",
	Example: "  ancla settings path",
	RunE: func(cmd *cobra.Command, args []string) error {
		globalPath, localPath := config.Paths()
		fmt.Printf("global: %s\n", globalPath)
		if localPath != "" {
			fmt.Printf("local:  %s\n", localPath)
		} else {
			fmt.Printf("local:  (none found)\n")
		}
		return nil
	},
}
