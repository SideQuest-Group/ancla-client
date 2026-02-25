package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configDeleteCmd)
	configCmd.AddCommand(configImportCmd)
	configImportCmd.Flags().StringP("file", "f", "", "Path to .env file to import")
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage application configuration",
}

var configListCmd = &cobra.Command{
	Use:   "list <app-id>",
	Short: "List configuration variables",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		req, _ := http.NewRequest("GET", apiURL("/configurations/"+args[0]), nil)
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

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tVALUE\tSECRET\tBUILDTIME")
		for _, c := range configs {
			fmt.Fprintf(w, "%s\t%s\t%v\t%v\n", c.Name, c.Value, c.Secret, c.Buildtime)
		}
		return w.Flush()
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <app-id> KEY=value",
	Short: "Set a configuration variable",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		parts := strings.SplitN(args[1], "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("expected KEY=value format")
		}

		payload, _ := json.Marshal(map[string]any{
			"name":  parts[0],
			"value": parts[1],
		})
		req, _ := http.NewRequest("POST", apiURL("/configurations/"+args[0]), bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		if _, err := doRequest(req); err != nil {
			return err
		}
		fmt.Printf("Set %s\n", parts[0])
		return nil
	},
}

var configDeleteCmd = &cobra.Command{
	Use:   "delete <app-id> <config-id>",
	Short: "Delete a configuration variable",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		req, _ := http.NewRequest("DELETE", apiURL("/configurations/"+args[0]+"/"+args[1]), nil)
		if _, err := doRequest(req); err != nil {
			return err
		}
		fmt.Println("Deleted.")
		return nil
	},
}

var configImportCmd = &cobra.Command{
	Use:   "import <app-id>",
	Short: "Bulk import configuration from a .env file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
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
		req, _ := http.NewRequest("POST", apiURL("/configurations/"+args[0]+"/bulk"), bytes.NewReader(payload))
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
		return nil
	},
}
