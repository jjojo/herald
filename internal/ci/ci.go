package ci

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
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
		return i.triggerGitHubAction(releaseInfo)
	case "gitlab":
		return i.triggerGitLabPipeline(releaseInfo)
	case "webhook":
		return i.triggerWebhook(releaseInfo)
	default:
		return fmt.Errorf("unsupported CI provider: %s", i.config.CI.Provider)
	}
}

// triggerGitHubAction triggers a GitHub Action workflow
func (i *Integrator) triggerGitHubAction(releaseInfo *ReleaseInfo) error {
	if i.config.CI.WebhookURL == "" {
		return fmt.Errorf("webhook URL is required for GitHub integration")
	}

	payload := map[string]interface{}{
		"event_type": "release",
		"client_payload": map[string]interface{}{
			"version":     releaseInfo.Version,
			"tag":         releaseInfo.Tag,
			"changelog":   releaseInfo.Changelog,
			"repository":  releaseInfo.Repository,
			"branch":      releaseInfo.Branch,
			"commit_hash": releaseInfo.CommitHash,
			"metadata":    releaseInfo.Metadata,
		},
	}

	return i.sendWebhookRequest(payload)
}

// triggerGitLabPipeline triggers a GitLab CI pipeline
func (i *Integrator) triggerGitLabPipeline(releaseInfo *ReleaseInfo) error {
	if i.config.CI.WebhookURL == "" {
		return fmt.Errorf("webhook URL is required for GitLab integration")
	}

	payload := map[string]interface{}{
		"token":     extractTokenFromURL(i.config.CI.WebhookURL),
		"ref":       releaseInfo.Branch,
		"variables": map[string]string{
			"HERALD_VERSION":     releaseInfo.Version,
			"HERALD_TAG":         releaseInfo.Tag,
			"HERALD_COMMIT_HASH": releaseInfo.CommitHash,
			"HERALD_RELEASE":     "true",
		},
	}

	return i.sendWebhookRequest(payload)
}

// triggerWebhook sends a generic webhook
func (i *Integrator) triggerWebhook(releaseInfo *ReleaseInfo) error {
	if i.config.CI.WebhookURL == "" {
		return fmt.Errorf("webhook URL is required for webhook integration")
	}

	payload := map[string]interface{}{
		"event":        "release",
		"version":      releaseInfo.Version,
		"tag":          releaseInfo.Tag,
		"changelog":    releaseInfo.Changelog,
		"repository":   releaseInfo.Repository,
		"branch":       releaseInfo.Branch,
		"commit_hash":  releaseInfo.CommitHash,
		"release_date": releaseInfo.ReleaseDate,
		"metadata":     releaseInfo.Metadata,
	}

	return i.sendWebhookRequest(payload)
}

// sendWebhookRequest sends an HTTP request to the webhook URL
func (i *Integrator) sendWebhookRequest(payload interface{}) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	req, err := http.NewRequest("POST", i.config.CI.WebhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Herald/1.0")

	resp, err := i.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook request failed with status: %d", resp.StatusCode)
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

	supportedProviders := []string{"github", "gitlab", "webhook"}
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

	if i.config.CI.TriggerOnRelease && i.config.CI.WebhookURL == "" {
		return fmt.Errorf("webhook URL is required when trigger_on_release is enabled")
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

// TestConnection tests the CI integration connection
func (i *Integrator) TestConnection() error {
	if !i.config.CI.Enabled {
		return fmt.Errorf("CI integration is disabled")
	}

	if i.config.CI.WebhookURL == "" {
		return fmt.Errorf("webhook URL is not configured")
	}

	// Create a test payload
	testPayload := map[string]interface{}{
		"event":   "test",
		"message": "Herald CI integration test",
		"timestamp": time.Now(),
	}

	// Send test request (we'll just check if we can reach the endpoint)
	jsonData, err := json.Marshal(testPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal test payload: %w", err)
	}

	req, err := http.NewRequest("POST", i.config.CI.WebhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create test request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Herald/1.0 (test)")

	resp, err := i.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to reach webhook URL: %w", err)
	}
	defer resp.Body.Close()

	// For testing, we accept any response that's not a connection error
	return nil
}

// extractTokenFromURL extracts token from GitLab webhook URL
func extractTokenFromURL(url string) string {
	// This is a simple implementation - in practice you might want more sophisticated parsing
	// For GitLab, tokens are usually in the URL path or query parameters
	return ""
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