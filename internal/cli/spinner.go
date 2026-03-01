package cli

import (
	"os"
	"time"

	"github.com/briandowns/spinner"
)

// newSpinner creates a spinner with the given message. The spinner is not
// started — call s.Start() to begin. The spinner is suppressed when stdout
// is not a terminal or when JSON output is requested.
func newSpinner(msg string) *spinner.Spinner {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond, spinner.WithWriter(os.Stderr), spinner.WithColor("fgHiCyan"))
	s.Suffix = " " + msg
	return s
}

// isTTY returns true when stderr is a terminal (spinners write to stderr).
func isTTY() bool {
	fi, err := os.Stderr.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// spin starts a spinner if stdout is a TTY and JSON output is not requested.
// Returns a stop function that should be deferred.
func spin(msg string) func() {
	if !isTTY() || isJSON() {
		return func() {}
	}
	s := newSpinner(msg)
	s.Start()
	return func() { s.Stop() }
}
