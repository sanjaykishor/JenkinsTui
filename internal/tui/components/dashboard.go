package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sanjaykishor/JenkinsTui.git/internal/utils"
)

// ServerInfo contains information about the Jenkins server
type ServerInfo struct {
	URL        string
	Version    string
	Connected  bool
	Mode       string
	Uptime     string
	TotalNodes int
	FreeNodes  int
}

// DashboardComponent represents the dashboard view
type DashboardComponent struct {
	width      int
	height     int
	keys       KeyMap
	help       help.Model
	serverInfo ServerInfo
}

// NewDashboard creates a new dashboard component
func NewDashboard() DashboardComponent {
	return DashboardComponent{
		keys: KeyMap{},
		help: help.New(),
		serverInfo: ServerInfo{
			Connected: false,
		},
	}
}

// WithServerInfo adds server information to the dashboard
func (d DashboardComponent) WithServerInfo(info ServerInfo) DashboardComponent {
	d.serverInfo = info
	return d
}

// Init initializes the dashboard component
func (d DashboardComponent) Init() tea.Cmd {
	return nil
}

// Update handles messages and user input
func (d DashboardComponent) Update(msg tea.Msg) (DashboardComponent, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.width = msg.Width
		d.height = msg.Height
		d.help.Width = msg.Width
	}
	return d, nil
}

// View renders the dashboard component
func (d DashboardComponent) View() string {
	// Create the dashboard view
	var sb strings.Builder

	title := utils.TitleStyle.Render("Jenkins TUI Dashboard")
	sb.WriteString(title)
	sb.WriteString("\n\n")

	// Server information
	serverTitle := "Server Information"
	if d.serverInfo.Connected {
		serverTitle += " " + utils.SuccessText.Render("● Connected")
	} else {
		serverTitle += " " + utils.FailureText.Render("● Disconnected")
	}

	var serverContent string
	if d.serverInfo.Connected {
		serverContent = fmt.Sprintf(
			"URL: %s\nVersion: %s\nMode: %s\nUptime: %s\nNodes: %d total, %d free",
			d.serverInfo.URL,
			d.serverInfo.Version,
			d.serverInfo.Mode,
			d.serverInfo.Uptime,
			d.serverInfo.TotalNodes,
			d.serverInfo.FreeNodes,
		)
	} else {
		serverContent = "Not connected to Jenkins server"
	}

	serverInfo := utils.ServerInfoStyle.Width(d.width / 2).Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			serverTitle,
			serverContent,
		),
	)

	sb.WriteString(serverInfo)
	sb.WriteString("\n\n")

	// Current time
	currentTime := fmt.Sprintf("Last updated: %s", time.Now().Format("2006-01-02 15:04:05"))
	sb.WriteString(currentTime)

	return sb.String()
}
