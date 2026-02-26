// Package client provides an HTTP API client for the Ancla PaaS platform.
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Client wraps net/http to communicate with the Ancla API.
type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

// New creates a new Ancla API client.
func New(baseURL, apiKey string) *Client {
	baseURL = strings.TrimRight(baseURL, "/")
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "https://" + baseURL
	}

	c := &Client{
		BaseURL: baseURL,
		APIKey:  apiKey,
	}
	c.HTTPClient = &http.Client{
		Transport: &apiKeyTransport{
			key:  apiKey,
			base: http.DefaultTransport,
		},
	}
	return c
}

type apiKeyTransport struct {
	key  string
	base http.RoundTripper
}

func (t *apiKeyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.key != "" {
		req.Header.Set("X-API-Key", t.key)
	}
	return t.base.RoundTrip(req)
}

// apiURL returns the full API v1 URL for the given path.
func (c *Client) apiURL(path string) string {
	return c.BaseURL + "/api/v1" + path
}

// doRequest performs an HTTP request and returns the response body bytes.
func (c *Client) doRequest(req *http.Request) ([]byte, error) {
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		switch resp.StatusCode {
		case 401:
			return nil, fmt.Errorf("not authenticated (401)")
		case 403:
			return nil, fmt.Errorf("permission denied (403)")
		case 404:
			return nil, &NotFoundError{Message: "not found (404)"}
		case 500:
			return nil, fmt.Errorf("server error (500)")
		default:
			var apiErr struct {
				Status  int    `json:"status"`
				Message string `json:"message"`
				Detail  string `json:"detail"`
			}
			if json.Unmarshal(body, &apiErr) == nil {
				msg := apiErr.Message
				if msg == "" {
					msg = apiErr.Detail
				}
				if msg != "" {
					return nil, fmt.Errorf("%s", msg)
				}
			}
			return nil, fmt.Errorf("request failed (%d)", resp.StatusCode)
		}
	}

	return body, nil
}

// NotFoundError indicates a 404 response from the API.
type NotFoundError struct {
	Message string
}

func (e *NotFoundError) Error() string {
	return e.Message
}

// IsNotFound returns true if the error is a 404 not-found error.
func IsNotFound(err error) bool {
	_, ok := err.(*NotFoundError)
	return ok
}

// --- Workspace API ---

// Workspace represents an Ancla workspace (formerly organization).
type Workspace struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Slug         string   `json:"slug"`
	MemberCount  int      `json:"member_count"`
	ProjectCount int      `json:"project_count"`
	ServiceCount int      `json:"service_count"`
	Members      []string `json:"members"`
}

// ListWorkspaces returns all workspaces the authenticated user belongs to.
func (c *Client) ListWorkspaces() ([]Workspace, error) {
	req, err := http.NewRequest("GET", c.apiURL("/workspaces/"), nil)
	if err != nil {
		return nil, err
	}
	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	var workspaces []Workspace
	if err := json.Unmarshal(body, &workspaces); err != nil {
		return nil, fmt.Errorf("parsing workspaces response: %w", err)
	}
	return workspaces, nil
}

// GetWorkspace returns a workspace by slug.
func (c *Client) GetWorkspace(slug string) (*Workspace, error) {
	req, err := http.NewRequest("GET", c.apiURL("/workspaces/"+slug), nil)
	if err != nil {
		return nil, err
	}
	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	var ws Workspace
	if err := json.Unmarshal(body, &ws); err != nil {
		return nil, fmt.Errorf("parsing workspace response: %w", err)
	}
	return &ws, nil
}

