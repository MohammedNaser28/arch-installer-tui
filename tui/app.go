package tui

import (
	"arch-installer/config"
	"arch-installer/tui/screens"

	tea "github.com/charmbracelet/bubbletea"
)

type App struct {
	cfg     *config.Config
	current config.Screen
	models  map[config.Screen]tea.Model
}

func New(cfg *config.Config) App {
	a := App{
		cfg:     cfg,
		current: config.ScreenWelcome,
		models:  make(map[config.Screen]tea.Model),
	}
	a.models[config.ScreenWelcome] = screens.NewWelcome(cfg)
	a.models[config.ScreenDiskSelect] = screens.NewDiskSelect(cfg)
	a.models[config.ScreenHostname] = screens.NewHostname(cfg)
	a.models[config.ScreenUser] = screens.NewUser(cfg)
	a.models[config.ScreenConfirm] = screens.NewConfirm(cfg)
	a.models[config.ScreenInstall] = screens.NewInstall(cfg)
	a.models[config.ScreenDone] = screens.NewDone(cfg)
	a.models[config.ScreenError] = screens.NewError(cfg)
	return a
}

func (a App) Init() tea.Cmd {
	return a.models[a.current].Init()
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Global quit
	if k, ok := msg.(tea.KeyMsg); ok && k.String() == "ctrl+c" {
		return a, tea.Quit
	}

	// Screen transitions
	if t, ok := msg.(screens.ToMsg); ok {
		a.current = t.Screen
		return a, a.models[a.current].Init()
	}

	// Delegate to current screen
	updated, cmd := a.models[a.current].Update(msg)
	a.models[a.current] = updated
	return a, cmd
}

func (a App) View() string {
	return a.models[a.current].View()
}
