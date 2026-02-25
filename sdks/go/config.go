package ancla

import "context"

// ListConfig returns all configuration variables for an application.
// The appID is the application's unique identifier.
func (c *Client) ListConfig(ctx context.Context, appID string) ([]ConfigVar, error) {
	var configs []ConfigVar
	if err := c.do(ctx, "GET", "/configurations/"+appID, nil, &configs); err != nil {
		return nil, err
	}
	return configs, nil
}

// GetConfig returns a single configuration variable by ID.
func (c *Client) GetConfig(ctx context.Context, appID, configID string) (*ConfigVar, error) {
	var config ConfigVar
	if err := c.do(ctx, "GET", "/configurations/"+appID+"/"+configID, nil, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// SetConfig creates or updates a configuration variable for an application.
func (c *Client) SetConfig(ctx context.Context, appID, key, value string, secret bool) (*ConfigVar, error) {
	body := SetConfigRequest{
		Name:   key,
		Value:  value,
		Secret: secret,
	}
	var config ConfigVar
	if err := c.do(ctx, "POST", "/configurations/"+appID, body, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// DeleteConfig deletes a configuration variable by ID.
func (c *Client) DeleteConfig(ctx context.Context, appID, configID string) error {
	return c.do(ctx, "DELETE", "/configurations/"+appID+"/"+configID, nil, nil)
}
