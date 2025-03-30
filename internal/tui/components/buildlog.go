package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	logStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1, 2)
)

// BuildLogComponent represents the build log view
type BuildLogComponent struct {
	jobName   string
	buildNum  int
	viewport  viewport.Model
	width     int
	height    int
	ready     bool
	keys      KeyMap
	log       string
	logFilter string
}

// NewBuildLog creates a new build log component
func NewBuildLog() BuildLogComponent {
	return BuildLogComponent{
		keys: DefaultKeyMap(),
	}
}

// Init initializes the build log component
func (b BuildLogComponent) Init() tea.Cmd {
	return nil
}

// WithLog adds log content to the build log component
func (b BuildLogComponent) WithLog(log string) BuildLogComponent {
	b.log = log

	// If viewport is already initialized, update its content
	if b.ready {
		b.viewport.SetContent(b.formatLog())
	}

	return b
}

// WithJobAndBuild sets the job and build information
func (b BuildLogComponent) WithJobAndBuild(jobName string, buildNum int) BuildLogComponent {
	b.jobName = jobName
	b.buildNum = buildNum
	return b
}

// Update handles messages
func (b BuildLogComponent) Update(msg tea.Msg) (BuildLogComponent, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		b.width = msg.Width
		b.height = msg.Height

		if !b.ready {
			// Initialize viewport now that we know the terminal dimensions
			b.viewport = viewport.New(msg.Width-4, msg.Height-10)
			b.viewport.Style = logStyle
			b.viewport.SetContent(b.formatLog())
			b.ready = true
		} else {
			// Resize the viewport
			b.viewport.Width = msg.Width - 4
			b.viewport.Height = msg.Height - 10
		}
	}

	// Handle viewport messages
	if b.ready {
		b.viewport, cmd = b.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return b, tea.Batch(cmds...)
}

// View renders the build log component
func (b BuildLogComponent) View() string {
	if !b.ready {
		return "Loading..."
	}

	var sb strings.Builder

	// Add the title
	title := titleStyle.Render(fmt.Sprintf("Build Log: %s #%d", b.jobName, b.buildNum))
	sb.WriteString(title)
	sb.WriteString("\n\n")

	// Add viewport with log content
	sb.WriteString(b.viewport.View())

	// Add footer with controls
	footerHelp := fmt.Sprintf(
		"%s scroll up/down | %s page up/down | %s back",
		lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("↑/↓"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("PgUp/PgDown"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("ESC"),
	)

	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(footerHelp)

	sb.WriteString("\n\n")
	sb.WriteString(footer)

	return sb.String()
}

// formatLog formats the log content for display
func (b BuildLogComponent) formatLog() string {
	if b.log == "" {
		return "No log data available for this build."
	}

	// Apply any log filters here if needed
	log := b.log

	// Colorize certain log patterns
	log = colorizeLogOutput(log)

	return log
}

// Helper function to colorize log output
func colorizeLogOutput(log string) string {
	// Split log into lines
	lines := strings.Split(log, "\n")

	// Process each line
	for i, line := range lines {
		// Colorize error lines
		if strings.Contains(strings.ToLower(line), "error") ||
			strings.Contains(strings.ToLower(line), "exception") ||
			strings.Contains(strings.ToLower(line), "failed") {
			lines[i] = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(line)
		} else if strings.Contains(strings.ToLower(line), "warning") {
			// Colorize warning lines
			lines[i] = lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Render(line)
		} else if strings.HasPrefix(line, "+") || strings.HasPrefix(line, ">") {
			// Colorize command execution lines
			lines[i] = lipgloss.NewStyle().Foreground(lipgloss.Color("33")).Render(line)
		} else if strings.Contains(strings.ToLower(line), "success") ||
			strings.Contains(strings.ToLower(line), "passed") ||
			strings.Contains(strings.ToLower(line), "completed") {
			// Colorize success lines
			lines[i] = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render(line)
		}
	}

	return strings.Join(lines, "\n")
}
