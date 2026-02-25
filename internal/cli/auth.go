package cli

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/SideQuest-Group/ancla-client/internal/config"
)

func init() {
	loginCmd.Flags().Bool("manual", false, "Skip browser login and enter an API key manually")
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(whoamiCmd)
}

var loginCmd = &cobra.Command{
	Use:     "login",
	Short:   "Authenticate with the Ancla server",
	Long:    "Log in to the Ancla server via your browser and store the API key.",
	Example: "  ancla login\n  ancla login --manual",
	GroupID: "auth",
	RunE: func(cmd *cobra.Command, args []string) error {
		manual, _ := cmd.Flags().GetBool("manual")
		if manual {
			return loginManual()
		}
		return loginBrowser()
	},
}

// loginBrowser opens the browser, starts a local callback server, and waits
// for the server to redirect back with an API key.
func loginBrowser() error {
	// Generate a session code: 8 hex chars displayed as XXXX-XXXX
	codeBytes := make([]byte, 4)
	if _, err := rand.Read(codeBytes); err != nil {
		return fmt.Errorf("generating session code: %w", err)
	}
	raw := hex.EncodeToString(codeBytes)
	sessionCode := strings.ToUpper(raw[:4] + "-" + raw[4:])

	// Start a temporary HTTP server on a random port, bound to localhost only
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("starting callback server: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port

	type callbackResult struct {
		apiKey   string
		code     string
		username string
		email    string
	}
	resultCh := make(chan callbackResult, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		resultCh <- callbackResult{
			apiKey:   r.URL.Query().Get("api_key"),
			code:     r.URL.Query().Get("code"),
			username: r.URL.Query().Get("username"),
			email:    r.URL.Query().Get("email"),
		}
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<!DOCTYPE html><html><body style="font-family:system-ui;text-align:center;padding:4rem">
<h2>CLI Authorized</h2><p>You can close this tab and return to your terminal.</p>
</body></html>`)
	})

	srv := &http.Server{Handler: mux}
	go srv.Serve(listener)
	defer srv.Shutdown(context.Background())

	// Open the browser
	loginURL := fmt.Sprintf("%s/cli-auth?code=%s&port=%d", serverURL(), sessionCode, port)

	fmt.Println("Opening browser to log in...")
	fmt.Printf("Confirmation code: %s\n\n", sessionCode)

	if err := openBrowser(loginURL); err != nil {
		fmt.Printf("Could not open browser: %v\n", err)
		fmt.Printf("Open this URL manually:\n  %s\n\n", loginURL)
	}

	fmt.Println("Waiting for authentication... (press Ctrl+C to cancel)")

	// Wait for callback or timeout (5 minutes)
	timeout := time.After(5 * time.Minute)
	select {
	case result := <-resultCh:
		if result.code != sessionCode {
			return fmt.Errorf("session code mismatch — possible CSRF attack, aborting")
		}
		if result.apiKey == "" {
			return fmt.Errorf("no API key received from server")
		}
		// Key was just created by the server — save directly without re-validation
		cfg.APIKey = result.apiKey
		cfg.Username = result.username
		cfg.Email = result.email
		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}
		if result.username != "" {
			fmt.Printf("\n  Logged in as %s (%s)\n", result.username, result.email)
		} else {
			fmt.Printf("\n  Logged in successfully.\n")
		}
		fmt.Printf("  API key saved to %s\n", config.FilePath())
		return nil

	case <-timeout:
		fmt.Println("\nBrowser login timed out after 5 minutes.")
		fmt.Print("Falling back to manual API key entry...\n\n")
		return loginManual()
	}
}

// loginManual prompts the user for an API key directly.
func loginManual() error {
	fmt.Print("API Key: ")
	keyBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return fmt.Errorf("reading API key: %w", err)
	}
	apiKey := strings.TrimSpace(string(keyBytes))
	if apiKey == "" {
		return fmt.Errorf("API key cannot be empty")
	}
	return saveAndVerifyKey(apiKey)
}

// saveAndVerifyKey validates an API key against the server and saves it to the
// config file. Uses /organizations/ to verify the key since /auth/session only
// supports cookie-based auth.
func saveAndVerifyKey(apiKey string) error {
	client := &http.Client{
		Transport: &apiKeyTransport{key: apiKey, base: http.DefaultTransport},
	}
	req, err := http.NewRequest("GET", apiURL("/organizations/"), nil)
	if err != nil {
		return fmt.Errorf("invalid server URL: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("cannot reach server %s: %w", serverURL(), err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("API key was not accepted by %s", serverURL())
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned %d — check your API key", resp.StatusCode)
	}

	cfg.APIKey = apiKey
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Println("\n  Logged in successfully.")
	fmt.Printf("  API key saved to %s\n", config.FilePath())
	return nil
}

var whoamiCmd = &cobra.Command{
	Use:     "whoami",
	Short:   "Show the current authenticated user",
	Example: "  ancla whoami",
	GroupID: "auth",
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfg.APIKey == "" {
			fmt.Println("Not authenticated. Run 'ancla login' to authenticate.")
			return nil
		}

		// Verify the key still works by hitting an authenticated endpoint
		req, _ := http.NewRequest("GET", apiURL("/organizations/"), nil)
		_, err := doRequest(req)
		if err != nil {
			fmt.Println("Not authenticated (API key is invalid or expired). Run 'ancla login' to re-authenticate.")
			return nil
		}

		if isJSON() {
			return printJSON(map[string]string{
				"username": cfg.Username,
				"email":    cfg.Email,
			})
		}

		if cfg.Username != "" {
			fmt.Printf("Username: %s\n", cfg.Username)
		}
		if cfg.Email != "" {
			fmt.Printf("Email:    %s\n", cfg.Email)
		}
		if cfg.Username == "" && cfg.Email == "" {
			fmt.Println("Authenticated (re-login to populate user details)")
		}
		return nil
	},
}
