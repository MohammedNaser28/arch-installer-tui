package screens

import (
	"arch-installer/config"
	"arch-installer/tui/components"
	"strings"
	"unicode"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Hostname struct {
	cfg   *config.Config
	input textinput.Model
	err   string
}

func NewHostname(cfg *config.Config) Hostname {
	t := textinput.New()
	t.Placeholder = "archlinux"
	t.Width = 32
	t.CharLimit = 63
	t.Focus()
	if cfg.Hostname != "" {
		t.SetValue(cfg.Hostname)
	}
	return Hostname{cfg: cfg, input: t}
}

func (m Hostname) Init() tea.Cmd { return textinput.Blink }

func (m Hostname) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			v := strings.TrimSpace(m.input.Value())
			if v == "" {
				m.err = "hostname cannot be empty"
				return m, nil
			}
			if !validHost(v) {
				m.err = "only letters, digits, hyphens — no spaces"
				return m, nil
			}
			m.cfg.Hostname = v
			return m, GoTo(config.ScreenUser)
		case "esc":
			return m, GoTo(config.ScreenDiskSelect)
		case "ctrl+c":
			return m, tea.Quit
		}
	}
	m.input, cmd = m.input.Update(msg)
	m.err = ""
	return m, cmd
}

func (m Hostname) View() string {
	title := components.Title.Render("Hostname")
	sub := components.Subtitle.Render("Name for this machine on the network")

	box := components.ActiveBox.Render(m.input.View())

	errLine := ""
	if m.err != "" {
		errLine = "\n" + components.Err.Render("  "+m.err)
	}

	help := components.Help("enter", "next", "esc", "back")

	return lipgloss.NewStyle().Padding(2, 4).Render(
		lipgloss.JoinVertical(lipgloss.Left, title, sub, box, errLine, help),
	)
}

func validHost(s string) bool {
	for _, c := range s {
		if !unicode.IsLetter(c) && !unicode.IsDigit(c) && c != '-' {
			return false
		}
	}
	return true
}
