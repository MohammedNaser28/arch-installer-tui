package screens

import (
	"arch-installer/config"
	"arch-installer/tui/components"
	"fmt"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Done struct{ cfg *config.Config }

func NewDone(cfg *config.Config) Done { return Done{cfg} }
func (m Done) Init() tea.Cmd         { return nil }

func (m Done) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "r":
			return m, tea.ExecProcess(exec.Command("reboot"), nil)
		case "q", "enter":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Done) View() string {
	banner := lipgloss.NewStyle().Foreground(components.Green).Bold(true).Render(`
  ██████╗  ██████╗ ███╗   ██╗███████╗
  ██╔══██╗██╔═══██╗████╗  ██║██╔════╝
  ██║  ██║██║   ██║██╔██╗ ██║█████╗
  ██║  ██║██║   ██║██║╚██╗██║██╔══╝
  ██████╔╝╚██████╔╝██║ ╚████║███████╗
  ╚═════╝  ╚═════╝ ╚═╝  ╚═══╝╚══════╝`)

	info := components.Normal.Render(fmt.Sprintf(
		"\n  Installed on %s\n  User: %s   Hostname: %s\n",
		m.cfg.TargetDisk, m.cfg.Username, m.cfg.Hostname,
	))

	help := components.Help("r", "reboot", "q", "back to live session")

	return lipgloss.NewStyle().Padding(2, 4).Render(
		lipgloss.JoinVertical(lipgloss.Center, banner, info, help),
	)
}
