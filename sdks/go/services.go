package ancla

import (
	"context"
	"fmt"
)

// envPathSDK builds the path prefix up to the environment level.
func envPathSDK(ws, proj, env string) string {
	return fmt.Sprintf("/workspaces/%s/projects/%s/envs/%s", ws, proj, env)
}

// servicePath builds the base path for service operations within an environment.
func servicePath(ws, proj, env string) string {
	return envPathSDK(ws, proj, env) + "/services/"
}

// ListServices returns all services within an environment.
func (c *Client) ListServices(ctx context.Context, ws, proj, env string) ([]Service, error) {
	var services []Service
	if err := c.do(ctx, "GET", servicePath(ws, proj, env), nil, &services); err != nil {
		return nil, err
	}
	return services, nil
}

// GetService returns details for a single service.
func (c *Client) GetService(ctx context.Context, ws, proj, env, slug string) (*Service, error) {
	var svc Service
	if err := c.do(ctx, "GET", servicePath(ws, proj, env)+slug, nil, &svc); err != nil {
		return nil, err
	}
	return &svc, nil
}

// CreateService creates a new service in an environment.
func (c *Client) CreateService(ctx context.Context, ws, proj, env, name, platform string) (*Service, error) {
	var svc Service
	body := CreateServiceRequest{Name: name, Platform: platform}
	if err := c.do(ctx, "POST", servicePath(ws, proj, env), body, &svc); err != nil {
		return nil, err
	}
	return &svc, nil
}

// UpdateService updates a service. The opts fields that are non-nil will be sent.
func (c *Client) UpdateService(ctx context.Context, ws, proj, env, slug string, opts UpdateServiceOptions) (*Service, error) {
	var svc Service
	if err := c.do(ctx, "PATCH", servicePath(ws, proj, env)+slug, opts, &svc); err != nil {
		return nil, err
	}
	return &svc, nil
}

// DeleteService deletes a service.
func (c *Client) DeleteService(ctx context.Context, ws, proj, env, slug string) error {
	return c.do(ctx, "DELETE", servicePath(ws, proj, env)+slug, nil, nil)
}

// DeployService triggers a full deploy for a service.
// The svcID is the service's unique identifier (not the slug path).
func (c *Client) DeployService(ctx context.Context, ws, proj, env, svcID string) (*BuildResult, error) {
	var result BuildResult
	if err := c.do(ctx, "POST", servicePath(ws, proj, env)+svcID+"/deploy", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ScaleService scales a service's processes.
// The counts map specifies the desired count for each process type (e.g., {"web": 2, "worker": 1}).
func (c *Client) ScaleService(ctx context.Context, ws, proj, env, svcID string, counts map[string]int) error {
	return c.do(ctx, "POST", servicePath(ws, proj, env)+svcID+"/scale", ScaleRequest{ProcessCounts: counts}, nil)
}

// GetServiceStatus returns the pipeline status for a service.
func (c *Client) GetServiceStatus(ctx context.Context, ws, proj, env, svcID string) (*PipelineStatus, error) {
	var status PipelineStatus
	if err := c.do(ctx, "GET", servicePath(ws, proj, env)+svcID+"/pipeline-status", nil, &status); err != nil {
		return nil, err
	}
	return &status, nil
}
