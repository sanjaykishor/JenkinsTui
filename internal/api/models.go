package api

import "time"

// ServerInfo represents information about a Jenkins server
type ServerInfo struct {
	URL       string
	Version   string
	Mode      string
	Connected bool
	Username  string
	Nodes     []Node
	Uptime    time.Duration
}

// Node represents a Jenkins node (agent)
type Node struct {
	Name         string
	DisplayName  string
	Description  string
	Online       bool
	Idle         bool
	NumExecutors int
}

// Job represents a Jenkins job
type Job struct {
	Name        string
	URL         string
	Class       string
	Color       string
	Description string
	Status      string
	InProgress  bool
}

// JobDetail represents detailed information about a Jenkins job
type JobDetail struct {
	Name        string
	URL         string
	Description string
	Buildable   bool
	Builds      []Build
	LastBuild   *Build
	Parameters  []JobParameter
}

// JobParameter represents a parameter for a Jenkins job
type JobParameter struct {
	Name         string
	Type         string
	DefaultValue string
	Description  string
	Choices      []string
}

// Build represents a Jenkins build
type Build struct {
	Number      int
	URL         string
	Status      string
	StartTime   int64
	Duration    int64
	Result      string
	Description string
}

// BuildDetail represents detailed information about a Jenkins build
type BuildDetail struct {
	Number      int
	URL         string
	StartTime   int64
	Duration    int64
	Building    bool
	Result      string
	Description string
	Parameters  map[string]string
}

// GetStatusFromColor converts a Jenkins color to a status string
func GetStatusFromColor(color string) (status string, inProgress bool) {
	switch color {
	case "blue", "blue_anime":
		status = "success"
		inProgress = color == "blue_anime"
	case "red", "red_anime":
		status = "failure"
		inProgress = color == "red_anime"
	case "yellow", "yellow_anime":
		status = "unstable"
		inProgress = color == "yellow_anime"
	case "grey", "grey_anime", "disabled", "disabled_anime":
		status = "disabled"
		inProgress = color == "grey_anime" || color == "disabled_anime"
	case "aborted", "aborted_anime":
		status = "aborted"
		inProgress = color == "aborted_anime"
	default:
		status = "unknown"
		inProgress = false
	}
	return
}
