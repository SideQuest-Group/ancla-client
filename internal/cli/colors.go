package cli

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/mattn/go-runewidth"
)

// colorStatus returns the status string with a colored dot prefix.
// Respects NO_COLOR automatically via lipgloss color profile detection.
func colorStatus(status string) string {
	return statusDot(status) + " " + status
}

var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// visLen returns the visible display width of a string, stripping ANSI escape
// codes and accounting for wide Unicode characters (e.g. ● is 1 column).
func visLen(s string) int {
	return runewidth.StringWidth(ansiRe.ReplaceAllString(s, ""))
}

// table writes rows with ANSI-aware column alignment.
// Column widths are computed from visible string lengths so that ANSI escape
// codes (e.g. from colorStatus) don't break alignment.
func table(headers []string, rows [][]string) {
	cols := len(headers)
	widths := make([]int, cols)
	for i, h := range headers {
		if n := visLen(h); n > widths[i] {
			widths[i] = n
		}
	}
	for _, row := range rows {
		for i := 0; i < cols && i < len(row); i++ {
			if n := visLen(row[i]); n > widths[i] {
				widths[i] = n
			}
		}
	}

	const gap = 2
	padCell := func(cell string, width int) string {
		pad := width - visLen(cell) + gap
		if pad < gap {
			pad = gap
		}
		return cell + strings.Repeat(" ", pad)
	}

	var hdr strings.Builder
	for i, h := range headers {
		hdr.WriteString(padCell(h, widths[i]))
	}
	fmt.Println(stTableHeader.Render(hdr.String()))

	for _, row := range rows {
		var line strings.Builder
		for i := 0; i < cols; i++ {
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			line.WriteString(padCell(cell, widths[i]))
		}
		fmt.Println(line.String())
	}
}
