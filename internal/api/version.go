package api

import (
	"context"
	"net/http"
)

// GetVersion fetches core version info.
func (c *Client) GetVersion(ctx context.Context) (Version, error) {
	var v Version
	if err := c.doJSON(ctx, http.MethodGet, "/version", nil, &v); err != nil {
		return v, err
	}
	return v, nil
}
