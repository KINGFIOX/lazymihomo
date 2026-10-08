package panels

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"lazymihomo/internal/api"
	"lazymihomo/internal/format"
	"lazymihomo/internal/ui/components"
	"lazymihomo/internal/ui/theme"
)

type (
	// PxLoadedMsg carries a full proxies snapshot.
	PxLoadedMsg struct {
		Groups  []api.Proxy
		Proxies map[string]api.Proxy
		Err     error
	}
	// PxSelectedMsg reports a node switch result.
	PxSelectedMsg struct {
		Group, Node string
		Err         error
	}
	// PxDelayMsg reports a single-node delay test.
	PxDelayMsg struct {
		Name  string
		Delay uint16
		Err   error
	}
	// PxGroupDelayMsg reports a whole-group delay test.
	PxGroupDelayMsg struct {
		Group string
		Delay map[string]uint16
		Err   error
	}
)

// Proxies is the group/node management panel.
type Proxies struct {
	client *api.Client
	th     *theme.Theme
	w, h   int

	groups []api.Proxy
	byName map[string]api.Proxy

	gList, nList components.List
	focusGroups  bool
	sortByDelay  bool

	err    error
	status string
}

// NewProxies builds the proxies panel.
func NewProxies(c *api.Client, t *theme.Theme) *Proxies {
	return &Proxies{client: c, th: t}
}

// Title implements Panel.
func (p *Proxies) Title() string { return "Proxies" }

// ShortHelp implements Panel.
func (p *Proxies) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("enter", " "), key.WithHelp("enter", "switch node")),
		key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "test node")),
		key.NewBinding(key.WithKeys("t"), key.WithHelp("t", "test group")),
		key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "sort by delay")),
		key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "group/node")),
	}
}

// SetSize implements Panel.
func (p *Proxies) SetSize(w, h int) { p.w, p.h = w, h }

// Refresh implements the loader for the panel.
func (p *Proxies) Refresh() tea.Cmd {
	c := p.client
	return func() tea.Msg {
		ctx := context.Background()
		raw, e1 := c.GetGroups(ctx)
		// Only manually switchable groups make sense here: URLTest and
		// friends pick their node automatically (and reject manual Set),
		// and GLOBAL is mihomo's synthetic catch-all selector.
		groups := make([]api.Proxy, 0, len(raw))
		for _, g := range raw {
			if g.Name != "GLOBAL" && g.Type == "Selector" {
				groups = append(groups, g)
			}
		}
		px, e2 := c.GetAllProxies(ctx)
		err := e1
		if err == nil {
			err = e2
		}
		return PxLoadedMsg{Groups: groups, Proxies: px, Err: err}
	}
}

func (p *Proxies) currentGroup() api.Proxy {
	if p.gList.Cursor < 0 || p.gList.Cursor >= len(p.groups) {
		return api.Proxy{}
	}
	return p.groups[p.gList.Cursor]
}

func (p *Proxies) lookup(name string) api.Proxy {
	if pr, ok := p.byName[name]; ok {
		return pr
	}
	return api.Proxy{Name: name}
}

func (p *Proxies) nodeNames() []string {
	g := p.currentGroup()
	names := make([]string, len(g.All))
	copy(names, g.All)
	if p.sortByDelay {
		sort.SliceStable(names, func(i, j int) bool {
			di, dj := p.lookup(names[i]).LastDelay(), p.lookup(names[j]).LastDelay()
			switch {
			case di == 0:
				return false
			case dj == 0:
				return true
			default:
				return di < dj
			}
		})
	}
	return names
}

func (p *Proxies) currentNodeName() string {
	names := p.nodeNames()
	if p.nList.Cursor < 0 || p.nList.Cursor >= len(names) {
		return ""
	}
	return names[p.nList.Cursor]
}

// Update implements Panel.
func (p *Proxies) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case PxLoadedMsg:
		if msg.Err != nil {
			p.err = msg.Err
			break
		}
		p.err = nil
		p.groups, p.byName = msg.Groups, msg.Proxies
		if len(p.groups) == 0 {
			p.gList.Cursor = 0
		} else if p.gList.Cursor > len(p.groups)-1 {
			p.gList.Cursor = len(p.groups) - 1
		}
		if nn := len(p.nodeNames()); p.nList.Cursor > nn-1 {
			p.nList.Cursor = max(nn-1, 0)
		}
	case PxSelectedMsg:
		if msg.Err != nil {
			p.err = msg.Err
			break
		}
		p.err = nil
		p.status = "switched → " + msg.Node
		for i := range p.groups {
			if p.groups[i].Name == msg.Group {
				p.groups[i].Now = msg.Node
			}
		}
		if pr, ok := p.byName[msg.Group]; ok {
			pr.Now = msg.Node
			p.byName[msg.Group] = pr
		}
		return p.Refresh()
	case PxDelayMsg:
		if msg.Err != nil {
			p.err = msg.Err
			break
		}
		p.err = nil
		p.status = msg.Name + " · " + format.Delay(msg.Delay)
		if pr, ok := p.byName[msg.Name]; ok {
			pr.RecordDelay(msg.Delay)
			p.byName[msg.Name] = pr
		}
	case PxGroupDelayMsg:
		if msg.Err != nil {
			p.err = msg.Err
			break
		}
		p.err = nil
		p.status = fmt.Sprintf("tested %d nodes in %s", len(msg.Delay), msg.Group)
		return p.Refresh()
	case tea.KeyMsg:
		return p.handleKey(msg)
	}
	return nil
}

