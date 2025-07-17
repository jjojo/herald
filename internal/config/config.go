package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	Types                   map[string]CommitType `yaml:"types"`
	BreakingChangeKeywords []string              `yaml:"breaking_change_keywords"`
}

// CommitType defines a commit type with its display title and semver bump level
type CommitType struct {
	Title  string `yaml:"title"`
	Semver string `yaml:"semver"` // "major", "minor", "patch", "none"
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
	Enabled          bool             `yaml:"enabled"`
	Provider         string           `yaml:"provider"`
	TriggerOnRelease bool             `yaml:"trigger_on_release"`
	GitLab           GitLabConfig     `yaml:"gitlab,omitempty"`
	GitHub           GitHubConfig     `yaml:"github,omitempty"`
}

// GitLabConfig holds GitLab-specific CI settings
type GitLabConfig struct {
	ProjectID     string `yaml:"project_id"`
	AccessToken   string `yaml:"access_token"`
	CreateRelease bool   `yaml:"create_release"`
}

// GitHubConfig holds GitHub-specific CI settings
type GitHubConfig struct {
	Repository    string `yaml:"repository"`    // owner/repo
	AccessToken   string `yaml:"access_token"`
	CreateRelease bool   `yaml:"create_release"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Version: VersionConfig{
			Initial: "0.1.0",
			Prefix:  "v",
		},
		Commits: CommitsConfig{
			Types: map[string]CommitType{
				"feat": {
					Title:  "Features",
					Semver: "minor",
				},
				"fix": {
					Title:  "Bug Fixes",
					Semver: "patch",
				},
				"docs": {
					Title:  "Documentation",
					Semver: "none",
				},
				"style": {
					Title:  "Styles",
					Semver: "none",
				},
				"refactor": {
					Title:  "Code Refactoring",
					Semver: "none",
				},
				"test": {
					Title:  "Tests",
					Semver: "none",
				},
				"chore": {
					Title:  "Chores",
					Semver: "none",
				},
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
			GitHub: GitHubConfig{
				Repository:    "",
				AccessToken:   "",
				CreateRelease: true,
			},
			GitLab: GitLabConfig{
				ProjectID:     "",
				AccessToken:   "",
				CreateRelease: true,
			},
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

	// Generate documented config content
	configContent := generateDocumentedConfig()

	// Write to file
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("Initialized %s with default configuration\n", configFile)
	return nil
}

// generateDocumentedConfig creates a YAML config with comprehensive comments
func generateDocumentedConfig() string {
	return `# Herald Configuration File
# This file configures Herald's behavior for automated release management
# using conventional commits and semantic versioning.

# Version Configuration
version:
  # The initial version to use when no git tags exist
  # Must be a valid semantic version (e.g., "0.1.0", "1.0.0")
  initial: "0.1.0"
  
  # Prefix for git tags (e.g., "v" creates tags like "v1.0.0")
  # Set to empty string "" for no prefix
  prefix: "v"

# Conventional Commits Configuration
commits:
  # Define commit types, their display titles, and version bump behavior
  # Each commit type can specify:
  #   title: Display name in changelog
  #   semver: Version bump level ("major", "minor", "patch", "none")
  types:
    # New features - typically bump minor version
    feat:
      title: "Features"
      semver: "minor"
    
    # Bug fixes - typically bump patch version
    fix:
      title: "Bug Fixes"
      semver: "patch"
    
    # Documentation changes - no version bump by default
    docs:
      title: "Documentation"
      semver: "none"
    
    # Code style changes (formatting, etc.) - no version bump
    style:
      title: "Styles"
      semver: "none"
    
    # Code refactoring without functional changes - no version bump
    refactor:
      title: "Code Refactoring"
      semver: "none"
    
    # Test additions or modifications - no version bump
    test:
      title: "Tests"
      semver: "none"
    
    # Build process, tooling, dependencies - no version bump by default
    chore:
      title: "Chores"
      semver: "none"
  
  # Keywords that indicate breaking changes (triggers major version bump)
  # These can appear in commit body or footer
  breaking_change_keywords:
    - "BREAKING CHANGE"
    - "BREAKING-CHANGE"

# Changelog Configuration
changelog:
  # Path to the changelog file (relative to repository root)
  file: "CHANGELOG.md"
  
  # Template to use for changelog generation
  # Currently only "default" is supported
  template: "default"
  
  # Whether to include all commit types in changelog
  # true: Include all configured commit types
  # false: Only include "feat" and "fix" (plus breaking changes)
  include_all: false

# Git Configuration
git:
  # Message template for git tags
  # {version} will be replaced with the actual version
  tag_message: "Release {version}"
  
  # Whether to commit the changelog file after updating it
  commit_changelog: true
  
  # Commit message template when committing changelog
  # {version} will be replaced with the actual version
  commit_message: "chore: update changelog for {version}"

# CI/CD Integration (Optional)
ci:
  # Enable or disable CI integration
  enabled: false
  
  # CI provider type
  # Supported values: "github", "gitlab"
  provider: "github"
  
  # Whether to trigger CI pipeline after creating a release
  trigger_on_release: true
  
  # GitHub-specific configuration (only used when provider is "github")
  github:
    # GitHub repository in "owner/repo" format
    repository: ""
    
    # GitHub access token with repo permissions
    # Can also be set via GITHUB_TOKEN environment variable
    access_token: ""
    
    # Whether to create GitHub releases automatically
    create_release: true
  
  # GitLab-specific configuration (only used when provider is "gitlab")
  gitlab:
    # GitLab project ID (numeric ID or "group/project-name")
    project_id: ""
    
    # GitLab access token with API permissions
    # Can also be set via GITLAB_ACCESS_TOKEN environment variable
    access_token: ""
    
    # Whether to create GitLab releases automatically
    create_release: true
`
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

	// Validate semver levels for commit types
	validSemverLevels := []string{"major", "minor", "patch", "none"}
	for commitType, typeConfig := range c.Commits.Types {
		if typeConfig.Title == "" {
			return fmt.Errorf("commit type '%s' must have a title", commitType)
		}
		
		validSemver := false
		for _, validLevel := range validSemverLevels {
			if strings.ToLower(typeConfig.Semver) == validLevel {
				validSemver = true
				break
			}
		}
		if !validSemver {
			return fmt.Errorf("commit type '%s' has invalid semver level '%s' (must be: major, minor, patch, or none)", commitType, typeConfig.Semver)
		}
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