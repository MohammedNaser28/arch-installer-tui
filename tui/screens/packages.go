package screens

import (
	"arch-installer/config"
	"arch-installer/tui/components"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Package struct {
	Name        string
	PkgNames    []string
	Selected    bool
	Required    bool
	IsJetbrains bool
	JBName      string
}

type Group struct {
	Title    string
	Packages []Package
}

type Packages struct {
	cfg      *config.Config
	groups   []Group
	groupIdx int
	pkgIdx   int
	width    int
	height   int
	viewport viewport.Model
	ready    bool
}

func NewPackages(cfg *config.Config) Packages {
	return Packages{
		cfg:    cfg,
		width:  80,
		height: 24,
		groups: []Group{
			{
				Title: "Desktop",
				Packages: []Package{
					{Name: "Hyprland + Waybar", PkgNames: []string{"hyprland", "waybar", "hyprpaper", "dunst", "wofi"}, Selected: true, Required: true},
					{Name: "Kitty Terminal", PkgNames: []string{"kitty"}, Selected: true, Required: true},
					{Name: "Fish + Starship", PkgNames: []string{"fish", "starship"}, Selected: true, Required: true},
					{Name: "Thunar (File Manager)", PkgNames: []string{"thunar", "thunar-archive-plugin", "gvfs"}, Selected: true},
					{Name: "Nerd Fonts", PkgNames: []string{"ttf-jetbrains-mono-nerd", "noto-fonts", "noto-fonts-emoji"}, Selected: true, Required: true},
				},
			},
			{
				Title: "Apps",
				Packages: []Package{
					{Name: "Firefox", PkgNames: []string{"firefox"}, Selected: true},
					{Name: "VS Code", PkgNames: []string{"code"}, Selected: false},
					{Name: "Neovim", PkgNames: []string{"neovim"}, Selected: false},
					{Name: "VLC", PkgNames: []string{"vlc"}, Selected: false},
					{Name: "Discord", PkgNames: []string{"discord"}, Selected: false},
				},
			},
			{
				Title: "Programming Languages",
				Packages: []Package{
					{Name: "Go", PkgNames: []string{"go"}, Selected: false},
					{Name: "Python", PkgNames: []string{"python", "python-pip"}, Selected: false},
					{Name: "Node.js + npm", PkgNames: []string{"nodejs", "npm"}, Selected: false},
					{Name: "Rust", PkgNames: []string{"rust"}, Selected: false},
					{Name: "Java (JDK 21)", PkgNames: []string{"jdk21-openjdk"}, Selected: false},
					{Name: "C/C++ (GCC + CMake)", PkgNames: []string{"gcc", "cmake", "make"}, Selected: false},
					{Name: "Clang + LLVM", PkgNames: []string{"clang", "llvm"}, Selected: false},
				},
			},
			{
				Title: "JetBrains IDEs  (downloaded during install)",
				Packages: []Package{
					{Name: "IntelliJ IDEA", IsJetbrains: true, JBName: "IntelliJ IDEA"},
					{Name: "PyCharm", IsJetbrains: true, JBName: "PyCharm"},
					{Name: "GoLand", IsJetbrains: true, JBName: "GoLand"},
					{Name: "CLion", IsJetbrains: true, JBName: "CLion"},
					{Name: "WebStorm", IsJetbrains: true, JBName: "WebStorm"},
					{Name: "Rider", IsJetbrains: true, JBName: "Rider"},
				},
			},
			{
				Title: "Tools",
				Packages: []Package{
					{Name: "Git", PkgNames: []string{"git"}, Selected: true, Required: true},
					{Name: "Docker + Compose", PkgNames: []string{"docker", "docker-compose"}},
					{Name: "Fastfetch", PkgNames: []string{"fastfetch"}, Selected: true},
					{Name: "btop", PkgNames: []string{"btop"}},
					{Name: "ripgrep + fd + bat", PkgNames: []string{"ripgrep", "fd", "bat"}},
					{Name: "curl + wget", PkgNames: []string{"curl", "wget"}, Selected: true, Required: true},
				},
			},
		},
	}
}

func (m Packages) Init() tea.Cmd { return nil }

func (m Packages) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width  = msg.Width
		m.height = msg.Height
		// header = progress(3) + title(1) + sub(1) + footer(3) = ~8 lines
		vpHeight := m.height - 14
		if vpHeight < 5 {
			vpHeight = 5
		}
		if !m.ready {
			m.viewport = viewport.New(m.width-8, vpHeight)
			m.ready = true
		} else {
			m.viewport.Width  = m.width - 8
			m.viewport.Height = vpHeight
		}
		m.viewport.SetContent(m.buildList())
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.pkgIdx > 0 {
				m.pkgIdx--
			} else if m.groupIdx > 0 {
				m.groupIdx--
				m.pkgIdx = len(m.groups[m.groupIdx].Packages) - 1
			}
			m.viewport.SetContent(m.buildList())
			m.scrollToCursor()

		case "down", "j":
			if m.pkgIdx < len(m.groups[m.groupIdx].Packages)-1 {
				m.pkgIdx++
			} else if m.groupIdx < len(m.groups)-1 {
				m.groupIdx++
				m.pkgIdx = 0
			}
			m.viewport.SetContent(m.buildList())
			m.scrollToCursor()

		case " ":
			pkg := &m.groups[m.groupIdx].Packages[m.pkgIdx]
			if !pkg.Required {
				pkg.Selected = !pkg.Selected
			}
			m.viewport.SetContent(m.buildList())

		case "enter":
			m.collect()
			return m, GoTo(config.ScreenConfirm)

		case "esc":
			return m, GoTo(config.ScreenUser)

		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// buildList renders all groups and packages into a string for the viewport
func (m Packages) buildList() string {
	var body strings.Builder
	for gi, group := range m.groups {
		body.WriteString(
			lipgloss.NewStyle().
				Foreground(components.Mauve).Bold(true).
				Render("  "+group.Title) + "\n",
		)
		for pi, pkg := range group.Packages {
			active := gi == m.groupIdx && pi == m.pkgIdx

			var check string
			switch {
			case pkg.Required:
				check = lipgloss.NewStyle().Foreground(components.Blue).Render("[✓]")
			case pkg.Selected:
				check = lipgloss.NewStyle().Foreground(components.Green).Render("[✓]")
			default:
				check = components.Dim.Render("[ ]")
			}

			label := pkg.Name
			if pkg.Required {
				label += components.Dim.Render(" (required)")
			}
			if pkg.IsJetbrains && pkg.Selected {
				label += components.Warn.Render(" ~1GB")
			}

			var line string
			if active {
				line = components.Selected.Render("❯ ") + check + " " + components.Selected.Render(label)
			} else {
				line = "  " + check + " " + components.Normal.Render(label)
			}
			body.WriteString(line + "\n")
		}
		body.WriteString("\n")
	}
	return body.String()
}

// scrollToCursor keeps the selected item visible in the viewport
func (m *Packages) scrollToCursor() {
	line := 0
	for gi, g := range m.groups {
		line++ // group header
		for pi := range g.Packages {
			if gi == m.groupIdx && pi == m.pkgIdx {
				// scroll so cursor is visible
				if line < m.viewport.YOffset {
					m.viewport.YOffset = line
				}
				if line >= m.viewport.YOffset+m.viewport.Height {
					m.viewport.YOffset = line - m.viewport.Height + 1
				}
				return
			}
			line++
		}
		line++ // blank line after group
	}
}

func (m *Packages) collect() {
	var pkgs []string
	var jbIDEs []string
	seen := map[string]bool{}
	for _, g := range m.groups {
		for _, p := range g.Packages {
			if !p.Selected {
				continue
			}
			if p.IsJetbrains {
				jbIDEs = append(jbIDEs, p.JBName)
				continue
			}
			for _, name := range p.PkgNames {
				if !seen[name] {
					pkgs = append(pkgs, name)
					seen[name] = true
				}
			}
		}
	}
	m.cfg.ExtraPackages = pkgs
	m.cfg.JetbrainsIDEs = jbIDEs
}

func (m Packages) View() string {
	if !m.ready {
		return components.Page(m.width, m.height,
			components.Dim.Render("Loading..."),
		)
	}

	progress := components.ProgressBar(m.width, 3)
	title    := components.Title.Render("Select Packages")
	sub      := components.Subtitle.Render("space = toggle  ·  enter = confirm  ·  locked = required")

	// Count selected
	total, selected := 0, 0
	for _, g := range m.groups {
		for _, p := range g.Packages {
			total++
			if p.Selected {
				selected++
			}
		}
	}

	jbCount := 0
	for _, p := range m.groups[3].Packages {
		if p.Selected {
			jbCount++
		}
	}
	var jbWarn string
	if jbCount > 0 {
		jbWarn = components.Warn.Render(fmt.Sprintf(
			"  ⚠  %d JetBrains IDE(s) — ~%dGB download required", jbCount, jbCount,
		))
	}

	count := components.Dim.Render(fmt.Sprintf("  %d / %d selected", selected, total))
	scroll := components.Dim.Render(fmt.Sprintf("  %d%%", int(m.viewport.ScrollPercent()*100)))
	help   := components.Help("↑/↓", "navigate", "space", "toggle", "enter", "confirm", "esc", "back")

	footer := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Left, count, "   ", scroll),
		jbWarn,
		help,
	)

	content := lipgloss.JoinVertical(lipgloss.Left,
		progress, "\n",
		title, sub,
		m.viewport.View(),
		footer,
	)

	return components.Page(m.width, m.height, content)
}
