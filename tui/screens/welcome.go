package screens

import (
	"arch-installer/config"
	"arch-installer/tui/components"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Welcome struct{ cfg *config.Config }

func NewWelcome(cfg *config.Config) Welcome { return Welcome{cfg} }
func (m Welcome) Init() tea.Cmd             { return nil }

func (m Welcome) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "enter", " ":
			return m, GoTo(config.ScreenDiskSelect)
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Welcome) View() string {
	logo := lipgloss.NewStyle().Foreground(components.Mauve).Bold(true).Render(`
   █████╗ ██████╗  ██████╗██╗  ██╗
  ██╔══██╗██╔══██╗██╔════╝██║  ██║
  ███████║██████╔╝██║     ███████║
  ██╔══██║██╔══██╗██║     ██╔══██║
  ██║  ██║██║  ██║╚██████╗██║  ██║
  ╚═╝  ╚═╝╚═╝  ╚═╝ ╚═════╝╚═╝  ╚═╝`)

	sub := components.Subtitle.Render("  Linux Committee — Arch Installer")

	warn := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(components.Yellow).
		Foreground(components.Yellow).
		Padding(0, 2).
		Render("⚠  This will erase the selected disk entirely")

	help := components.Help("enter", "start", "q", "quit")

	return lipgloss.NewStyle().Padding(2, 4).Render(
		lipgloss.JoinVertical(lipgloss.Left, logo, sub, "\n", warn, help),
	)
}
