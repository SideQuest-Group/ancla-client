// Package cli implements the CLI commands for the ancla client.
package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/spf13/cobra"

	"github.com/SideQuest-Group/ancla-client/internal/config"
)

var (
	cfgFile      string
	outputFormat string
	jsonFlag     bool
	quietFlag    bool
	cfg          *config.Config
)

var rootCmd = &cobra.Command{
	Use:   "ancla",
	Short: "Ancla CLI — manage your Ancla PaaS deployments",
	Long: `Ancla CLI is a command-line client for the Ancla deployment platform.
It communicates with the Ancla API to manage organizations, projects,
applications, images, releases, deployments, and configuration.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		cfg, err = config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		// CLI flags override config file and env vars
		if s, _ := cmd.Flags().GetString("server"); s != "" {
			cfg.Server = s
		}
		if k, _ := cmd.Flags().GetString("api-key"); k != "" {
			cfg.APIKey = k
		}
		return nil
	},
}

// RootCmd returns the root cobra.Command for documentation generation.
func RootCmd() *cobra.Command {
	return rootCmd
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ~/.ancla/config.yaml)")
	rootCmd.PersistentFlags().String("server", "", "Ancla server URL (dev only)")
	rootCmd.PersistentFlags().String("api-key", "", "API key for authentication")
	_ = rootCmd.PersistentFlags().MarkHidden("server")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "Output format: table or json")
	rootCmd.PersistentFlags().BoolVar(&jsonFlag, "json", false, "Shorthand for --output json")
	rootCmd.PersistentFlags().BoolVarP(&quietFlag, "quiet", "q", false, "Suppress non-essential output")

	rootCmd.AddGroup(
		&cobra.Group{ID: "auth", Title: "Auth & Identity:"},
		&cobra.Group{ID: "workflow", Title: "Workflow:"},
		&cobra.Group{ID: "resources", Title: "Resources:"},
		&cobra.Group{ID: "config", Title: "Configuration:"},
	)
}

// isJSON returns true when the user requested JSON output.
func isJSON() bool {
	return jsonFlag || outputFormat == "json"
}

// isQuiet returns true when the user requested quiet/scripting mode.
// In quiet mode, only essential output (IDs, errors) is printed.
func isQuiet() bool {
	return quietFlag
}

// printJSON marshals v as indented JSON and writes it to stdout.
func printJSON(v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

// apiClient returns an *http.Client with the API key header set.
func apiClient() *http.Client {
	return &http.Client{
		Transport: &apiKeyTransport{
			key:  cfg.APIKey,
			base: http.DefaultTransport,
		},
	}
}

type apiKeyTransport struct {
	key  string
	base http.RoundTripper
}

func (t *apiKeyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.key != "" {
		req.Header.Set("X-API-Key", t.key)
	}
	return t.base.RoundTrip(req)
}

// serverURL returns the configured server base URL, ensuring it has a scheme.
func serverURL() string {
	s := cfg.Server
	if !strings.HasPrefix(s, "http://") && !strings.HasPrefix(s, "https://") {
		s = "http://" + s
	}
	return strings.TrimRight(s, "/")
}

// apiURL returns the full API v1 URL for the given path.
func apiURL(path string) string {
	return serverURL() + "/api/v1" + path
}

// doRequest performs an HTTP request and returns the response body.
// It checks for error status codes and formats API error messages.
func doRequest(req *http.Request) ([]byte, error) {
	resp, err := apiClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		switch resp.StatusCode {
		case 401:
			return nil, fmt.Errorf("not authenticated — run `ancla login` first")
		case 403:
			return nil, fmt.Errorf("permission denied")
		case 404:
			return nil, fmt.Errorf("not found")
		case 500:
			return nil, fmt.Errorf("server error — try again or check server logs")
		default:
			var apiErr struct {
				Status  int    `json:"status"`
				Message string `json:"message"`
				Detail  string `json:"detail"`
			}
			if json.Unmarshal(body, &apiErr) == nil {
				msg := apiErr.Message
				if msg == "" {
					msg = apiErr.Detail
				}
				if msg != "" {
					return nil, fmt.Errorf("%s", msg)
				}
			}
			return nil, fmt.Errorf("request failed (%d)", resp.StatusCode)
		}
	}

	return body, nil
}
