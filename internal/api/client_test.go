package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAuthHeaderAndVersion(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/version" {
			t.Errorf("path = %q, want /version", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer s3cr3t" {
			t.Errorf("Authorization = %q", got)
		}
		w.Write([]byte(`{"meta":true,"version":"v1.19.32"}`))
	}))
	defer srv.Close()

	c := New(srv.URL, "s3cr3t")
	v, err := c.GetVersion(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !v.Meta || v.Version != "v1.19.32" {
		t.Errorf("version = %+v", v)
	}
}

func TestSelectProxyError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("method = %s", r.Method)
		}
		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "node-a" {
			t.Errorf("body = %v", body)
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"Must be a Selector"}`))
	}))
	defer srv.Close()

	c := New(srv.URL, "")
	err := c.SelectProxy(context.Background(), "🚀 Group", "node-a")
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("err type = %T", err)
	}
	if apiErr.Status != 400 || apiErr.Message != "Must be a Selector" {
		t.Errorf("apiErr = %+v", apiErr)
	}
}

func TestSelectProxyEscapesName(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := New(srv.URL, "")
	if err := c.SelectProxy(context.Background(), "🚀 节点选择", "n"); err != nil {
		t.Fatal(err)
	}
	if want := "/proxies/" + esc("🚀 节点选择"); gotPath != want {
		t.Errorf("escaped path = %q, want %q", gotPath, want)
	}
	_ = strings.TrimSpace("")
}

func TestGetConfigs(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/configs" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Write([]byte(`{"mode":"rule","mixed-port":7890,"allow-lan":true}`))
	}))
	defer srv.Close()

	c := New(srv.URL, "")
	cfg, err := c.GetConfigs(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if cfg["mode"] != "rule" {
		t.Errorf("mode = %v", cfg["mode"])
	}
}

func TestPatchConfigsSendsBody(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("method = %s", r.Method)
		}
		json.NewDecoder(r.Body).Decode(&body)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := New(srv.URL, "")
	if err := c.PatchConfigs(context.Background(), map[string]any{"mode": "global"}); err != nil {
		t.Fatal(err)
	}
	if body["mode"] != "global" {
		t.Errorf("body = %v", body)
	}
}
