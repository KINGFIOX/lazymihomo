// Package theme holds shared lipgloss styles for lazymihomo.
package theme

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	colPrimary = lipgloss.Color("#61afef")
	colAccent  = lipgloss.Color("#c678dd")
	colGreen   = lipgloss.Color("#98c379")
	colYellow  = lipgloss.Color("#e5c07b")
	colRed     = lipgloss.Color("#e06c75")
	colGray    = lipgloss.Color("#5c6370")
	colFg      = lipgloss.Color("#abb2bf")
	colBgSel   = lipgloss.Color("#2d323b")
)

// Theme carries the styles shared by every panel.
type Theme struct {
	Panel      lipgloss.Style
	PanelFocus lipgloss.Style
	Title      lipgloss.Style
	TitleFocus lipgloss.Style
	Cursor     lipgloss.Style // focused list cursor row
	CursorDim  lipgloss.Style // unfocused list cursor row
	Dim        lipgloss.Style
	Err        lipgloss.Style
	OK         lipgloss.Style
	Warn       lipgloss.Style
	Accent     lipgloss.Style
	Primary    lipgloss.Style
	Brand      lipgloss.Style
}

// New builds the default theme.
func New() *Theme {
	return &Theme{
		Panel:      lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(colGray),
		PanelFocus: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(colPrimary),
		Title:      lipgloss.NewStyle().Foreground(colGray).Bold(true).Padding(0, 1),
		TitleFocus: lipgloss.NewStyle().Foreground(colPrimary).Bold(true).Padding(0, 1),
		Cursor:     lipgloss.NewStyle().Background(colBgSel).Foreground(colFg),
		CursorDim:  lipgloss.NewStyle().Foreground(colGray),
		Dim:        lipgloss.NewStyle().Foreground(colGray),
		Err:        lipgloss.NewStyle().Foreground(colRed),
		OK:         lipgloss.NewStyle().Foreground(colGreen),
		Warn:       lipgloss.NewStyle().Foreground(colYellow),
		Accent:     lipgloss.NewStyle().Foreground(colAccent),
		Primary:    lipgloss.NewStyle().Foreground(colPrimary),
		Brand:      lipgloss.NewStyle().Foreground(colAccent).Bold(true),
	}
}

// DelayStyle colors a latency value.
func (t *Theme) DelayStyle(ms uint16) lipgloss.Style {
	switch {
	case ms == 0:
		return t.Err
	case ms < 200:
		return t.OK
	case ms < 500:
		return t.Warn
	default:
		return lipgloss.NewStyle().Foreground(colRed)
	}
}

// Box wraps body in a rounded border with the title spliced into the top
// border, sized w x h (outer). focused picks the border color.
func (t *Theme) Box(focused bool, title, body string, w, h int) string {
	if w < 6 || h < 3 {
		return body
	}
	st := lipgloss.NewStyle().Border(lipgloss.RoundedBorder(), true)
	ts := t.Title
	if focused {
		st = st.BorderForeground(colPrimary)
		ts = t.TitleFocus
	}
	boxed := st.Width(w - 2).Height(h - 2).Render(body)
	parts := strings.SplitN(boxed, "\n", 2)
	if len(parts) != 2 {
		return boxed
	}
	label := ts.Render(title)
	pad := w - lipgloss.Width(label) - 4
	if pad < 0 {
		pad = 0
	}
	parts[0] = "╭─" + label + strings.Repeat("─", pad) + "╮"
	return parts[0] + "\n" + parts[1]
}
