package ancla

import "context"

// GetDeployment returns details for a specific deployment.
func (c *Client) GetDeployment(ctx context.Context, deploymentID string) (*Deployment, error) {
	var dpl Deployment
	if err := c.do(ctx, "GET", "/deployments/"+deploymentID+"/detail", nil, &dpl); err != nil {
		return nil, err
	}
	return &dpl, nil
}

// GetDeploymentLog returns the log for a specific deployment.
func (c *Client) GetDeploymentLog(ctx context.Context, deploymentID string) (*DeploymentLog, error) {
	var result DeploymentLog
	if err := c.do(ctx, "GET", "/deployments/"+deploymentID+"/log", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
