package screens

import (
	"arch-installer/config"
	"arch-installer/installer"
	"arch-installer/tui/components"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type stepDone struct {
	idx int
	err error
}

type Install struct {
	cfg     *config.Config
	steps   []installer.Step
	current int
	log     []string
	done    bool
	spin    spinner.Model
}

func NewInstall(cfg *config.Config) Install {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(components.Mauve)
	return Install{
		cfg:   cfg,
		steps: installer.Pipeline(cfg),
		spin:  s,
	}
}

func (m Install) Init() tea.Cmd {
	return tea.Batch(m.spin.Tick, runStep(m.steps, 0, m.cfg))
}

func runStep(steps []installer.Step, idx int, cfg *config.Config) tea.Cmd {
	if idx >= len(steps) {
		return nil
	}
	return func() tea.Msg {
		err := steps[idx].Run(cfg)
		return stepDone{idx: idx, err: err}
	}
}

func (m Install) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case stepDone:
		name := m.steps[msg.idx].Name
		if msg.err != nil {
			m.log = append(m.log, "✗ "+name+": "+msg.err.Error())
			m.done = true
			m.cfg.LastErr = msg.err.Error()
			return m, GoTo(config.ScreenError)
		}
		m.log = append(m.log, "✓ "+name)
		m.current = msg.idx + 1
		if m.current >= len(m.steps) {
			m.done = true
			return m, GoTo(config.ScreenDone)
		}
		return m, runStep(m.steps, m.current, m.cfg)

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spin, cmd = m.spin.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m Install) View() string {
	title := components.Title.Render("Installing")
	sub := components.Subtitle.Render("Do not power off")

	var rows strings.Builder
	for i, s := range m.steps {
		var icon string
		switch {
		case i < m.current:
			icon = components.Success.Render("✓")
		case i == m.current && !m.done:
			icon = m.spin.View()
		default:
			icon = components.Dim.Render("○")
		}
		rows.WriteString(fmt.Sprintf(" %s  %s\n", icon, s.Name))
	}

	box := components.Box.Render(rows.String())

	tail := m.log
	if len(tail) > 3 {
		tail = tail[len(tail)-3:]
	}
	var logStr string
	for _, l := range tail {
		logStr += components.Dim.Render("  "+l) + "\n"
	}

	progress := components.Dim.Render(
		fmt.Sprintf("Step %d / %d", clamp(m.current+1, len(m.steps)), len(m.steps)),
	)

	return lipgloss.NewStyle().Padding(2, 4).Render(
		lipgloss.JoinVertical(lipgloss.Left, title, sub, box, progress, "\n", logStr),
	)
}

func clamp(v, max int) int {
	if v > max {
		return max
	}
	return v
}
