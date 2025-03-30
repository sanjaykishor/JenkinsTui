package components

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// StatusType represents different types of status
type StatusType int

const (
	StatusInfo StatusType = iota
	StatusSuccess
	StatusWarning
	StatusError
)

// StatusComponent displays status messages
type StatusComponent struct {
	message string
	status  StatusType
	width   int
}

// NewStatus creates a new status component
func NewStatus() StatusComponent {
	return StatusComponent{
		message: "",
		status:  StatusInfo,
	}
}

// WithMessage sets the status message
func (s StatusComponent) WithMessage(message string, status StatusType) StatusComponent {
	s.message = message
	s.status = status
	return s
}

// Update handles window size changes
func (s StatusComponent) Update(msg tea.Msg) (StatusComponent, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
	}
	return s, nil
}

// View renders the status component
func (s StatusComponent) View() string {
	if s.message == "" {
		return ""
	}

	var style lipgloss.Style
	var prefix string

	switch s.status {
	case StatusSuccess:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
		prefix = "✓ "
	case StatusWarning:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
		prefix = "! "
	case StatusError:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		prefix = "✗ "
	default: // StatusInfo
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("33"))
		prefix = "ℹ "
	}

	return style.Render(fmt.Sprintf("%s%s", prefix, s.message))
}
