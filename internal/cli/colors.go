package cli

import (
	"bytes"
	"fmt"
	"strings"
	"text/tabwriter"
)

// colorStatus returns the status string with a colored dot prefix.
// Respects NO_COLOR automatically via lipgloss color profile detection.
func colorStatus(status string) string {
	return statusDot(status) + " " + status
}

// table writes rows through tabwriter with a styled header line.
// The first row is treated as the header and rendered with the brand style.
func table(headers []string, rows [][]string) {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, strings.Join(headers, "\t"))
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
			fmt.Println(stTableHeader.Render(line))
		} else {
			fmt.Println(line)
		}
	}
}
