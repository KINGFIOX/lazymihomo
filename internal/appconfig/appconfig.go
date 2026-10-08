// Package appconfig resolves lazymihomo's launch configuration.
//
// There is deliberately no default controller address: flags win over the
// LAZYMIHOMO_URL / LAZYMIHOMO_SECRET environment variables, and when the URL
// or secret is still missing the UI shows an interactive dialog.
package appconfig

import "os"

// Config points at a mihomo external-controller.
type Config struct {
	URL    string
	Secret string
}

// Environment variable names read by Load.
const (
	EnvURL    = "LAZYMIHOMO_URL"
	EnvSecret = "LAZYMIHOMO_SECRET"
)

// Load resolves per-field priority flag > env. The returned source label is
// "flag", "env" or "" (nothing configured yet — caller should prompt).
func Load(urlFlag, secretFlag string) (Config, string) {
	cfg := Config{URL: urlFlag, Secret: secretFlag}
	src := ""
	if cfg.URL != "" || cfg.Secret != "" {
		src = "flag"
	}
	if cfg.URL == "" {
		cfg.URL = os.Getenv(EnvURL)
	}
	if cfg.Secret == "" {
		cfg.Secret = os.Getenv(EnvSecret)
	}
	if src == "" && (cfg.URL != "" || cfg.Secret != "") {
		src = "env"
	}
	return cfg, src
}
