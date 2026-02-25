package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(orgsCmd)
	orgsCmd.AddCommand(orgsListCmd)
	orgsCmd.AddCommand(orgsGetCmd)
}

var orgsCmd = &cobra.Command{
	Use:     "orgs",
	Aliases: []string{"org", "o"},
	Short:   "Manage organizations",
	Example: "  ancla orgs list\n  ancla orgs get my-org",
	GroupID: "resources",
}

var orgsListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List organizations",
	Example: "  ancla orgs list",
	RunE: func(cmd *cobra.Command, args []string) error {
		req, _ := http.NewRequest("GET", apiURL("/organizations/"), nil)
		body, err := doRequest(req)
		if err != nil {
			return err
		}

		var orgs []struct {
			ID           string `json:"id"`
			Name         string `json:"name"`
			Slug         string `json:"slug"`
			MemberCount  int    `json:"member_count"`
			ProjectCount int    `json:"project_count"`
		}
		if err := json.Unmarshal(body, &orgs); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		if isJSON() {
			return printJSON(orgs)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", colorHeader("SLUG"), colorHeader("NAME"), colorHeader("MEMBERS"), colorHeader("PROJECTS"))
		for _, o := range orgs {
			fmt.Fprintf(w, "%s\t%s\t%d\t%d\n", o.Slug, o.Name, o.MemberCount, o.ProjectCount)
		}
		return w.Flush()
	},
}

var orgsGetCmd = &cobra.Command{
	Use:     "get <slug>",
	Short:   "Get organization details",
	Example: "  ancla orgs get my-org",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		req, _ := http.NewRequest("GET", apiURL("/organizations/"+args[0]), nil)
		body, err := doRequest(req)
		if err != nil {
			return err
		}

		var org struct {
			Name             string `json:"name"`
			Slug             string `json:"slug"`
			ProjectCount     int    `json:"project_count"`
			ApplicationCount int    `json:"application_count"`
			Members          []struct {
				Username string `json:"username"`
				Email    string `json:"email"`
				Admin    bool   `json:"admin"`
			} `json:"members"`
		}
		if err := json.Unmarshal(body, &org); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		if isJSON() {
			return printJSON(org)
		}

		fmt.Printf("Organization: %s (%s)\n", org.Name, org.Slug)
		fmt.Printf("Projects: %d  Applications: %d\n\n", org.ProjectCount, org.ApplicationCount)
		fmt.Println("Members:")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintf(w, "  %s\t%s\t%s\n", colorHeader("USERNAME"), colorHeader("EMAIL"), colorHeader("ADMIN"))
		for _, m := range org.Members {
			fmt.Fprintf(w, "  %s\t%s\t%v\n", m.Username, m.Email, m.Admin)
		}
		return w.Flush()
	},
}
