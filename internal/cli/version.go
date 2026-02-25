package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Version is set at build time via -ldflags.
	Version = "dev"
	// Commit is set at build time via -ldflags.
	Commit = "none"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show CLI version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("ancla %s (%s)\n", Version, Commit)
	},
}
