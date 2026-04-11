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

func getTermSize() (int, int) {
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

func newScreen(cfg *config.Config, s config.Screen) tea.Model {
	switch s {
	case config.ScreenWelcome:
		return screens.NewWelcome(cfg)
	case config.ScreenNetwork:
		return screens.NewNetwork(cfg)
	case config.ScreenDiskSelect:
		return screens.NewDiskSelect(cfg)
	case config.ScreenPackages:
		return screens.NewPackages(cfg)
	case config.ScreenHostname:
		return screens.NewHostname(cfg)
	case config.ScreenUser:
		return screens.NewUser(cfg)
	case config.ScreenConfirm:
		return screens.NewConfirm(cfg)
	case config.ScreenInstall:
		return screens.NewInstall(cfg)
	case config.ScreenDone:
		return screens.NewDone(cfg)
	case config.ScreenError:
		return screens.NewError(cfg)
	}
	return screens.NewWelcome(cfg)
}

func New(cfg *config.Config) App {
	w, h := getTermSize()
	a := App{
		cfg:     cfg,
		current: config.ScreenWelcome,
		models:  make(map[config.Screen]tea.Model),
		width:   w,
		height:  h,
	}

	all := []config.Screen{
		config.ScreenWelcome, config.ScreenNetwork, config.ScreenDiskSelect,
		config.ScreenPackages, config.ScreenHostname, config.ScreenUser,
		config.ScreenConfirm, config.ScreenInstall, config.ScreenDone,
		config.ScreenError,
	}

	initMsg := tea.WindowSizeMsg{Width: w, Height: h}
	for _, s := range all {
		m := newScreen(cfg, s)
		updated, _ := m.Update(initMsg)
		a.models[s] = updated
	}
	return a
}

func (a App) Init() tea.Cmd {
	return tea.Batch(tea.EnterAltScreen, a.models[a.current].Init())
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok && k.String() == "ctrl+c" {
		return a, tea.Quit
	}

	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		a.width = ws.Width
		a.height = ws.Height
		for k, m := range a.models {
			updated, _ := m.Update(msg)
			a.models[k] = updated
		}
		return a, nil
	}

	if t, ok := msg.(screens.ToMsg); ok {
		a.current = t.Screen
		a.models[t.Screen] = newScreen(a.cfg, t.Screen)
		updated, cmd := a.models[t.Screen].Update(
			tea.WindowSizeMsg{Width: a.width, Height: a.height},
		)
		initCmd := updated.Init()
		a.models[t.Screen] = updated
		return a, tea.Batch(tea.ClearScreen, cmd, initCmd)
	}

	updated, cmd := a.models[a.current].Update(msg)
	a.models[a.current] = updated
	return a, cmd
}

func (a App) View() string {
	return lipgloss.NewStyle().
		Width(a.width).
		Height(a.height).
		Render(a.models[a.current].View())
}
