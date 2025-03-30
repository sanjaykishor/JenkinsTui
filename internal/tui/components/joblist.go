package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// JobListItem represents an item in the job list
type JobListItem struct {
	Name      string
	Status    string
	LastBuild time.Time
	JobDesc   string
	URL       string
}

// FilterValue returns the value to filter on
func (i JobListItem) FilterValue() string {
	return i.Name
}

// Title returns the title of the job item
func (i JobListItem) Title() string {
	return i.Name
}

// Description returns the description of the job item
func (i JobListItem) Description() string {
	statusColor := getStatusColor(i.Status)
	status := lipgloss.NewStyle().Foreground(lipgloss.Color(statusColor)).Render(i.Status)

	var lastBuildStr string
	if !i.LastBuild.IsZero() {
		lastBuildStr = fmt.Sprintf(" | Last build: %s", formatTimeAgo(i.LastBuild))
	}

	return fmt.Sprintf("%s%s | %s", status, lastBuildStr, i.JobDesc)
}

// JobListComponent represents the job list view
type JobListComponent struct {
	list     list.Model
	keys     KeyMap
	width    int
	height   int
	selected *JobListItem
}

// NewJobList creates a new job list component
func NewJobList() JobListComponent {
	// Set up list
	delegate := list.NewDefaultDelegate()
	jobList := list.New([]list.Item{}, delegate, 0, 0)
	jobList.Title = "Jenkins Jobs"
	jobList.SetShowStatusBar(true)
	jobList.SetFilteringEnabled(true)
	jobList.Styles.Title = titleStyle
	jobList.SetShowHelp(true)

	return JobListComponent{
		list: jobList,
		keys: DefaultKeyMap(),
	}
}

// WithJobs adds jobs to the job list
func (j JobListComponent) WithJobs(jobs []JobListItem) JobListComponent {
	items := make([]list.Item, len(jobs))
	for i, job := range jobs {
		items[i] = job
	}
	j.list.SetItems(items)
	return j
}

// GetSelected returns the selected job
func (j JobListComponent) GetSelected() *JobListItem {
	if j.list.SelectedItem() == nil {
		return nil
	}

	selected := j.list.SelectedItem().(JobListItem)
	return &selected
}

// Init initializes the job list component
func (j JobListComponent) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (j JobListComponent) Update(msg tea.Msg) (JobListComponent, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		j.width = msg.Width
		j.height = msg.Height
		j.list.SetWidth(msg.Width)
		j.list.SetHeight(msg.Height - 10) // Allow space for header and footer

	case tea.KeyMsg:
		// Don't handle keys if list is filtering
		if j.list.FilterState() == list.Filtering {
			break
		}

		switch {
		case key.Matches(msg, j.keys.Quit):
			return j, tea.Quit
		}
	}

	// Handle list updates
	j.list, cmd = j.list.Update(msg)
	cmds = append(cmds, cmd)

	return j, tea.Batch(cmds...)
}

// View renders the job list component
func (j JobListComponent) View() string {
	return j.list.View()
}

// Helper functions

// getStatusColor returns the appropriate color for a job status
func getStatusColor(status string) string {
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

// formatTimeAgo formats a time as a human readable "ago" string
func formatTimeAgo(t time.Time) string {
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
