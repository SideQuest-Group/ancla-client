package ancla

import "context"

// ListReleases returns all releases for an application.
func (c *Client) ListReleases(ctx context.Context, appID string) (*ReleaseList, error) {
	var result ReleaseList
	if err := c.do(ctx, "GET", "/releases/"+appID, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetRelease returns details for a specific release by listing releases
// and finding the one with the matching ID.
func (c *Client) GetRelease(ctx context.Context, appID, releaseID string) (*Release, error) {
	list, err := c.ListReleases(ctx, appID)
	if err != nil {
		return nil, err
	}
	for _, r := range list.Items {
		if r.ID == releaseID {
			return &r, nil
		}
	}
	return nil, &APIError{StatusCode: 404, Message: "release not found"}
}

// CreateRelease creates a new release for an application.
func (c *Client) CreateRelease(ctx context.Context, appID string) (*CreateReleaseResult, error) {
	var result CreateReleaseResult
	if err := c.do(ctx, "POST", "/releases/"+appID+"/create", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DeployRelease deploys a release by its ID.
func (c *Client) DeployRelease(ctx context.Context, releaseID string) (*DeployReleaseResult, error) {
	var result DeployReleaseResult
	if err := c.do(ctx, "POST", "/releases/"+releaseID+"/deploy", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
