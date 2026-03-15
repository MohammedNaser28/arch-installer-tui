package screens

import (
	"arch-installer/config"
	"errors"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type ToMsg struct{ Screen config.Screen }
type ErrMsg struct{ Err error }

func GoTo(s config.Screen) tea.Cmd {
	return func() tea.Msg { return ToMsg{s} }
}

func Fail(msg string) tea.Cmd {
	return func() tea.Msg { return ErrMsg{errors.New(msg)} }
}

// clearErrAfter sends errClearMsg after n seconds
// so error messages stay visible long enough to read
func clearErrAfter(seconds int) tea.Cmd {
	return tea.Tick(
		time.Duration(seconds)*time.Second,
		func(t time.Time) tea.Msg { return errClearMsg{} },
	)
}