// Package ancla provides a Go client for the Ancla PaaS platform API.
//
// Create a client with your API key and call methods to manage organizations,
// projects, applications, configuration, images, releases, and deployments.
//
//	client := ancla.New("your-api-key")
//	orgs, err := client.ListOrgs(ctx)
package ancla

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const defaultServer = "https://ancla.dev"

// Client is the Ancla API client.
type Client struct {
	server     string
	apiKey     string
	httpClient *http.Client
}

// Option configures a Client.
type Option func(*Client)

// WithServer sets a custom server URL.
func WithServer(url string) Option {
	return func(c *Client) {
		c.server = strings.TrimRight(url, "/")
	}
}

// WithHTTPClient sets a custom http.Client as the underlying transport.
// The provided client's Transport will be wrapped to inject the API key header.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		c.httpClient = hc
	}
}

// New creates a new Ancla API client with the given API key and options.
func New(apiKey string, opts ...Option) *Client {
	c := &Client{
		server: defaultServer,
		apiKey: apiKey,
	}
	for _, opt := range opts {
		opt(c)
	}
	if c.httpClient == nil {
		c.httpClient = &http.Client{}
	}
	// Wrap the transport to inject the API key header.
	base := c.httpClient.Transport
	if base == nil {
		base = http.DefaultTransport
	}
	c.httpClient.Transport = &apiKeyTransport{
		key:  c.apiKey,
		base: base,
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
	return c.server + "/api/v1" + path
}

// do performs an HTTP request and decodes the JSON response into dst.
// If dst is nil, the response body is discarded (useful for DELETE/POST with no response body).
func (c *Client) do(ctx context.Context, method, path string, body any, dst any) error {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encoding request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.apiURL(path), bodyReader)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return c.parseError(resp.StatusCode, respBody)
	}

	if dst != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, dst); err != nil {
			return fmt.Errorf("decoding response: %w", err)
		}
	}
	return nil
}

// parseError converts an HTTP error response into an *APIError.
func (c *Client) parseError(statusCode int, body []byte) error {
	apiErr := &APIError{StatusCode: statusCode}

	switch statusCode {
	case 401:
		apiErr.Message = "not authenticated"
	case 403:
		apiErr.Message = "permission denied"
	case 404:
		apiErr.Message = "not found"
	case 500:
		apiErr.Message = "server error"
	default:
		// Try to extract message from API error response body.
		var errResp struct {
			Status  int    `json:"status"`
			Message string `json:"message"`
			Detail  string `json:"detail"`
		}
		if json.Unmarshal(body, &errResp) == nil {
			msg := errResp.Message
			if msg == "" {
				msg = errResp.Detail
			}
			if msg != "" {
				apiErr.Message = msg
			}
		}
	}

	return apiErr
}
