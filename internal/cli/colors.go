package cli

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"
)

var bold = color.New(color.Bold)

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

// table writes rows through tabwriter with a bold header line.
// The first row is treated as the header and rendered in bold.
// Any cell value can use colorStatus() — colors are applied after alignment.
func table(headers []string, rows [][]string) {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	// Write header
	fmt.Fprintln(w, strings.Join(headers, "\t"))
	// Write data rows
	for _, row := range rows {
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}
	w.Flush()

	lines := strings.Split(buf.String(), "\n")
	for i, line := range lines {
		if line == "" {
			continue
		}
		if i == 0 {
			bold.Fprintln(os.Stdout, line)
		} else {
			fmt.Println(line)
		}
	}
}
