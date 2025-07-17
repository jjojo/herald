package commits

import (
	"regexp"
	"strings"

	"herald/internal/config"
	"herald/internal/git"
)

// ConventionalCommit represents a parsed conventional commit
type ConventionalCommit struct {
	Type             string
	Scope            string
	Description      string
	Body             string
	IsBreakingChange bool
	BreakingChanges  []string
	Original         *git.Commit
}

// BumpType represents the type of version bump needed
type BumpType int

const (
	None BumpType = iota
	Patch
	Minor
	Major
)

func (bt BumpType) String() string {
	switch bt {
	case Patch:
		return "patch"
	case Minor:
		return "minor"
	case Major:
		return "major"
	default:
		return "none"
	}
}

// Parser handles parsing of conventional commits
type Parser struct {
	config *config.Config
	regex  *regexp.Regexp
}

// NewParser creates a new conventional commits parser
func NewParser(cfg *config.Config) *Parser {
	// Conventional commit regex pattern
	// Matches: type(scope): description
	pattern := `^(\w+)(?:\(([^)]+)\))?: (.+)$`
	regex := regexp.MustCompile(pattern)

	return &Parser{
		config: cfg,
		regex:  regex,
	}
}

// ParseCommit parses a single git commit into a conventional commit
func (p *Parser) ParseCommit(commit *git.Commit) (*ConventionalCommit, error) {
	cc := &ConventionalCommit{
		Original: commit,
		Body:     commit.Body,
	}

	// Parse the commit subject line
	matches := p.regex.FindStringSubmatch(commit.Subject)
	if len(matches) != 4 {
		// Not a conventional commit, treat as unknown type
		cc.Type = "other"
		cc.Description = commit.Subject
	} else {
		cc.Type = matches[1]
		cc.Scope = matches[2]
		cc.Description = matches[3]
	}

	// Check for breaking changes
	cc.IsBreakingChange = p.hasBreakingChange(commit)
	cc.BreakingChanges = p.extractBreakingChanges(commit)

	return cc, nil
}

// ParseCommits parses multiple git commits
func (p *Parser) ParseCommits(commits []*git.Commit) ([]*ConventionalCommit, error) {
	var result []*ConventionalCommit

	for _, commit := range commits {
		cc, err := p.ParseCommit(commit)
		if err != nil {
			return nil, err
		}
		result = append(result, cc)
	}

	return result, nil
}

// GroupCommitsByType groups conventional commits by their type
func (p *Parser) GroupCommitsByType(commits []*ConventionalCommit) map[string][]*ConventionalCommit {
	groups := make(map[string][]*ConventionalCommit)

	for _, commit := range commits {
		groups[commit.Type] = append(groups[commit.Type], commit)
	}

	return groups
}

// FilterCommitsForChangelog filters commits that should be included in the changelog
func (p *Parser) FilterCommitsForChangelog(commits []*ConventionalCommit) []*ConventionalCommit {
	var result []*ConventionalCommit

	for _, commit := range commits {
		// Include all types if configured to do so
		if p.config.Changelog.IncludeAll {
			result = append(result, commit)
			continue
		}

		// Otherwise, only include certain types
		switch commit.Type {
		case "feat", "fix":
			result = append(result, commit)
		case "docs", "style", "refactor", "test", "chore":
			// Only include if it's a breaking change
			if commit.IsBreakingChange {
				result = append(result, commit)
			}
		default:
			// Include unknown types
			result = append(result, commit)
		}
	}

	return result
}

// CalculateBumpType determines the type of version bump needed based on commits
func (p *Parser) CalculateBumpType(commits []*ConventionalCommit) BumpType {
	bumpType := None

	for _, commit := range commits {
		// Breaking changes always require major bump
		if commit.IsBreakingChange {
			return Major
		}

		// Features require minor bump
		if commit.Type == "feat" && bumpType < Minor {
			bumpType = Minor
		}

		// Bug fixes require patch bump
		if commit.Type == "fix" && bumpType < Patch {
			bumpType = Patch
		}
	}

	return bumpType
}

// hasBreakingChange checks if a commit contains breaking changes
func (p *Parser) hasBreakingChange(commit *git.Commit) bool {
	// Check for exclamation mark in subject (feat!: or feat(scope)!:)
	if strings.Contains(commit.Subject, "!:") {
		return true
	}

	// Check for breaking change keywords in body
	fullMessage := commit.Subject + "\n" + commit.Body
	for _, keyword := range p.config.Commits.BreakingChangeKeywords {
		if strings.Contains(fullMessage, keyword) {
			return true
		}
	}

	return false
}

// extractBreakingChanges extracts breaking change descriptions from commit
func (p *Parser) extractBreakingChanges(commit *git.Commit) []string {
	var breakingChanges []string
	fullMessage := commit.Subject + "\n" + commit.Body

	for _, keyword := range p.config.Commits.BreakingChangeKeywords {
		if strings.Contains(fullMessage, keyword) {
			// Find the line with the keyword and extract the description
			lines := strings.Split(fullMessage, "\n")
			for _, line := range lines {
				if strings.Contains(line, keyword) {
					// Extract text after the keyword
					parts := strings.SplitN(line, keyword, 2)
					if len(parts) > 1 {
						description := strings.TrimSpace(parts[1])
						if strings.HasPrefix(description, ":") {
							description = strings.TrimSpace(description[1:])
						}
						if description != "" {
							breakingChanges = append(breakingChanges, description)
						}
					}
				}
			}
		}
	}

	return breakingChanges
}

// GetCommitTypeTitle returns the display title for a commit type
func (p *Parser) GetCommitTypeTitle(commitType string) string {
	if title, exists := p.config.Commits.Types[commitType]; exists {
		return title
	}
	// Fallback to capitalized type
	return strings.Title(commitType)
}

// IsValidCommitType checks if a commit type is recognized
func (p *Parser) IsValidCommitType(commitType string) bool {
	_, exists := p.config.Commits.Types[commitType]
	return exists
}

// GetBreakingChanges returns all breaking changes from a list of commits
func (p *Parser) GetBreakingChanges(commits []*ConventionalCommit) []*ConventionalCommit {
	var result []*ConventionalCommit

	for _, commit := range commits {
		if commit.IsBreakingChange {
			result = append(result, commit)
		}
	}

	return result
}

// SortCommitsByType sorts commits by type priority for changelog display
func (p *Parser) SortCommitsByType(groups map[string][]*ConventionalCommit) []string {
	// Define priority order for commit types
	priority := []string{"feat", "fix", "docs", "style", "refactor", "test", "chore"}
	
	var sortedTypes []string
	
	// Add types in priority order if they exist
	for _, commitType := range priority {
		if _, exists := groups[commitType]; exists {
			sortedTypes = append(sortedTypes, commitType)
		}
	}
	
	// Add any remaining types
	for commitType := range groups {
		found := false
		for _, existing := range sortedTypes {
			if existing == commitType {
				found = true
				break
			}
		}
		if !found {
			sortedTypes = append(sortedTypes, commitType)
		}
	}
	
	return sortedTypes
} 