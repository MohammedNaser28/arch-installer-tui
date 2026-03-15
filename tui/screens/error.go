package screens

import (
	"arch-installer/config"
	"arch-installer/tui/components"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Error struct{ cfg *config.Config }

func NewError(cfg *config.Config) Error { return Error{cfg} }
func (m Error) Init() tea.Cmd           { return nil }

func (m Error) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok {
		if k.String() == "q" || k.String() == "enter" || k.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Error) View() string {
	title := components.Err.Render("Installation Failed")

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(components.Red).
		Padding(1, 2).
		Render(components.Normal.Render(m.cfg.LastErr))

	hint := components.Dim.Render("Log saved to: " + m.cfg.LogPath)
	help := components.Help("q", "quit")

	return lipgloss.NewStyle().Padding(2, 4).Render(
		lipgloss.JoinVertical(lipgloss.Left, title, "\n", box, "\n", hint, help),
	)
}
