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
	width  int
	height int
}

func NewDiskSelect(cfg *config.Config) DiskSelect {
	return DiskSelect{cfg: cfg, width: 80, height: 24}
}

func (m DiskSelect) Init() tea.Cmd { return fetchDisks() }

func fetchDisks() tea.Cmd {
	return func() tea.Msg {
		out, err := exec.Command("lsblk", "-d", "-o", "NAME,SIZE,MODEL", "--noheadings").Output()
		if err != nil {
			return disksLoaded{disks: []Disk{}}
		}
		var disks []Disk
		scanner := bufio.NewScanner(bytes.NewReader(out))
		for scanner.Scan() {
			f := strings.Fields(scanner.Text())
			if len(f) < 1 {
				continue
			}
			d := Disk{Device: "/dev/" + f[0]}
			if len(f) >= 2 { d.Size = f[1] }
			if len(f) >= 3 { d.Model = strings.Join(f[2:], " ") }
			disks = append(disks, d)
		}
		return disksLoaded{disks: disks}
	}
}

func (m DiskSelect) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width  = msg.Width
		m.height = msg.Height

	case disksLoaded:
		m.disks = msg.disks
		m.ready = true
		if _, err := os.Stat("/sys/firmware/efi"); err == nil {
			m.cfg.BootMode = "uefi"
		} else {
			m.cfg.BootMode = "bios"
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, GoTo(config.ScreenWelcome)
		case "ctrl+c", "q":
			return m, tea.Quit
		}
		// Only handle navigation if ready and has disks
		if !m.ready || len(m.disks) == 0 {
			return m, nil
		}
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 { m.cursor-- }
		case "down", "j":
			if m.cursor < len(m.disks)-1 { m.cursor++ }
		case "enter":
			d := m.disks[m.cursor]
			m.cfg.TargetDisk    = d.Device
			m.cfg.EFIPartition  = partName(d.Device, 1)
			m.cfg.RootPartition = partName(d.Device, 2)
			return m, GoTo(config.ScreenPackages)
		}
	}
	return m, nil
}

func (m DiskSelect) View() string {
	progress := components.ProgressBar(m.width, 0)
	title    := components.Title.Render("Select Disk")
	sub      := components.Subtitle.Render("Auto-partition: 512MB EFI (UEFI) + rest as root (ext4)")

	var rows string
	if !m.ready {
		rows = components.Dim.Render("  Scanning disks...")
	} else if len(m.disks) == 0 {
		rows = components.Err.Render("  No disks found.") + "\n" +
			components.Dim.Render("  Attach a disk and restart the installer.")
	} else {
		for i, d := range m.disks {
			model := d.Model
			if model == "" { model = "Unknown" }
			line := fmt.Sprintf("%-16s %-8s %s", d.Device, d.Size, model)
			if i == m.cursor {
				rows += components.Selected.Render("❯ "+line) + "\n"
			} else {
				rows += components.Normal.Render("  "+line) + "\n"
			}
		}
	}

	box  := components.BoxWithWidth(m.width).Render(rows)
	warn := components.Warn.Render("⚠  All data on selected disk will be destroyed")
	help := components.Help("↑/↓", "navigate", "enter", "select", "esc", "back")

	content := lipgloss.JoinVertical(lipgloss.Left,
		progress, "\n", title, sub, box, warn, help,
	)

	return components.Page(m.width, m.height, content)
}

func partName(disk string, n int) string {
	if strings.Contains(disk, "nvme") || strings.Contains(disk, "mmcblk") {
		return fmt.Sprintf("%sp%d", disk, n)
	}
	return fmt.Sprintf("%s%d", disk, n)
}
