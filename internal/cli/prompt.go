package cli

import (
	"fmt"

	"github.com/charmbracelet/huh"
)

// promptItem represents a selectable item in an interactive list.
type promptItem struct {
	Label string // unused legacy field
	Slug  string // machine identifier returned on selection
	Name  string // human-friendly name
}

const (
	createNewSlug = "__create_new__"
	skipSlug      = "__skip__"
)

// promptSelect shows an interactive arrow-key selector and returns the chosen slug.
func promptSelect(label string, items []promptItem, defaultSlug string) (string, error) {
	if len(items) == 0 {
		return "", fmt.Errorf("no items to select")
	}

	opts := make([]huh.Option[string], 0, len(items))
	for _, it := range items {
		display := it.Name
		if display == "" {
			display = it.Slug
		}
		opt := huh.NewOption(display, it.Slug)
		if it.Slug == defaultSlug {
			opt = opt.Selected(true)
		}
		opts = append(opts, opt)
	}

	var selected string
	err := themed(
		huh.NewSelect[string]().
			Title(label).
			Options(opts...).
			Value(&selected),
	).Run()
	if err != nil {
		return "", err
	}
	return selected, nil
}

// promptSelectOrCreate shows an interactive selector with an extra "Create new…" option.
// Returns (slug, true) for an existing item, or ("", false) for create-new.
func promptSelectOrCreate(label string, items []promptItem, createLabel string) (string, bool, error) {
	opts := make([]huh.Option[string], 0, len(items)+1)
	for _, it := range items {
		display := it.Name
		if display == "" {
			display = it.Slug
		}
		opts = append(opts, huh.NewOption(display, it.Slug))
	}
	opts = append(opts, huh.NewOption(createLabel, createNewSlug))

	var selected string
	err := themed(
		huh.NewSelect[string]().
			Title(label).
			Options(opts...).
			Value(&selected),
	).Run()
	if err != nil {
		return "", false, err
	}
	if selected == createNewSlug {
		return "", false, nil
	}
	return selected, true, nil
}

// promptSelectCreateSkip shows a selector with existing items, a "Create new…" option,
// and a "Skip" option. Returns the action taken: "existing" (slug set), "create", or "skip".
func promptSelectCreateSkip(label string, items []promptItem, createLabel, skipLabel string) (slug, action string, err error) {
	opts := make([]huh.Option[string], 0, len(items)+2)
	for _, it := range items {
		display := it.Name
		if display == "" {
			display = it.Slug
		}
		opts = append(opts, huh.NewOption(display, it.Slug))
	}
	opts = append(opts, huh.NewOption(createLabel, createNewSlug))
	opts = append(opts, huh.NewOption(skipLabel, skipSlug))

	var selected string
	err = themed(
		huh.NewSelect[string]().
			Title(label).
			Options(opts...).
			Value(&selected),
	).Run()
	if err != nil {
		return "", "", err
	}
	switch selected {
	case createNewSlug:
		return "", "create", nil
	case skipSlug:
		return "", "skip", nil
	default:
		return selected, "existing", nil
	}
}

// promptInput asks for a text value with an optional default.
func promptInput(label, defaultVal string) (string, error) {
	var value string
	input := huh.NewInput().
		Title(label).
		Value(&value)
	if defaultVal != "" {
		value = defaultVal
		input = input.Placeholder(defaultVal)
	}
	if err := themed(input).Run(); err != nil {
		return "", err
	}
	if value == "" {
		return defaultVal, nil
	}
	return value, nil
}

// promptConfirm asks a yes/no question, defaulting to yes.
func promptConfirm(message string) bool {
	confirmed := true
	err := themed(
		huh.NewConfirm().
			Title(message).
			Affirmative("Yes").
			Negative("No").
			Value(&confirmed),
	).Run()
	if err != nil {
		return false
	}
	return confirmed
}
