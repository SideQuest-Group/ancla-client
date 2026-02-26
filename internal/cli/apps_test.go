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

func TestServicesListCmd_ArgValidation(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"no args", []string{}, true},
		{"one arg", []string{"my-ws/my-proj/staging"}, false},
		{"two args", []string{"my-ws/my-proj/staging", "extra"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			argErr := cobra.ExactArgs(1)(servicesListCmd, tt.args)
			if tt.wantErr && argErr == nil {
				t.Error("expected arg validation error, got nil")
			}
			if !tt.wantErr && argErr != nil {
				t.Errorf("unexpected arg validation error: %v", argErr)
			}
		})
	}
}

func TestServicesListCmd_RunE(t *testing.T) {
	origCfg := cfg
	defer func() { cfg = origCfg }()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[]`))
	}))
	defer ts.Close()

	cfg = &config.Config{Server: ts.URL}

	cmd := servicesListCmd
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.RunE(cmd, []string{"my-ws/my-proj/staging"})
	if err != nil {
		t.Errorf("unexpected RunE error: %v", err)
	}
}

func TestServicesGetCmd_RequiresExactlyOneArg(t *testing.T) {
	argErr := cobra.ExactArgs(1)(servicesGetCmd, []string{})
	if argErr == nil {
		t.Error("expected error for zero args, got nil")
	}

	argErr = cobra.ExactArgs(1)(servicesGetCmd, []string{"my-ws/my-proj/staging/my-svc"})
	if argErr != nil {
		t.Errorf("unexpected error for one arg: %v", argErr)
	}

	argErr = cobra.ExactArgs(1)(servicesGetCmd, []string{"a", "b"})
	if argErr == nil {
		t.Error("expected error for two args, got nil")
	}
}

func TestServicesDeployCmd_RequiresExactlyOneArg(t *testing.T) {
	argErr := cobra.ExactArgs(1)(servicesDeployCmd, []string{})
	if argErr == nil {
		t.Error("expected error for zero args")
	}
	argErr = cobra.ExactArgs(1)(servicesDeployCmd, []string{"my-ws/my-proj/staging/my-svc"})
	if argErr != nil {
		t.Errorf("unexpected error: %v", argErr)
	}
}

func TestServicesScaleCmd_RequiresMinimumTwoArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"no args", []string{}, true},
		{"one arg", []string{"my-ws/my-proj/staging/my-svc"}, true},
		{"two args", []string{"my-ws/my-proj/staging/my-svc", "web=2"}, false},
		{"three args", []string{"my-ws/my-proj/staging/my-svc", "web=2", "worker=1"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			argErr := cobra.MinimumNArgs(2)(servicesScaleCmd, tt.args)
			if tt.wantErr && argErr == nil {
				t.Error("expected arg validation error, got nil")
			}
			if !tt.wantErr && argErr != nil {
				t.Errorf("unexpected error: %v", argErr)
			}
		})
	}
}

func TestServicesScaleCmd_InvalidScaleFormat(t *testing.T) {
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
		{"missing equals", []string{"my-ws/my-proj/staging/my-svc", "web2"}},
		{"non-numeric count", []string{"my-ws/my-proj/staging/my-svc", "web=abc"}},
	}

	for _, tt := range badFormats {
		t.Run(tt.name, func(t *testing.T) {
			cmd := servicesScaleCmd
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})
			err := cmd.RunE(cmd, tt.args)
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestServicesListCmd_ParsesServerResponse(t *testing.T) {
	origCfg := cfg
	origFormat := outputFormat
	origFlag := jsonFlag
	defer func() {
		cfg = origCfg
		outputFormat = origFormat
		jsonFlag = origFlag
	}()

	svcs := []map[string]string{
		{"name": "My Svc", "slug": "my-svc", "platform": "go"},
		{"name": "Other Svc", "slug": "other-svc", "platform": "python"},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(svcs)
	}))
	defer ts.Close()

	cfg = &config.Config{Server: ts.URL}
	outputFormat = "table"
	jsonFlag = false

	cmd := servicesListCmd
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.RunE(cmd, []string{"my-ws/my-proj/staging"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestServicesGetCmd_ParsesServerResponse(t *testing.T) {
	origCfg := cfg
	origFormat := outputFormat
	origFlag := jsonFlag
	defer func() {
		cfg = origCfg
		outputFormat = origFormat
		jsonFlag = origFlag
	}()

	svc := map[string]any{
		"name":               "My Svc",
		"slug":               "my-svc",
		"platform":           "go",
		"github_repository":  "org/repo",
		"auto_deploy_branch": "main",
		"process_counts":     map[string]int{"web": 2},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(svc)
	}))
	defer ts.Close()

	cfg = &config.Config{Server: ts.URL}
	outputFormat = "table"
	jsonFlag = false

	cmd := servicesGetCmd
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.RunE(cmd, []string{"my-ws/my-proj/staging/my-svc"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestServicesListCmd_HTTPError(t *testing.T) {
	origCfg := cfg
	defer func() { cfg = origCfg }()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer ts.Close()

	cfg = &config.Config{Server: ts.URL}

	cmd := servicesListCmd
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.RunE(cmd, []string{"my-ws/my-proj/staging"})
	if err == nil {
		t.Fatal("expected error for 401 response, got nil")
	}
	if got := err.Error(); got != "not authenticated — run `ancla login` first" {
		t.Errorf("error = %q, want auth error message", got)
	}
}

func TestServicesStatusCmd_RequiresExactlyOneArg(t *testing.T) {
	argErr := cobra.ExactArgs(1)(servicesStatusCmd, []string{})
	if argErr == nil {
		t.Error("expected error for zero args")
	}
	argErr = cobra.ExactArgs(1)(servicesStatusCmd, []string{"my-ws/my-proj/staging/my-svc"})
	if argErr != nil {
		t.Errorf("unexpected error: %v", argErr)
	}
}
