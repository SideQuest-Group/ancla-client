package cli

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// stdinReader is a shared buffered reader for interactive prompts.
var stdinReader = bufio.NewReader(os.Stdin)

// promptItem represents a selectable item in an interactive list.
type promptItem struct {
	Label string // displayed as "[N] Label (Slug)"
	Slug  string // machine identifier returned on selection
	Name  string // human-friendly name
}

// promptSelect shows a numbered list and returns the selected item's slug.
// If defaultSlug is non-empty and found in items, it is highlighted and used
// when the user presses Enter without input.
func promptSelect(label string, items []promptItem, defaultSlug string) (string, error) {
	if len(items) == 0 {
		return "", fmt.Errorf("no items to select")
	}

	fmt.Println()
	fmt.Println(label)
	defaultIdx := -1
	for i, it := range items {
		marker := " "
		display := it.Name
		if display == "" {
			display = it.Slug
		}
		if it.Slug == defaultSlug {
			defaultIdx = i
			marker = "*"
		}
		fmt.Printf("  %s[%d] %s (%s)\n", marker, i+1, display, it.Slug)
	}

	prompt := "Enter number or slug"
	if defaultIdx >= 0 {
		prompt += fmt.Sprintf(" [%d]", defaultIdx+1)
	}
	fmt.Print(prompt + ": ")

	input, _ := stdinReader.ReadString('\n')
	input = strings.TrimSpace(input)

	// Empty input → use default
	if input == "" && defaultIdx >= 0 {
		return items[defaultIdx].Slug, nil
	}

	// Try as number
	if n, err := strconv.Atoi(input); err == nil {
		if n < 1 || n > len(items) {
			return "", fmt.Errorf("invalid selection: %d (must be 1-%d)", n, len(items))
		}
		return items[n-1].Slug, nil
	}

	// Try as slug
	for _, it := range items {
		if it.Slug == input {
			return it.Slug, nil
		}
	}
	return "", fmt.Errorf("selection %q not found", input)
}

// promptSelectOrCreate shows a numbered list with an extra "[N] Create new..."
// option. Returns (slug, true) if user chose an existing item, or ("", false)
// if user chose to create new.
func promptSelectOrCreate(label string, items []promptItem, createLabel string) (string, bool, error) {
	fmt.Println()
	fmt.Println(label)
	for i, it := range items {
		display := it.Name
		if display == "" {
			display = it.Slug
		}
		fmt.Printf("  [%d] %s (%s)\n", i+1, display, it.Slug)
	}
	createIdx := len(items) + 1
	fmt.Printf("  [%d] %s\n", createIdx, createLabel)
	fmt.Print("Enter number or slug: ")

	input, _ := stdinReader.ReadString('\n')
	input = strings.TrimSpace(input)

	if n, err := strconv.Atoi(input); err == nil {
		if n == createIdx {
			return "", false, nil
		}
		if n < 1 || n > len(items) {
			return "", false, fmt.Errorf("invalid selection: %d (must be 1-%d)", n, createIdx)
		}
		return items[n-1].Slug, true, nil
	}

	for _, it := range items {
		if it.Slug == input {
			return it.Slug, true, nil
		}
	}
	return "", false, fmt.Errorf("selection %q not found", input)
}

// promptInput asks for a text value with an optional default.
func promptInput(label, defaultVal string) (string, error) {
	if defaultVal != "" {
		fmt.Printf("%s [%s]: ", label, defaultVal)
	} else {
		fmt.Printf("%s: ", label)
	}

	input, _ := stdinReader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultVal, nil
	}
	return input, nil
}

// promptConfirm asks a yes/no question, defaulting to yes (Y/n).
func promptConfirm(message string) bool {
	fmt.Printf("%s [Y/n]: ", message)
	input, _ := stdinReader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	return input == "" || input == "y" || input == "yes"
}
