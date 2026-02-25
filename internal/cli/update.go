package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
)

// checkForUpdate runs a non-blocking check against the GitHub releases API
// to see if a newer version of the CLI is available. It prints a notice to
// stderr if an update is found. Errors are silently ignored.
func checkForUpdate() {
	if Version == "dev" || isQuiet() {
		return
	}

	go func() {
		client := &http.Client{Timeout: 2 * time.Second}
		resp, err := client.Get("https://api.github.com/repos/SideQuest-Group/ancla-client/releases/latest")
		if err != nil {
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return
		}

		var release struct {
			TagName string `json:"tag_name"`
			HTMLURL string `json:"html_url"`
		}
		if json.NewDecoder(resp.Body).Decode(&release) != nil {
			return
		}

		latest := strings.TrimPrefix(release.TagName, "v")
		current := strings.TrimPrefix(Version, "v")
		if latest != "" && current != "" && latest != current {
			notice := fmt.Sprintf("Update available: %s → %s  (%s)", current, latest, release.HTMLURL)
			fmt.Fprintln(os.Stderr, color.YellowString(notice))
		}
	}()
}
