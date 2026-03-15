package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Steps defines the installer progress steps in order
var Steps = []string{
	"Disk", "Hostname", "User", "Packages", "Confirm", "Install", "Done",
}

// ProgressBar renders a filled bar showing current step
// e.g. step=2 out of 7 → [████░░░░░░░░░░]  Step 3 of 7
func ProgressBar(width, currentStep int) string {
	totalSteps := len(Steps)

	// Bar width — leave room for label
	barWidth := width - 20
	if barWidth < 20 {
		barWidth = 20
	}

	filled := int(float64(barWidth) * float64(currentStep) / float64(totalSteps))
	empty := barWidth - filled

	bar := lipgloss.NewStyle().Foreground(Mauve).Render(strings.Repeat("█", filled)) +
		lipgloss.NewStyle().Foreground(Surface).Render(strings.Repeat("░", empty))

	stepLabel := lipgloss.NewStyle().
		Foreground(Subtext).
		Render(fmt.Sprintf("  %s  ·  Step %d of %d",
			Steps[min(currentStep, totalSteps-1)],
			currentStep+1,
			totalSteps,
		))

	barBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Surface).
		Padding(0, 1).
		Width(width - 8).
		Render(bar)

	return lipgloss.JoinVertical(lipgloss.Left, barBox, stepLabel)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
