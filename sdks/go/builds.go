package ancla

import (
	"context"
	"fmt"
)

// ListBuilds returns all builds for a service.
func (c *Client) ListBuilds(ctx context.Context, ws, proj, env, svc string) (*BuildList, error) {
	var result BuildList
	if err := c.do(ctx, "GET", servicePath(ws, proj, env)+svc+"/builds/", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetBuildLog returns build log details by version number.
func (c *Client) GetBuildLog(ctx context.Context, ws, proj, env, svc string, version int) (*BuildLog, error) {
	var result BuildLog
	path := fmt.Sprintf("%s%s/builds/%d/log", servicePath(ws, proj, env), svc, version)
	if err := c.do(ctx, "GET", path, nil, &result); err != nil {
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
