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
	rootCmd.AddCommand(dbshellCmd)
}

var dbshellCmd = &cobra.Command{
	Use:   "dbshell [ws/proj/env/svc]",
	Short: "Open an interactive database shell for the service",
	Long: `Open an interactive database shell for the linked service.

Connects to the service's primary database using credentials from the
platform. Automatically detects the database type (PostgreSQL, MySQL) and
launches the appropriate client (psql, mysql).`,
	Example: `  ancla dbshell
  ancla dbshell my-ws/my-proj/staging/my-svc`,
	GroupID: "workflow",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var arg string
		if len(args) == 1 {
			arg = args[0]
		}
		ws, proj, env, svc, err := config.ResolveServicePath(arg, cfg)
		if err != nil {
			return err
		}
		if ws == "" || proj == "" || env == "" || svc == "" {
			return fmt.Errorf("no service specified — provide an argument or run `ancla link` first")
		}

		// Fetch database connection info from the API
		svcPath := "/workspaces/" + ws + "/projects/" + proj + "/envs/" + env + "/services/" + svc
		req, _ := http.NewRequest("GET", apiURL(svcPath+"/database"), nil)
		stop := spin("Fetching database credentials...")
		body, err := doRequest(req)
		stop()
		if err != nil {
			return fmt.Errorf("no database found: %w", err)
		}

		var db struct {
			Engine   string `json:"engine"`
			Host     string `json:"host"`
			Port     int    `json:"port"`
			Name     string `json:"name"`
			User     string `json:"user"`
			Password string `json:"password"`
			URL      string `json:"url"`
		}
		if err := json.Unmarshal(body, &db); err != nil {
			return fmt.Errorf("parsing database info: %w", err)
		}

		if isJSON() {
			// Omit password in JSON output
			return printJSON(map[string]any{
				"engine": db.Engine,
				"host":   db.Host,
				"port":   db.Port,
				"name":   db.Name,
				"user":   db.User,
			})
		}

		var c *exec.Cmd
		switch db.Engine {
		case "postgresql", "postgres":
			if db.URL != "" {
				c = exec.Command("psql", db.URL)
			} else {
				c = exec.Command("psql",
					"-h", db.Host,
					"-p", fmt.Sprintf("%d", db.Port),
					"-U", db.User,
					"-d", db.Name,
				)
				c.Env = append(os.Environ(), "PGPASSWORD="+db.Password)
			}
		case "mysql":
			c = exec.Command("mysql",
				"-h", db.Host,
				"-P", fmt.Sprintf("%d", db.Port),
				"-u", db.User,
				fmt.Sprintf("-p%s", db.Password),
				db.Name,
			)
		default:
			if db.URL != "" {
				fmt.Printf("Database URL: %s\n", db.URL)
				return nil
			}
			return fmt.Errorf("unsupported database engine %q — connect manually using host=%s port=%d", db.Engine, db.Host, db.Port)
		}

		if !isQuiet() {
			fmt.Printf("Connecting to %s database %q on %s...\n", db.Engine, db.Name, db.Host)
		}

		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	},
}
