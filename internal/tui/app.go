package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sanjaykishor/JenkinsTui.git/internal/api"
	"github.com/sanjaykishor/JenkinsTui.git/internal/tui/components"
)

// ViewType represents the various views available in our application
type ViewType int

const (
	DashboardView ViewType = iota
	JobListView
	JobDetailView
	BuildLogView
	HelpView
)

// KeyMap defines the keybindings for the application
type KeyMap struct {
	Up        key.Binding
	Down      key.Binding
	Left      key.Binding
	Right     key.Binding
	Help      key.Binding
	Quit      key.Binding
	Enter     key.Binding
	Back      key.Binding
	Dashboard key.Binding
	Jobs      key.Binding
	Refresh   key.Binding
}

// DefaultKeyMap returns the default keybindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "right"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q/ctrl+c", "quit"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		Dashboard: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "dashboard"),
		),
		Jobs: key.NewBinding(
			key.WithKeys("j"),
			key.WithHelp("j", "jobs"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
	}
}

// ShortHelp returns keybindings to be shown in the mini help view
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit, k.Enter, k.Back}
}

// FullHelp returns keybindings for the expanded help view
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.Enter, k.Back, k.Help, k.Quit},
		{k.Dashboard, k.Jobs, k.Refresh},
	}
}

// Custom tea.Msg types for asynchronous operations
type connectMsg struct {
	err        error
	serverInfo *api.ServerInfo
}

type fetchJobsMsg struct {
	jobs []api.Job
	err  error
}

type fetchJobDetailMsg struct {
	jobDetail *api.JobDetail
	err       error
}

type fetchBuildDetailMsg struct {
	buildDetail *api.BuildDetail
	err         error
}

type fetchBuildLogMsg struct {
	buildLog string
	err      error
}

// RefreshTickMsg is sent when it's time to refresh the UI
type RefreshTickMsg time.Time

// Model represents the state of our application
type Model struct {
	keys           KeyMap
	help           help.Model
	currentView    ViewType
	width          int
	height         int
	showFullHelp   bool
	connected      bool
	serverURL      string
	errorMsg       string
	statusMessage  string
	loadingMessage string
	service        *JenkinsService
	selectedJob    string
	selectedBuild  int

	// View components
	dashboard components.DashboardComponent
	jobList   components.JobListComponent
	jobDetail components.JobDetailComponent
	buildLog  components.BuildLogComponent
	helpView  components.HelpComponent
}

// New returns a new instance of our application model
func New() (Model, error) {
	keys := DefaultKeyMap()
	h := help.New()
	h.ShowAll = false

	// Initialize the Jenkins service
	service, err := NewJenkinsService()
	if err != nil {
		return Model{}, fmt.Errorf("failed to initialize Jenkins service: %v", err)
	}

	m := Model{
		keys:           keys,
		help:           h,
		currentView:    DashboardView,
		connected:      false,
		serverURL:      "",
		statusMessage:  "Welcome to Jenkins TUI",
		loadingMessage: "",
		dashboard:      components.NewDashboard(),
		jobList:        components.NewJobList(),
		jobDetail:      components.NewJobDetail(),
		buildLog:       components.NewBuildLog(),
		helpView:       components.NewHelp(),
		service:        service,
	}

	return m, nil
}

// Connect initiates a connection to the Jenkins server
func (m Model) Connect() tea.Cmd {
	return func() tea.Msg {
		err := m.service.Connect()
		if err != nil {
			return connectMsg{err: err}
		}
		return connectMsg{serverInfo: m.service.GetServerInfo()}
	}
}

// FetchJobs retrieves the list of Jenkins jobs
func (m Model) FetchJobs() tea.Cmd {
	return func() tea.Msg {
		jobs, err := m.service.GetJobs()
		return fetchJobsMsg{jobs: jobs, err: err}
	}
}

// FetchJobDetail retrieves detailed information about a specific job
func (m Model) FetchJobDetail(jobName string) tea.Cmd {
	return func() tea.Msg {
		jobDetail, err := m.service.GetJobDetails(jobName)
		return fetchJobDetailMsg{jobDetail: jobDetail, err: err}
	}
}

// FetchBuildDetail retrieves detailed information about a specific build
func (m Model) FetchBuildDetail(jobName string, buildNumber int) tea.Cmd {
	return func() tea.Msg {
		buildDetail, err := m.service.GetBuildDetails(jobName, buildNumber)
		return fetchBuildDetailMsg{buildDetail: buildDetail, err: err}
	}
}

// FetchBuildLog retrieves the console output for a specific build
func (m Model) FetchBuildLog(jobName string, buildNumber int) tea.Cmd {
	return func() tea.Msg {
		log, err := m.service.GetBuildLog(jobName, buildNumber)
		return fetchBuildLogMsg{buildLog: log, err: err}
	}
}

