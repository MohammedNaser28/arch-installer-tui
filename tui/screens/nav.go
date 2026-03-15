package screens

import (
	"arch-installer/config"
	"errors"

	tea "github.com/charmbracelet/bubbletea"
)

// ToMsg triggers a screen transition
type ToMsg struct{ Screen config.Screen }

// ErrMsg sends an error to the error screen
type ErrMsg struct{ Err error }

func GoTo(s config.Screen) tea.Cmd {
	return func() tea.Msg { return ToMsg{s} }
}

func Fail(msg string) tea.Cmd {
	return func() tea.Msg { return ErrMsg{errors.New(msg)} }
}
