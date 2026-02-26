package ancla

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newTestClient creates a Client pointed at the given httptest.Server.
func newTestClient(t *testing.T, server *httptest.Server) *Client {
	t.Helper()
	return New("test-api-key", WithServer(server.URL))
}

func TestNewDefaults(t *testing.T) {
	c := New("my-key")
	if c.server != defaultServer {
		t.Errorf("expected default server %q, got %q", defaultServer, c.server)
	}
	if c.apiKey != "my-key" {
		t.Errorf("expected apiKey %q, got %q", "my-key", c.apiKey)
	}
	if c.httpClient == nil {
		t.Fatal("expected httpClient to be set")
	}
}

func TestWithServer(t *testing.T) {
	c := New("k", WithServer("https://custom.example.com/"))
	if c.server != "https://custom.example.com" {
		t.Errorf("expected trailing slash stripped, got %q", c.server)
	}
}

func TestWithHTTPClient(t *testing.T) {
	custom := &http.Client{}
	c := New("k", WithHTTPClient(custom))
	// The client should be the same pointer (transport is wrapped in place).
	if c.httpClient != custom {
		t.Error("expected custom http client to be used")
	}
}

func TestAPIKeyHeader(t *testing.T) {
	var gotKey string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get("X-API-Key")
		w.WriteHeader(200)
		w.Write([]byte("[]"))
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	_, _ = c.ListWorkspaces(context.Background())
	if gotKey != "test-api-key" {
		t.Errorf("expected X-API-Key %q, got %q", "test-api-key", gotKey)
	}
}

// --- Workspace CRUD tests ---

