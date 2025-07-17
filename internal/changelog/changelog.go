package changelog

import (
	"fmt"
	"os"
	"strings"
	"time"

	"herald/internal/commits"
	"herald/internal/config"
	"herald/internal/version"
)

// Generator handles changelog generation
type Generator struct {
	config *config.Config
}

// Release represents a release entry in the changelog
type Release struct {
	Version     *version.Version
	Date        time.Time
	Commits     []*commits.ConventionalCommit
	GroupedCommits map[string][]*commits.ConventionalCommit
	BreakingChanges []*commits.ConventionalCommit
}

// NewGenerator creates a new changelog generator
func NewGenerator(cfg *config.Config) *Generator {
	return &Generator{
		config: cfg,
	}
}

// GenerateRelease creates a release entry from commits
func (g *Generator) GenerateRelease(ver *version.Version, conventionalCommits []*commits.ConventionalCommit) *Release {
	parser := commits.NewParser(g.config)
	
	// Filter commits for changelog
	filteredCommits := parser.FilterCommitsForChangelog(conventionalCommits)
	
	// Group commits by type
	groupedCommits := parser.GroupCommitsByType(filteredCommits)
	
	// Get breaking changes
	breakingChanges := parser.GetBreakingChanges(conventionalCommits)

	return &Release{
		Version:         ver,
		Date:            time.Now(),
		Commits:         filteredCommits,
		GroupedCommits:  groupedCommits,
		BreakingChanges: breakingChanges,
	}
}

// FormatRelease formats a release entry as markdown
func (g *Generator) FormatRelease(release *Release) string {
	var builder strings.Builder
	parser := commits.NewParser(g.config)

	// Release header
	builder.WriteString(fmt.Sprintf("## [%s]", release.Version.String()))
	builder.WriteString(fmt.Sprintf(" - %s\n\n", release.Date.Format("2006-01-02")))

	// Breaking changes section (if any)
	if len(release.BreakingChanges) > 0 {
		builder.WriteString("### ⚠ BREAKING CHANGES\n\n")
		for _, commit := range release.BreakingChanges {
			builder.WriteString(fmt.Sprintf("* %s", commit.Description))
			if commit.Scope != "" {
				builder.WriteString(fmt.Sprintf(" (**%s**)", commit.Scope))
			}
			builder.WriteString("\n")
			
			// Add breaking change details if available
			for _, bc := range commit.BreakingChanges {
				if bc != "" {
					builder.WriteString(fmt.Sprintf("  %s\n", bc))
				}
			}
		}
		builder.WriteString("\n")
	}

	// Sort commit types for consistent ordering
	sortedTypes := parser.SortCommitsByType(release.GroupedCommits)

	// Generate sections for each commit type
	for _, commitType := range sortedTypes {
		commits := release.GroupedCommits[commitType]
		if len(commits) == 0 {
			continue
		}

		// Section header
		typeTitle := parser.GetCommitTypeTitle(commitType)
		builder.WriteString(fmt.Sprintf("### %s\n\n", typeTitle))

		// List commits
		for _, commit := range commits {
			builder.WriteString("* ")
			
			// Add scope if present
			if commit.Scope != "" {
				builder.WriteString(fmt.Sprintf("**%s:** ", commit.Scope))
			}
			
			builder.WriteString(commit.Description)
			
			// Add commit hash (short)
			if len(commit.Original.Hash) >= 7 {
				shortHash := commit.Original.Hash[:7]
				builder.WriteString(fmt.Sprintf(" ([%s])", shortHash))
			}
			
			builder.WriteString("\n")
		}
		builder.WriteString("\n")
	}

	return builder.String()
}

// ReadExistingChangelog reads the existing changelog file
func (g *Generator) ReadExistingChangelog() (string, error) {
	content, err := os.ReadFile(g.config.Changelog.File)
	if os.IsNotExist(err) {
		return "", nil // File doesn't exist, return empty string
	}
	if err != nil {
		return "", fmt.Errorf("failed to read changelog file: %w", err)
	}
	return string(content), nil
}

// WriteChangelog writes the changelog to file
func (g *Generator) WriteChangelog(content string) error {
	err := os.WriteFile(g.config.Changelog.File, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write changelog file: %w", err)
	}
	return nil
}

