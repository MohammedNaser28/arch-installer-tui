package screens

import (
	"arch-installer/config"
	"arch-installer/tui/components"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	fUser = iota
	fPass
	fRoot
	fTotal
)

type User struct {
	cfg    *config.Config
	inputs [fTotal]textinput.Model
	focus  int
	err    string
	width  int
	height int
}

func NewUser(cfg *config.Config) User {
	labels := [fTotal]string{"Username", "Password", "Root Password"}
	var inputs [fTotal]textinput.Model
	for i := range inputs {
		t := textinput.New()
		t.Placeholder = labels[i]
		t.Width = 32
		t.CharLimit = 64
		if i > 0 {
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
		}
		inputs[i] = t
	}
	inputs[fUser].Focus()
	if cfg.Username != "" {
		inputs[fUser].SetValue(cfg.Username)
	}
	return User{cfg: cfg, inputs: inputs, width: 80, height: 24}
}

func (m User) Init() tea.Cmd { return textinput.Blink }

func (m User) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		for i := range m.inputs {
			m.inputs[i].Width = m.width / 3
		}
		return m, nil

	case errClearMsg:
		m.err = ""
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "down":
			m.inputs[m.focus].Blur()
			m.focus = (m.focus + 1) % fTotal
			m.inputs[m.focus].Focus()
			return m, textinput.Blink
		case "shift+tab", "up":
			m.inputs[m.focus].Blur()
			m.focus = (m.focus - 1 + fTotal) % fTotal
			m.inputs[m.focus].Focus()
			return m, textinput.Blink
		case "enter":
			if m.focus < fTotal-1 {
				m.inputs[m.focus].Blur()
				m.focus++
				m.inputs[m.focus].Focus()
				return m, textinput.Blink
			}
			return m, m.submit()
		case "esc":
			return m, GoTo(config.ScreenHostname)
		case "ctrl+c":
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.inputs[m.focus], cmd = m.inputs[m.focus].Update(msg)
	return m, cmd
}

func (m *User) submit() tea.Cmd {
	u := strings.TrimSpace(m.inputs[fUser].Value())
	p := m.inputs[fPass].Value()
	r := m.inputs[fRoot].Value()

	switch {
	case u == "":
		m.err = "username cannot be empty"
		return clearErrAfter(4)
	case strings.ContainsAny(u, " /"):
		m.err = "username cannot have spaces or slashes"
		return clearErrAfter(4)
	case len(p) < 6:
		m.err = "password must be at least 6 characters"
		return clearErrAfter(4)
	case len(r) < 6:
		m.err = "root password must be at least 6 characters"
		return clearErrAfter(4)
	default:
		m.cfg.Username = u
		m.cfg.UserPassword = p
		m.cfg.RootPassword = r
		m.err = ""
		return GoTo(config.ScreenPackages)
	}
}

func (m User) View() string {
	progress := components.ProgressBar(m.width, 2)
	title := components.Title.Render("User Account")
	sub := components.Subtitle.Render("Create your user — gets sudo access automatically")

	labels := [fTotal]string{"Username", "Password", "Root Password"}
	var form string
	for i, inp := range m.inputs {
		inp.Width = m.width / 3
		label := components.Dim.Render(labels[i])
		var box string
		if i == m.focus {
			box = components.ActiveBoxWithWidth(m.width).Render(inp.View())
		} else {
			box = components.BoxWithWidth(m.width).Render(inp.View())
		}
		form += label + "\n" + box + "\n\n"
	}

	errLine := ""
	if m.err != "" {
		errLine = components.Err.Render("  ✗  " + m.err)
	}

	help := components.Help("tab", "next field", "enter", "confirm", "esc", "back")

	content := lipgloss.JoinVertical(lipgloss.Left,
		title, sub, form, errLine, help, progress,
	)

	return components.Page(m.width, m.height, content)
}
