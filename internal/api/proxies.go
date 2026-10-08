package api

import (
	"context"
	"net/http"
	"net/url"
)

// DefaultTestURL is the latency probe target used across the UI.
const DefaultTestURL = "https://www.gstatic.com/generate_204"

// DefaultTestTimeoutMS is the per-probe timeout in milliseconds.
const DefaultTestTimeoutMS = 5000

func esc(s string) string { return url.PathEscape(s) }

func delayQuery() string {
	return "url=" + url.QueryEscape(DefaultTestURL) + "&timeout=5000"
}

// GetProxies fetches every proxy and group, keyed by name.
func (c *Client) GetProxies(ctx context.Context) (map[string]Proxy, error) {
	var r proxiesResp
	if err := c.doJSON(ctx, http.MethodGet, "/proxies", nil, &r); err != nil {
		return nil, err
	}
	return r.Proxies, nil
}

// GetGroups fetches the ordered list of policy groups.
func (c *Client) GetGroups(ctx context.Context) ([]Proxy, error) {
	var r groupsResp
	if err := c.doJSON(ctx, http.MethodGet, "/group", nil, &r); err != nil {
		return nil, err
	}
	return r.Proxies, nil
}

// SelectProxy switches the selected node of a selector group.
func (c *Client) SelectProxy(ctx context.Context, group, node string) error {
	return c.doJSON(ctx, http.MethodPut, "/proxies/"+esc(group), map[string]string{"name": node}, nil)
}

// ProxyDelay probes one node and returns the latency in milliseconds.
// Subscription nodes are not addressable under /proxies/:name/delay (mihomo
// keeps them out of that map), so provider nodes go through the
// provider-scoped healthcheck route instead.
func (c *Client) ProxyDelay(ctx context.Context, provider, name string) (uint16, error) {
	var r struct {
		Delay uint16 `json:"delay"`
	}
	var path string
	if provider != "" {
		path = "/providers/proxies/" + esc(provider) + "/" + esc(name) + "/healthcheck?" + delayQuery()
	} else {
		path = "/proxies/" + esc(name) + "/delay?" + delayQuery()
	}
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &r); err != nil {
		return 0, err
	}
	return r.Delay, nil
}

// GroupDelay probes every node in a group, returning node name -> delay(ms).
func (c *Client) GroupDelay(ctx context.Context, group string) (map[string]uint16, error) {
	r := map[string]uint16{}
	path := "/group/" + esc(group) + "/delay?" + delayQuery()
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &r); err != nil {
		return nil, err
	}
	return r, nil
}
