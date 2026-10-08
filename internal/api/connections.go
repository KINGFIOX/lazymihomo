package api

import (
	"context"
	"net/http"
)

// GetConnections fetches the current connection snapshot.
func (c *Client) GetConnections(ctx context.Context) (*ConnectionsSnapshot, error) {
	var r ConnectionsSnapshot
	if err := c.doJSON(ctx, http.MethodGet, "/connections", nil, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// CloseConnection closes one connection by ID.
func (c *Client) CloseConnection(ctx context.Context, id string) error {
	return c.doJSON(ctx, http.MethodDelete, "/connections/"+esc(id), nil, nil)
}

// CloseAllConnections closes every live connection.
func (c *Client) CloseAllConnections(ctx context.Context) error {
	return c.doJSON(ctx, http.MethodDelete, "/connections", nil, nil)
}
