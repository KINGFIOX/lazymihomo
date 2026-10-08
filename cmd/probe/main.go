// probe exercises the app's API client against a live controller; it is a
// debugging aid for connectivity issues like provider-node lookups.
package main

import (
	"context"
	"fmt"

	"lazymihomo/internal/api"
)

func main() {
	c := api.New("http://127.0.0.1:9090", "2ffa0b2a63dd8541223db9e0cb78ae99")
	ctx := context.Background()
	px, err := c.GetAllProxies(ctx)
	if err != nil {
		fmt.Println("proxies:", err)
		return
	}
	gs, err := c.GetGroups(ctx)
	if err != nil {
		fmt.Println("group:", err)
		return
	}
	for _, g := range gs {
		if g.Name == "GLOBAL" || g.Type != "Selector" {
			continue
		}
		fmt.Printf("== group %q (%s), now=%q\n", g.Name, g.Type, g.Now)
		for i, n := range g.All {
			p, ok := px[n]
			if !ok {
				fmt.Printf("  %-40q MISSING even after provider merge\n", n)
				continue
			}
			d, err := c.ProxyDelay(ctx, p.ProviderName, n)
			res := fmt.Sprintf("%dms", d)
			if err != nil {
				res = "ERR: " + err.Error()
			}
			hist := "no-history"
			if len(p.History) > 0 {
				hist = fmt.Sprintf("hist=%dms", p.LastDelay())
			}
			fmt.Printf("  %-40q %-18s %s\n", n, hist, res)
			if i >= 4 {
				fmt.Println("  ... (前5个)")
				break
			}
		}
	}
}
