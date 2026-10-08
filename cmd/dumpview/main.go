// dumpview renders one lazymihomo frame to stdout. It is a layout debugging
// aid: no terminal is required.
package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"lazymihomo/internal/api"
	"lazymihomo/internal/appconfig"
	"lazymihomo/internal/ui"
)

func run(m tea.Model, cmd tea.Cmd) tea.Model {
	for i := 0; cmd != nil && i < 6; i++ {
		msg := cmd()
		if msg == nil {
			break
		}
		var c tea.Cmd
		m, c = m.Update(msg)
		cmd = c
	}
	return m
}

func main() {
	cfg, src := appconfig.Load("", "")
	client := api.New(cfg.URL, cfg.Secret)

	model := ui.New(client, "dump", src)
	m0, _ := model.Update(tea.WindowSizeMsg{Width: 110, Height: 34})
	m := run(m0, m0.(*ui.Model).Init())
	fmt.Println(m.View())
}
