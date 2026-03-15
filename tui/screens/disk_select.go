package screens

import (
	"arch-installer/config"
	"arch-installer/tui/components"
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Disk struct {
	Device string
	Size   string
	Model  string
}

type disksLoaded struct{ disks []Disk }

type DiskSelect struct {
	cfg    *config.Config
	disks  []Disk
	cursor int
	ready  bool
}

func NewDiskSelect(cfg *config.Config) DiskSelect { return DiskSelect{cfg: cfg} }
func (m DiskSelect) Init() tea.Cmd                { return fetchDisks() }

func fetchDisks() tea.Cmd {
	return func() tea.Msg {
		out, err := exec.Command("lsblk", "-d", "-o", "NAME,SIZE,MODEL", "--noheadings").Output()
		if err != nil {
			// dry-run fallback
			return disksLoaded{disks: []Disk{
				{"/dev/sda", "256G", "MOCK DISK (dry-run)"},
				{"/dev/sdb", "128G", "MOCK USB (dry-run)"},
			}}
		}
		var disks []Disk
		scanner := bufio.NewScanner(bytes.NewReader(out))
		for scanner.Scan() {
			f := strings.Fields(scanner.Text())
			if len(f) < 1 {
				continue
			}
			d := Disk{Device: "/dev/" + f[0]}
			if len(f) >= 2 {
				d.Size = f[1]
			}
			if len(f) >= 3 {
				d.Model = strings.Join(f[2:], " ")
			}
			disks = append(disks, d)
		}
		return disksLoaded{disks: disks}
	}
}

func (m DiskSelect) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case disksLoaded:
		m.disks = msg.disks
		m.ready = true
		// auto-detect boot mode
		if _, err := os.Stat("/sys/firmware/efi"); err == nil {
			m.cfg.BootMode = "uefi"
		} else {
			m.cfg.BootMode = "bios"
		}
		return m, nil
	case tea.KeyMsg:
		if !m.ready {
			return m, nil
		}
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.disks)-1 {
				m.cursor++
			}
		case "enter":
			if len(m.disks) == 0 {
				return m, nil
			}
			d := m.disks[m.cursor]
			m.cfg.TargetDisk = d.Device
			// set partition names based on device type
			m.cfg.EFIPartition = partName(d.Device, 1)
			m.cfg.RootPartition = partName(d.Device, 2)
			return m, GoTo(config.ScreenHostname)
		case "esc":
			return m, GoTo(config.ScreenWelcome)
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m DiskSelect) View() string {
	title := components.Title.Render("Select Disk")
	sub := components.Subtitle.Render("Auto-partition: 512MB EFI (if UEFI) + rest as root")

	var rows string
	if !m.ready {
		rows = components.Dim.Render("  Scanning...")
	} else {
		for i, d := range m.disks {
			model := d.Model
			if model == "" {
				model = "Unknown"
			}
			line := fmt.Sprintf("%-16s %-8s %s", d.Device, d.Size, model)
			if i == m.cursor {
				rows += components.Selected.Render("❯ "+line) + "\n"
			} else {
				rows += components.Normal.Render("  "+line) + "\n"
			}
		}
	}

	box := components.Box.Render(rows)

	warn := components.Warn.Render("⚠  All data on selected disk will be destroyed")
	help := components.Help("↑/↓", "navigate", "enter", "select", "esc", "back")

	return lipgloss.NewStyle().Padding(2, 4).Render(
		lipgloss.JoinVertical(lipgloss.Left, title, sub, box, warn, help),
	)
}

// partName handles sda→sda1 and nvme0n1→nvme0n1p1
func partName(disk string, n int) string {
	if strings.Contains(disk, "nvme") || strings.Contains(disk, "mmcblk") {
		return fmt.Sprintf("%sp%d", disk, n)
	}
	return fmt.Sprintf("%s%d", disk, n)
}
