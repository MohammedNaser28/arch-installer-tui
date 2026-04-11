package screens

import (
	"arch-installer/config"
	"arch-installer/tui/components"
	"context"
	"fmt"
	"net"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type netState int

const (
	netChecking netState = iota
	netConnected
	netDisconnected
	netScanning
	netWifiList
	netWifiPassword
	netConnecting
	netFailed
)

type Network struct {
	cfg          *config.Config
	state        netState
	spin         spinner.Model
	networks     []string
	cursor       int
	passInput    textinput.Model
	selectedSSID string
	err          string
	width        int
	height       int
}

type netCheckResult struct{ connected bool }
type netScanResult struct{ networks []string }
type netConnectResult struct{ err error }

func NewNetwork(cfg *config.Config) Network {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(components.Mauve)

	pass := textinput.New()
	pass.Placeholder = "WiFi password"
	pass.EchoMode = textinput.EchoPassword
	pass.EchoCharacter = '•'
	pass.Width = 32

	return Network{
		cfg:       cfg,
		state:     netChecking,
		spin:      s,
		passInput: pass,
		width:     80,
		height:    24,
	}
}

func (m Network) Init() tea.Cmd {
	return tea.Batch(m.spin.Tick, checkInternet())
}

func checkInternet() tea.Cmd {
	return func() tea.Msg {
		conn, err := net.DialTimeout("tcp", "1.1.1.1:53", 3*time.Second)
		if err == nil {
			conn.Close()
			return netCheckResult{connected: true}
		}
		return netCheckResult{connected: false}
	}
}

func getWifiInterface() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "wlan0"
	}
	for _, i := range interfaces {
		if strings.HasPrefix(i.Name, "wl") || strings.HasPrefix(i.Name, "wlan") {
			return i.Name
		}
	}
	return "wlan0"
}

func scanWifi() tea.Cmd {
	return func() tea.Msg {
		// Fallback to nmcli if iwctl is not available
		if _, err := exec.LookPath("iwctl"); err != nil {
			out, nmErr := exec.Command("nmcli", "-t", "-f", "SSID", "dev", "wifi").Output()
			if nmErr != nil {
				return netScanResult{networks: []string{}}
			}
			var networks []string
			for _, line := range strings.Split(string(out), "\n") {
				line = strings.TrimSpace(line)
				if line != "" {
					found := false
					for _, n := range networks {
						if n == line {
							found = true
							break
						}
					}
					if !found {
						networks = append(networks, line)
					}
				}
			}
			return netScanResult{networks: networks}
		}

		iface := getWifiInterface()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		exec.CommandContext(ctx, "iwctl", "station", iface, "scan").Run()

		time.Sleep(2 * time.Second)

		ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel2()
		out, err := exec.CommandContext(ctx2, "iwctl", "station", iface, "get-networks").Output()
		if err != nil {
			return netScanResult{networks: []string{}}
		}
		var networks []string
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			cleanLine := strings.TrimSpace(line)
			if cleanLine == "" || strings.HasPrefix(cleanLine, "-") ||
				strings.HasPrefix(cleanLine, "Network") {
				continue
			}

			cleanLine = strings.TrimPrefix(cleanLine, ">")
			cleanLine = strings.TrimSpace(cleanLine)

			fields := strings.Fields(cleanLine)
			if len(fields) >= 3 {
				ssid := strings.Join(fields[:len(fields)-2], " ")
				networks = append(networks, ssid)
			} else if len(fields) > 0 {
				networks = append(networks, fields[0])
			}
		}
		return netScanResult{networks: networks}
	}
}

func connectWifi(ssid, password string) tea.Cmd {
	return func() tea.Msg {
		var err error
		if _, lookErr := exec.LookPath("iwctl"); lookErr != nil {
			if password == "" {
				err = exec.Command("nmcli", "dev", "wifi", "connect", ssid).Run()
			} else {
				err = exec.Command("nmcli", "dev", "wifi", "connect", ssid, "password", password).Run()
			}
		} else {
			iface := getWifiInterface()
			if password == "" {
				err = exec.Command("iwctl", "station", iface, "connect", ssid).Run()
			} else {
				err = exec.Command("iwctl", "station", iface,
					"connect", ssid, "--passphrase", password).Run()
			}
		}
		if err != nil {
			return netConnectResult{err: err}
		}
		time.Sleep(3 * time.Second)
		conn, dialErr := net.DialTimeout("tcp", "1.1.1.1:53", 3*time.Second)
		if dialErr != nil {
			return netConnectResult{err: fmt.Errorf("connected to WiFi but no internet")}
		}
		conn.Close()
		return netConnectResult{err: nil}
	}
}

