package utils

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// Styles for different build statuses
var (
	SuccessStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#00CC00"))
	FailureStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
	UnstableStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFCC00"))
	AbortedStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#999999"))
	InProgressStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00CCFF"))
	DisabledStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
)

// FormatBuildStatus returns a styled string for the build status
func FormatBuildStatus(status string, inProgress bool) string {
	if inProgress {
		return InProgressStyle.Render(fmt.Sprintf("%s (Building)", status))
	}

	switch strings.ToLower(status) {
	case "success":
		return SuccessStyle.Render(status)
	case "failure":
		return FailureStyle.Render(status)
	case "unstable":
		return UnstableStyle.Render(status)
	case "aborted":
		return AbortedStyle.Render(status)
	case "disabled":
		return DisabledStyle.Render(status)
	default:
		return status
	}
}

// FormatDuration formats a duration in milliseconds as a human-readable string
func FormatDuration(durationMs int64) string {
	duration := time.Duration(durationMs) * time.Millisecond
	return FormatTimeDuration(duration)
}

// FormatTimeDuration formats a time.Duration as a human-readable string
func FormatTimeDuration(duration time.Duration) string {
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

// FormatTimestamp formats a Unix timestamp as a human-readable string
func FormatTimestamp(timestamp int64) string {
	t := time.Unix(timestamp/1000, 0)
	return t.Format("Jan 02, 2006 15:04:05")
}

// FormatTimeElapsed formats the time elapsed since a timestamp
func FormatTimeElapsed(timestamp int64) string {
	elapsed := time.Since(time.Unix(timestamp/1000, 0))
	if elapsed.Hours() >= 24 {
		days := int(elapsed.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	} else if elapsed.Hours() >= 1 {
		hours := int(elapsed.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	} else if elapsed.Minutes() >= 1 {
		minutes := int(elapsed.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	}
	return "just now"
}

// TruncateString truncates a string to the specified length, adding ellipsis if needed
func TruncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	return s[:maxLength-3] + "..."
}

// FormatPercentage formats a percentage value with % symbol
func FormatPercentage(value float64) string {
	return fmt.Sprintf("%.1f%%", value)
}

// FormatBuildNumber formats a build number as "#123"
func FormatBuildNumber(number int) string {
	return fmt.Sprintf("#%d", number)
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