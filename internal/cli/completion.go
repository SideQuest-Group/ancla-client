package cli

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(completionCmd)
}

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion script",
	Long: `Generate a shell completion script for ancla.

To load completions:

  bash:
    source <(ancla completion bash)
    # Or permanently: ancla completion bash > /etc/bash_completion.d/ancla

  zsh:
    source <(ancla completion zsh)
    # Or permanently: ancla completion zsh > "${fpath[1]}/_ancla"

  fish:
    ancla completion fish | source
    # Or permanently: ancla completion fish > ~/.config/fish/completions/ancla.fish

  powershell:
    ancla completion powershell | Out-String | Invoke-Expression
`,
	Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			return rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			return rootCmd.GenFishCompletion(os.Stdout, true)
		case "powershell":
			return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
		}
		return nil
	},
}

// completeOrgs fetches organization slugs from the API for shell completion.
func completeOrgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if cfg == nil || cfg.APIKey == "" {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	req, err := http.NewRequest("GET", apiURL("/organizations/"), nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	body, err := doRequest(req)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	var orgs []struct {
		Slug string `json:"slug"`
		Name string `json:"name"`
	}
	if json.Unmarshal(body, &orgs) != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	var completions []string
	for _, o := range orgs {
		completions = append(completions, o.Slug+"\t"+o.Name)
	}
	return completions, cobra.ShellCompDirectiveNoFileComp
}

// completeProjects fetches org/project slugs from the API for shell completion.
func completeProjects(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if cfg == nil || cfg.APIKey == "" {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	req, err := http.NewRequest("GET", apiURL("/projects/"), nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	body, err := doRequest(req)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	var projects []struct {
		Slug             string `json:"slug"`
		Name             string `json:"name"`
		OrganizationSlug string `json:"organization_slug"`
	}
	if json.Unmarshal(body, &projects) != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	var completions []string
	for _, p := range projects {
		completions = append(completions, p.OrganizationSlug+"/"+p.Slug+"\t"+p.Name)
	}
	return completions, cobra.ShellCompDirectiveNoFileComp
}
