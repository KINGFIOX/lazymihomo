package appconfig

import "testing"

func clearEnv(t *testing.T) {
	t.Helper()
	t.Setenv(EnvURL, "")
	t.Setenv(EnvSecret, "")
}

func TestLoadNoDefault(t *testing.T) {
	clearEnv(t)
	cfg, src := Load("", "")
	if cfg.URL != "" || cfg.Secret != "" {
		t.Errorf("expected nothing configured, got %+v", cfg)
	}
	if src != "" {
		t.Errorf("src = %q, want empty", src)
	}
}

func TestLoadFromEnv(t *testing.T) {
	t.Setenv(EnvURL, "http://env:9090")
	t.Setenv(EnvSecret, "envsec")
	cfg, src := Load("", "")
	if cfg.URL != "http://env:9090" || cfg.Secret != "envsec" {
		t.Errorf("cfg = %+v", cfg)
	}
	if src != "env" {
		t.Errorf("src = %q, want env", src)
	}
}

func TestLoadFlagWinsPerField(t *testing.T) {
	t.Setenv(EnvURL, "http://env:9090")
	t.Setenv(EnvSecret, "envsec")
	// flag URL + env secret, and flag secret + env URL
	cfg, src := Load("http://flag:9090", "")
	if cfg.URL != "http://flag:9090" || cfg.Secret != "envsec" {
		t.Errorf("cfg = %+v", cfg)
	}
	if src != "flag" {
		t.Errorf("src = %q, want flag", src)
	}
	cfg, src = Load("", "flagsec")
	if cfg.URL != "http://env:9090" || cfg.Secret != "flagsec" {
		t.Errorf("cfg = %+v", cfg)
	}
}
