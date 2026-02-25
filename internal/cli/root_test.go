package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SideQuest-Group/ancla-client/internal/config"
)

func TestApiURL(t *testing.T) {
	// Save and restore package-level cfg.
	origCfg := cfg
	defer func() { cfg = origCfg }()

	tests := []struct {
		name   string
		server string
		path   string
		want   string
	}{
		{
			name:   "https server with path",
			server: "https://ancla.dev",
			path:   "/applications/my-org/my-project",
			want:   "https://ancla.dev/api/v1/applications/my-org/my-project",
		},
		{
			name:   "http server",
			server: "http://localhost:8000",
			path:   "/orgs",
			want:   "http://localhost:8000/api/v1/orgs",
		},
		{
			name:   "server without scheme gets http prefix",
			server: "localhost:8000",
			path:   "/orgs",
			want:   "http://localhost:8000/api/v1/orgs",
		},
		{
			name:   "trailing slash stripped",
			server: "https://ancla.dev/",
			path:   "/applications",
			want:   "https://ancla.dev/api/v1/applications",
		},
		{
			name:   "empty path",
			server: "https://ancla.dev",
			path:   "",
			want:   "https://ancla.dev/api/v1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg = &config.Config{Server: tt.server}
			got := apiURL(tt.path)
			if got != tt.want {
				t.Errorf("apiURL(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestApiKeyTransport(t *testing.T) {
	// Verify the custom RoundTripper injects X-API-Key header.
	var gotHeader string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("X-API-Key")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	t.Run("key is injected", func(t *testing.T) {
		gotHeader = ""
		client := &http.Client{
			Transport: &apiKeyTransport{key: "test-key-123", base: http.DefaultTransport},
		}
		req, _ := http.NewRequest("GET", ts.URL, nil)
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		resp.Body.Close()

		if gotHeader != "test-key-123" {
			t.Errorf("X-API-Key = %q, want %q", gotHeader, "test-key-123")
		}
	})

	t.Run("empty key omits header", func(t *testing.T) {
		gotHeader = ""
		client := &http.Client{
			Transport: &apiKeyTransport{key: "", base: http.DefaultTransport},
		}
		req, _ := http.NewRequest("GET", ts.URL, nil)
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		resp.Body.Close()

		if gotHeader != "" {
			t.Errorf("X-API-Key = %q, want empty (header should not be set)", gotHeader)
		}
	})
}

func TestDoRequest_Success(t *testing.T) {
	origCfg := cfg
	defer func() { cfg = origCfg }()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer ts.Close()

	cfg = &config.Config{Server: ts.URL, APIKey: "key"}

	req, _ := http.NewRequest("GET", ts.URL+"/api/v1/test", nil)
	body, err := doRequest(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(body) != `{"status":"ok"}` {
		t.Errorf("body = %q, want %q", string(body), `{"status":"ok"}`)
	}
}

func TestDoRequest_HTTPErrors(t *testing.T) {
	origCfg := cfg
	defer func() { cfg = origCfg }()

	tests := []struct {
		name       string
		statusCode int
		body       string
		wantErr    string
	}{
		{
			name:       "401 unauthorized",
			statusCode: http.StatusUnauthorized,
			body:       "",
			wantErr:    "not authenticated — run `ancla login` first",
		},
		{
			name:       "403 forbidden",
			statusCode: http.StatusForbidden,
			body:       "",
			wantErr:    "permission denied",
		},
		{
			name:       "404 not found",
			statusCode: http.StatusNotFound,
			body:       "",
			wantErr:    "not found",
		},
		{
			name:       "500 server error",
			statusCode: http.StatusInternalServerError,
			body:       "",
			wantErr:    "server error — try again or check server logs",
		},
		{
			name:       "422 with JSON message",
			statusCode: http.StatusUnprocessableEntity,
			body:       `{"status":422,"message":"validation failed"}`,
			wantErr:    "validation failed",
		},
		{
			name:       "422 with JSON detail",
			statusCode: http.StatusUnprocessableEntity,
			body:       `{"status":422,"detail":"missing field"}`,
			wantErr:    "missing field",
		},
		{
			name:       "400 with no parseable body",
			statusCode: http.StatusBadRequest,
			body:       "not json",
			wantErr:    "request failed (400)",
		},
		{
			name:       "400 with empty JSON message",
			statusCode: http.StatusBadRequest,
			body:       `{"status":400}`,
			wantErr:    "request failed (400)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				if tt.body != "" {
					w.Write([]byte(tt.body))
				}
			}))
			defer ts.Close()

			cfg = &config.Config{Server: ts.URL}

			req, _ := http.NewRequest("GET", ts.URL+"/test", nil)
			_, err := doRequest(req)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if got := err.Error(); got != tt.wantErr {
				t.Errorf("error = %q, want %q", got, tt.wantErr)
			}
		})
	}
}

func TestDoRequest_NetworkError(t *testing.T) {
	origCfg := cfg
	defer func() { cfg = origCfg }()

	// Point to an address that will refuse connections.
	cfg = &config.Config{Server: "http://127.0.0.1:1"}

	req, _ := http.NewRequest("GET", "http://127.0.0.1:1/api/v1/test", nil)
	_, err := doRequest(req)
	if err == nil {
		t.Fatal("expected error for unreachable server, got nil")
	}
	// The error should be wrapped with "request failed:"
	want := "request failed:"
	if got := err.Error(); len(got) < len(want) || got[:len(want)] != want {
		t.Errorf("error = %q, want prefix %q", got, want)
	}
}

func TestDoRequest_200WithJSONBody(t *testing.T) {
	origCfg := cfg
	defer func() { cfg = origCfg }()

	payload := map[string]string{"name": "my-app", "slug": "my-app"}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(payload)
	}))
	defer ts.Close()

	cfg = &config.Config{Server: ts.URL}

	req, _ := http.NewRequest("GET", ts.URL+"/api/v1/applications/test", nil)
	body, err := doRequest(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]string
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if result["name"] != "my-app" {
		t.Errorf("name = %q, want %q", result["name"], "my-app")
	}
}

func TestIsJSON(t *testing.T) {
	// Save and restore globals.
	origFormat := outputFormat
	origFlag := jsonFlag
	defer func() {
		outputFormat = origFormat
		jsonFlag = origFlag
	}()

	tests := []struct {
		name         string
		outputFormat string
		jsonFlag     bool
		want         bool
	}{
		{"default table", "table", false, false},
		{"output json", "json", false, true},
		{"json flag", "table", true, true},
		{"both set", "json", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputFormat = tt.outputFormat
			jsonFlag = tt.jsonFlag
			if got := isJSON(); got != tt.want {
				t.Errorf("isJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}
