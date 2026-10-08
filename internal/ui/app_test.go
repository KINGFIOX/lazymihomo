package ui

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"lazymihomo/internal/api"
	"lazymihomo/internal/ui/panels"
)

func fixtureServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/group", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"proxies":[
			{"type":"Selector","name":"🚀 节点选择","now":"香港 01","all":["♻️ 自动选择","香港 01","香港 02"]},
			{"type":"URLTest","name":"♻️ 自动选择","now":"香港 02","all":["香港 01","香港 02"]},
			{"type":"Selector","name":"GLOBAL","now":"🚀 节点选择","all":["🚀 节点选择"]}]}`)
	})
	mux.HandleFunc("/proxies", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"proxies":{
			"香港 01":{"type":"Hysteria2","name":"香港 01","history":[{"delay":120}]},
			"香港 02":{"type":"Hysteria2","name":"香港 02","history":[{"delay":355}]}}}`)
	})
	mux.HandleFunc("/providers/proxies", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"providers":{}}`)
	})
	return httptest.NewServer(mux)
}

// execCmds runs tea.Cmd chains manually (no real program in tests) and feeds
// each produced message back into the model.
func execCmds(t *testing.T, mm *Model, cmd tea.Cmd) {
	t.Helper()
	for i := 0; cmd != nil && i < 4; i++ {
		msg := cmd()
		if msg == nil {
			return
		}
		next, c := mm.Update(msg)
		mm = next.(*Model)
		cmd = c
	}
}

func TestRenderProxiesPanel(t *testing.T) {
	srv := fixtureServer(t)
	defer srv.Close()

	model := New(api.New(srv.URL, ""), "test", "test")
	m0, _ := model.Update(tea.WindowSizeMsg{Width: 110, Height: 30})
	mm := m0.(*Model)
	execCmds(t, mm, mm.Init())

	view := mm.View()
	for _, want := range []string{"GROUPS", "NODES", "🚀 节点选择", "香港 01", "120ms", "355ms"} {
		if !strings.Contains(view, want) {
			t.Errorf("view missing %q", want)
		}
	}
	// Auto groups and the synthetic GLOBAL group must be hidden.
	for _, unwanted := range []string{"GROUPS (3)", "(URLTest)", "GLOBAL"} {
		if strings.Contains(view, unwanted) {
			t.Errorf("view should not contain %q", unwanted)
		}
	}
	if strings.Count(view, "\n") < 10 {
		t.Errorf("suspiciously short render:\n%s", view)
	}

	// Help overlay renders and closes.
	mm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	if !strings.Contains(mm.View(), "help") {
		t.Error("help overlay missing")
	}

	// Narrow terminal must not panic.
	mm.Update(tea.WindowSizeMsg{Width: 40, Height: 12})
	_ = mm.View()
}

func TestGroupDelayRefresh(t *testing.T) {
	srv := fixtureServer(t)
	defer srv.Close()

	model := New(api.New(srv.URL, ""), "test", "test")
	m0, _ := model.Update(tea.WindowSizeMsg{Width: 110, Height: 30})
	mm := m0.(*Model)
	execCmds(t, mm, mm.Init())
	// A group-delay result triggers a refresh cycle and must not panic.
	next, cmd := mm.Update(panels.PxGroupDelayMsg{
		Group: "🚀 节点选择",
		Delay: map[string]uint16{"香港 01": 100, "香港 02": 200},
	})
	execCmds(t, next.(*Model), cmd)
	_ = next.(*Model).View()
}
