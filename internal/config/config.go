package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Services  []ServiceConfig  `mapstructure:"services"`
	Detection DetectionConfig  `mapstructure:"detection"`
	Git       GitConfig        `mapstructure:"git"`
	Claude    ClaudeConfig     `mapstructure:"claude"`
}

// ServiceConfig represents a service configuration
type ServiceConfig struct {
	Name     string `mapstructure:"name"`
	LogPath  string `mapstructure:"log_path"`
	Format   string `mapstructure:"format"`
	RepoPath string `mapstructure:"repo_path"`
}

// DetectionConfig represents error detection configuration
type DetectionConfig struct {
	Patterns         []string            `mapstructure:"patterns"`
	SeverityKeywords SeverityKeywords    `mapstructure:"severity_keywords"`
}

// SeverityKeywords maps severity levels to keywords
type SeverityKeywords struct {
	Critical []string `mapstructure:"critical"`
	High     []string `mapstructure:"high"`
	Medium   []string `mapstructure:"medium"`
}

// GitConfig represents Git configuration
type GitConfig struct {
	BranchPrefix string `mapstructure:"branch_prefix"`
	CommitPrefix string `mapstructure:"commit_prefix"`
}

// ClaudeConfig represents Claude AI configuration
type ClaudeConfig struct {
	Model     string `mapstructure:"model"`
	MaxTokens int    `mapstructure:"max_tokens"`
	APIKey    string `mapstructure:"api_key"`
}

// Load loads the configuration from the specified file
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set config file
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		// Look for config in common locations
		v.SetConfigName("tracehub")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("$HOME/.config/tracehub")
		v.AddConfigPath("/etc/tracehub")
	}

	// Environment variables override
	v.AutomaticEnv()
	v.SetEnvPrefix("TRACEHUB")

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Override API key from environment if set
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		cfg.Claude.APIKey = apiKey
	}

	// Validate config
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if len(c.Services) == 0 {
		return fmt.Errorf("at least one service must be configured")
	}

	for _, svc := range c.Services {
		if svc.Name == "" {
			return fmt.Errorf("service name cannot be empty")
		}
		if svc.LogPath == "" {
			return fmt.Errorf("log_path cannot be empty for service %s", svc.Name)
		}
	}

	if c.Claude.APIKey == "" {
		return fmt.Errorf("ANTHROPIC_API_KEY must be set")
	}

	// Set defaults
	if c.Claude.Model == "" {
		c.Claude.Model = "claude-sonnet-4-20250514"
	}
	if c.Claude.MaxTokens == 0 {
		c.Claude.MaxTokens = 4000
	}
	if c.Git.BranchPrefix == "" {
		c.Git.BranchPrefix = "autofix/"
	}
	if c.Git.CommitPrefix == "" {
		c.Git.CommitPrefix = "[AutoFix]"
	}

	return nil
}

// GetService returns the configuration for a specific service
func (c *Config) GetService(name string) *ServiceConfig {
	for _, svc := range c.Services {
		if svc.Name == name {
			return &svc
		}
	}
	return nil
}
