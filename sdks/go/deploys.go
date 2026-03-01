package ancla

import "context"

// ListDeploys returns all deploys for a service.
func (c *Client) ListDeploys(ctx context.Context, ws, proj, env, svc string) (*DeployList, error) {
	var result DeployList
	if err := c.do(ctx, "GET", servicePath(ws, proj, env)+svc+"/deploys/", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetDeploy returns details for a specific deploy (env-level endpoint).
func (c *Client) GetDeploy(ctx context.Context, ws, proj, env, deployID string) (*Deploy, error) {
	var dpl Deploy
	if err := c.do(ctx, "GET", envPathSDK(ws, proj, env)+"/deploys/"+deployID, nil, &dpl); err != nil {
		return nil, err
	}
	return &dpl, nil
}

// GetDeployLog returns the log for a specific deploy (env-level endpoint).
func (c *Client) GetDeployLog(ctx context.Context, ws, proj, env, deployID string) (*DeployLog, error) {
	var result DeployLog
	if err := c.do(ctx, "GET", envPathSDK(ws, proj, env)+"/deploys/"+deployID+"/log", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
