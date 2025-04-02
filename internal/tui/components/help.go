package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sanjaykishor/JenkinsTui.git/internal/utils"
)

// HelpComponent represents the help view
type HelpComponent struct {
	width  int
	height int
	keys   KeyMap
	help   help.Model
}

// NewHelp creates a new help component
func NewHelp() HelpComponent {
	h := help.New()
	h.ShowAll = true

	return HelpComponent{
		keys: DefaultKeyMap(),
		help: h,
	}
}

// Init initializes the help component
func (h HelpComponent) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (h HelpComponent) Update(msg tea.Msg) (HelpComponent, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h.width = msg.Width
		h.height = msg.Height
		h.help.Width = msg.Width
	}
	return h, nil
}

// View renders the help component
func (h HelpComponent) View() string {
	var sb strings.Builder

	// Title
	title := utils.HelpTitleStyle.Render("Jenkins TUI Help")
	sb.WriteString(title)
	sb.WriteString("\n\n")

	// Keyboard shortcuts section
	shortcutsContent := "Keyboard Shortcuts:\n\n" + h.help.View(h.keys)
	shortcuts := utils.HelpSectionStyle.Width(h.width - 4).Render(shortcutsContent)
	sb.WriteString(shortcuts)
	sb.WriteString("\n")

	// Usage section
	usageContent := fmt.Sprintf(`
Navigation:
• Use arrow keys to navigate in lists and logs
• Press Enter to select an item or action
• Press Esc to go back to the previous view
• Press j to go to the job list


Views:
• Dashboard: Overview of Jenkins server status
• Job List: List of all Jenkins jobs
• Job Detail: Information about a specific job
• Build Log: Console output for a specific build

Filtering:
• Press / to filter jobs in the job list
• Type your search term and press Enter
• Press Esc to cancel filtering

Tips:
• Press r to refresh data
• Logs will automatically colorize common patterns
`)

	usage := utils.HelpSectionStyle.Width(h.width - 4).Render("Usage Guide:" + usageContent)
	sb.WriteString(usage)
	sb.WriteString("\n")

	// About section
	aboutContent := fmt.Sprintf(`
Jenkins TUI is a terminal user interface for Jenkins CI/CD server.
Built with Go and the Bubble Tea framework.

Version: 0.1.0
Source: github.com/sanjaykishor/JenkinsTui
`)

	about := utils.HelpSectionStyle.Width(h.width - 4).Render("About:" + aboutContent)
	sb.WriteString(about)

	return sb.String()
}
