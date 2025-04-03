package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sanjaykishor/JenkinsTui.git/internal/utils"
)

// BuildInfo represents a build in the build list
type BuildInfo struct {
	Number    int
	Status    string
	StartTime time.Time
	Duration  time.Duration
}

// Build represents detailed information about a build
type Build struct {
	Number      int
	Status      string
	StartTime   time.Time
	Duration    time.Duration
	Description string
	Parameters  map[string]string
}

// JobDetailComponent represents the job detail view
type JobDetailComponent struct {
	jobName     string
	jobURL      string
	description string
	buildList   list.Model
	lastBuild   *Build
	width       int
	height      int
	keys        KeyMap
}

// FilterValue implements list.Item
func (b BuildInfo) FilterValue() string {
	return fmt.Sprintf("#%d", b.Number)
}

// Title implements list.Item
func (b BuildInfo) Title() string {
	return fmt.Sprintf("Build #%d", b.Number)
}

// Description implements list.Item
func (b BuildInfo) Description() string {
	statusColor := utils.GetStatusColor(b.Status)
	status := lipgloss.NewStyle().Foreground(lipgloss.Color(statusColor)).Render(b.Status)

	return fmt.Sprintf("%s | %s | Duration: %s",
		status,
		utils.FormatTimeAgo(b.StartTime),
		utils.FormatDuration(int64(b.Duration)),
	)
}

// NewJobDetail creates a new job detail component
func NewJobDetail() JobDetailComponent {
	// Set up the build list
	delegate := list.NewDefaultDelegate()
	buildList := list.New([]list.Item{}, delegate, 0, 0)
	buildList.Title = "Builds"
	buildList.SetShowStatusBar(true)
	buildList.SetFilteringEnabled(true)
	buildList.Styles.Title = utils.TitleStyle
	buildList.SetShowHelp(true)

	return JobDetailComponent{
		buildList: buildList,
		keys:      DefaultKeyMap(),
	}
}

// WithJobDetail adds job details to the component
func (j JobDetailComponent) WithJobDetail(name, description, url string) JobDetailComponent {
	j.jobName = name
	j.description = description
	j.jobURL = url
	j.buildList.Title = fmt.Sprintf("Builds for %s", name)
	return j
}

// WithBuilds adds builds to the job detail component
func (j JobDetailComponent) WithBuilds(builds []BuildInfo) JobDetailComponent {
	// Convert builds to list items
	items := make([]list.Item, len(builds))
	for i, build := range builds {
		items[i] = build
	}
	j.buildList.SetItems(items)
	return j
}

// WithLastBuildInfo adds the last build information
func (j JobDetailComponent) WithLastBuildInfo(build Build) JobDetailComponent {
	j.lastBuild = &build
	return j
}

// GetSelectedBuild returns the currently selected build
func (j JobDetailComponent) GetSelectedBuild() *BuildInfo {
	if j.buildList.SelectedItem() == nil {
		return nil
	}

	selected := j.buildList.SelectedItem().(BuildInfo)
	return &selected
}

// Init initializes the component
func (j JobDetailComponent) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (j JobDetailComponent) Update(msg tea.Msg) (JobDetailComponent, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		j.width = msg.Width
		j.height = msg.Height
		j.buildList.SetWidth(msg.Width)
		j.buildList.SetHeight(msg.Height - 15) // Leave space for job info
	}

	// Handle build list updates
	j.buildList, cmd = j.buildList.Update(msg)

	return j, cmd
}

// View renders the job detail component
func (j JobDetailComponent) View() string {
	if j.width == 0 {
		return "Loading..."
	}

	var sb strings.Builder

	// Render job title and details
	title := utils.TitleStyle.Render(fmt.Sprintf("Job: %s", j.jobName))
	sb.WriteString(title)
	sb.WriteString("\n\n")

	// Job details section
	jobDetailsStyle := utils.InfoBlockStyle.Copy().Width(j.width - 4)

	var jobDetails strings.Builder
	jobDetails.WriteString(fmt.Sprintf("URL: %s\n", j.jobURL))
	if j.description != "" {
		jobDetails.WriteString(fmt.Sprintf("Description: %s\n", j.description))
	}

	// Last build info if available
	if j.lastBuild != nil {
		statusColor := utils.GetStatusColor(j.lastBuild.Status)
		status := lipgloss.NewStyle().Foreground(lipgloss.Color(statusColor)).Render(j.lastBuild.Status)

		jobDetails.WriteString(fmt.Sprintf("\nLast Build (#%d):\n", j.lastBuild.Number))
		jobDetails.WriteString(fmt.Sprintf("Status: %s\n", status))
		jobDetails.WriteString(fmt.Sprintf("Started: %s (%s)\n",
			j.lastBuild.StartTime.Format("2006-01-02 15:04:05"),
			utils.FormatTimeAgo(j.lastBuild.StartTime),
		))
		jobDetails.WriteString(fmt.Sprintf("Duration: %s\n", utils.FormatDuration(int64(j.lastBuild.Duration))))

		// Show parameters if any
		if len(j.lastBuild.Parameters) > 0 {
			jobDetails.WriteString("\nParameters:\n")
			for key, value := range j.lastBuild.Parameters {
				jobDetails.WriteString(fmt.Sprintf("- %s: %s\n", key, value))
			}
		}
	}

	sb.WriteString(jobDetailsStyle.Render(jobDetails.String()))
	sb.WriteString("\n\n")

	// Build list
	sb.WriteString(j.buildList.View())

	return sb.String()
}
