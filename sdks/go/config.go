package ancla

import "context"

// ListConfig returns all configuration variables for a service.
func (c *Client) ListConfig(ctx context.Context, ws, proj, env, svc string) ([]ConfigVar, error) {
	var configs []ConfigVar
	if err := c.do(ctx, "GET", servicePath(ws, proj, env)+svc+"/config/", nil, &configs); err != nil {
		return nil, err
	}
	return configs, nil
}

// SetConfig creates or updates a configuration variable for a service.
func (c *Client) SetConfig(ctx context.Context, ws, proj, env, svc, key, value string, secret bool) (*ConfigVar, error) {
	body := SetConfigRequest{
		Name:   key,
		Value:  value,
		Secret: secret,
	}
	var config ConfigVar
	if err := c.do(ctx, "POST", servicePath(ws, proj, env)+svc+"/config/", body, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// DeleteConfig deletes a configuration variable by ID.
func (c *Client) DeleteConfig(ctx context.Context, ws, proj, env, svc, configID string) error {
	return c.do(ctx, "DELETE", servicePath(ws, proj, env)+svc+"/config/"+configID, nil, nil)
}
