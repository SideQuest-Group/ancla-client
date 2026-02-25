package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// confirmAction prompts the user with "Are you sure? [y/N]" and returns true
// only if they type "y" or "yes". It defaults to No on empty input or any
// other response. If the --yes flag is set on the command, it skips the prompt
// and returns true immediately.
func confirmAction(cmd *cobra.Command, message string) bool {
	yes, _ := cmd.Flags().GetBool("yes")
	if yes {
		return true
	}

	fmt.Fprintf(os.Stderr, "%s Are you sure? [y/N] ", message)
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))
	return answer == "y" || answer == "yes"
}
