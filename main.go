package main

import (
	"arch-installer/config"
	"arch-installer/tui"
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	dry := flag.Bool("dry-run", false, "simulate install without touching disk")
	flag.Parse()

	cfg := config.New()
	cfg.DryRun = *dry

	if cfg.DryRun {
		fmt.Fprintln(os.Stderr, "DRY-RUN: no changes will be made")
	}

	p := tea.NewProgram(
		tui.New(cfg),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
