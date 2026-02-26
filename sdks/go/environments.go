package ancla

import "context"

// ListEnvironments returns all environments within a project.
func (c *Client) ListEnvironments(ctx context.Context, ws, proj string) ([]Environment, error) {
	var envs []Environment
	if err := c.do(ctx, "GET", "/workspaces/"+ws+"/projects/"+proj+"/envs/", nil, &envs); err != nil {
		return nil, err
	}
	return envs, nil
}

// GetEnvironment returns details for a single environment.
func (c *Client) GetEnvironment(ctx context.Context, ws, proj, slug string) (*Environment, error) {
	var env Environment
	if err := c.do(ctx, "GET", "/workspaces/"+ws+"/projects/"+proj+"/envs/"+slug, nil, &env); err != nil {
		return nil, err
	}
	return &env, nil
}

// CreateEnvironment creates a new environment within a project.
func (c *Client) CreateEnvironment(ctx context.Context, ws, proj, name string) (*Environment, error) {
	var env Environment
	if err := c.do(ctx, "POST", "/workspaces/"+ws+"/projects/"+proj+"/envs/", CreateEnvironmentRequest{Name: name}, &env); err != nil {
		return nil, err
	}
	return &env, nil
}
