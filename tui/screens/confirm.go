package screens

import (
	"arch-installer/config"
	"arch-installer/tui/components"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Confirm struct{ cfg *config.Config }

func NewConfirm(cfg *config.Config) Confirm { return Confirm{cfg} }
func (m Confirm) Init() tea.Cmd             { return nil }

func (m Confirm) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "enter":
			return m, GoTo(config.ScreenInstall)
		case "esc":
			return m, GoTo(config.ScreenUser)
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Confirm) View() string {
	title := components.Title.Render("Confirm")
	sub := components.Subtitle.Render("Last chance — press enter to begin")

	row := func(k, v string) string {
		return fmt.Sprintf("  %-18s %s\n",
			components.Dim.Render(k),
			components.Selected.Render(v),
		)
	}

	efi := "none (BIOS)"
	if m.cfg.BootMode == "uefi" {
		efi = m.cfg.EFIPartition
	}

	summary := strings.Join([]string{
		row("Disk", m.cfg.TargetDisk),
		row("Boot mode", strings.ToUpper(m.cfg.BootMode)),
		row("EFI partition", efi),
		row("Root partition", m.cfg.RootPartition),
		row("Hostname", m.cfg.Hostname),
		row("Username", m.cfg.Username),
		row("Bootloader", "GRUB"),
		row("Filesystem", "ext4"),
	}, "")

	box := components.ActiveBox.Render(summary)

	warn := components.Err.Render(
		fmt.Sprintf("⚠  %s will be completely erased", m.cfg.TargetDisk),
	)

	help := components.Help("enter", "install", "esc", "back", "q", "quit")

	return lipgloss.NewStyle().Padding(2, 4).Render(
		lipgloss.JoinVertical(lipgloss.Left, title, sub, box, "\n", warn, help),
	)
}
