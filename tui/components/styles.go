package components

import "github.com/charmbracelet/lipgloss"

var (
	Mauve   = lipgloss.Color("#cba6f7")
	Blue    = lipgloss.Color("#89b4fa")
	Green   = lipgloss.Color("#a6e3a1")
	Red     = lipgloss.Color("#f38ba8")
	Yellow  = lipgloss.Color("#f9e2af")
	Text    = lipgloss.Color("#cdd6f4")
	Subtext = lipgloss.Color("#a6adc8")
	Surface = lipgloss.Color("#313244")
)

var (
	Title = lipgloss.NewStyle().
		Foreground(Mauve).Bold(true).MarginBottom(1)

	Subtitle = lipgloss.NewStyle().
			Foreground(Subtext).MarginBottom(1)

	Selected = lipgloss.NewStyle().
			Foreground(Blue).Bold(true)

	Normal = lipgloss.NewStyle().
		Foreground(Text)

	Dim = lipgloss.NewStyle().
		Foreground(Subtext)

	Success = lipgloss.NewStyle().
		Foreground(Green).Bold(true)

	Err = lipgloss.NewStyle().
		Foreground(Red).Bold(true)

	Warn = lipgloss.NewStyle().
		Foreground(Yellow)

	Box = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Surface).
		Padding(1, 2)

	ActiveBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Mauve).
			Padding(1, 2)
)

// Help renders key hint bar at bottom of every screen
func Help(pairs ...string) string {
	out := ""
	for i := 0; i+1 < len(pairs); i += 2 {
		if i > 0 {
			out += Dim.Render("  ·  ")
		}
		out += lipgloss.NewStyle().Foreground(Blue).Render(pairs[i]) +
			" " + Dim.Render(pairs[i+1])
	}
	return lipgloss.NewStyle().MarginTop(1).Render(out)
}
