package tui

import (
	"arch-installer/config"
	"arch-installer/tui/screens"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type App struct {
	cfg     *config.Config
	current config.Screen
	models  map[config.Screen]tea.Model
	width   int
	height  int
}

func New(cfg *config.Config) App {
	a := App{
		cfg:     cfg,
		current: config.ScreenWelcome,
		models:  make(map[config.Screen]tea.Model),
		width:   80,
		height:  24,
	}
	a.models[config.ScreenWelcome] = screens.NewWelcome(cfg)
	a.models[config.ScreenDiskSelect] = screens.NewDiskSelect(cfg)
	a.models[config.ScreenPackages] = screens.NewPackages(cfg)
	a.models[config.ScreenHostname] = screens.NewHostname(cfg)
	a.models[config.ScreenUser] = screens.NewUser(cfg)
	a.models[config.ScreenConfirm] = screens.NewConfirm(cfg)
	a.models[config.ScreenInstall] = screens.NewInstall(cfg)
	a.models[config.ScreenDone] = screens.NewDone(cfg)
	a.models[config.ScreenError] = screens.NewError(cfg)
	return a
}

func (a App) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		a.models[a.current].Init(),
	)
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok && k.String() == "ctrl+c" {
		return a, tea.Quit
	}

	// Propagate window size to all screens
	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		a.width = ws.Width
		a.height = ws.Height
		for k, m := range a.models {
			updated, _ := m.Update(msg)
			a.models[k] = updated
		}
		return a, nil
	}

	// Screen transition — reinitialize the target screen fresh
	if t, ok := msg.(screens.ToMsg); ok {
		a.current = t.Screen
		// Reinitialize with fresh model to clear any stale state
		switch t.Screen {
		case config.ScreenWelcome:
			a.models[t.Screen] = screens.NewWelcome(a.cfg)
		case config.ScreenDiskSelect:
			a.models[t.Screen] = screens.NewDiskSelect(a.cfg)
		case config.ScreenPackages:
			a.models[t.Screen] = screens.NewPackages(a.cfg)
		case config.ScreenHostname:
			a.models[t.Screen] = screens.NewHostname(a.cfg)
		case config.ScreenUser:
			a.models[t.Screen] = screens.NewUser(a.cfg)
		case config.ScreenConfirm:
			a.models[t.Screen] = screens.NewConfirm(a.cfg)
		case config.ScreenInstall:
			a.models[t.Screen] = screens.NewInstall(a.cfg)
		case config.ScreenDone:
			a.models[t.Screen] = screens.NewDone(a.cfg)
		case config.ScreenError:
			a.models[t.Screen] = screens.NewError(a.cfg)
		}
		// Send current window size to new screen immediately
		updated, cmd := a.models[t.Screen].Update(tea.WindowSizeMsg{
			Width:  a.width,
			Height: a.height,
		})
		a.models[t.Screen] = updated
		return a, tea.Batch(tea.ClearScreen, cmd)
	}

	updated, cmd := a.models[a.current].Update(msg)
	a.models[a.current] = updated
	return a, cmd
}

func (a App) View() string {
	content := a.models[a.current].View()
	return lipgloss.NewStyle().
		Width(a.width).
		Height(a.height).
		Render(content)
}
