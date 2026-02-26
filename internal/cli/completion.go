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

// completeWorkspaces fetches workspace slugs from the API for shell completion.
func completeWorkspaces(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if cfg == nil || cfg.APIKey == "" {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	req, err := http.NewRequest("GET", apiURL("/workspaces/"), nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	body, err := doRequest(req)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	var workspaces []struct {
		Slug string `json:"slug"`
		Name string `json:"name"`
	}
	if json.Unmarshal(body, &workspaces) != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	var completions []string
	for _, w := range workspaces {
		completions = append(completions, w.Slug+"\t"+w.Name)
	}
	return completions, cobra.ShellCompDirectiveNoFileComp
}

// completeProjects fetches project slugs for the linked workspace for shell completion.
func completeProjects(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if cfg == nil || cfg.APIKey == "" || cfg.Workspace == "" {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	req, err := http.NewRequest("GET", apiURL("/workspaces/"+cfg.Workspace+"/projects/"), nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	body, err := doRequest(req)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	var projects []struct {
		Slug string `json:"slug"`
		Name string `json:"name"`
	}
	if json.Unmarshal(body, &projects) != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	var completions []string
	for _, p := range projects {
		completions = append(completions, p.Slug+"\t"+p.Name)
	}
	return completions, cobra.ShellCompDirectiveNoFileComp
}

// completeEnvs fetches environment slugs for the linked workspace/project for shell completion.
func completeEnvs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if cfg == nil || cfg.APIKey == "" || cfg.Workspace == "" || cfg.Project == "" {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	req, err := http.NewRequest("GET", apiURL("/workspaces/"+cfg.Workspace+"/projects/"+cfg.Project+"/envs/"), nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	body, err := doRequest(req)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	var envs []struct {
		Slug string `json:"slug"`
		Name string `json:"name"`
	}
	if json.Unmarshal(body, &envs) != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	var completions []string
	for _, e := range envs {
		completions = append(completions, e.Slug+"\t"+e.Name)
	}
	return completions, cobra.ShellCompDirectiveNoFileComp
}

// completeServices fetches service slugs for the linked workspace/project/env for shell completion.
func completeServices(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if cfg == nil || cfg.APIKey == "" || cfg.Workspace == "" || cfg.Project == "" || cfg.Env == "" {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	req, err := http.NewRequest("GET", apiURL("/workspaces/"+cfg.Workspace+"/projects/"+cfg.Project+"/envs/"+cfg.Env+"/services/"), nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	body, err := doRequest(req)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	var services []struct {
		Slug string `json:"slug"`
		Name string `json:"name"`
	}
	if json.Unmarshal(body, &services) != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	var completions []string
	for _, s := range services {
		completions = append(completions, s.Slug+"\t"+s.Name)
	}
	return completions, cobra.ShellCompDirectiveNoFileComp
}
