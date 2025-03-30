package api

import (
	"context"
	"fmt"
	"sync"

	"github.com/jenkins-zh/jenkins-cli/app/cmd"
	"github.com/jenkins-zh/jenkins-cli/app/helper"
	"github.com/jenkins-zh/jenkins-cli/client"
)

// JobStatus represents the status of a Jenkins job or build
type JobStatus string

const (
	// Job/Build statuses
	StatusSuccess JobStatus = "success"
	StatusFailed  JobStatus = "failed"
	StatusAborted JobStatus = "aborted"
	StatusRunning JobStatus = "running"
	StatusWaiting JobStatus = "waiting"
	StatusUnknown JobStatus = "unknown"
)

// GetStatusFromResult converts Jenkins result strings to JobStatus
func GetStatusFromResult(result string, building bool) JobStatus {
	if building {
		return StatusRunning
	}

	switch result {
	case "SUCCESS":
		return StatusSuccess
	case "FAILURE":
		return StatusFailed
	case "ABORTED":
		return StatusAborted
	case "UNSTABLE":
		return StatusFailed
	case "NOT_BUILT":
		return StatusWaiting
	default:
		return StatusUnknown
	}
}

// JenkinsClient is a wrapper around jenkins-cli that provides methods
// for interacting with a Jenkins server
type JenkinsClient struct {
	jclient  *client.JenkinsCore
	config   *helper.JenkinsServer
	rootCore cmd.Core
	mutex    sync.Mutex
	apiErr   error
}

