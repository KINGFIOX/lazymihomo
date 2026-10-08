// lazymihomo is a lazygit-style TUI dashboard for the mihomo proxy core.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"lazymihomo/internal/api"
	"lazymihomo/internal/appconfig"
	"lazymihomo/internal/ui"
)

var version = "0.3.0"

func main() {
	urlFlag := flag.String("url", "", "mihomo external-controller base URL")
	secretFlag := flag.String("secret", "", "external-controller API secret")
	check := flag.Bool("check", false, "test controller connectivity and exit")
	showVer := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVer {
		fmt.Println("lazymihomo " + version)
		return
	}

	cfg, src := appconfig.Load(*urlFlag, *secretFlag)

	if *check {
		if cfg.URL == "" {
			fmt.Fprintf(os.Stderr, "no controller configured; set %s or pass --url\n", appconfig.EnvURL)
			os.Exit(1)
		}
		os.Exit(runCheck(api.New(cfg.URL, cfg.Secret), cfg, src))
	}

	// No defaults: anything the flags and environment did not provide is
	// asked for in an interactive dialog (secret may stay empty).
	if cfg.URL == "" || cfg.Secret == "" {
		url, secret, err := ui.RunPrompt(cfg)
		if err != nil {
			fmt.Fprintln(os.Stderr, "lazymihomo:", err)
			fmt.Fprintf(os.Stderr, "hint: export %s / %s or pass --url / --secret\n",
				appconfig.EnvURL, appconfig.EnvSecret)
			os.Exit(1)
		}
		if cfg.URL == "" {
			cfg.URL = url
		}
		if cfg.Secret == "" {
			cfg.Secret = secret
		}
		if src == "" {
			src = "prompt"
		}
	}

	client := api.New(cfg.URL, cfg.Secret)
	m := ui.New(client, version, src)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "lazymihomo:", err)
		os.Exit(1)
	}
}

func runCheck(c *api.Client, cfg appconfig.Config, src string) int {
	fmt.Printf("controller: %s (secret set: %v, from: %s)\n", cfg.URL, cfg.Secret != "", src)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	v, err := c.GetVersion(ctx)
	if err != nil {
		fmt.Println("✗ version:", err)
		return 1
	}
	meta := ""
	if v.Meta {
		meta = " (meta)"
	}
	fmt.Printf("✓ core: mihomo %s%s\n", v.Version, meta)

	if gs, err := c.GetGroups(ctx); err == nil {
		if px, err2 := c.GetProxies(ctx); err2 == nil {
			nodes := 0
			for _, p := range px {
				if !p.IsGroup() {
					nodes++
				}
			}
			fmt.Printf("✓ proxies: %d nodes, %d groups\n", nodes, len(gs))
		}
	}
	fmt.Println("✓ ok")
	return 0
}
