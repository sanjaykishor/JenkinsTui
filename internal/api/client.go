package api

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
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

// JenkinsConfig represents a configuration entry for a Jenkins server
type JenkinsConfig struct {
	Name               string `yaml:"name"`
	URL                string `yaml:"url"`
	Username           string `yaml:"username"`
	Token              string `yaml:"token"`
	Proxy              string `yaml:"proxy"`
	InsecureSkipVerify bool   `yaml:"insecureSkipVerify"`
}

// JenkinsConfigFile represents the Jenkins CLI config file
type JenkinsConfigFile struct {
	Current        string          `yaml:"current"`
	JenkinsServers []JenkinsConfig `yaml:"jenkins_servers"`
}

// JenkinsClient is a client for interacting with a Jenkins server
type JenkinsClient struct {
	client     *http.Client
	config     *JenkinsConfig
	configPath string
	mutex      sync.Mutex
}

// NewClient creates a new JenkinsClient with the given config
func NewClient(configPath string) (*JenkinsClient, error) {
	if configPath == "" {
		// Default to ~/.jenkins-cli.yaml if no config file is provided
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %v", err)
		}
		configPath = filepath.Join(homeDir, ".jenkins-cli.yaml")
	}

	// Read and parse the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var configFile JenkinsConfigFile
	if err := yaml.Unmarshal(data, &configFile); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	// Find the current server configuration
	var serverConfig *JenkinsConfig
	for i, server := range configFile.JenkinsServers {
		if server.Name == configFile.Current {
			serverConfig = &configFile.JenkinsServers[i]
			break
		}
	}

	if serverConfig == nil {
		return nil, fmt.Errorf("no Jenkins server found in config")
	}

	// Create an HTTP client with the appropriate settings
	transport := &http.Transport{}

	// Configure TLS if needed
	if serverConfig.InsecureSkipVerify {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	// Configure proxy if specified
	if serverConfig.Proxy != "" {
		proxyURL, err := url.Parse(serverConfig.Proxy)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy URL: %v", err)
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	return &JenkinsClient{
		client:     client,
		config:     serverConfig,
		configPath: configPath,
	}, nil
}

// GetServerInfo retrieves information about the Jenkins server
func (c *JenkinsClient) GetServerInfo(ctx context.Context) (*ServerInfo, error) {
	// Lock to ensure thread safety
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Create API URL for server info
	apiURL := fmt.Sprintf("%s/api/json", c.config.URL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.SetBasicAuth(c.config.Username, c.config.Token)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Jenkins: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var serverData struct {
		NodeDescription string `json:"nodeDescription"`
		Mode            string `json:"mode"`
		NodeName        string `json:"nodeName"`
		Version         string `json:"version"`
	}

	if err := json.Unmarshal(bodyBytes, &serverData); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	serverInfo := &ServerInfo{
		URL:       c.config.URL,
		Connected: true,
		Username:  c.config.Username,
		Version:   serverData.Version,
		Mode:      serverData.Mode,
	}

	return serverInfo, nil
}

// GetJobs retrieves a list of all jobs from the Jenkins server
func (c *JenkinsClient) GetJobs(ctx context.Context) ([]Job, error) {
	// Lock to ensure thread safety
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Create API URL for jobs
	apiURL := fmt.Sprintf("%s/api/json?tree=jobs[name,url,color,description]", c.config.URL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.SetBasicAuth(c.config.Username, c.config.Token)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var jobsResponse struct {
		Jobs []struct {
			Name        string `json:"name"`
			URL         string `json:"url"`
			Color       string `json:"color"`
			Description string `json:"description"`
			Class       string `json:"_class"`
		} `json:"jobs"`
	}

	if err := json.Unmarshal(bodyBytes, &jobsResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	// Convert the Jenkins API jobs to our model
	var jobs []Job
	for _, jobData := range jobsResponse.Jobs {
		job := Job{
			Name:        jobData.Name,
			URL:         jobData.URL,
			Class:       jobData.Class,
			Color:       jobData.Color,
			Description: jobData.Description,
		}

		// Determine the job status based on the color
		switch jobData.Color {
		case "blue", "blue_anime":
			job.Status = "success"
			job.InProgress = jobData.Color == "blue_anime"
		case "red", "red_anime":
			job.Status = "failure"
			job.InProgress = jobData.Color == "red_anime"
		case "yellow", "yellow_anime":
			job.Status = "unstable"
			job.InProgress = jobData.Color == "yellow_anime"
		case "grey", "grey_anime", "disabled", "disabled_anime":
			job.Status = "disabled"
			job.InProgress = jobData.Color == "grey_anime" || jobData.Color == "disabled_anime"
		case "aborted", "aborted_anime":
			job.Status = "aborted"
			job.InProgress = jobData.Color == "aborted_anime"
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

	// URL encode the job name
	encodedJobName := url.PathEscape(jobName)

	// Create API URL for job details
	apiURL := fmt.Sprintf("%s/job/%s/api/json?depth=1", c.config.URL, encodedJobName)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.SetBasicAuth(c.config.Username, c.config.Token)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get job details: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var jobDetails struct {
		Name        string `json:"name"`
		URL         string `json:"url"`
		Description string `json:"description"`
		Buildable   bool   `json:"buildable"`
		Builds      []struct {
			Number int    `json:"number"`
			URL    string `json:"url"`
		} `json:"builds"`
		LastBuild *struct {
			Number int    `json:"number"`
			URL    string `json:"url"`
		} `json:"lastBuild"`
	}

	if err := json.Unmarshal(bodyBytes, &jobDetails); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
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

	// URL encode the job name
	encodedJobName := url.PathEscape(jobName)

	// Create API URL for build details
	apiURL := fmt.Sprintf("%s/job/%s/%d/api/json", c.config.URL, encodedJobName, buildNumber)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.SetBasicAuth(c.config.Username, c.config.Token)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get build details: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var buildData struct {
		Number      int    `json:"number"`
		URL         string `json:"url"`
		Timestamp   int64  `json:"timestamp"`
		Duration    int64  `json:"duration"`
		Building    bool   `json:"building"`
		Result      string `json:"result"`
		Description string `json:"description"`
		Actions     []struct {
			Parameters []struct {
				Name  string      `json:"name"`
				Value interface{} `json:"value"`
			} `json:"parameters"`
		} `json:"actions"`
	}

	if err := json.Unmarshal(bodyBytes, &buildData); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	// Create a BuildDetail object
	build := &BuildDetail{
		Number:      buildData.Number,
		URL:         buildData.URL,
		StartTime:   buildData.Timestamp,
		Duration:    buildData.Duration,
		Building:    buildData.Building,
		Result:      buildData.Result,
		Description: buildData.Description,
		Parameters:  make(map[string]string),
	}

	// Extract parameters if available
	if len(buildData.Actions) > 0 {
		for _, action := range buildData.Actions {
			if len(action.Parameters) > 0 {
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

	// URL encode the job name
	encodedJobName := url.PathEscape(jobName)

	// Create API URL for build log
	apiURL := fmt.Sprintf("%s/job/%s/%d/consoleText", c.config.URL, encodedJobName, buildNumber)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.SetBasicAuth(c.config.Username, c.config.Token)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get build log: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	return string(bodyBytes), nil
}

// TriggerBuild starts a build for a specific job
func (c *JenkinsClient) TriggerBuild(ctx context.Context, jobName string, parameters map[string]string) error {
	// Lock to ensure thread safety
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// URL encode the job name
	encodedJobName := url.PathEscape(jobName)

	var apiURL string
	var req *http.Request
	var err error

	if len(parameters) > 0 {
		// Create API URL for triggering a build with parameters
		apiURL = fmt.Sprintf("%s/job/%s/buildWithParameters", c.config.URL, encodedJobName)

		// Build the form values
		formValues := url.Values{}
		for key, value := range parameters {
			formValues.Add(key, value)
		}

		// Create request with form body
		req, err = http.NewRequestWithContext(ctx, http.MethodPost, apiURL, strings.NewReader(formValues.Encode()))
		if err != nil {
			return fmt.Errorf("failed to create request: %v", err)
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		// Create API URL for triggering a build without parameters
		apiURL = fmt.Sprintf("%s/job/%s/build", c.config.URL, encodedJobName)

		// Create request without body
		req, err = http.NewRequestWithContext(ctx, http.MethodPost, apiURL, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %v", err)
		}
	}

	req.SetBasicAuth(c.config.Username, c.config.Token)

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to trigger build: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("failed to trigger build, status code: %d", resp.StatusCode)
	}

	return nil
}
