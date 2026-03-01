package cli

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ─── Error Card System ─────────────────────────────────────────
// Styled error output inspired by nautical signal flags.
// A colored left-edge stripe signals severity; information flows
// in strict hierarchy: what → why → where → what next.

// errKind classifies a pipeline failure for contextual messaging.
type errKind int

const (
	errBuild errKind = iota
	errDeploy
	errTimeout
	errAuth
)

// pipelineError holds structured context about a pipeline failure.
type pipelineError struct {
	Kind      errKind
	Detail    string // raw error detail from the server
	Workspace string
	Project   string
	Env       string
	Service   string
}

// dashboardURL returns a direct link into the Ancla web dashboard
// for the failed resource.
func (e *pipelineError) dashboardURL() string {
	base := serverURL()
	switch e.Kind {
	case errBuild:
		return fmt.Sprintf("%s/workspaces/%s/%s/services/%s", base, e.Workspace, e.Project, e.Service)
	case errDeploy:
		return fmt.Sprintf("%s/workspaces/%s/%s/envs/%s", base, e.Workspace, e.Project, e.Env)
	case errAuth:
		return base + "/login"
	default:
		return fmt.Sprintf("%s/workspaces/%s/%s", base, e.Workspace, e.Project)
	}
}

// title returns the error headline.
func (e *pipelineError) title() string {
	switch e.Kind {
	case errBuild:
		return "Build failed"
	case errDeploy:
		return "Deploy failed"
	case errTimeout:
		return "Pipeline timed out"
	case errAuth:
		return "Authentication failed"
	default:
		return "Pipeline error"
	}
}

// hint returns contextual next-step suggestions.
func (e *pipelineError) hints() []string {
	switch e.Kind {
	case errBuild:
		hints := []string{"Check the build log for details"}
		if strings.Contains(strings.ToLower(e.Detail), "procfile") {
			hints = append(hints, "Add a Procfile to your project root")
		}
		if strings.Contains(strings.ToLower(e.Detail), "dockerfile") {
			hints = append(hints, "Verify your Dockerfile builds locally")
		}
		if strings.Contains(strings.ToLower(e.Detail), "timeout") {
			hints = append(hints, "Build exceeded time limit — try optimizing layers or caching")
		}
		hints = append(hints, "Run "+stAccent.Render("ancla builds log")+" to view full output")
		return hints
	case errDeploy:
		return []string{
			"Check deploy logs for crash details",
			"Verify health check endpoint responds",
			"Run " + stAccent.Render("ancla deploys log") + " to view output",
		}
	case errTimeout:
		return []string{
			"The build or deploy did not complete in time",
			"Check server status and try again",
			"Run " + stAccent.Render("ancla deploys list") + " to check current state",
		}
	case errAuth:
		return []string{
			"Your session may have expired",
			"Run " + stAccent.Render("ancla login") + " to re-authenticate",
		}
	default:
		return []string{"Check the dashboard for details"}
	}
}

// renderErrorCard prints a styled error card to stderr.
// Layout:
//
//	▌ ✗ Build failed
//	▌
//	▌ Buildpack build failed!
//	▌
//	▌ Next steps
//	▌   → Check the build log for details
//	▌   → Run ancla builds log to view full output
//	▌
//	▌ Dashboard
//	▌   https://ancla.dev/ws/proj/services/svc/builds
func renderErrorCard(e *pipelineError) {
	// The left-edge stripe — 1 char wide, colored by severity.
	barColor := brandError
	if e.Kind == errTimeout {
		barColor = brandWarning
	}
	bar := lipgloss.NewStyle().Foreground(barColor).Render("▌")

	var lines []string

	// ── Header: ▌ ✗ Build failed ──
	header := stError.Bold(true).Render(symCross + " " + e.title())
	lines = append(lines, bar+" "+header)
	lines = append(lines, bar)

	// ── Detail (what happened) ──
	if e.Detail != "" {
		// Wrap long detail text, indent each line behind the bar.
		for _, dl := range strings.Split(strings.TrimSpace(e.Detail), "\n") {
			dl = strings.TrimSpace(dl)
			if dl != "" {
				lines = append(lines, bar+"  "+stDim.Render(dl))
			}
		}
		lines = append(lines, bar)
	}

	// ── Next steps ──
	hints := e.hints()
	if len(hints) > 0 {
		lines = append(lines, bar+"  "+stMuted.Bold(true).Render("Next steps"))
		for _, h := range hints {
			lines = append(lines, bar+"    "+stDim.Render(symArrow)+" "+h)
		}
		lines = append(lines, bar)
	}

	// ── Dashboard link ──
	url := e.dashboardURL()
	lines = append(lines, bar+"  "+stMuted.Bold(true).Render("Dashboard"))
	lines = append(lines, bar+"    "+stAccent.Underline(true).Render(url))

	// Print with a blank line above for breathing room.
	fmt.Println()
	for _, l := range lines {
		fmt.Println(l)
	}
	fmt.Println()
}
