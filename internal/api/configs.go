package api

import (
	"context"
	"net/http"
)

// GetConfigs fetches the runtime general config.
func (c *Client) GetConfigs(ctx context.Context) (map[string]any, error) {
	r := map[string]any{}
	if err := c.doJSON(ctx, http.MethodGet, "/configs", nil, &r); err != nil {
		return nil, err
	}
	return r, nil
}

// PatchConfigs partially updates the runtime config, e.g. {"mode":"global"}.
func (c *Client) PatchConfigs(ctx context.Context, patch map[string]any) error {
	return c.doJSON(ctx, http.MethodPatch, "/configs", patch, nil)
}