// PrependRelease adds a new release to the beginning of the changelog
func (g *Generator) PrependRelease(release *Release) error {
	// Read existing changelog
	existingContent, err := g.ReadExistingChangelog()
	if err != nil {
		return err
	}

	// Format the new release
	newRelease := g.FormatRelease(release)

	// Create the new changelog content
	var newContent strings.Builder

	// Add changelog header if it doesn't exist
	if existingContent == "" || !strings.Contains(existingContent, "# Changelog") {
		newContent.WriteString("# Changelog\n\n")
		newContent.WriteString("All notable changes to this project will be documented in this file.\n\n")
		newContent.WriteString("The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),\n")
		newContent.WriteString("and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).\n\n")
	}

	// Add the new release
	newContent.WriteString(newRelease)

	// Add existing content (but skip the header if we just added it)
	if existingContent != "" {
		if strings.Contains(existingContent, "# Changelog") {
			// Find the end of the header section
			lines := strings.Split(existingContent, "\n")
			var contentStartIndex int
			headerEndFound := false
			
			for i, line := range lines {
				if headerEndFound && strings.TrimSpace(line) != "" {
					contentStartIndex = i
					break
				}
				if strings.HasPrefix(line, "## ") {
					contentStartIndex = i
					break
				}
				if strings.Contains(line, "Semantic Versioning") {
					headerEndFound = true
				}
			}
			
			if contentStartIndex > 0 {
				remainingContent := strings.Join(lines[contentStartIndex:], "\n")
				newContent.WriteString(remainingContent)
			}
		} else {
			newContent.WriteString(existingContent)
		}
	}

	// Write the new changelog
	return g.WriteChangelog(newContent.String())
}

// GenerateFullChangelog generates a complete changelog from scratch
func (g *Generator) GenerateFullChangelog(releases []*Release) error {
	var content strings.Builder

	// Header
	content.WriteString("# Changelog\n\n")
	content.WriteString("All notable changes to this project will be documented in this file.\n\n")
	content.WriteString("The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),\n")
	content.WriteString("and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).\n\n")

	// Add each release
	for _, release := range releases {
		content.WriteString(g.FormatRelease(release))
	}

	return g.WriteChangelog(content.String())
}

// ValidateChangelogPath checks if the changelog path is valid
func (g *Generator) ValidateChangelogPath() error {
	dir := strings.TrimSuffix(g.config.Changelog.File, "/"+g.config.Changelog.File)
	if dir != g.config.Changelog.File {
		// Check if directory exists
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return fmt.Errorf("changelog directory does not exist: %s", dir)
		}
	}
	return nil
}

// PreviewRelease returns a preview of what the release would look like
func (g *Generator) PreviewRelease(release *Release) string {
	preview := "=== CHANGELOG PREVIEW ===\n\n"
	preview += g.FormatRelease(release)
	preview += "\n=== END PREVIEW ===\n"
	return preview
}

// GetChangelogStats returns statistics about the changelog
func (g *Generator) GetChangelogStats(release *Release) map[string]int {
	stats := make(map[string]int)
	
	for commitType, commits := range release.GroupedCommits {
		stats[commitType] = len(commits)
	}
	
	stats["total"] = len(release.Commits)
	stats["breaking_changes"] = len(release.BreakingChanges)
	
	return stats
}

// HasSignificantChanges checks if the release has significant changes worth releasing
func (g *Generator) HasSignificantChanges(release *Release) bool {
	// Check for breaking changes
	if len(release.BreakingChanges) > 0 {
		return true
	}
	
	// Check for features or fixes
	for commitType, commits := range release.GroupedCommits {
		if (commitType == "feat" || commitType == "fix") && len(commits) > 0 {
			return true
		}
	}
	
	return false
}

// FormatCommitList formats a list of commits for display
func (g *Generator) FormatCommitList(commits []*commits.ConventionalCommit) string {
	var builder strings.Builder
	
	for i, commit := range commits {
		builder.WriteString(fmt.Sprintf("%d. ", i+1))
		
		if commit.Scope != "" {
			builder.WriteString(fmt.Sprintf("**%s:** ", commit.Scope))
		}
		
		builder.WriteString(commit.Description)
		
		if commit.IsBreakingChange {
			builder.WriteString(" ⚠️")
		}
		
		builder.WriteString("\n")
	}
	
	return builder.String()
} 