// NewClient creates a new JenkinsClient with the given config
func NewClient(configPath string) (*JenkinsClient, error) {
	// Create a new JenkinsClient
	c := &JenkinsClient{
		jclient: &client.JenkinsCore{},
	}

	// Initialize the jenkins-cli core
	c.rootCore = cmd.Core{}
	err := c.rootCore.SetupViperWithoutFlagParsing(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	// Get the current Jenkins server from config
	config := c.rootCore.GetCurrentJenkinsFromOptions()
	if config == nil {
		return nil, fmt.Errorf("no Jenkins server found in config")
	}

	c.config = config
	c.jclient.URL = config.URL
	c.jclient.UserName = config.Username
	c.jclient.Token = config.Token
	c.jclient.Proxy = config.Proxy
	c.jclient.InsecureSkipVerify = config.InsecureSkipVerify

	return c, nil
}

// GetServerInfo retrieves information about the Jenkins server
func (c *JenkinsClient) GetServerInfo(ctx context.Context) (*ServerInfo, error) {
	// Lock to ensure thread safety
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Get the Jenkins server info
	info, err := c.jclient.GetCrumb(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Jenkins: %v", err)
	}

	serverInfo := &ServerInfo{
		URL:       c.config.URL,
		Connected: true,
		Username:  c.config.Username,
	}

	// Get the Jenkins version
	versionData, err := c.jclient.GetVersion(ctx)
	if err == nil && versionData != nil {
		serverInfo.Version = versionData.Jenkins.Version
		serverInfo.Mode = versionData.Jenkins.Mode
	}

	return serverInfo, nil
}

// GetJobs retrieves a list of all jobs from the Jenkins server
func (c *JenkinsClient) GetJobs(ctx context.Context) ([]Job, error) {
	// Lock to ensure thread safety
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Get jobs from the Jenkins server
	jobsMeta, err := c.jclient.GetJobs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs: %v", err)
	}

	// Convert the jenkins-cli jobs to our model
	var jobs []Job
	for _, jobMeta := range jobsMeta.Jobs {
		job := Job{
			Name:        jobMeta.Name,
			URL:         jobMeta.URL,
			Class:       jobMeta.Class,
			Color:       jobMeta.Color,
			Description: jobMeta.Description,
		}

		// Determine the job status based on the color
		switch jobMeta.Color {
		case "blue", "blue_anime":
			job.Status = "success"
			job.InProgress = jobMeta.Color == "blue_anime"
		case "red", "red_anime":
			job.Status = "failure"
			job.InProgress = jobMeta.Color == "red_anime"
		case "yellow", "yellow_anime":
			job.Status = "unstable"
			job.InProgress = jobMeta.Color == "yellow_anime"
		case "grey", "grey_anime", "disabled", "disabled_anime":
			job.Status = "disabled"
			job.InProgress = jobMeta.Color == "grey_anime" || jobMeta.Color == "disabled_anime"
		case "aborted", "aborted_anime":
			job.Status = "aborted"
			job.InProgress = jobMeta.Color == "aborted_anime"
		default:
			job.Status = "unknown"
			job.InProgress = false
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
}

// GetJobDetails retrieves detailed information about a specific job
func (c *JenkinsClient) GetJobDetails(ctx context.Context, jobName string) (*JobDetail, error) {
	// Lock to ensure thread safety
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Get job details from the Jenkins server
	jobDetails, err := c.jclient.GetJob(ctx, jobName)
	if err != nil {
		return nil, fmt.Errorf("failed to get job details: %v", err)
	}

	// Create a JobDetail object
	job := &JobDetail{
		Name:        jobDetails.Name,
		URL:         jobDetails.URL,
		Description: jobDetails.Description,
		Buildable:   jobDetails.Buildable,
		Builds:      make([]Build, 0),
	}

	// Add builds
	for _, build := range jobDetails.Builds {
		job.Builds = append(job.Builds, Build{
			Number: build.Number,
			URL:    build.URL,
		})
	}

	// Set the last build info if available
	if jobDetails.LastBuild != nil {
		job.LastBuild = &Build{
			Number: jobDetails.LastBuild.Number,
			URL:    jobDetails.LastBuild.URL,
		}
	}

	return job, nil
}

// GetBuildDetails retrieves details about a specific build
func (c *JenkinsClient) GetBuildDetails(ctx context.Context, jobName string, buildNumber int) (*BuildDetail, error) {
	// Lock to ensure thread safety
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Get build details from the Jenkins server
	buildDetails, err := c.jclient.GetBuild(ctx, jobName, buildNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get build details: %v", err)
	}

	// Create a BuildDetail object
	build := &BuildDetail{
		Number:      buildDetails.Number,
		URL:         buildDetails.URL,
		StartTime:   buildDetails.Timestamp,
		Duration:    buildDetails.Duration,
		Building:    buildDetails.Building,
		Result:      buildDetails.Result,
		Description: buildDetails.Description,
		Parameters:  make(map[string]string),
	}

	// Extract parameters if available
	if buildDetails.Actions != nil {
		for _, action := range buildDetails.Actions {
			if action.Parameters != nil {
				for _, param := range action.Parameters {
					if param.Name != "" {
						build.Parameters[param.Name] = fmt.Sprintf("%v", param.Value)
					}
				}
			}
		}
	}

	return build, nil
}

// GetBuildLog retrieves the console output for a specific build
func (c *JenkinsClient) GetBuildLog(ctx context.Context, jobName string, buildNumber int) (string, error) {
	// Lock to ensure thread safety
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Get build log from the Jenkins server
	log, err := c.jclient.GetBuildLog(ctx, jobName, buildNumber)
	if err != nil {
		return "", fmt.Errorf("failed to get build log: %v", err)
	}

	return log, nil
}

// TriggerBuild starts a build for a specific job
func (c *JenkinsClient) TriggerBuild(ctx context.Context, jobName string, parameters map[string]string) error {
	// Lock to ensure thread safety
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Create a job request
	jobReq := client.JobBuildReq{
		Parameters: parameters,
	}

	// Trigger the build
	err := c.jclient.TriggerJob(ctx, jobName, jobReq)
	if err != nil {
		return fmt.Errorf("failed to trigger build: %v", err)
	}

	return nil
}
