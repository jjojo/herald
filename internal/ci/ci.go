package ci

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"herald/internal/config"
	"herald/internal/version"
)

// Integrator handles CI/CD integrations
type Integrator struct {
	config *config.Config
	client *http.Client
}

// ReleaseInfo contains information about a release for CI integration
type ReleaseInfo struct {
	Version     string            `json:"version"`
	Tag         string            `json:"tag"`
	Changelog   string            `json:"changelog"`
	Repository  string            `json:"repository"`
	Branch      string            `json:"branch"`
	CommitHash  string            `json:"commit_hash"`
	ReleaseDate time.Time         `json:"release_date"`
	Metadata    map[string]string `json:"metadata"`
}

// NewIntegrator creates a new CI integrator
func NewIntegrator(cfg *config.Config) *Integrator {
	return &Integrator{
		config: cfg,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// TriggerRelease triggers a CI pipeline for a release
func (i *Integrator) TriggerRelease(releaseInfo *ReleaseInfo) error {
	if !i.config.CI.Enabled || !i.config.CI.TriggerOnRelease {
		return nil // CI integration is disabled
	}

	switch i.config.CI.Provider {
	case "github":
		return i.triggerGitHubRelease(releaseInfo)
	case "gitlab":
		return i.triggerGitLabPipeline(releaseInfo)
	default:
		return fmt.Errorf("unsupported CI provider: %s (supported: github, gitlab)", i.config.CI.Provider)
	}
}

// triggerGitHubRelease creates a GitHub release
func (i *Integrator) triggerGitHubRelease(releaseInfo *ReleaseInfo) error {
	// Skip if GitHub release creation is disabled
	if !i.config.CI.GitHub.CreateRelease {
		return nil
	}

	// Get repository from config
	repository := i.config.CI.GitHub.Repository
	if repository == "" {
		return fmt.Errorf("GitHub repository is required for release creation")
	}

	// Get access token from config or environment
	accessToken := i.config.CI.GitHub.AccessToken
	if accessToken == "" {
		accessToken = os.Getenv("GITHUB_TOKEN")
	}
	if accessToken == "" {
		return fmt.Errorf("GitHub access token is required for release creation (set in config or GITHUB_TOKEN env var)")
	}

	return i.createGitHubRelease(releaseInfo, repository, accessToken)
}

// triggerGitLabPipeline triggers a GitLab CI pipeline and creates a GitLab release
func (i *Integrator) triggerGitLabPipeline(releaseInfo *ReleaseInfo) error {
	// Create GitLab release
	err := i.createGitLabRelease(releaseInfo)
	if err != nil {
		// Log but don't fail
		fmt.Printf("Warning: Failed to create GitLab release: %v\n", err)
	}

	return nil
}

// createGitHubRelease creates a release using GitHub's Release API
func (i *Integrator) createGitHubRelease(releaseInfo *ReleaseInfo, repository, accessToken string) error {
	// GitHub Release API URL
	releaseURL := fmt.Sprintf("https://api.github.com/repos/%s/releases", repository)

	// Create release payload
	releasePayload := map[string]interface{}{
		"tag_name":         releaseInfo.Tag,
		"target_commitish": releaseInfo.Branch,
		"name":            fmt.Sprintf("Release %s", releaseInfo.Version),
		"body":            releaseInfo.Changelog,
		"draft":           false,
		"prerelease":      false,
	}

	// Marshal payload
	jsonData, err := json.Marshal(releasePayload)
	if err != nil {
		return fmt.Errorf("failed to marshal GitHub release payload: %w", err)
	}

	// Create request
	req, err := http.NewRequest("POST", releaseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create GitHub release request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Authorization", fmt.Sprintf("token %s", accessToken))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Herald/1.0")

	// Send request
	resp, err := i.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send GitHub release request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("GitHub release creation failed with status: %d", resp.StatusCode)
	}

	return nil
}

// createGitLabRelease creates a release using GitLab's Release API
func (i *Integrator) createGitLabRelease(releaseInfo *ReleaseInfo) error {
	// Skip if GitLab release creation is disabled
	if !i.config.CI.GitLab.CreateRelease {
		return nil
	}

	// Get project ID from config
	projectID := i.config.CI.GitLab.ProjectID
	if projectID == "" {
		return fmt.Errorf("GitLab project ID is required for release creation")
	}

	// Get access token from config or environment
	accessToken := i.config.CI.GitLab.AccessToken
	if accessToken == "" {
		accessToken = os.Getenv("GITLAB_ACCESS_TOKEN")
	}
	if accessToken == "" {
		return fmt.Errorf("GitLab access token is required for release creation (set in config or GITLAB_ACCESS_TOKEN env var)")
	}

	// GitLab Release API URL
	releaseURL := fmt.Sprintf("https://gitlab.com/api/v4/projects/%s/releases", projectID)

	// Create release payload
	releasePayload := map[string]interface{}{
		"name":        fmt.Sprintf("Release %s", releaseInfo.Version),
		"tag_name":    releaseInfo.Tag,
		"description": releaseInfo.Changelog,
		"released_at": releaseInfo.ReleaseDate.Format(time.RFC3339),
	}

	// Marshal payload
	jsonData, err := json.Marshal(releasePayload)
	if err != nil {
		return fmt.Errorf("failed to marshal GitLab release payload: %w", err)
	}

	// Create request
	req, err := http.NewRequest("POST", releaseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create GitLab release request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("User-Agent", "Herald/1.0")

	// Send request
	resp, err := i.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send GitLab release request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("GitLab release creation failed with status: %d", resp.StatusCode)
	}

	return nil
}

// CreateReleaseInfo creates release information from version and other data
func (i *Integrator) CreateReleaseInfo(ver *version.Version, changelog, repository, branch, commitHash string) *ReleaseInfo {
	return &ReleaseInfo{
		Version:     ver.String(),
		Tag:         ver.String(),
		Changelog:   changelog,
		Repository:  repository,
		Branch:      branch,
		CommitHash:  commitHash,
		ReleaseDate: time.Now(),
		Metadata: map[string]string{
			"herald_version": "1.0.0", // Herald tool version
			"provider":       i.config.CI.Provider,
		},
	}
}

// ValidateConfiguration validates the CI configuration
func (i *Integrator) ValidateConfiguration() error {
	if !i.config.CI.Enabled {
		return nil // No validation needed if disabled
	}

	if i.config.CI.Provider == "" {
		return fmt.Errorf("CI provider must be specified when CI is enabled")
	}

	supportedProviders := []string{"github", "gitlab"}
	validProvider := false
	for _, provider := range supportedProviders {
		if i.config.CI.Provider == provider {
			validProvider = true
			break
		}
	}

	if !validProvider {
		return fmt.Errorf("unsupported CI provider: %s (supported: %v)", i.config.CI.Provider, supportedProviders)
	}

	// Validate provider-specific configuration
	switch i.config.CI.Provider {
	case "github":
		if i.config.CI.GitHub.CreateRelease && i.config.CI.GitHub.Repository == "" {
			return fmt.Errorf("GitHub repository is required when create_release is enabled")
		}
	case "gitlab":
		if i.config.CI.GitLab.CreateRelease && i.config.CI.GitLab.ProjectID == "" {
			return fmt.Errorf("GitLab project ID is required when create_release is enabled")
		}
	}

	return nil
}

// IsEnabled returns true if CI integration is enabled
func (i *Integrator) IsEnabled() bool {
	return i.config.CI.Enabled
}

// GetProvider returns the configured CI provider
func (i *Integrator) GetProvider() string {
	return i.config.CI.Provider
}

// SetCustomClient allows setting a custom HTTP client (useful for testing)
func (i *Integrator) SetCustomClient(client *http.Client) {
	i.client = client
}

// AddMetadata adds custom metadata to release info
func (ri *ReleaseInfo) AddMetadata(key, value string) {
	if ri.Metadata == nil {
		ri.Metadata = make(map[string]string)
	}
	ri.Metadata[key] = value
} 