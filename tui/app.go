package tui

import (
	"arch-installer/config"
	"arch-installer/tui/screens"
	"os"
	"strconv"
	"strings"

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

func getTerminalSize() (int, int) {
	// Try to get real terminal size from environment
	w, h := 80, 24
	if cols := os.Getenv("COLUMNS"); cols != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(cols)); err == nil {
			w = n
		}
	}
	if rows := os.Getenv("LINES"); rows != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(rows)); err == nil {
			h = n
		}
	}
	return w, h
}

func New(cfg *config.Config) App {
	w, h := getTerminalSize()
	a := App{
		cfg:     cfg,
		current: config.ScreenWelcome,
		models:  make(map[config.Screen]tea.Model),
		width:   w,
		height:  h,
	}
	a.models[config.ScreenWelcome]    = screens.NewWelcome(cfg)
	a.models[config.ScreenDiskSelect] = screens.NewDiskSelect(cfg)
	a.models[config.ScreenPackages]   = screens.NewPackages(cfg)
	a.models[config.ScreenHostname]   = screens.NewHostname(cfg)
	a.models[config.ScreenUser]       = screens.NewUser(cfg)
	a.models[config.ScreenConfirm]    = screens.NewConfirm(cfg)
	a.models[config.ScreenInstall]    = screens.NewInstall(cfg)
	a.models[config.ScreenDone]       = screens.NewDone(cfg)
	a.models[config.ScreenError]      = screens.NewError(cfg)

	// Send initial window size to all screens
	initMsg := tea.WindowSizeMsg{Width: w, Height: h}
	for k, m := range a.models {
		updated, _ := m.Update(initMsg)
		a.models[k] = updated
	}
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

	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		a.width  = ws.Width
		a.height = ws.Height
		for k, m := range a.models {
			updated, _ := m.Update(msg)
			a.models[k] = updated
		}
		return a, nil
	}

	if t, ok := msg.(screens.ToMsg); ok {
		a.current = t.Screen
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
		updated, cmd := a.models[t.Screen].Update(tea.WindowSizeMsg{
			Width: a.width, Height: a.height,
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
