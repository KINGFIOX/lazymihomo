// The connection dialog: asks for the controller URL and secret when flags
// and environment variables do not provide them.
package ui

import (
	"errors"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"lazymihomo/internal/appconfig"
	"lazymihomo/internal/ui/theme"
)

// ErrPromptCancelled is returned when the user aborts the connection dialog.
var ErrPromptCancelled = errors.New("cancelled")

type promptModel struct {
	th        *theme.Theme
	inputs    []textinput.Model
	focus     int
	err       string
	w, h      int
	submitted bool
	cancelled bool
}

func newPromptModel(prefill appconfig.Config) *promptModel {
	urlIn := textinput.New()
	urlIn.Placeholder = "e.g. http://127.0.0.1:9090"
	urlIn.Prompt = "url › "
	urlIn.SetValue(prefill.URL)
	urlIn.Width = 44
	urlIn.Focus()

	secIn := textinput.New()
	secIn.Placeholder = "leave empty if no secret"
	secIn.Prompt = "secret › "
	secIn.EchoMode = textinput.EchoPassword
	secIn.SetValue(prefill.Secret)
	secIn.Width = 44

	return &promptModel{th: theme.New(), inputs: []textinput.Model{urlIn, secIn}, w: 80, h: 24}
}

// RunPrompt shows the connection dialog. Known values (from flags/env) are
// prefilled; the secret may be left empty.
func RunPrompt(prefill appconfig.Config) (url, secret string, err error) {
	m := newPromptModel(prefill)
	p := tea.NewProgram(m, tea.WithAltScreen())
	out, err := p.Run()
	if err != nil {
		return "", "", err
	}
	fin, ok := out.(*promptModel)
	if !ok || fin.cancelled || !fin.submitted {
		return "", "", ErrPromptCancelled
	}
	return strings.TrimSpace(fin.inputs[0].Value()), fin.inputs[1].Value(), nil
}

// Init implements tea.Model.
func (m *promptModel) Init() tea.Cmd { return textinput.Blink }

// Update implements tea.Model.
func (m *promptModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if sz, ok := msg.(tea.WindowSizeMsg); ok {
		m.w, m.h = sz.Width, sz.Height
		return m, nil
	}
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "ctrl+c", "esc":
			m.cancelled = true
			return m, tea.Quit
		case "tab", "down":
			m.switchFocus(m.focus + 1)
			return m, nil
		case "shift+tab", "up":
			m.switchFocus(m.focus - 1)
			return m, nil
		case "enter":
			if strings.TrimSpace(m.inputs[0].Value()) == "" {
				m.err = "url is required"
				m.switchFocus(0)
				return m, nil
			}
			if m.focus == 0 {
				m.err = ""
				m.switchFocus(1)
				return m, nil
			}
			m.submitted = true
			return m, tea.Quit
		}
	}
	var cmds []tea.Cmd
	for i := range m.inputs {
		var cmd tea.Cmd
		m.inputs[i], cmd = m.inputs[i].Update(msg)
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

func (m *promptModel) switchFocus(i int) {
	if i < 0 || i >= len(m.inputs) {
		return
	}
	m.inputs[m.focus].Blur()
	m.focus = i
	m.inputs[m.focus].Focus()
}

// View implements tea.Model.
func (m *promptModel) View() string {
	rows := []string{
		m.th.Brand.Render("connect to mihomo"),
		m.th.Dim.Render("external-controller address and secret"),
		"",
		m.inputs[0].View(),
		m.inputs[1].View(),
		"",
	}
	if m.err != "" {
		rows = append(rows, m.th.Err.Render(m.err))
	} else {
		rows = append(rows, m.th.Dim.Render("enter confirm · tab switch field · esc cancel"))
	}
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#61afef")).
		Padding(1, 2).
		Render(strings.Join(rows, "\n"))
	return lipgloss.Place(m.w, m.h, lipgloss.Center, lipgloss.Center, box)
}
