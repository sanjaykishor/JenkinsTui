package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sanjaykishor/JenkinsTui.git/internal/api"
	"github.com/sanjaykishor/JenkinsTui.git/internal/config"
)

// JenkinsService provides high-level Jenkins operations for the UI
type JenkinsService struct {
	client      *api.JenkinsClient
	config      *config.Manager
	configPath  string
	connected   bool
	lastError   error
	serverInfo  *api.ServerInfo
	lastRefresh time.Time
}

// NewJenkinsService creates a new JenkinsService
func NewJenkinsService() (*JenkinsService, error) {
	// Determine the config file path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %v", err)
	}

	configPath := filepath.Join(homeDir, ".jenkins-cli.yaml")

	// Create the config manager
	configManager := config.New(configPath)
	err = configManager.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	// Create the Jenkins client
	client, err := api.NewClient(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create Jenkins client: %v", err)
	}

	return &JenkinsService{
		client:     client,
		config:     configManager,
		configPath: configPath,
		connected:  false,
	}, nil
}

// Connect establishes a connection to the Jenkins server
func (s *JenkinsService) Connect() error {
	ctx := context.Background()

	// Get the server info to check connection
	info, err := s.client.GetServerInfo(ctx)
	if err != nil {
		s.connected = false
		s.lastError = err
		return err
	}

	// Get the nodes
	nodes, err := s.client.GetNodes(ctx)
	if err != nil {
		// Log the error but don't fail the connection
		s.lastError = err
	} else {
		// Add nodes to server info
		info.Nodes = nodes
	}

	s.connected = true
	s.serverInfo = info
	s.lastRefresh = time.Now()
	return nil
}

// IsConnected returns the connection status
func (s *JenkinsService) IsConnected() bool {
	return s.connected
}

// GetServerInfo returns information about the Jenkins server
func (s *JenkinsService) GetServerInfo() *api.ServerInfo {
	return s.serverInfo
}

// GetNodes returns a list of all Jenkins nodes
func (s *JenkinsService) GetNodes() ([]api.Node, error) {
	if !s.connected {
		return nil, fmt.Errorf("not connected to Jenkins server")
	}

	ctx := context.Background()
	nodes, err := s.client.GetNodes(ctx)
	if err != nil {
		s.lastError = err
		return nil, err
	}

	return nodes, nil
}

// GetJobs returns a list of all Jenkins jobs
func (s *JenkinsService) GetJobs() ([]api.Job, error) {
	if !s.connected {
		return nil, fmt.Errorf("not connected to Jenkins server")
	}

	ctx := context.Background()
	jobs, err := s.client.GetJobs(ctx)
	if err != nil {
		s.lastError = err
		return nil, err
	}

	return jobs, nil
}

// GetJobDetails returns detailed information about a specific job
func (s *JenkinsService) GetJobDetails(jobName string) (*api.JobDetail, error) {
	if !s.connected {
		return nil, fmt.Errorf("not connected to Jenkins server")
	}

	ctx := context.Background()
	jobDetail, err := s.client.GetJobDetails(ctx, jobName)
	if err != nil {
		s.lastError = err
		return nil, err
	}

	return jobDetail, nil
}

// GetBuildDetails returns detailed information about a specific build
func (s *JenkinsService) GetBuildDetails(jobName string, buildNumber int) (*api.BuildDetail, error) {
	if !s.connected {
		return nil, fmt.Errorf("not connected to Jenkins server")
	}

	ctx := context.Background()
	buildDetail, err := s.client.GetBuildDetails(ctx, jobName, buildNumber)
	if err != nil {
		s.lastError = err
		return nil, err
	}

	return buildDetail, nil
}

// GetBuildLog returns the console output for a specific build
func (s *JenkinsService) GetBuildLog(jobName string, buildNumber int) (string, error) {
	if !s.connected {
		return "", fmt.Errorf("not connected to Jenkins server")
	}

	ctx := context.Background()
	log, err := s.client.GetBuildLog(ctx, jobName, buildNumber)
	if err != nil {
		s.lastError = err
		return "", err
	}

	return log, nil
}

// TriggerBuild starts a build for a specific job
func (s *JenkinsService) TriggerBuild(jobName string, parameters map[string]string) error {
	if !s.connected {
		return fmt.Errorf("not connected to Jenkins server")
	}

	ctx := context.Background()
	err := s.client.TriggerBuild(ctx, jobName, parameters)
	if err != nil {
		s.lastError = err
		return err
	}

	return nil
}

// DeleteJob deletes a job from the Jenkins server
func (s *JenkinsService) DeleteJob(jobName string) error {
	if !s.connected {
		return fmt.Errorf("not connected to Jenkins server")
	}

	ctx := context.Background()
	err := s.client.DeleteJob(ctx, jobName)
	if err != nil {
		s.lastError = err
		return err
	}

	return nil
}

// StopBuild stops a running build
func (s *JenkinsService) StopBuild(jobName string, buildNumber int) error {
	if !s.connected {
		return fmt.Errorf("not connected to Jenkins server")
	}

	ctx := context.Background()
	err := s.client.StopBuild(ctx, jobName, buildNumber)
	if err != nil {
		s.lastError = err
		return err
	}

	return nil
}

// GetLastError returns the last error encountered
func (s *JenkinsService) GetLastError() error {
	return s.lastError
}

// Refresh refreshes the connection and server information
func (s *JenkinsService) Refresh() error {
	return s.Connect()
}

// ShouldRefresh returns true if it's time to refresh based on the configured interval
func (s *JenkinsService) ShouldRefresh() bool {
	if s.config == nil || s.config.Config == nil {
		return true
	}

	refreshInterval := s.config.Config.UI.RefreshInterval
	if refreshInterval <= 0 {
		refreshInterval = 30
	}

	return time.Since(s.lastRefresh) > time.Duration(refreshInterval)*time.Second
}
