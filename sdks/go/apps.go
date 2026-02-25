package ancla

import "context"

// ListApps returns all applications in a project.
func (c *Client) ListApps(ctx context.Context, org, project string) ([]App, error) {
	var apps []App
	if err := c.do(ctx, "GET", "/applications/"+org+"/"+project, nil, &apps); err != nil {
		return nil, err
	}
	return apps, nil
}

// GetApp returns details for a single application.
func (c *Client) GetApp(ctx context.Context, org, project, slug string) (*App, error) {
	var app App
	if err := c.do(ctx, "GET", "/applications/"+org+"/"+project+"/"+slug, nil, &app); err != nil {
		return nil, err
	}
	return &app, nil
}

// CreateApp creates a new application in a project.
func (c *Client) CreateApp(ctx context.Context, org, project, name, platform string) (*App, error) {
	var app App
	body := CreateAppRequest{Name: name, Platform: platform}
	if err := c.do(ctx, "POST", "/applications/"+org+"/"+project, body, &app); err != nil {
		return nil, err
	}
	return &app, nil
}

// UpdateApp updates an application. The opts fields that are non-nil will be sent.
func (c *Client) UpdateApp(ctx context.Context, org, project, slug string, opts UpdateAppOptions) (*App, error) {
	var app App
	if err := c.do(ctx, "PATCH", "/applications/"+org+"/"+project+"/"+slug, opts, &app); err != nil {
		return nil, err
	}
	return &app, nil
}

// DeleteApp deletes an application.
func (c *Client) DeleteApp(ctx context.Context, org, project, slug string) error {
	return c.do(ctx, "DELETE", "/applications/"+org+"/"+project+"/"+slug, nil, nil)
}

// DeployApp triggers a full deploy for an application.
// The appID is the application's unique identifier (not the slug path).
func (c *Client) DeployApp(ctx context.Context, appID string) (*DeployResult, error) {
	var result DeployResult
	if err := c.do(ctx, "POST", "/applications/"+appID+"/deploy", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ScaleApp scales an application's processes.
// The appID is the application's unique identifier. The counts map specifies
// the desired count for each process type (e.g., {"web": 2, "worker": 1}).
func (c *Client) ScaleApp(ctx context.Context, appID string, counts map[string]int) error {
	return c.do(ctx, "POST", "/applications/"+appID+"/scale", ScaleRequest{ProcessCounts: counts}, nil)
}

// GetAppStatus returns the pipeline status for an application.
func (c *Client) GetAppStatus(ctx context.Context, appID string) (*PipelineStatus, error) {
	var status PipelineStatus
	if err := c.do(ctx, "GET", "/applications/"+appID+"/pipeline-status", nil, &status); err != nil {
		return nil, err
	}
	return &status, nil
}
