package ancla

import "context"

// ListProjects returns all projects within an organization.
// The org parameter is the organization slug.
func (c *Client) ListProjects(ctx context.Context, org string) ([]Project, error) {
	var projects []Project
	if err := c.do(ctx, "GET", "/projects/"+org, nil, &projects); err != nil {
		return nil, err
	}
	return projects, nil
}

// GetProject returns details for a single project.
// The path is constructed as org/slug.
func (c *Client) GetProject(ctx context.Context, org, slug string) (*Project, error) {
	var project Project
	if err := c.do(ctx, "GET", "/projects/"+org+"/"+slug, nil, &project); err != nil {
		return nil, err
	}
	return &project, nil
}

// CreateProject creates a new project within an organization.
func (c *Client) CreateProject(ctx context.Context, org, name string) (*Project, error) {
	var project Project
	if err := c.do(ctx, "POST", "/projects/"+org, CreateProjectRequest{Name: name}, &project); err != nil {
		return nil, err
	}
	return &project, nil
}

// UpdateProject updates the project identified by org and slug.
func (c *Client) UpdateProject(ctx context.Context, org, slug, name string) (*Project, error) {
	var project Project
	if err := c.do(ctx, "PATCH", "/projects/"+org+"/"+slug, UpdateProjectRequest{Name: name}, &project); err != nil {
		return nil, err
	}
	return &project, nil
}

// DeleteProject deletes the project identified by org and slug.
func (c *Client) DeleteProject(ctx context.Context, org, slug string) error {
	return c.do(ctx, "DELETE", "/projects/"+org+"/"+slug, nil, nil)
}
