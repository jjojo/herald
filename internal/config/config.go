package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the Herald configuration
type Config struct {
	Version   VersionConfig   `yaml:"version"`
	Commits   CommitsConfig   `yaml:"commits"`
	Changelog ChangelogConfig `yaml:"changelog"`
	Git       GitConfig       `yaml:"git"`
	CI        CIConfig        `yaml:"ci"`
}

// VersionConfig holds version-related settings
type VersionConfig struct {
	Initial string `yaml:"initial"`
	Prefix  string `yaml:"prefix"`
}

// CommitsConfig holds conventional commits settings
type CommitsConfig struct {
	Types                   map[string]string `yaml:"types"`
	BreakingChangeKeywords []string          `yaml:"breaking_change_keywords"`
}

// ChangelogConfig holds changelog generation settings
type ChangelogConfig struct {
	File       string `yaml:"file"`
	Template   string `yaml:"template"`
	IncludeAll bool   `yaml:"include_all"`
}

// GitConfig holds git operation settings
type GitConfig struct {
	TagMessage      string `yaml:"tag_message"`
	CommitChangelog bool   `yaml:"commit_changelog"`
	CommitMessage   string `yaml:"commit_message"`
}

// CIConfig holds CI integration settings
type CIConfig struct {
	Enabled          bool   `yaml:"enabled"`
	Provider         string `yaml:"provider"`
	TriggerOnRelease bool   `yaml:"trigger_on_release"`
	WebhookURL       string `yaml:"webhook_url"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Version: VersionConfig{
			Initial: "0.1.0",
			Prefix:  "v",
		},
		Commits: CommitsConfig{
			Types: map[string]string{
				"feat":     "Features",
				"fix":      "Bug Fixes",
				"docs":     "Documentation",
				"style":    "Styles",
				"refactor": "Code Refactoring",
				"test":     "Tests",
				"chore":    "Chores",
			},
			BreakingChangeKeywords: []string{"BREAKING CHANGE", "BREAKING-CHANGE"},
		},
		Changelog: ChangelogConfig{
			File:       "CHANGELOG.md",
			Template:   "default",
			IncludeAll: false,
		},
		Git: GitConfig{
			TagMessage:      "Release {version}",
			CommitChangelog: true,
			CommitMessage:   "chore: update changelog for {version}",
		},
		CI: CIConfig{
			Enabled:          false,
			Provider:         "github",
			TriggerOnRelease: true,
			WebhookURL:       "",
		},
	}
}

// LoadConfig loads configuration from a file or returns default config
func LoadConfig(configFile string) (*Config, error) {
	// Use provided config file or look for .heraldrc
	if configFile == "" {
		configFile = ".heraldrc"
	}

	// Check if config file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	// Read config file
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	config := DefaultConfig()
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}

// InitializeConfig creates a default .heraldrc file
func InitializeConfig() error {
	configFile := ".heraldrc"

	// Check if config file already exists
	if _, err := os.Stat(configFile); err == nil {
		return fmt.Errorf("config file %s already exists", configFile)
	}

	// Create default config
	config := DefaultConfig()

	// Marshal to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("Initialized %s with default configuration\n", configFile)
	return nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Version.Initial == "" {
		return fmt.Errorf("version.initial cannot be empty")
	}

	if c.Changelog.File == "" {
		return fmt.Errorf("changelog.file cannot be empty")
	}

	if len(c.Commits.Types) == 0 {
		return fmt.Errorf("commits.types cannot be empty")
	}

	return nil
}

// GetConfigPath returns the path to the config file
func GetConfigPath(configFile string) string {
	if configFile != "" {
		return configFile
	}

	// Look for .heraldrc in current directory
	if _, err := os.Stat(".heraldrc"); err == nil {
		return ".heraldrc"
	}

	// Look for .heraldrc in home directory
	if home, err := os.UserHomeDir(); err == nil {
		homeConfig := filepath.Join(home, ".heraldrc")
		if _, err := os.Stat(homeConfig); err == nil {
			return homeConfig
		}
	}

	return ".heraldrc" // Default fallback
} 