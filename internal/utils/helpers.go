package utils

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/sanjaykishor/JenkinsTui.git/internal/api"
)

// CountFreeNodes returns the number of online and idle nodes
func CountFreeNodes(nodes []api.Node) int {
	count := 0
	for _, node := range nodes {
		if node.Online && node.Idle {
			count++
		}
	}
	return count
}

// GetStatusColor returns the appropriate color for a job status
func GetStatusColor(status string) string {
	switch strings.ToLower(status) {
	case "success":
		return "42" // Green
	case "failed", "failure":
		return "196" // Red
	case "aborted":
		return "208" // Orange
	case "running":
		return "33" // Blue
	case "waiting":
		return "247" // Gray
	default:
		return "247" // Gray for unknown status
	}
}

// ColorizeLogOutput colors the log output
func ColorizeLogOutput(log string) string {
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
