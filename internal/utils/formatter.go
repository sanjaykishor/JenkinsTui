package utils

import (
	"fmt"
	"time"
	"strings"
	"github.com/charmbracelet/lipgloss"
)

// FormatDuration formats a duration in milliseconds as a human-readable string
func FormatDuration(durationMs int64) string {
	duration := time.Duration(durationMs) * time.Millisecond
	return formatTimeDuration(duration)
}

// FormatTimeDuration formats a time.Duration as a human-readable string
func formatTimeDuration(duration time.Duration) string {
	if duration < time.Second {
		return "Less than a second"
	}

	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	seconds := int(duration.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

// FormatTimeAgo formats a time as a human readable "ago" string
func FormatTimeAgo(t time.Time) string {
	duration := time.Since(t)

	if duration < time.Minute {
		return "just now"
	} else if duration < time.Hour {
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	} else if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	} else if duration < 7*24*time.Hour {
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	} else if duration < 30*24*time.Hour {
		weeks := int(duration.Hours() / 24 / 7)
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	} else {
		months := int(duration.Hours() / 24 / 30)
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	}
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

// Helper function to colorize log output
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