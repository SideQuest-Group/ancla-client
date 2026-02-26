package ancla

import "context"

// ListWorkspaces returns all workspaces the authenticated user belongs to.
func (c *Client) ListWorkspaces(ctx context.Context) ([]Workspace, error) {
	var workspaces []Workspace
	if err := c.do(ctx, "GET", "/workspaces/", nil, &workspaces); err != nil {
		return nil, err
	}
	return workspaces, nil
}

// GetWorkspace returns details for a single workspace by slug.
func (c *Client) GetWorkspace(ctx context.Context, slug string) (*Workspace, error) {
	var ws Workspace
	if err := c.do(ctx, "GET", "/workspaces/"+slug, nil, &ws); err != nil {
		return nil, err
	}
	return &ws, nil
}

// CreateWorkspace creates a new workspace with the given name.
func (c *Client) CreateWorkspace(ctx context.Context, name string) (*Workspace, error) {
	var ws Workspace
	if err := c.do(ctx, "POST", "/workspaces/", CreateWorkspaceRequest{Name: name}, &ws); err != nil {
		return nil, err
	}
	return &ws, nil
}

// UpdateWorkspace updates the workspace identified by slug.
func (c *Client) UpdateWorkspace(ctx context.Context, slug string, name string) (*Workspace, error) {
	var ws Workspace
	if err := c.do(ctx, "PATCH", "/workspaces/"+slug, UpdateWorkspaceRequest{Name: name}, &ws); err != nil {
		return nil, err
	}
	return &ws, nil
}

// DeleteWorkspace deletes the workspace identified by slug.
func (c *Client) DeleteWorkspace(ctx context.Context, slug string) error {
	return c.do(ctx, "DELETE", "/workspaces/"+slug, nil, nil)
}
