package ancla

import "context"

// ListOrgs returns all organizations the authenticated user belongs to.
func (c *Client) ListOrgs(ctx context.Context) ([]Org, error) {
	var orgs []Org
	if err := c.do(ctx, "GET", "/organizations/", nil, &orgs); err != nil {
		return nil, err
	}
	return orgs, nil
}

// GetOrg returns details for a single organization by slug.
func (c *Client) GetOrg(ctx context.Context, slug string) (*Org, error) {
	var org Org
	if err := c.do(ctx, "GET", "/organizations/"+slug, nil, &org); err != nil {
		return nil, err
	}
	return &org, nil
}

// CreateOrg creates a new organization with the given name.
func (c *Client) CreateOrg(ctx context.Context, name string) (*Org, error) {
	var org Org
	if err := c.do(ctx, "POST", "/organizations/", CreateOrgRequest{Name: name}, &org); err != nil {
		return nil, err
	}
	return &org, nil
}

// UpdateOrg updates the organization identified by slug.
func (c *Client) UpdateOrg(ctx context.Context, slug string, name string) (*Org, error) {
	var org Org
	if err := c.do(ctx, "PATCH", "/organizations/"+slug, UpdateOrgRequest{Name: name}, &org); err != nil {
		return nil, err
	}
	return &org, nil
}

// DeleteOrg deletes the organization identified by slug.
func (c *Client) DeleteOrg(ctx context.Context, slug string) error {
	return c.do(ctx, "DELETE", "/organizations/"+slug, nil, nil)
}
