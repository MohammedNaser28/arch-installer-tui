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

type errClearMsg struct{}

type Hostname struct {
	cfg    *config.Config
	input  textinput.Model
	err    string
	width  int
	height int
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
	return Hostname{cfg: cfg, input: t, width: 80, height: 24}
}

func (m Hostname) Init() tea.Cmd { return textinput.Blink }

func (m Hostname) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width  = msg.Width
		m.height = msg.Height
		m.input.Width = m.width / 3
		return m, nil
	case errClearMsg:
		m.err = ""
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			v := strings.TrimSpace(m.input.Value())
			if v == "" {
				m.err = "hostname cannot be empty"
				return m, clearErrAfter(3)
			}
			if !validHost(v) {
				m.err = "only letters, digits, hyphens — no spaces"
				return m, clearErrAfter(3)
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
	return m, cmd
}

func (m Hostname) View() string {
	progress := components.ProgressBar(m.width, 1)
	title    := components.Title.Render("Hostname")
	sub      := components.Subtitle.Render("Name for this machine on the network")

	m.input.Width = m.width / 3
	box := components.ActiveBoxWithWidth(m.width).Render(m.input.View())

	errLine := ""
	if m.err != "" {
		errLine = "\n" + components.Err.Render("  ✗  "+m.err)
	}

	hint := components.Dim.Render("Only lowercase letters, numbers, hyphens (e.g. my-arch)")
	help := components.Help("enter", "next", "esc", "back")

	content := lipgloss.JoinVertical(lipgloss.Left,
		progress, "\n", title, sub, box, errLine, hint, help,
	)

	return components.Page(m.width, m.height, content)
}

func validHost(s string) bool {
	for _, c := range s {
		if !unicode.IsLetter(c) && !unicode.IsDigit(c) && c != '-' {
			return false
		}
	}
	return true
}
