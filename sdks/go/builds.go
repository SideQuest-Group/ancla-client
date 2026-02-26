package ancla

import "context"

// ListBuilds returns all builds for a service.
func (c *Client) ListBuilds(ctx context.Context, ws, proj, env, svc string) (*BuildList, error) {
	var result BuildList
	if err := c.do(ctx, "GET", servicePath(ws, proj, env)+svc+"/builds/", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetBuildLog returns build log details for a specific build.
func (c *Client) GetBuildLog(ctx context.Context, buildID string) (*BuildLog, error) {
	var result BuildLog
	if err := c.do(ctx, "GET", "/builds/"+buildID+"/log", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// TriggerBuild triggers a new build for a service.
func (c *Client) TriggerBuild(ctx context.Context, ws, proj, env, svc string) (*BuildResult, error) {
	var result BuildResult
	if err := c.do(ctx, "POST", servicePath(ws, proj, env)+svc+"/builds/", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
