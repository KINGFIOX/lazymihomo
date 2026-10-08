package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetGroupsAndProxies(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/group":
			w.Write([]byte(`{"proxies":[
				{"type":"Selector","name":"🚀 节点选择","now":"香港 01","all":["香港 01","香港 02"]},
				{"type":"URLTest","name":"♻️ 自动选择","all":["香港 01"]}
			]}`))
		case "/proxies":
			w.Write([]byte(`{"proxies":{
				"DIRECT":{"type":"Direct","name":"DIRECT"},
				"香港 01":{"type":"Hysteria2","name":"香港 01","history":[{"delay":120}]},
				"香港 02":{"type":"Hysteria2","name":"香港 02","history":[{"delay":355}]}
			}}`))
		default:
			t.Errorf("unexpected path %q", r.URL.Path)
		}
	}))
	defer srv.Close()

	c := New(srv.URL, "")
	gs, err := c.GetGroups(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(gs) != 2 || gs[0].Name != "🚀 节点选择" || gs[0].Now != "香港 01" {
		t.Errorf("groups = %+v", gs)
	}
	px, err := c.GetProxies(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if d := px["香港 02"].LastDelay(); d != 355 {
		t.Errorf("LastDelay = %d", d)
	}
}

func TestProxyAndGroupDelay(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("url") != DefaultTestURL {
			t.Errorf("url param = %q", q.Get("url"))
		}
		if q.Get("timeout") != "5000" {
			t.Errorf("timeout param = %q", q.Get("timeout"))
		}
		path := r.URL.EscapedPath()
		switch {
		case path == "/proxies/"+esc("香港 01")+"/delay":
			w.Write([]byte(`{"delay":233}`))
		case path == "/providers/proxies/nosub/"+esc("香港 02")+"/healthcheck":
			w.Write([]byte(`{"delay":140}`))
		case path == "/group/G/delay":
			w.Write([]byte(`{"香港 01":120,"香港 02":350}`))
		default:
			t.Errorf("unexpected path %q", path)
		}
	}))
	defer srv.Close()

	c := New(srv.URL, "")
	d, err := c.ProxyDelay(context.Background(), "", "香港 01")
	if err != nil || d != 233 {
		t.Errorf("ProxyDelay(config proxy) = %d, %v", d, err)
	}
	d, err = c.ProxyDelay(context.Background(), "nosub", "香港 02")
	if err != nil || d != 140 {
		t.Errorf("ProxyDelay(provider node) = %d, %v", d, err)
	}
	m, err := c.GroupDelay(context.Background(), "G")
	if err != nil || len(m) != 2 || m["香港 02"] != 350 {
		t.Errorf("GroupDelay = %v, %v", m, err)
	}
}

func TestGetAllProxiesMergesProviders(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/proxies":
			// Like real mihomo: only config-defined proxies are listed here.
			w.Write([]byte(`{"proxies":{
				"DIRECT":{"type":"Direct","name":"DIRECT"},
				"🚀 节点选择":{"type":"Selector","name":"🚀 节点选择","all":["香港 01"]}}}`))
		case "/providers/proxies":
			w.Write([]byte(`{"providers":{"nosub":{"name":"nosub","vehicleType":"HTTP",
				"proxies":[{"type":"Hysteria2","name":"香港 01","history":[{"delay":120}]}]}}}`))
		default:
			t.Errorf("unexpected path %q", r.URL.Path)
		}
	}))
	defer srv.Close()

	c := New(srv.URL, "")
	px, err := c.GetAllProxies(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	node, ok := px["香港 01"]
	if !ok {
		t.Fatal("provider node not merged into proxies map")
	}
	if node.ProviderName != "nosub" {
		t.Errorf("provider-name = %q, want nosub", node.ProviderName)
	}
	if node.LastDelay() != 120 {
		t.Errorf("LastDelay = %d", node.LastDelay())
	}
	if _, ok := px["🚀 节点选择"]; !ok {
		t.Error("config proxies lost in merge")
	}
}

func TestConnections(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method + " " + r.URL.Path {
		case "GET /connections":
			w.Write([]byte(`{"downloadTotal":100,"uploadTotal":50,"memory":1024,"connections":[
				{"id":"abc","upload":10,"download":20,"chains":["香港 01","🚀 节点选择"],
				 "rule":"GEOIP","rulePayload":"CN",
				 "metadata":{"network":"tcp","host":"example.com","destinationIP":"1.2.3.4","destinationPort":"443"}}
			]}`))
		case "DELETE /connections/abc":
			w.WriteHeader(http.StatusNoContent)
		case "DELETE /connections":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected %s %q", r.Method, r.URL.Path)
		}
	}))
	defer srv.Close()

	c := New(srv.URL, "")
	snap, err := c.GetConnections(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if snap.DownloadTotal != 100 || len(snap.Connections) != 1 {
		t.Errorf("snap = %+v", snap)
	}
	if hp := snap.Connections[0].Metadata.Hostport(); hp != "example.com:443" {
		t.Errorf("Hostport = %q", hp)
	}
	if err := c.CloseConnection(context.Background(), "abc"); err != nil {
		t.Error(err)
	}
	if err := c.CloseAllConnections(context.Background()); err != nil {
		t.Error(err)
	}
}

func TestProviders(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method + " " + r.URL.Path {
		case "GET /providers/proxies":
			w.Write([]byte(`{"providers":{
				"default":{"name":"default","vehicleType":"Compatible","proxies":[{"name":"DIRECT"}]},
				"nosub":{"name":"nosub","vehicleType":"HTTP","updatedAt":"2026-10-07T21:42:00Z",
				 "subscriptionInfo":{"Upload":100,"Download":900,"Total":2000,"Expire":1762000000},
				 "proxies":[{"name":"香港 01"},{"name":"香港 02"}]}
			}}`))
		case "PUT /providers/proxies/nosub":
			w.WriteHeader(http.StatusNoContent)
		case "GET /providers/proxies/nosub/healthcheck":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected %s %q", r.Method, r.URL.Path)
		}
	}))
	defer srv.Close()

	c := New(srv.URL, "")
	pv, err := c.GetProviders(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	sub := pv["nosub"]
	if sub.VehicleType != "HTTP" || len(sub.Proxies) != 2 {
		t.Errorf("sub = %+v", sub)
	}
	if si := sub.SubscriptionInfo; si == nil || si.Total != 2000 {
		t.Errorf("subscriptionInfo = %+v", sub.SubscriptionInfo)
	}
	if err := c.UpdateProvider(context.Background(), "nosub"); err != nil {
		t.Error(err)
	}
	if err := c.HealthcheckProvider(context.Background(), "nosub"); err != nil {
		t.Error(err)
	}
}

func TestRules(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"rules":[
			{"index":0,"type":"GEOIP","payload":"LAN","proxy":"DIRECT","size":2048,
			 "extra":{"hitCount":42}},
			{"index":1,"type":"MATCH","payload":"","proxy":"🚀 节点选择","size":-1}
		]}`))
	}))
	defer srv.Close()

	c := New(srv.URL, "")
	rs, err := c.GetRules(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(rs) != 2 || rs[0].Extra == nil || rs[0].Extra.HitCount != 42 {
		t.Errorf("rules = %+v", rs)
	}
}
