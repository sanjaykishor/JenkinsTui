package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// JenkinsServer represents a Jenkins server configuration
type JenkinsServer struct {
	Name               string `yaml:"name"`
	URL                string `yaml:"url"`
	Username           string `yaml:"username"`
	Token              string `yaml:"token"`
	Proxy              string `yaml:"proxy"`
	InsecureSkipVerify bool   `yaml:"insecureSkipVerify"`
}

// UISettings represents the UI configuration
type UISettings struct {
	Theme           string `yaml:"theme"`
	RefreshInterval int    `yaml:"refreshInterval"`
	MaxLogLines     int    `yaml:"maxLogLines"`
	CompactMode     bool   `yaml:"compactMode"`
}

// KeyBindings represents custom keybindings
type KeyBindings struct {
	Quit      string `yaml:"quit"`
	Help      string `yaml:"help"`
	Dashboard string `yaml:"dashboard"`
	Jobs      string `yaml:"jobs"`
	Builds    string `yaml:"builds"`
	Nodes     string `yaml:"nodes"`
}

// Config represents the application configuration
type Config struct {
	Current        string          `yaml:"current"`
	JenkinsServers []JenkinsServer `yaml:"jenkins_servers"`
	UI             UISettings      `yaml:"ui"`
	KeyBindings    KeyBindings     `yaml:"keybindings"`
}

// Manager handles configuration loading and saving
type Manager struct {
	Config     *Config
	ConfigPath string
}

// DefaultConfig creates a default configuration
func DefaultConfig() *Config {
	return &Config{
		Current: "default",
		JenkinsServers: []JenkinsServer{
			{
				Name:               "default",
				URL:                "http://localhost:8080",
				Username:           "",
				Token:              "",
				Proxy:              "",
				InsecureSkipVerify: false,
			},
		},
		UI: UISettings{
			Theme:           "default",
			RefreshInterval: 30,
			MaxLogLines:     1000,
			CompactMode:     false,
		},
		KeyBindings: KeyBindings{
			Quit:      "q",
			Help:      "?",
			Dashboard: "d",
			Jobs:      "j",
			Builds:    "b",
			Nodes:     "n",
		},
	}
}

// New creates a new config manager with the specified config file
func New(configPath string) *Manager {
	return &Manager{
		ConfigPath: configPath,
	}
}

// Load loads the configuration from the file
func (m *Manager) Load() error {
	// Check if the config file exists
	if _, err := os.Stat(m.ConfigPath); os.IsNotExist(err) {
		// Create default config
		m.Config = DefaultConfig()
		return m.Save()
	}

	// Read the config file
	data, err := os.ReadFile(m.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}

	// Parse the config
	config := &Config{}
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return fmt.Errorf("failed to parse config file: %v", err)
	}

	m.Config = config
	return nil
}

// Save saves the configuration to the file
func (m *Manager) Save() error {
	// Create the directory if it doesn't exist
	dir := filepath.Dir(m.ConfigPath)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	// Marshal the config
	data, err := yaml.Marshal(m.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	// Write the config file
	err = os.WriteFile(m.ConfigPath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}

// GetCurrentServer returns the currently selected Jenkins server
func (m *Manager) GetCurrentServer() *JenkinsServer {
	if m.Config == nil {
		return nil
	}

	for _, server := range m.Config.JenkinsServers {
		if server.Name == m.Config.Current {
			return &server
		}
	}

	// If no current server is found, use the first one if available
	if len(m.Config.JenkinsServers) > 0 {
		return &m.Config.JenkinsServers[0]
	}

	return nil
}

// SetCurrentServer sets the current Jenkins server
func (m *Manager) SetCurrentServer(name string) error {
	if m.Config == nil {
		return fmt.Errorf("config not loaded")
	}

	// Check if the server exists
	found := false
	for _, server := range m.Config.JenkinsServers {
		if server.Name == name {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("server %q not found", name)
	}

	m.Config.Current = name
	return m.Save()
}

// AddServer adds a new Jenkins server
func (m *Manager) AddServer(server JenkinsServer) error {
	if m.Config == nil {
		return fmt.Errorf("config not loaded")
	}

	// Check if the server already exists
	for i, s := range m.Config.JenkinsServers {
		if s.Name == server.Name {
			// Update the existing server
			m.Config.JenkinsServers[i] = server
			return m.Save()
		}
	}

	// Add the new server
	m.Config.JenkinsServers = append(m.Config.JenkinsServers, server)
	return m.Save()
}

// RemoveServer removes a Jenkins server
func (m *Manager) RemoveServer(name string) error {
	if m.Config == nil {
		return fmt.Errorf("config not loaded")
	}

	// Check if the server exists
	found := false
	for i, server := range m.Config.JenkinsServers {
		if server.Name == name {
			// Remove the server
			m.Config.JenkinsServers = append(m.Config.JenkinsServers[:i], m.Config.JenkinsServers[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("server %q not found", name)
	}

	// If the removed server was the current one, set the current to the first server
	if m.Config.Current == name && len(m.Config.JenkinsServers) > 0 {
		m.Config.Current = m.Config.JenkinsServers[0].Name
	}

	return m.Save()
}