func (m Network) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case spinner.TickMsg:
		m.spin, cmd = m.spin.Update(msg)
		return m, cmd

	case netCheckResult:
		if msg.connected {
			m.state = netConnected
		} else {
			m.state = netDisconnected
		}

	case netScanResult:
		m.networks = msg.networks
		m.state = netWifiList
		if len(m.networks) == 0 {
			m.err = "No networks found."
		}

	case netConnectResult:
		if msg.err != nil {
			m.state = netFailed
			m.err = msg.err.Error()
		} else {
			m.state = netConnected
		}

	case tea.KeyMsg:
		switch m.state {
		case netChecking, netScanning:
			switch msg.String() {
			case "s":
				return m, GoTo(config.ScreenDiskSelect)
			case "esc":
				return m, GoTo(config.ScreenWelcome)
			}

		case netDisconnected:
			switch msg.String() {
			case "r":
				m.state = netChecking
				return m, tea.Batch(m.spin.Tick, checkInternet())
			case "w":
				m.state = netScanning
				m.err = ""
				return m, tea.Batch(m.spin.Tick, scanWifi())
			case "s":
				return m, GoTo(config.ScreenDiskSelect)
			case "esc":
				return m, GoTo(config.ScreenWelcome)
			}

		case netConnected:
			switch msg.String() {
			case "enter":
				return m, GoTo(config.ScreenDiskSelect)
			case "esc":
				return m, GoTo(config.ScreenWelcome)
			}

		case netWifiList:
			switch msg.String() {
			case "s":
				return m, GoTo(config.ScreenDiskSelect)
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.networks)-1 {
					m.cursor++
				}
			case "enter":
				if len(m.networks) > 0 {
					m.selectedSSID = m.networks[m.cursor]
					m.state = netWifiPassword
					m.passInput.Focus()
					return m, textinput.Blink
				}
			case "esc":
				m.state = netDisconnected
			}

		case netWifiPassword:
			switch msg.String() {
			case "s":
				return m, GoTo(config.ScreenDiskSelect)
			case "enter":
				m.state = netConnecting
				return m, tea.Batch(m.spin.Tick,
					connectWifi(m.selectedSSID, m.passInput.Value()))
			case "esc":
				m.state = netWifiList
				m.passInput.SetValue("")
				return m, nil
			}
			m.passInput, cmd = m.passInput.Update(msg)
			return m, cmd

		case netConnecting:
			switch msg.String() {
			case "s":
				return m, GoTo(config.ScreenDiskSelect)
			case "esc":
				m.state = netWifiList
				return m, nil
			}

		case netFailed:
			switch msg.String() {
			case "r":
				m.state = netWifiList
				m.err = ""
			case "s":
				return m, GoTo(config.ScreenDiskSelect)
			case "esc":
				m.state = netDisconnected
			}
		}
	}
	return m, cmd
}

func (m Network) View() string {
	title := components.Title.Render("Network Connection")

	var body string
	switch m.state {

	case netChecking:
		body = lipgloss.JoinVertical(lipgloss.Left,
			components.Subtitle.Render("Checking internet connection..."),
			"\n"+m.spin.View()+" Verifying connectivity (1.1.1.1)",
			"\n\n"+components.Help("s", "skip", "esc", "back"),
		)

	case netConnected:
		body = lipgloss.JoinVertical(lipgloss.Left,
			components.Success.Render("✓  Connected to the internet"),
			"\n"+components.Dim.Render("Ready to install"),
			"\n"+components.Help("enter", "continue", "esc", "back"),
		)

	case netDisconnected:
		body = lipgloss.JoinVertical(lipgloss.Left,
			components.Err.Render("✗  No internet connection"),
			"\n"+components.Dim.Render("Internet required for pacstrap and package downloads"),
			"\n",
			"  "+components.Selected.Render("w")+"  "+components.Normal.Render("Scan for WiFi networks"),
			"  "+components.Selected.Render("r")+"  "+components.Normal.Render("Retry connection check"),
			"  "+components.Selected.Render("s")+"  "+components.Warn.Render("Skip (not recommended)"),
			"\n"+components.Help("esc", "back"),
		)

	case netScanning:
		body = lipgloss.JoinVertical(lipgloss.Left,
			components.Subtitle.Render("Scanning for WiFi networks..."),
			"\n"+m.spin.View()+" Using iwctl (this may take a few seconds)",
			"\n\n"+components.Help("s", "skip", "esc", "back"),
		)

	case netWifiList:
		sub := components.Subtitle.Render("Select a WiFi network")
		var rows string
		if len(m.networks) == 0 {
			rows = components.Err.Render("  No networks found. Try 'r' to rescan.")
		} else {
			for i, n := range m.networks {
				if i == m.cursor {
					rows += components.Selected.Render("❯ "+n) + "\n"
				} else {
					rows += components.Normal.Render("  "+n) + "\n"
				}
			}
		}
		box := components.BoxWithWidth(m.width).Render(rows)
		help := components.Help("s", "skip", "↑/↓", "navigate", "enter", "select", "esc", "back")
		body = lipgloss.JoinVertical(lipgloss.Left, sub, box, help)

	case netWifiPassword:
		sub := components.Subtitle.Render("Password for: " + components.Selected.Render(m.selectedSSID))
		box := components.ActiveBoxWithWidth(m.width).Render(m.passInput.View())
		hint := components.Dim.Render("Leave empty if open network")
		help := components.Help("s", "skip", "enter", "connect", "esc", "back")
		body = lipgloss.JoinVertical(lipgloss.Left, sub, "\n", box, hint, "\n"+help)

	case netConnecting:
		body = lipgloss.JoinVertical(lipgloss.Left,
			components.Subtitle.Render("Connecting to "+m.selectedSSID+"..."),
			"\n"+m.spin.View()+" Please wait",
			"\n\n"+components.Help("s", "skip", "esc", "back"),
		)

	case netFailed:
		body = lipgloss.JoinVertical(lipgloss.Left,
			components.Err.Render("✗  "+m.err),
			"\n",
			"  "+components.Selected.Render("r")+"  "+components.Normal.Render("Try another network"),
			"  "+components.Selected.Render("s")+"  "+components.Warn.Render("Skip (not recommended)"),
			"\n"+components.Help("esc", "back"),
		)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, title, "\n", body)
	return components.Page(m.width, m.height, content)
}