// CreateWorkspace creates a new workspace.
func (c *Client) CreateWorkspace(name string) (*Workspace, error) {
	payload, _ := json.Marshal(map[string]string{"name": name})
	req, err := http.NewRequest("POST", c.apiURL("/workspaces/"), bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	var ws Workspace
	if err := json.Unmarshal(body, &ws); err != nil {
		return nil, fmt.Errorf("parsing workspace response: %w", err)
	}
	return &ws, nil
}

// UpdateWorkspace updates a workspace by slug.
func (c *Client) UpdateWorkspace(slug string, name string) (*Workspace, error) {
	payload, _ := json.Marshal(map[string]string{"name": name})
	req, err := http.NewRequest("PATCH", c.apiURL("/workspaces/"+slug), bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	var ws Workspace
	if err := json.Unmarshal(body, &ws); err != nil {
		return nil, fmt.Errorf("parsing workspace response: %w", err)
	}
	return &ws, nil
}

// DeleteWorkspace deletes a workspace by slug.
func (c *Client) DeleteWorkspace(slug string) error {
	req, err := http.NewRequest("DELETE", c.apiURL("/workspaces/"+slug), nil)
	if err != nil {
		return err
	}
	_, err = c.doRequest(req)
	return err
}

// --- Project API ---

// Project represents an Ancla project.
type Project struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Slug          string `json:"slug"`
	WorkspaceSlug string `json:"workspace_slug"`
	WorkspaceName string `json:"workspace_name"`
	ServiceCount  int    `json:"service_count"`
	Created       string `json:"created"`
	Updated       string `json:"updated"`
}

// ListProjects returns all projects in a workspace.
func (c *Client) ListProjects(ws string) ([]Project, error) {
	req, err := http.NewRequest("GET", c.apiURL("/workspaces/"+ws+"/projects/"), nil)
	if err != nil {
		return nil, err
	}
	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	var projects []Project
	if err := json.Unmarshal(body, &projects); err != nil {
		return nil, fmt.Errorf("parsing projects response: %w", err)
	}
	return projects, nil
}

// GetProject returns a project by workspace slug and project slug.
func (c *Client) GetProject(ws, projectSlug string) (*Project, error) {
	req, err := http.NewRequest("GET", c.apiURL("/workspaces/"+ws+"/projects/"+projectSlug), nil)
	if err != nil {
		return nil, err
	}
	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	var project Project
	if err := json.Unmarshal(body, &project); err != nil {
		return nil, fmt.Errorf("parsing project response: %w", err)
	}
	return &project, nil
}

// CreateProject creates a new project under a workspace.
func (c *Client) CreateProject(ws, name string) (*Project, error) {
	payload, _ := json.Marshal(map[string]string{"name": name})
	req, err := http.NewRequest("POST", c.apiURL("/workspaces/"+ws+"/projects/"), bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	var project Project
	if err := json.Unmarshal(body, &project); err != nil {
		return nil, fmt.Errorf("parsing project response: %w", err)
	}
	return &project, nil
}

// UpdateProject updates a project by workspace slug and project slug.
func (c *Client) UpdateProject(ws, projectSlug, name string) (*Project, error) {
	payload, _ := json.Marshal(map[string]string{"name": name})
	req, err := http.NewRequest("PATCH", c.apiURL("/workspaces/"+ws+"/projects/"+projectSlug), bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	var project Project
	if err := json.Unmarshal(body, &project); err != nil {
		return nil, fmt.Errorf("parsing project response: %w", err)
	}
	return &project, nil
}

// DeleteProject deletes a project by workspace slug and project slug.
func (c *Client) DeleteProject(ws, projectSlug string) error {
	req, err := http.NewRequest("DELETE", c.apiURL("/workspaces/"+ws+"/projects/"+projectSlug), nil)
	if err != nil {
		return err
	}
	_, err = c.doRequest(req)
	return err
}

// --- Environment API ---

// Environment represents an Ancla environment within a project.
type Environment struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Slug         string `json:"slug"`
	ServiceCount int    `json:"service_count"`
	Created      string `json:"created"`
}

// ListEnvironments returns all environments in a project.
func (c *Client) ListEnvironments(ws, proj string) ([]Environment, error) {
	req, err := http.NewRequest("GET", c.apiURL("/workspaces/"+ws+"/projects/"+proj+"/envs/"), nil)
	if err != nil {
		return nil, err
	}
	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	var envs []Environment
	if err := json.Unmarshal(body, &envs); err != nil {
		return nil, fmt.Errorf("parsing environments response: %w", err)
	}
	return envs, nil
}

// GetEnvironment returns an environment by workspace, project, and environment slug.
func (c *Client) GetEnvironment(ws, proj, envSlug string) (*Environment, error) {
	req, err := http.NewRequest("GET", c.apiURL("/workspaces/"+ws+"/projects/"+proj+"/envs/"+envSlug), nil)
	if err != nil {
		return nil, err
	}
	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	var env Environment
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, fmt.Errorf("parsing environment response: %w", err)
	}
	return &env, nil
}

// CreateEnvironment creates a new environment under a project.
func (c *Client) CreateEnvironment(ws, proj, name string) (*Environment, error) {
	payload, _ := json.Marshal(map[string]string{"name": name})
	req, err := http.NewRequest("POST", c.apiURL("/workspaces/"+ws+"/projects/"+proj+"/envs/"), bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	var env Environment
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, fmt.Errorf("parsing environment response: %w", err)
	}
	return &env, nil
}

// UpdateEnvironment updates an environment by workspace, project, and environment slug.
func (c *Client) UpdateEnvironment(ws, proj, envSlug, name string) (*Environment, error) {
	payload, _ := json.Marshal(map[string]string{"name": name})
	req, err := http.NewRequest("PATCH", c.apiURL("/workspaces/"+ws+"/projects/"+proj+"/envs/"+envSlug), bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	var env Environment
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, fmt.Errorf("parsing environment response: %w", err)
	}
	return &env, nil
}

// DeleteEnvironment deletes an environment by workspace, project, and environment slug.
func (c *Client) DeleteEnvironment(ws, proj, envSlug string) error {
	req, err := http.NewRequest("DELETE", c.apiURL("/workspaces/"+ws+"/projects/"+proj+"/envs/"+envSlug), nil)
	if err != nil {
		return err
	}
	_, err = c.doRequest(req)
	return err
}

// --- Service API ---

// Service represents an Ancla service (formerly application).
type Service struct {
	ID               string         `json:"id"`
	Name             string         `json:"name"`
	Slug             string         `json:"slug"`
	WorkspaceSlug    string         `json:"workspace_slug"`
	ProjectSlug      string         `json:"project_slug"`
	EnvSlug          string         `json:"env_slug"`
	Platform         string         `json:"platform"`
	GithubRepository string         `json:"github_repository"`
	AutoDeployBranch string         `json:"auto_deploy_branch"`
	ProcessCounts    map[string]int `json:"process_counts"`
}

// ListServices returns all services in an environment.
func (c *Client) ListServices(ws, proj, env string) ([]Service, error) {
	req, err := http.NewRequest("GET", c.apiURL("/workspaces/"+ws+"/projects/"+proj+"/envs/"+env+"/services/"), nil)
	if err != nil {
		return nil, err
	}
	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	var services []Service
	if err := json.Unmarshal(body, &services); err != nil {
		return nil, fmt.Errorf("parsing services response: %w", err)
	}
	return services, nil
}

// GetService returns a service by workspace, project, environment, and service slug.
func (c *Client) GetService(ws, proj, env, svcSlug string) (*Service, error) {
	req, err := http.NewRequest("GET", c.apiURL("/workspaces/"+ws+"/projects/"+proj+"/envs/"+env+"/services/"+svcSlug), nil)
	if err != nil {
		return nil, err
	}
	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	var svc Service
	if err := json.Unmarshal(body, &svc); err != nil {
		return nil, fmt.Errorf("parsing service response: %w", err)
	}
	return &svc, nil
}

// CreateService creates a new service under an environment.
func (c *Client) CreateService(ws, proj, env, name, platform string) (*Service, error) {
	payload, _ := json.Marshal(map[string]string{
		"name":     name,
		"platform": platform,
	})
	req, err := http.NewRequest("POST", c.apiURL("/workspaces/"+ws+"/projects/"+proj+"/envs/"+env+"/services/"), bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	var svc Service
	if err := json.Unmarshal(body, &svc); err != nil {
		return nil, fmt.Errorf("parsing service response: %w", err)
	}
	return &svc, nil
}

// UpdateService updates a service.
func (c *Client) UpdateService(ws, proj, env, svcSlug string, fields map[string]any) (*Service, error) {
	payload, _ := json.Marshal(fields)
	req, err := http.NewRequest("PATCH", c.apiURL("/workspaces/"+ws+"/projects/"+proj+"/envs/"+env+"/services/"+svcSlug), bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	var svc Service
	if err := json.Unmarshal(body, &svc); err != nil {
		return nil, fmt.Errorf("parsing service response: %w", err)
	}
	return &svc, nil
}

// DeleteService deletes a service.
func (c *Client) DeleteService(ws, proj, env, svcSlug string) error {
	req, err := http.NewRequest("DELETE", c.apiURL("/workspaces/"+ws+"/projects/"+proj+"/envs/"+env+"/services/"+svcSlug), nil)
	if err != nil {
		return err
	}
	_, err = c.doRequest(req)
	return err
}

// ScaleService sets process counts for a service.
func (c *Client) ScaleService(ws, proj, env, svcSlug string, processCounts map[string]int) error {
	payload, _ := json.Marshal(map[string]any{"process_counts": processCounts})
	req, err := http.NewRequest("POST", c.apiURL("/workspaces/"+ws+"/projects/"+proj+"/envs/"+env+"/services/"+svcSlug+"/scale"), bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	_, err = c.doRequest(req)
	return err
}

// --- Configuration API ---

// ConfigVar represents a configuration variable with scope.
type ConfigVar struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Value     string `json:"value"`
	Secret    bool   `json:"secret"`
	Buildtime bool   `json:"buildtime"`
	Scope     string `json:"scope"`
}

// configBasePath returns the API path for config variables based on scope.
// For "service" scope: /workspaces/{ws}/projects/{proj}/envs/{env}/services/{svc}/config/
// For "environment" scope: /workspaces/{ws}/projects/{proj}/envs/{env}/config/
// For "project" scope: /workspaces/{ws}/projects/{proj}/config/
// For "workspace" scope: /workspaces/{ws}/config/
func (c *Client) configBasePath(ws, proj, env, svc, scope string) string {
	switch scope {
	case "workspace":
		return "/workspaces/" + ws + "/config/"
	case "project":
		return "/workspaces/" + ws + "/projects/" + proj + "/config/"
	case "environment":
		return "/workspaces/" + ws + "/projects/" + proj + "/envs/" + env + "/config/"
	default:
		// "service" scope is the default.
		return "/workspaces/" + ws + "/projects/" + proj + "/envs/" + env + "/services/" + svc + "/config/"
	}
}

// ListConfig returns all configuration variables at the given scope.
func (c *Client) ListConfig(ws, proj, env, svc, scope string) ([]ConfigVar, error) {
	req, err := http.NewRequest("GET", c.apiURL(c.configBasePath(ws, proj, env, svc, scope)), nil)
	if err != nil {
		return nil, err
	}
	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	var configs []ConfigVar
	if err := json.Unmarshal(body, &configs); err != nil {
		return nil, fmt.Errorf("parsing config response: %w", err)
	}
	return configs, nil
}

// SetConfig creates or updates a configuration variable.
func (c *Client) SetConfig(ws, proj, env, svc, scope, name, value string, secret, buildtime bool) (*ConfigVar, error) {
	payload, _ := json.Marshal(map[string]any{
		"name":      name,
		"value":     value,
		"secret":    secret,
		"buildtime": buildtime,
	})
	req, err := http.NewRequest("POST", c.apiURL(c.configBasePath(ws, proj, env, svc, scope)), bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	var cfg ConfigVar
	if err := json.Unmarshal(body, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config response: %w", err)
	}
	return &cfg, nil
}

// DeleteConfig deletes a configuration variable by ID.
func (c *Client) DeleteConfig(ws, proj, env, svc, scope, configID string) error {
	req, err := http.NewRequest("DELETE", c.apiURL(c.configBasePath(ws, proj, env, svc, scope)+configID), nil)
	if err != nil {
		return err
	}
	_, err = c.doRequest(req)
	return err
}
