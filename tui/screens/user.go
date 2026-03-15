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
	return User{cfg: cfg, inputs: inputs}
}

func (m User) Init() tea.Cmd { return textinput.Blink }

func (m User) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
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
	m.err = ""
	return m, cmd
}

func (m *User) submit() tea.Cmd {
	u := strings.TrimSpace(m.inputs[fUser].Value())
	p := m.inputs[fPass].Value()
	r := m.inputs[fRoot].Value()

	switch {
	case u == "":
		m.err = "username cannot be empty"
	case strings.ContainsAny(u, " /"):
		m.err = "username cannot have spaces or slashes"
	case len(p) < 6:
		m.err = "password must be at least 6 characters"
	case len(r) < 6:
		m.err = "root password must be at least 6 characters"
	default:
		m.cfg.Username = u
		m.cfg.UserPassword = p
		m.cfg.RootPassword = r
		m.err = ""
		return GoTo(config.ScreenConfirm)
	}
	return nil
}

func (m User) View() string {
	title := components.Title.Render("User Account")
	sub := components.Subtitle.Render("Create your user — user gets sudo access")

	labels := [fTotal]string{"Username", "Password", "Root Password"}
	var form string
	for i, inp := range m.inputs {
		label := components.Dim.Render(labels[i])
		var box string
		if i == m.focus {
			box = components.ActiveBox.Render(inp.View())
		} else {
			box = components.Box.Render(inp.View())
		}
		form += label + "\n" + box + "\n\n"
	}

	errLine := ""
	if m.err != "" {
		errLine = components.Err.Render("  " + m.err)
	}

	help := components.Help("tab", "next field", "enter", "confirm", "esc", "back")

	return lipgloss.NewStyle().Padding(2, 4).Render(
		lipgloss.JoinVertical(lipgloss.Left, title, sub, form, errLine, help),
	)
}