func TestListWorkspaces(t *testing.T) {
	workspaces := []Workspace{
		{ID: "1", Name: "Acme", Slug: "acme", MemberCount: 3, ProjectCount: 2},
		{ID: "2", Name: "Beta", Slug: "beta", MemberCount: 1, ProjectCount: 0},
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/workspaces/" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(workspaces)
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	result, err := c.ListWorkspaces(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 workspaces, got %d", len(result))
	}
	if result[0].Slug != "acme" {
		t.Errorf("expected slug %q, got %q", "acme", result[0].Slug)
	}
}

func TestGetWorkspace(t *testing.T) {
	ws := Workspace{
		Name:         "Acme",
		Slug:         "acme",
		ProjectCount: 5,
		Members: []WorkspaceMember{
			{Username: "alice", Email: "alice@example.com", Admin: true},
		},
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/workspaces/acme" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(ws)
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	result, err := c.GetWorkspace(context.Background(), "acme")
	if err != nil {
		t.Fatal(err)
	}
	if result.Name != "Acme" {
		t.Errorf("expected name %q, got %q", "Acme", result.Name)
	}
	if len(result.Members) != 1 {
		t.Fatalf("expected 1 member, got %d", len(result.Members))
	}
	if result.Members[0].Username != "alice" {
		t.Errorf("expected member %q, got %q", "alice", result.Members[0].Username)
	}
}

func TestCreateWorkspace(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/workspaces/" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		var body CreateWorkspaceRequest
		json.NewDecoder(r.Body).Decode(&body)
		if body.Name != "New Workspace" {
			t.Errorf("expected name %q, got %q", "New Workspace", body.Name)
		}
		json.NewEncoder(w).Encode(Workspace{ID: "3", Name: "New Workspace", Slug: "new-workspace"})
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	result, err := c.CreateWorkspace(context.Background(), "New Workspace")
	if err != nil {
		t.Fatal(err)
	}
	if result.Slug != "new-workspace" {
		t.Errorf("expected slug %q, got %q", "new-workspace", result.Slug)
	}
}

func TestUpdateWorkspace(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/workspaces/acme" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(Workspace{Name: "Acme Corp", Slug: "acme"})
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	result, err := c.UpdateWorkspace(context.Background(), "acme", "Acme Corp")
	if err != nil {
		t.Fatal(err)
	}
	if result.Name != "Acme Corp" {
		t.Errorf("expected name %q, got %q", "Acme Corp", result.Name)
	}
}

func TestDeleteWorkspace(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/workspaces/acme" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		w.WriteHeader(204)
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	err := c.DeleteWorkspace(context.Background(), "acme")
	if err != nil {
		t.Fatal(err)
	}
}

// --- Service and Config tests ---

func TestListServices(t *testing.T) {
	services := []Service{
		{Name: "Web", Slug: "web", Platform: "docker"},
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/workspaces/acme/projects/myproj/envs/production/services/" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(services)
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	result, err := c.ListServices(context.Background(), "acme", "myproj", "production")
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 || result[0].Slug != "web" {
		t.Errorf("unexpected result: %+v", result)
	}
}

func TestScaleService(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/workspaces/acme/projects/myproj/envs/production/services/svc123/scale" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		var body ScaleRequest
		json.NewDecoder(r.Body).Decode(&body)
		if body.ProcessCounts["web"] != 3 {
			t.Errorf("expected web=3, got %d", body.ProcessCounts["web"])
		}
		w.WriteHeader(200)
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	err := c.ScaleService(context.Background(), "acme", "myproj", "production", "svc123", map[string]int{"web": 3})
	if err != nil {
		t.Fatal(err)
	}
}

func TestSetConfig(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/workspaces/acme/projects/myproj/envs/production/services/web/config/" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		var body SetConfigRequest
		json.NewDecoder(r.Body).Decode(&body)
		if body.Name != "DB_URL" || body.Value != "postgres://localhost" {
			t.Errorf("unexpected body: %+v", body)
		}
		json.NewEncoder(w).Encode(ConfigVar{ID: "c1", Name: "DB_URL", Value: "postgres://localhost"})
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	result, err := c.SetConfig(context.Background(), "acme", "myproj", "production", "web", "DB_URL", "postgres://localhost", false)
	if err != nil {
		t.Fatal(err)
	}
	if result.Name != "DB_URL" {
		t.Errorf("expected name %q, got %q", "DB_URL", result.Name)
	}
}

func TestGetDeploy(t *testing.T) {
	dpl := Deploy{
		ID:       "dep-1",
		Complete: true,
		Created:  "2025-01-01T00:00:00Z",
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/deploys/dep-1/detail" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(dpl)
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	result, err := c.GetDeploy(context.Background(), "dep-1")
	if err != nil {
		t.Fatal(err)
	}
	if !result.Complete {
		t.Error("expected deploy to be complete")
	}
}

// --- Error handling tests ---

func TestError401(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	_, err := c.ListWorkspaces(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsUnauthorized(err) {
		t.Errorf("expected unauthorized error, got %v", err)
	}
}

func TestError404(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	_, err := c.GetWorkspace(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsNotFound(err) {
		t.Errorf("expected not found error, got %v", err)
	}
}

func TestError500(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	_, err := c.ListWorkspaces(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 500 {
		t.Errorf("expected status 500, got %d", apiErr.StatusCode)
	}
	if apiErr.Message != "server error" {
		t.Errorf("expected message %q, got %q", "server error", apiErr.Message)
	}
}

func TestErrorCustomMessage(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(422)
		json.NewEncoder(w).Encode(map[string]any{
			"status":  422,
			"message": "validation failed",
		})
	}))
	defer ts.Close()

	c := newTestClient(t, ts)
	_, err := c.ListWorkspaces(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 422 {
		t.Errorf("expected status 422, got %d", apiErr.StatusCode)
	}
	if apiErr.Message != "validation failed" {
		t.Errorf("expected message %q, got %q", "validation failed", apiErr.Message)
	}
}

func TestAPIErrorString(t *testing.T) {
	err := &APIError{StatusCode: 403, Message: "permission denied"}
	expected := "ancla api: 403 permission denied"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}

	err2 := &APIError{StatusCode: 418}
	expected2 := "ancla api: 418"
	if err2.Error() != expected2 {
		t.Errorf("expected %q, got %q", expected2, err2.Error())
	}
}
