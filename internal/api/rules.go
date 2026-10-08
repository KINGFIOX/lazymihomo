package api

import (
	"context"
	"net/http"
)

// GetRules fetches the parsed routing rules.
func (c *Client) GetRules(ctx context.Context) ([]Rule, error) {
	var r rulesResp
	if err := c.doJSON(ctx, http.MethodGet, "/rules", nil, &r); err != nil {
		return nil, err
	}
	return r.Rules, nil
}
