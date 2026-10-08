// Package components holds small reusable UI widgets.
package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// List renders a scrollable, cursor-highlighted column of pre-styled lines.
// Lines may contain ANSI sequences; they are truncated by cell width.
type List struct {
	Width       int
	Height      int
	CursorStyle lipgloss.Style
	Cursor      int
	offset      int
}

// Render draws exactly Height rows (padded), keeping the cursor visible.
func (l *List) Render(lines []string) string {
	if l.Width < 4 || l.Height <= 0 {
		return ""
	}
	if len(lines) == 0 {
		l.Cursor = 0
		return strings.Repeat(lipgloss.NewStyle().Width(l.Width).Render("")+"\n", l.Height-1) +
			lipgloss.NewStyle().Width(l.Width).Render("")
	}
	if l.Cursor < 0 {
		l.Cursor = 0
	}
	if l.Cursor > len(lines)-1 {
		l.Cursor = len(lines) - 1
	}
	if l.Cursor < l.offset {
		l.offset = l.Cursor
	}
	if l.Cursor > l.offset+l.Height-1 {
		l.offset = l.Cursor - l.Height + 1
	}
	if l.offset > len(lines)-l.Height {
		l.offset = len(lines) - l.Height
	}
	if l.offset < 0 {
		l.offset = 0
	}

	var rows []string
	for i := l.offset; i < len(lines) && len(rows) < l.Height; i++ {
		content := ansi.Truncate(lines[i], l.Width-2, "…")
		var row string
		if i == l.Cursor {
			row = l.CursorStyle.Width(l.Width).Render("❯ " + content)
		} else {
			row = lipgloss.NewStyle().Width(l.Width).Render("  " + content)
		}
		rows = append(rows, row)
	}
	blank := lipgloss.NewStyle().Width(l.Width).Render("")
	for len(rows) < l.Height {
		rows = append(rows, blank)
	}
	return strings.Join(rows, "\n")
}

// Up moves the cursor one row up.
func (l *List) Up() { l.Cursor-- }

// Down moves the cursor one row down.
func (l *List) Down() { l.Cursor++ }

// Top jumps to the first row.
func (l *List) Top() { l.Cursor, l.offset = 0, 0 }

// Bottom jumps to the last row.
func (l *List) Bottom() { l.Cursor = 1 << 30 }
