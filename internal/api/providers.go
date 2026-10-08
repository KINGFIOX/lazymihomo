package api

import (
	"context"
	"fmt"
	"net/http"
)

// GetProviders fetches all proxy providers keyed by name.
func (c *Client) GetProviders(ctx context.Context) (map[string]Provider, error) {
	var r providersResp
	if err := c.doJSON(ctx, http.MethodGet, "/providers/proxies", nil, &r); err != nil {
		return nil, err
	}
	return r.Providers, nil
}

// GetAllProxies returns the /proxies map merged with every provider's nodes.
// mihomo only lists config-defined proxies under /proxies; subscription nodes
// live in the provider objects, so both sources are needed for name lookups.
func (c *Client) GetAllProxies(ctx context.Context) (map[string]Proxy, error) {
	var pr proxiesResp
	if err := c.doJSON(ctx, http.MethodGet, "/proxies", nil, &pr); err != nil {
		return nil, err
	}
	pv, err := c.GetProviders(ctx)
	if err != nil {
		return nil, fmt.Errorf("providers: %w", err)
	}
	for _, p := range pv {
		for _, node := range p.Proxies {
			if node.ProviderName == "" {
				node.ProviderName = p.Name
			}
			pr.Proxies[node.Name] = node
		}
	}
	return pr.Proxies, nil
}

// UpdateProvider triggers a subscription update.
func (c *Client) UpdateProvider(ctx context.Context, name string) error {
	return c.doJSON(ctx, http.MethodPut, "/providers/proxies/"+esc(name), nil, nil)
}

// HealthcheckProvider triggers a health check over a provider's nodes.
func (c *Client) HealthcheckProvider(ctx context.Context, name string) error {
	return c.doJSON(ctx, http.MethodGet, "/providers/proxies/"+esc(name)+"/healthcheck", nil, nil)
}
