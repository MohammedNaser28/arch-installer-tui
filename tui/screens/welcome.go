package screens

import (
	"arch-installer/config"
	"arch-installer/tui/components"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Welcome struct {
	cfg    *config.Config
	width  int
	height int
}

func NewWelcome(cfg *config.Config) Welcome {
	return Welcome{cfg: cfg, width: 80, height: 24}
}

func (m Welcome) Init() tea.Cmd { return nil }

func (m Welcome) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width  = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "enter", " ":
			return m, GoTo(config.ScreenNetwork)
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Welcome) View() string {
	logo := lipgloss.NewStyle().
		Foreground(components.Mauve).Bold(true).
		Render(
			"░█████╗░░██████╗░█████╗░\n" +
			"██╔══██╗██╔════╝██╔══██╗\n" +
			"██║░░██║╚█████╗░██║░░╚═╝\n" +
			"██║░░██║░╚═══██╗██║░░██╗\n" +
			"╚█████╔╝██████╔╝╚█████╔╝\n" +
			"░╚════╝░╚═════╝░░╚════╝░",
		)

	title := lipgloss.NewStyle().Foreground(components.Blue).Bold(true).
		Render("Open Source Community")

	sub  := components.Subtitle.Render("Linux Committee — Arch Installer")
	div  := components.Dim.Render("────────────────────────────")
	warn := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(components.Yellow).
		Foreground(components.Yellow).
		Padding(0, 2).
		Render("⚠  This will erase the selected disk entirely")
	help := components.Help("enter", "start", "q", "quit")

	content := lipgloss.JoinVertical(lipgloss.Center,
		logo, "\n", title, sub, div, "\n", warn, "\n", help,
	)

	return components.Page(m.width, m.height, content)
}
