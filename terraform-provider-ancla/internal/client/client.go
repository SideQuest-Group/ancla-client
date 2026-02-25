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

// --- Organization API ---

// Org represents an Ancla organization.
type Org struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Slug             string `json:"slug"`
	MemberCount      int    `json:"member_count"`
	ProjectCount     int    `json:"project_count"`
	ApplicationCount int    `json:"application_count"`
}

// ListOrgs returns all organizations the authenticated user belongs to.
func (c *Client) ListOrgs() ([]Org, error) {
	req, err := http.NewRequest("GET", c.apiURL("/organizations/"), nil)
	if err != nil {
		return nil, err
	}
	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	var orgs []Org
	if err := json.Unmarshal(body, &orgs); err != nil {
		return nil, fmt.Errorf("parsing orgs response: %w", err)
	}
	return orgs, nil
}

// GetOrg returns an organization by slug.
func (c *Client) GetOrg(slug string) (*Org, error) {
	req, err := http.NewRequest("GET", c.apiURL("/organizations/"+slug), nil)
	if err != nil {
		return nil, err
	}
	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	var org Org
	if err := json.Unmarshal(body, &org); err != nil {
		return nil, fmt.Errorf("parsing org response: %w", err)
	}
	return &org, nil
}

// CreateOrg creates a new organization.
func (c *Client) CreateOrg(name string) (*Org, error) {
	payload, _ := json.Marshal(map[string]string{"name": name})
	req, err := http.NewRequest("POST", c.apiURL("/organizations/"), bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	var org Org
	if err := json.Unmarshal(body, &org); err != nil {
		return nil, fmt.Errorf("parsing org response: %w", err)
	}
	return &org, nil
}

// UpdateOrg updates an organization by slug.
func (c *Client) UpdateOrg(slug string, name string) (*Org, error) {
	payload, _ := json.Marshal(map[string]string{"name": name})
	req, err := http.NewRequest("PATCH", c.apiURL("/organizations/"+slug), bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	var org Org
	if err := json.Unmarshal(body, &org); err != nil {
		return nil, fmt.Errorf("parsing org response: %w", err)
	}
	return &org, nil
}

// DeleteOrg deletes an organization by slug.
func (c *Client) DeleteOrg(slug string) error {
	req, err := http.NewRequest("DELETE", c.apiURL("/organizations/"+slug), nil)
	if err != nil {
		return err
	}
	_, err = c.doRequest(req)
	return err
}

// --- Project API ---

// Project represents an Ancla project.
type Project struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Slug             string `json:"slug"`
	OrganizationSlug string `json:"organization_slug"`
	OrganizationName string `json:"organization_name"`
	ApplicationCount int    `json:"application_count"`
	Created          string `json:"created"`
	Updated          string `json:"updated"`
}

// GetProject returns a project by org slug and project slug.
func (c *Client) GetProject(orgSlug, projectSlug string) (*Project, error) {
	req, err := http.NewRequest("GET", c.apiURL("/projects/"+orgSlug+"/"+projectSlug), nil)
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

// CreateProject creates a new project under an organization.
func (c *Client) CreateProject(orgSlug, name string) (*Project, error) {
	payload, _ := json.Marshal(map[string]string{"name": name})
	req, err := http.NewRequest("POST", c.apiURL("/projects/"+orgSlug), bytes.NewReader(payload))
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

// UpdateProject updates a project by org slug and project slug.
func (c *Client) UpdateProject(orgSlug, projectSlug, name string) (*Project, error) {
	payload, _ := json.Marshal(map[string]string{"name": name})
	req, err := http.NewRequest("PATCH", c.apiURL("/projects/"+orgSlug+"/"+projectSlug), bytes.NewReader(payload))
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

// DeleteProject deletes a project by org slug and project slug.
func (c *Client) DeleteProject(orgSlug, projectSlug string) error {
	req, err := http.NewRequest("DELETE", c.apiURL("/projects/"+orgSlug+"/"+projectSlug), nil)
	if err != nil {
		return err
	}
	_, err = c.doRequest(req)
	return err
}

// --- Application API ---

// App represents an Ancla application.
type App struct {
	ID               string         `json:"id"`
	Name             string         `json:"name"`
	Slug             string         `json:"slug"`
	Platform         string         `json:"platform"`
	GithubRepository string         `json:"github_repository"`
	AutoDeployBranch string         `json:"auto_deploy_branch"`
	ProcessCounts    map[string]int `json:"process_counts"`
}

// GetApp returns an application by org/project/app slugs.
func (c *Client) GetApp(orgSlug, projectSlug, appSlug string) (*App, error) {
	req, err := http.NewRequest("GET", c.apiURL("/applications/"+orgSlug+"/"+projectSlug+"/"+appSlug), nil)
	if err != nil {
		return nil, err
	}
	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	var app App
	if err := json.Unmarshal(body, &app); err != nil {
		return nil, fmt.Errorf("parsing app response: %w", err)
	}
	return &app, nil
}

// CreateApp creates a new application under a project.
func (c *Client) CreateApp(orgSlug, projectSlug, name, platform string) (*App, error) {
	payload, _ := json.Marshal(map[string]string{
		"name":     name,
		"platform": platform,
	})
	req, err := http.NewRequest("POST", c.apiURL("/applications/"+orgSlug+"/"+projectSlug), bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	var app App
	if err := json.Unmarshal(body, &app); err != nil {
		return nil, fmt.Errorf("parsing app response: %w", err)
	}
	return &app, nil
}

// UpdateApp updates an application.
func (c *Client) UpdateApp(orgSlug, projectSlug, appSlug string, fields map[string]any) (*App, error) {
	payload, _ := json.Marshal(fields)
	req, err := http.NewRequest("PATCH", c.apiURL("/applications/"+orgSlug+"/"+projectSlug+"/"+appSlug), bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	var app App
	if err := json.Unmarshal(body, &app); err != nil {
		return nil, fmt.Errorf("parsing app response: %w", err)
	}
	return &app, nil
}

// DeleteApp deletes an application.
func (c *Client) DeleteApp(orgSlug, projectSlug, appSlug string) error {
	req, err := http.NewRequest("DELETE", c.apiURL("/applications/"+orgSlug+"/"+projectSlug+"/"+appSlug), nil)
	if err != nil {
		return err
	}
	_, err = c.doRequest(req)
	return err
}

// ScaleApp sets process counts for an application.
func (c *Client) ScaleApp(appID string, processCounts map[string]int) error {
	payload, _ := json.Marshal(map[string]any{"process_counts": processCounts})
	req, err := http.NewRequest("POST", c.apiURL("/applications/"+appID+"/scale"), bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	_, err = c.doRequest(req)
	return err
}

// --- Configuration API ---

// ConfigVar represents a configuration variable for an application.
type ConfigVar struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Value     string `json:"value"`
	Secret    bool   `json:"secret"`
	Buildtime bool   `json:"buildtime"`
}

// ListConfig returns all configuration variables for an application.
func (c *Client) ListConfig(appID string) ([]ConfigVar, error) {
	req, err := http.NewRequest("GET", c.apiURL("/configurations/"+appID), nil)
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
func (c *Client) SetConfig(appID, name, value string, secret, buildtime bool) (*ConfigVar, error) {
	payload, _ := json.Marshal(map[string]any{
		"name":      name,
		"value":     value,
		"secret":    secret,
		"buildtime": buildtime,
	})
	req, err := http.NewRequest("POST", c.apiURL("/configurations/"+appID), bytes.NewReader(payload))
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

// DeleteConfig deletes a configuration variable by app ID and config ID.
func (c *Client) DeleteConfig(appID, configID string) error {
	req, err := http.NewRequest("DELETE", c.apiURL("/configurations/"+appID+"/"+configID), nil)
	if err != nil {
		return err
	}
	_, err = c.doRequest(req)
	return err
}
