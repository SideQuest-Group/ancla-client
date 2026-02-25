package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(cacheCmd)
	cacheCmd.AddCommand(cacheInfoCmd)
	cacheCmd.AddCommand(cacheCliCmd)
	cacheCmd.AddCommand(cacheFlushCmd)
	cacheFlushCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
}

var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Manage the application's cache service (Redis, etc.)",
	Long: `Manage the cache service attached to your application.

Provides sub-commands to view cache info, open an interactive CLI session
(e.g. redis-cli), or flush the cache. Requires a linked app or explicit path.`,
	Example: `  ancla cache info
  ancla cache cli
  ancla cache flush`,
	GroupID: "workflow",
}

// resolveAppPath returns the app path from args or link context.
func resolveAppPath(args []string) (string, error) {
	if len(args) >= 1 {
		return args[0], nil
	}
	if cfg.Org != "" && cfg.Project != "" && cfg.App != "" {
		return cfg.Org + "/" + cfg.Project + "/" + cfg.App, nil
	}
	return "", fmt.Errorf("no app specified — provide an argument or run `ancla link` first")
}

type cacheInfo struct {
	Engine   string `json:"engine"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password,omitempty"`
	URL      string `json:"url"`
}

func fetchCacheInfo(appPath string) (*cacheInfo, error) {
	req, _ := http.NewRequest("GET", apiURL("/applications/"+appPath+"/cache"), nil)
	body, err := doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("no cache service found: %w", err)
	}
	var info cacheInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("parsing cache info: %w", err)
	}
	return &info, nil
}

var cacheInfoCmd = &cobra.Command{
	Use:     "info [app-path]",
	Short:   "Show cache service details",
	Example: "  ancla cache info",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		appPath, err := resolveAppPath(args)
		if err != nil {
			return err
		}

		stop := spin("Fetching cache info...")
		info, err := fetchCacheInfo(appPath)
		stop()
		if err != nil {
			return err
		}

		if isJSON() {
			return printJSON(map[string]any{
				"engine": info.Engine,
				"host":   info.Host,
				"port":   info.Port,
			})
		}

		fmt.Printf("Engine: %s\n", info.Engine)
		fmt.Printf("Host:   %s\n", info.Host)
		fmt.Printf("Port:   %d\n", info.Port)
		return nil
	},
}

var cacheCliCmd = &cobra.Command{
	Use:   "cli [app-path]",
	Short: "Open an interactive cache CLI session",
	Long: `Open an interactive CLI session to the application's cache service.

For Redis, launches redis-cli connected to the service. For other engines,
prints the connection URL.`,
	Example: "  ancla cache cli",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		appPath, err := resolveAppPath(args)
		if err != nil {
			return err
		}

		stop := spin("Connecting...")
		info, err := fetchCacheInfo(appPath)
		stop()
		if err != nil {
			return err
		}

		switch info.Engine {
		case "redis":
			cliArgs := []string{"-h", info.Host, "-p", fmt.Sprintf("%d", info.Port)}
			if info.Password != "" {
				cliArgs = append(cliArgs, "-a", info.Password)
			}
			c := exec.Command("redis-cli", cliArgs...)
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			if !isQuiet() {
				fmt.Printf("Connecting to Redis at %s:%d...\n", info.Host, info.Port)
			}
			return c.Run()
		default:
			if info.URL != "" {
				fmt.Printf("Cache URL: %s\n", info.URL)
				return nil
			}
			return fmt.Errorf("unsupported cache engine %q — connect manually at %s:%d", info.Engine, info.Host, info.Port)
		}
	},
}

var cacheFlushCmd = &cobra.Command{
	Use:     "flush [app-path]",
	Short:   "Flush the application cache",
	Example: "  ancla cache flush\n  ancla cache flush --yes",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		appPath, err := resolveAppPath(args)
		if err != nil {
			return err
		}

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			fmt.Printf("This will flush all cached data for %s.\n", appPath)
			fmt.Print("Continue? [y/N] ")
			var answer string
			fmt.Scanln(&answer)
			if answer != "y" && answer != "yes" {
				fmt.Println("Aborted.")
				return nil
			}
		}

		stop := spin("Flushing cache...")
		req, _ := http.NewRequest("POST", apiURL("/applications/"+appPath+"/cache/flush"), nil)
		_, err = doRequest(req)
		stop()
		if err != nil {
			return err
		}

		fmt.Println("Cache flushed.")
		return nil
	},
}