func (p *Proxies) handleKey(k tea.KeyMsg) tea.Cmd {
	switch k.String() {
	case "j", "down":
		if p.focusGroups {
			p.gList.Down()
		} else {
			p.nList.Down()
		}
	case "k", "up":
		if p.focusGroups {
			p.gList.Up()
		} else {
			p.nList.Up()
		}
	case "g":
		if p.focusGroups {
			p.gList.Top()
		} else {
			p.nList.Top()
		}
	case "G":
		if p.focusGroups {
			p.gList.Bottom()
		} else {
			p.nList.Bottom()
		}
	case "h", "left":
		p.focusGroups = true
	case "l", "right":
		p.focusGroups = false
	case "tab":
		p.focusGroups = !p.focusGroups
	case "enter", " ", "space":
		if p.focusGroups {
			p.focusGroups = false
			break
		}
		return p.selectCmd()
	case "d":
		if p.focusGroups {
			break
		}
		name := p.currentNodeName()
		if name == "" {
			break
		}
		provider := p.lookup(name).ProviderName
		c := p.client
		p.status = "testing " + name + " …"
		return func() tea.Msg {
			d, err := c.ProxyDelay(context.Background(), provider, name)
			return PxDelayMsg{Name: name, Delay: d, Err: err}
		}
	case "t":
		g := p.currentGroup()
		if g.Name == "" {
			break
		}
		c := p.client
		p.status = "testing group " + g.Name + " …"
		return func() tea.Msg {
			d, err := c.GroupDelay(context.Background(), g.Name)
			return PxGroupDelayMsg{Group: g.Name, Delay: d, Err: err}
		}
	case "s":
		p.sortByDelay = !p.sortByDelay
		if p.sortByDelay {
			p.status = "sorted by delay"
		} else {
			p.status = "original order"
		}
	}
	return nil
}

func (p *Proxies) selectCmd() tea.Cmd {
	g := p.currentGroup()
	name := p.currentNodeName()
	if g.Name == "" || name == "" || name == g.Now {
		return nil
	}
	c := p.client
	return func() tea.Msg {
		return PxSelectedMsg{Group: g.Name, Node: name, Err: c.SelectProxy(context.Background(), g.Name, name)}
	}
}

func (p *Proxies) header(focused bool, text string, w int) string {
	if focused {
		return rightPad(truncate(p.th.Primary.Render(text), w), w)
	}
	return rightPad(truncate(p.th.Dim.Render(text), w), w)
}

func (p *Proxies) cursorStyle(focused bool) lipgloss.Style {
	if focused {
		return p.th.Cursor
	}
	return p.th.CursorDim
}

// View implements Panel.
func (p *Proxies) View() string {
	if p.w < 24 {
		return ""
	}
	listH := p.h - 2
	if listH < 1 {
		listH = 1
	}
	gW := min(36, p.w*2/5)
	if gW < 12 {
		gW = 12
	}
	nW := p.w - gW - 3
	if nW < 12 {
		nW = 12
	}

	g := p.currentGroup()
	names := p.nodeNames()

	gLines := make([]string, 0, len(p.groups))
	for _, grp := range p.groups {
		tag := p.th.Dim.Render(fmt.Sprintf("(%s)", grp.Type))
		gLines = append(gLines, truncate(grp.Name, gW-8)+" "+tag)
	}

	nLines := make([]string, 0, len(names))
	nameW := nW - 11
	if nameW < 4 {
		nameW = 4
	}
	for _, name := range names {
		pr := p.lookup(name)
		delay := pr.LastDelay()
		var dstr string
		if len(pr.History) == 0 {
			dstr = p.th.Dim.Render(fmt.Sprintf("%7s", "—"))
		} else {
			dstr = p.th.DelayStyle(delay).Render(fmt.Sprintf("%7s", format.Delay(delay)))
		}
		mark := " "
		if name == g.Now {
			mark = "●"
		}
		nLines = append(nLines, rightPad(truncate(mark+" "+name, nameW), nameW)+" "+dstr)
	}

	headers := lipgloss.JoinHorizontal(lipgloss.Top,
		p.header(p.focusGroups, fmt.Sprintf("GROUPS (%d)", len(p.groups)), gW),
		"   ",
		p.header(!p.focusGroups, fmt.Sprintf("NODES · %s (%s) · %d", g.Name, g.Type, len(names)), nW),
	)

	p.gList.Width, p.gList.Height = gW, listH
	p.nList.Width, p.nList.Height = nW, listH
	p.gList.CursorStyle = p.cursorStyle(p.focusGroups)
	p.nList.CursorStyle = p.cursorStyle(!p.focusGroups)

	divider := p.th.Dim.Render(strings.TrimSuffix(strings.Repeat("│\n", listH), "\n"))

	body := lipgloss.JoinHorizontal(lipgloss.Top,
		p.gList.Render(gLines),
		divider,
		p.nList.Render(nLines),
	)

	return lipgloss.JoinVertical(lipgloss.Left,
		headers,
		body,
		statusLine(p.th, p.err, p.status, p.w),
	)
}
