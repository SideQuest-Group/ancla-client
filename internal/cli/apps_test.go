package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SideQuest-Group/ancla-client/internal/config"
	"github.com/spf13/cobra"
)

func TestAppsListCmd_ArgValidation(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"no args", []string{}, true},
		{"one arg", []string{"my-org/my-project"}, false},
		{"two args", []string{"my-org/my-project", "extra"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			argErr := cobra.ExactArgs(1)(appsListCmd, tt.args)
			if tt.wantErr && argErr == nil {
				t.Error("expected arg validation error, got nil")
			}
			if !tt.wantErr && argErr != nil {
				t.Errorf("unexpected arg validation error: %v", argErr)
			}
		})
	}
}

func TestAppsListCmd_RunE(t *testing.T) {
	origCfg := cfg
	defer func() { cfg = origCfg }()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[]`))
	}))
	defer ts.Close()

	cfg = &config.Config{Server: ts.URL}

	cmd := appsListCmd
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.RunE(cmd, []string{"my-org/my-project"})
	if err != nil {
		t.Errorf("unexpected RunE error: %v", err)
	}
}

func TestAppsGetCmd_RequiresExactlyOneArg(t *testing.T) {
	argErr := cobra.ExactArgs(1)(appsGetCmd, []string{})
	if argErr == nil {
		t.Error("expected error for zero args, got nil")
	}

	argErr = cobra.ExactArgs(1)(appsGetCmd, []string{"my-org/my-project/my-app"})
	if argErr != nil {
		t.Errorf("unexpected error for one arg: %v", argErr)
	}

	argErr = cobra.ExactArgs(1)(appsGetCmd, []string{"a", "b"})
	if argErr == nil {
		t.Error("expected error for two args, got nil")
	}
}

func TestAppsDeployCmd_RequiresExactlyOneArg(t *testing.T) {
	argErr := cobra.ExactArgs(1)(appsDeployCmd, []string{})
	if argErr == nil {
		t.Error("expected error for zero args")
	}
	argErr = cobra.ExactArgs(1)(appsDeployCmd, []string{"app-id"})
	if argErr != nil {
		t.Errorf("unexpected error: %v", argErr)
	}
}

func TestAppsScaleCmd_RequiresMinimumTwoArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"no args", []string{}, true},
		{"one arg", []string{"app-id"}, true},
		{"two args", []string{"app-id", "web=2"}, false},
		{"three args", []string{"app-id", "web=2", "worker=1"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			argErr := cobra.MinimumNArgs(2)(appsScaleCmd, tt.args)
			if tt.wantErr && argErr == nil {
				t.Error("expected arg validation error, got nil")
			}
			if !tt.wantErr && argErr != nil {
				t.Errorf("unexpected error: %v", argErr)
			}
		})
	}
}

func TestAppsScaleCmd_InvalidScaleFormat(t *testing.T) {
	origCfg := cfg
	defer func() { cfg = origCfg }()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	cfg = &config.Config{Server: ts.URL}

	badFormats := []struct {
		name string
		args []string
	}{
		{"missing equals", []string{"app-id", "web2"}},
		{"non-numeric count", []string{"app-id", "web=abc"}},
	}

	for _, tt := range badFormats {
		t.Run(tt.name, func(t *testing.T) {
			cmd := appsScaleCmd
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})
			err := cmd.RunE(cmd, tt.args)
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestAppsListCmd_ParsesServerResponse(t *testing.T) {
	origCfg := cfg
	origFormat := outputFormat
	origFlag := jsonFlag
	defer func() {
		cfg = origCfg
		outputFormat = origFormat
		jsonFlag = origFlag
	}()

	apps := []map[string]string{
		{"name": "My App", "slug": "my-app", "platform": "go"},
		{"name": "Other App", "slug": "other-app", "platform": "python"},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(apps)
	}))
	defer ts.Close()

	cfg = &config.Config{Server: ts.URL}
	outputFormat = "table"
	jsonFlag = false

	cmd := appsListCmd
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.RunE(cmd, []string{"my-org/my-project"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAppsGetCmd_ParsesServerResponse(t *testing.T) {
	origCfg := cfg
	origFormat := outputFormat
	origFlag := jsonFlag
	defer func() {
		cfg = origCfg
		outputFormat = origFormat
		jsonFlag = origFlag
	}()

	app := map[string]any{
		"name":               "My App",
		"slug":               "my-app",
		"platform":           "go",
		"github_repository":  "org/repo",
		"auto_deploy_branch": "main",
		"process_counts":     map[string]int{"web": 2},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(app)
	}))
	defer ts.Close()

	cfg = &config.Config{Server: ts.URL}
	outputFormat = "table"
	jsonFlag = false

	cmd := appsGetCmd
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.RunE(cmd, []string{"my-org/my-project/my-app"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAppsListCmd_HTTPError(t *testing.T) {
	origCfg := cfg
	defer func() { cfg = origCfg }()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer ts.Close()

	cfg = &config.Config{Server: ts.URL}

	cmd := appsListCmd
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.RunE(cmd, []string{"my-org/my-project"})
	if err == nil {
		t.Fatal("expected error for 401 response, got nil")
	}
	if got := err.Error(); got != "not authenticated — run `ancla login` first" {
		t.Errorf("error = %q, want auth error message", got)
	}
}

func TestAppsStatusCmd_RequiresExactlyOneArg(t *testing.T) {
	argErr := cobra.ExactArgs(1)(appsStatusCmd, []string{})
	if argErr == nil {
		t.Error("expected error for zero args")
	}
	argErr = cobra.ExactArgs(1)(appsStatusCmd, []string{"app-id"})
	if argErr != nil {
		t.Errorf("unexpected error: %v", argErr)
	}
}
