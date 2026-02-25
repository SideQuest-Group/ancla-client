package ancla

import "context"

// ListImages returns all images for an application.
func (c *Client) ListImages(ctx context.Context, appID string) (*ImageList, error) {
	var result ImageList
	if err := c.do(ctx, "GET", "/images/"+appID, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetImage returns build log details for a specific image.
func (c *Client) GetImage(ctx context.Context, imageID string) (*ImageLog, error) {
	var result ImageLog
	if err := c.do(ctx, "GET", "/images/"+imageID+"/log", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// BuildImage triggers a new image build for an application.
func (c *Client) BuildImage(ctx context.Context, appID string) (*BuildResult, error) {
	var result BuildResult
	if err := c.do(ctx, "POST", "/images/"+appID+"/build", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