// RefreshTick creates a command that will send a tick message after a duration
func RefreshTick(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return RefreshTickMsg(t)
	})
}

// Init implements bubbletea.Model
func (m Model) Init() tea.Cmd {
	// Initialize all components
	return tea.Batch(
		m.dashboard.Init(),
		m.jobList.Init(),
		m.jobDetail.Init(),
		m.buildLog.Init(),
		m.helpView.Init(),
		m.Connect(),
		RefreshTick(30*time.Second),
	)
}

// Update implements bubbletea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case connectMsg:
		if msg.err != nil {
			m.connected = false
			m.errorMsg = fmt.Sprintf("Connection error: %v", msg.err)
			m.statusMessage = "Connection failed"
		} else {
			m.connected = true
			m.serverURL = msg.serverInfo.URL
			m.statusMessage = "Connected to Jenkins"

			// Update the dashboard with server info
			serverInfo := components.ServerInfo{
				URL:        msg.serverInfo.URL,
				Version:    msg.serverInfo.Version,
				Connected:  true,
				Mode:       msg.serverInfo.Mode,
				Uptime:     msg.serverInfo.Uptime.String(),
				TotalNodes: len(msg.serverInfo.Nodes),
				FreeNodes:  countFreeNodes(msg.serverInfo.Nodes),
			}
			m.dashboard = m.dashboard.WithServerInfo(serverInfo)

			// Fetch jobs
			cmds = append(cmds, m.FetchJobs())
		}

	case fetchJobsMsg:
		if msg.err != nil {
			m.errorMsg = fmt.Sprintf("Failed to fetch jobs: %v", msg.err)
		} else {
			// Update the job list
			var jobItems []components.JobListItem
			for _, job := range msg.jobs {
				jobItem := components.JobListItem{
					Name:      job.Name,
					Status:    string(job.Status),
					LastBuild: time.Now().Add(-time.Hour), // This would come from the API
					JobDesc:   job.Description,
					URL:       job.URL,
				}
				jobItems = append(jobItems, jobItem)
			}
			m.jobList = m.jobList.WithJobs(jobItems)
		}

	case fetchJobDetailMsg:
		if msg.err != nil {
			m.errorMsg = fmt.Sprintf("Failed to fetch job details: %v", msg.err)
		} else {
			// Update the job detail view
			jobDetail := msg.jobDetail
			m.jobDetail = m.jobDetail.WithJobDetail(jobDetail.Name, jobDetail.Description, jobDetail.URL)

			// If there are builds, add them
			if len(jobDetail.Builds) > 0 {
				var builds []components.BuildInfo
				for _, build := range jobDetail.Builds {
					buildInfo := components.BuildInfo{
						Number:    build.Number,
						Status:    string(build.Status),
						StartTime: time.Unix(build.StartTime/1000, 0),
						Duration:  time.Duration(build.Duration) * time.Millisecond,
					}
					builds = append(builds, buildInfo)
				}
				m.jobDetail = m.jobDetail.WithBuilds(builds)

				// Fetch the last build details
				if jobDetail.LastBuild != nil {
					m.selectedBuild = jobDetail.LastBuild.Number
					cmds = append(cmds, m.FetchBuildDetail(jobDetail.Name, jobDetail.LastBuild.Number))
				}
			}
		}

	case fetchBuildDetailMsg:
		if msg.err != nil {
			m.errorMsg = fmt.Sprintf("Failed to fetch build details: %v", msg.err)
		} else {
			// Update the build detail
			buildDetail := msg.buildDetail
			m.jobDetail = m.jobDetail.WithLastBuildInfo(components.Build{
				Number:      buildDetail.Number,
				Status:      string(api.GetStatusFromResult(buildDetail.Result, buildDetail.Building)),
				StartTime:   time.Unix(buildDetail.StartTime/1000, 0),
				Duration:    time.Duration(buildDetail.Duration) * time.Millisecond,
				Description: buildDetail.Description,
				Parameters:  buildDetail.Parameters,
			})
		}

	case fetchBuildLogMsg:
		if msg.err != nil {
			m.errorMsg = fmt.Sprintf("Failed to fetch build log: %v", msg.err)
		} else {
			// Update the build log
			m.buildLog = m.buildLog.WithLog(msg.buildLog)
			// Also update the job and build number for display purposes
			m.buildLog = m.buildLog.WithJobAndBuild(m.selectedJob, m.selectedBuild)
		}

	case RefreshTickMsg:
		// Check if it's time to refresh
		if m.service.ShouldRefresh() {
			cmds = append(cmds, m.Connect())
		}

		// Schedule the next refresh
		cmds = append(cmds, RefreshTick(30*time.Second))

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keys.Help):
			if m.currentView == HelpView {
				// If we're already in help view, go back to previous view
				m.currentView = DashboardView
				m.statusMessage = "Dashboard View"
			} else {
				// Otherwise show help view
				m.currentView = HelpView
				m.statusMessage = "Help View"
			}
			return m, nil

		case key.Matches(msg, m.keys.Dashboard):
			m.currentView = DashboardView
			m.statusMessage = "Dashboard View"
			return m, nil

		case key.Matches(msg, m.keys.Jobs):
			m.currentView = JobListView
			m.statusMessage = "Job List View"

			// Refresh the job list when viewing it
			if m.connected {
				cmds = append(cmds, m.FetchJobs())
			}

			return m, tea.Batch(cmds...)

		case key.Matches(msg, m.keys.Enter):
			if m.currentView == JobListView {
				// Get the selected job
				selected := m.jobList.GetSelected()
				if selected != nil {
					m.selectedJob = selected.Name
					m.currentView = JobDetailView
					m.statusMessage = fmt.Sprintf("Job: %s", selected.Name)

					// Fetch job details
					if m.connected {
						cmds = append(cmds, m.FetchJobDetail(selected.Name))
					}

					return m, tea.Batch(cmds...)
				}
				return m, nil
			} else if m.currentView == JobDetailView {
				// Get the selected build
				selected := m.jobDetail.GetSelectedBuild()
				if selected != nil {
					m.selectedBuild = selected.Number
					m.currentView = BuildLogView
					m.statusMessage = fmt.Sprintf("Build #%d Logs", selected.Number)

					// Fetch build logs
					if m.connected && m.selectedJob != "" {
						cmds = append(cmds, m.FetchBuildLog(m.selectedJob, selected.Number))
					}

					return m, tea.Batch(cmds...)
				}
				return m, nil
			}

		case key.Matches(msg, m.keys.Back):
			// Handle navigation back
			switch m.currentView {
			case JobDetailView:
				m.currentView = JobListView
				m.statusMessage = "Job List View"
			case BuildLogView:
				m.currentView = JobDetailView
				m.statusMessage = "Job Detail View"
			case HelpView:
				m.currentView = DashboardView
				m.statusMessage = "Dashboard View"
			}
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width

		// Update component sizes
		var cmd tea.Cmd

		m.dashboard, cmd = m.dashboard.Update(msg)
		cmds = append(cmds, cmd)

		m.jobList, cmd = m.jobList.Update(msg)
		cmds = append(cmds, cmd)

		m.jobDetail, cmd = m.jobDetail.Update(msg)
		cmds = append(cmds, cmd)

		m.buildLog, cmd = m.buildLog.Update(msg)
		cmds = append(cmds, cmd)

		m.helpView, cmd = m.helpView.Update(msg)
		cmds = append(cmds, cmd)

		return m, tea.Batch(cmds...)
	}

	// Handle component updates based on current view
	switch m.currentView {
	case DashboardView:
		var cmd tea.Cmd
		m.dashboard, cmd = m.dashboard.Update(msg)
		cmds = append(cmds, cmd)
	case JobListView:
		var cmd tea.Cmd
		m.jobList, cmd = m.jobList.Update(msg)
		cmds = append(cmds, cmd)
	case JobDetailView:
		var cmd tea.Cmd
		m.jobDetail, cmd = m.jobDetail.Update(msg)
		cmds = append(cmds, cmd)
	case BuildLogView:
		var cmd tea.Cmd
		m.buildLog, cmd = m.buildLog.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View implements bubbletea.Model
func (m Model) View() string {
	// Status bar at the bottom
	statusBar := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render(m.statusMessage)

	// Error message
	var errorView string
	if m.errorMsg != "" {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Render(m.errorMsg)
		errorView = fmt.Sprintf("\n%s", errorStyle)
	}

	// Help at the bottom
	helpView := m.help.View(m.keys)

	// Main content
	var content string
	switch m.currentView {
	case DashboardView:
		content = m.dashboard.View()
	case JobListView:
		content = m.jobList.View()
	case JobDetailView:
		content = m.jobDetail.View()
	case BuildLogView:
		content = m.buildLog.View()
	case HelpView:
		content = m.helpView.View()
	}

	// Combine everything
	return fmt.Sprintf("%s\n\n%s\n\n%s%s", content, statusBar, helpView, errorView)
}

// Helper functions

// countFreeNodes returns the number of online and idle nodes
func countFreeNodes(nodes []api.Node) int {
	count := 0
	for _, node := range nodes {
		if node.Online && node.Idle {
			count++
		}
	}
	return count
}
