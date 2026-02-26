package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/SideQuest-Group/ancla-client/internal/config"
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
	Short: "Manage the service's cache (Redis, etc.)",
	Long: `Manage the cache service attached to your service.

Provides sub-commands to view cache info, open an interactive CLI session
(e.g. redis-cli), or flush the cache. Requires a linked service or explicit path.`,
	Example: `  ancla cache info
  ancla cache cli
  ancla cache flush`,
	GroupID: "workflow",
}

// resolveCachePath returns the API service path and display path from args or link context.
func resolveCachePath(args []string) (apiPath string, displayPath string, err error) {
	var arg string
	if len(args) >= 1 {
		arg = args[0]
	}
	ws, proj, env, svc, err := config.ResolveServicePath(arg, cfg)
	if err != nil {
		return "", "", err
	}
	if ws == "" || proj == "" || env == "" || svc == "" {
		return "", "", fmt.Errorf("no service specified — provide an argument or run `ancla link` first")
	}
	apiPath = "/workspaces/" + ws + "/projects/" + proj + "/envs/" + env + "/services/" + svc
	displayPath = ws + "/" + proj + "/" + env + "/" + svc
	return apiPath, displayPath, nil
}

type cacheInfo struct {
	Engine   string `json:"engine"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password,omitempty"`
	URL      string `json:"url"`
}

func fetchCacheInfo(svcAPIPath string) (*cacheInfo, error) {
	req, _ := http.NewRequest("GET", apiURL(svcAPIPath+"/cache"), nil)
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
	Use:     "info [ws/proj/env/svc]",
	Short:   "Show cache service details",
	Example: "  ancla cache info",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svcAPIPath, _, err := resolveCachePath(args)
		if err != nil {
			return err
		}

		stop := spin("Fetching cache info...")
		info, err := fetchCacheInfo(svcAPIPath)
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
	Use:   "cli [ws/proj/env/svc]",
	Short: "Open an interactive cache CLI session",
	Long: `Open an interactive CLI session to the service's cache.

For Redis, launches redis-cli connected to the service. For other engines,
prints the connection URL.`,
	Example: "  ancla cache cli",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svcAPIPath, _, err := resolveCachePath(args)
		if err != nil {
			return err
		}

		stop := spin("Connecting...")
		info, err := fetchCacheInfo(svcAPIPath)
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
	Use:     "flush [ws/proj/env/svc]",
	Short:   "Flush the service cache",
	Example: "  ancla cache flush\n  ancla cache flush --yes",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svcAPIPath, displayPath, err := resolveCachePath(args)
		if err != nil {
			return err
		}

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			fmt.Printf("This will flush all cached data for %s.\n", displayPath)
			fmt.Print("Continue? [y/N] ")
			var answer string
			fmt.Scanln(&answer)
			if answer != "y" && answer != "yes" {
				fmt.Println("Aborted.")
				return nil
			}
		}

		stop := spin("Flushing cache...")
		req, _ := http.NewRequest("POST", apiURL(svcAPIPath+"/cache/flush"), nil)
		_, err = doRequest(req)
		stop()
		if err != nil {
			return err
		}

		fmt.Println("Cache flushed.")
		return nil
	},
}
