package ancla

import "context"

// ListProjects returns all projects within a workspace.
func (c *Client) ListProjects(ctx context.Context, ws string) ([]Project, error) {
	var projects []Project
	if err := c.do(ctx, "GET", "/workspaces/"+ws+"/projects/", nil, &projects); err != nil {
		return nil, err
	}
	return projects, nil
}

// GetProject returns details for a single project.
func (c *Client) GetProject(ctx context.Context, ws, slug string) (*Project, error) {
	var project Project
	if err := c.do(ctx, "GET", "/workspaces/"+ws+"/projects/"+slug, nil, &project); err != nil {
		return nil, err
	}
	return &project, nil
}

// CreateProject creates a new project within a workspace.
func (c *Client) CreateProject(ctx context.Context, ws, name string) (*Project, error) {
	var project Project
	if err := c.do(ctx, "POST", "/workspaces/"+ws+"/projects/", CreateProjectRequest{Name: name}, &project); err != nil {
		return nil, err
	}
	return &project, nil
}

// UpdateProject updates the project identified by workspace and slug.
func (c *Client) UpdateProject(ctx context.Context, ws, slug, name string) (*Project, error) {
	var project Project
	if err := c.do(ctx, "PATCH", "/workspaces/"+ws+"/projects/"+slug, UpdateProjectRequest{Name: name}, &project); err != nil {
		return nil, err
	}
	return &project, nil
}

// DeleteProject deletes the project identified by workspace and slug.
func (c *Client) DeleteProject(ctx context.Context, ws, slug string) error {
	return c.do(ctx, "DELETE", "/workspaces/"+ws+"/projects/"+slug, nil, nil)
}
