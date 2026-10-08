// Package panels implements the proxies view of lazymihomo.
package panels

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"lazymihomo/internal/ui/theme"
)

// KeyNav is the shared list-navigation binding shown in the footer.
var KeyNav = key.NewBinding(key.WithKeys("j", "down", "k", "up"), key.WithHelp("j/k", "move"))

func truncate(s string, w int) string {
	if w < 1 {
		return ""
	}
	return ansi.Truncate(s, w, "…")
}

// statusLine renders the one-line footer of the panel.
func statusLine(t *theme.Theme, err error, status string, w int) string {
	if err != nil {
		return truncate(t.Err.Render("✗ "+err.Error()), w)
	}
	return truncate(t.Dim.Render("· "+status), w)
}

// rightPad pads a possibly-styled string with spaces up to cell width w.
func rightPad(s string, w int) string {
	gap := w - lipgloss.Width(s)
	if gap < 0 {
		gap = 0
	}
	return s + strings.Repeat(" ", gap)
}
