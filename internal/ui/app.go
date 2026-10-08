// Package ui hosts the root model: a single full-screen proxies panel.
package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"lazymihomo/internal/api"
	"lazymihomo/internal/ui/panels"
	"lazymihomo/internal/ui/theme"
)

var (
	keyRefresh = key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh"))
	keyHelp    = key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help"))
	keyQuit    = key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit"))
)

// Model is the root Bubble Tea model.
type Model struct {
	client  *api.Client
	th      *theme.Theme
	version string
	cfgSrc  string

	panel    *panels.Proxies
	help     help.Model
	width    int
	height   int
	ready    bool
	showHelp bool
}

// New builds the root model.
func New(c *api.Client, version, cfgSrc string) *Model {
	t := theme.New()
	return &Model{
		client: c, th: t, version: version, cfgSrc: cfgSrc,
		panel: panels.NewProxies(c, t),
		help:  help.New(),
	}
}

// Init implements tea.Model.
func (m *Model) Init() tea.Cmd { return m.panel.Refresh() }

// Update implements tea.Model.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.help.Width = m.width
		m.panel.SetSize(m.width-2, m.height-3)
		m.ready = m.width > 0
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, m.panel.Update(msg)
}

func (m *Model) handleKey(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	if k.Type == tea.KeyCtrlC {
		return m, tea.Quit
	}
	if m.showHelp {
		m.showHelp = false
		return m, nil
	}
	switch k.String() {
	case "q":
		return m, tea.Quit
	case "?":
		m.showHelp = true
		return m, nil
	case "r":
		return m, m.panel.Refresh()
	}
	return m, m.panel.Update(k)
}

// View implements tea.Model.
func (m *Model) View() string {
	if !m.ready {
		return m.th.Brand.Render("lazymihomo") + "  connecting to " + m.client.BaseURL() + " …"
	}
	if m.showHelp {
		return m.helpView()
	}
	box := m.th.Box(true, m.panel.Title(), m.panel.View(), m.width, m.height-1)
	return lipgloss.JoinVertical(lipgloss.Left, box, m.footerView())
}

// flatKeyMap adapts a []key.Binding to the help.KeyMap interface.
type flatKeyMap []key.Binding

func (f flatKeyMap) ShortHelp() []key.Binding  { return f }
func (f flatKeyMap) FullHelp() [][]key.Binding { return nil }

func (m *Model) footerView() string {
	binds := append([]key.Binding{panels.KeyNav}, m.panel.ShortHelp()...)
	binds = append(binds, keyRefresh, keyHelp, keyQuit)
	left := m.help.View(flatKeyMap(binds))
	right := m.th.Dim.Render(fmt.Sprintf("lazymihomo v%s · %s", m.version, m.client.BaseURL()))
	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	return left + strings.Repeat(" ", gap) + right
}

func (m *Model) helpView() string {
	rows := []string{m.th.Brand.Render("lazymihomo · help"), ""}
	add := func(k, d string) {
		rows = append(rows, "  "+pad(k, 16)+"  "+d)
	}
	add("j / k / g / G", "navigate lists")
	add("r", "refresh")
	add("?", "toggle help")
	add("q / ctrl+c", "quit")
	rows = append(rows, "", m.th.Primary.Render("Proxies keys"), "")
	for _, b := range m.panel.ShortHelp() {
		h := b.Help()
		add(h.Key, h.Desc)
	}
	rows = append(rows, "", m.th.Dim.Render("config source: "+m.cfgSrc))
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#61afef")).
		Padding(1, 2).
		Render(strings.Join(rows, "\n"))
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func pad(s string, w int) string {
	gap := w - lipgloss.Width(s)
	if gap < 0 {
		return s
	}
	return s + strings.Repeat(" ", gap)
}
