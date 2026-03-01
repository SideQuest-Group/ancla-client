package cli

import (
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// ─── Brand Palette ──────────────────────────────────────────────
// Sourced from the Ancla design system (ancla.dev themes.css).
// Deep-ocean dark theme: cyan primary, bright accents, slate neutrals.
var (
	brandPrimary = lipgloss.Color("#0891b2") // Cyan 600
	brandAccent  = lipgloss.Color("#22d3ee") // Cyan 400 — highlights
	brandSuccess = lipgloss.Color("#22c55e") // Green 500
	brandError   = lipgloss.Color("#ef4444") // Red 500
	brandWarning = lipgloss.Color("#f59e0b") // Amber 500
	brandInfo    = lipgloss.Color("#0284c7") // Blue 600
	brandDim     = lipgloss.Color("#64748b") // Slate 500
	brandMuted   = lipgloss.Color("#475569") // Slate 600
)

// ─── Symbols ────────────────────────────────────────────────────
const (
	symAnchor  = "⚓"
	symCheck   = "✓"
	symCross   = "✗"
	symDot     = "●"
	symArrow   = "→"
	symCircle  = "○"
	symPointer = "▸"
)

// ─── Styles ─────────────────────────────────────────────────────
var (
	stBold        = lipgloss.NewStyle().Bold(true)
	stHeading     = lipgloss.NewStyle().Bold(true).Foreground(brandAccent)
	stAccent      = lipgloss.NewStyle().Foreground(brandAccent)
	stPrimary     = lipgloss.NewStyle().Foreground(brandPrimary)
	stDim         = lipgloss.NewStyle().Foreground(brandDim)
	stMuted       = lipgloss.NewStyle().Foreground(brandMuted)
	stSuccess     = lipgloss.NewStyle().Foreground(brandSuccess)
	stError       = lipgloss.NewStyle().Foreground(brandError)
	stWarning     = lipgloss.NewStyle().Foreground(brandWarning)
	stLabel       = lipgloss.NewStyle().Foreground(brandDim).Width(14)
	stValue       = lipgloss.NewStyle().Bold(true)
	stTableHeader = lipgloss.NewStyle().Bold(true).Foreground(brandDim)
	stCmdName     = lipgloss.NewStyle().Foreground(brandAccent).Width(14)
)

// ─── Output Helpers ─────────────────────────────────────────────

// statusDot returns a colored ● for the given status.
func statusDot(status string) string {
	switch strings.ToLower(status) {
	case "success", "running", "complete", "built":
		return stSuccess.Render(symDot)
	case "error", "failed":
		return stError.Render(symDot)
	case "building", "pending", "in_progress", "in progress":
		return stWarning.Render(symDot)
	default:
		return stDim.Render(symDot)
	}
}

// stepDone renders a completed step: ✓ message
func stepDone(msg string) string {
	return "  " + stSuccess.Render(symCheck) + " " + msg
}

// stepActive renders an in-progress step: → message
func stepActive(msg string) string {
	return stAccent.Render(symArrow) + " " + msg
}

// kv renders an aligned label: value pair.
func kv(label, value string) string {
	return "  " + stLabel.Render(label) + stValue.Render(value)
}

// ─── Huh Theme ──────────────────────────────────────────────────

// anclaTheme returns a charmbracelet/huh theme matching the Ancla brand.
func anclaTheme() *huh.Theme {
	t := huh.ThemeBase()

	// Focused field styles
	t.Focused.Title = lipgloss.NewStyle().Bold(true).Foreground(brandAccent)
	t.Focused.Description = lipgloss.NewStyle().Foreground(brandDim)
	t.Focused.SelectSelector = lipgloss.NewStyle().Foreground(brandAccent).SetString(symPointer + " ")
	t.Focused.FocusedButton = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#0b1120")).
		Background(brandAccent).
		Padding(0, 1)
	t.Focused.BlurredButton = lipgloss.NewStyle().
		Foreground(brandDim).
		Padding(0, 1)

	// Blurred field styles
	t.Blurred.Title = lipgloss.NewStyle().Foreground(brandDim)
	t.Blurred.SelectSelector = lipgloss.NewStyle().SetString("  ")

	return t
}

// themed wraps huh fields in a form with the Ancla theme applied.
func themed(fields ...huh.Field) *huh.Form {
	return huh.NewForm(huh.NewGroup(fields...)).WithTheme(anclaTheme())
}
