package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"lazymihomo/internal/appconfig"
)

func TestPromptSubmit(t *testing.T) {
	m := newPromptModel(appconfig.Config{})
	m.inputs[0].SetValue("http://127.0.0.1:9090")
	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter}) // url → secret
	fin := out.(*promptModel)
	if fin.focus != 1 || fin.err != "" {
		t.Fatalf("focus=%d err=%q", fin.focus, fin.err)
	}
	fin.inputs[1].SetValue("s3cr3t")
	out, _ = fin.Update(tea.KeyMsg{Type: tea.KeyEnter})
	fin = out.(*promptModel)
	if !fin.submitted {
		t.Fatal("not submitted")
	}
	if fin.inputs[0].Value() != "http://127.0.0.1:9090" || fin.inputs[1].Value() != "s3cr3t" {
		t.Errorf("values = %q / %q", fin.inputs[0].Value(), fin.inputs[1].Value())
	}
}

func TestPromptRequiresURL(t *testing.T) {
	m := newPromptModel(appconfig.Config{})
	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter}) // empty url → error, stay on field 0
	fin := out.(*promptModel)
	if fin.err == "" || fin.focus != 0 || fin.submitted {
		t.Errorf("err=%q focus=%d submitted=%v", fin.err, fin.focus, fin.submitted)
	}
}

func TestPromptCancel(t *testing.T) {
	m := newPromptModel(appconfig.Config{})
	m.inputs[0].SetValue("http://x:9090")
	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	fin := out.(*promptModel)
	if !fin.cancelled || fin.submitted {
		t.Errorf("cancelled=%v submitted=%v", fin.cancelled, fin.submitted)
	}
}

func TestPromptViewRenders(t *testing.T) {
	m := newPromptModel(appconfig.Config{})
	m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	v := m.View()
	for _, want := range []string{"connect to mihomo", "url ›", "secret ›"} {
		if !strings.Contains(v, want) {
			t.Errorf("view missing %q", want)
		}
	}
}
