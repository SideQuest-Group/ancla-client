package cli

import (
	"strings"

	"github.com/fatih/color"
)

// colorStatus returns the status string colorized based on its meaning.
// Green for success states, red for error states, yellow for in-progress states.
// Respects NO_COLOR environment variable automatically via fatih/color.
func colorStatus(status string) string {
	switch strings.ToLower(status) {
	case "success", "running", "complete", "built":
		return color.GreenString(status)
	case "error", "failed":
		return color.RedString(status)
	case "building", "pending", "in_progress", "in progress":
		return color.YellowString(status)
	default:
		return status
	}
}

// colorHeader returns the string rendered in bold, suitable for table headers.
func colorHeader(s string) string {
	return color.New(color.Bold).Sprint(s)
}
