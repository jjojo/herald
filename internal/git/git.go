package git

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Repository wraps git repository operations using git commands
type Repository struct {
	path string
}

// Commit represents a git commit with parsed information
type Commit struct {
	Hash      string
	Message   string
	Author    string
	Email     string
	Date      time.Time
	Subject   string
	Body      string
}

// Tag represents a git tag
type Tag struct {
	Name    string
	Hash    string
	Date    time.Time
	Message string
}

// OpenRepository opens a git repository from the current or specified directory
func OpenRepository(path string) (*Repository, error) {
	if path == "" {
		path = "."
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Check if we're in a git repository
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = absPath
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("not a git repository (or any of the parent directories): %s", absPath)
	}

	return &Repository{
		path: absPath,
	}, nil
}

// runGitCommand executes a git command and returns the output
func (r *Repository) runGitCommand(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = r.path
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git command failed: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// GetLatestTag returns the latest tag in the repository
func (r *Repository) GetLatestTag() (*Tag, error) {
	// Get all tags sorted by creation date
	output, err := r.runGitCommand("tag", "-l", "--sort=-creatordate", "--format=%(refname:short)|%(creatordate:iso)|%(objectname)|%(contents:subject)")
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}

	if output == "" {
		return nil, fmt.Errorf("no tags found")
	}

	// Parse the first (latest) tag
	lines := strings.Split(output, "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("no tags found")
	}

	parts := strings.Split(lines[0], "|")
	if len(parts) < 4 {
		return nil, fmt.Errorf("invalid tag format")
	}

	date, err := time.Parse("2006-01-02 15:04:05 -0700", parts[1])
	if err != nil {
		// Fallback to simpler format
		date = time.Now()
	}

	return &Tag{
		Name:    parts[0],
		Hash:    parts[2],
		Date:    date,
		Message: parts[3],
	}, nil
}

// GetCommitsSinceTag returns all commits since the specified tag
func (r *Repository) GetCommitsSinceTag(tagName string) ([]*Commit, error) {
	var args []string
	if tagName != "" {
		args = []string{"log", "--oneline", "--format=%H|%an|%ae|%at|%s|%b", tagName + "..HEAD"}
	} else {
		args = []string{"log", "--oneline", "--format=%H|%an|%ae|%at|%s|%b"}
	}

	output, err := r.runGitCommand(args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get commits: %w", err)
	}

	if output == "" {
		return []*Commit{}, nil
	}

	var commits []*Commit
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) < 5 {
			continue
		}

		// Parse timestamp
		timestamp, err := strconv.ParseInt(parts[3], 10, 64)
		if err != nil {
			timestamp = time.Now().Unix()
		}

		var body string
		if len(parts) > 5 {
			body = strings.Join(parts[5:], "|")
		}

		commit := &Commit{
			Hash:    parts[0],
			Author:  parts[1],
			Email:   parts[2],
			Date:    time.Unix(timestamp, 0),
			Subject: parts[4],
			Body:    body,
			Message: parts[4] + "\n\n" + body,
		}

		commits = append(commits, commit)
	}

	return commits, nil
}

// GetAllCommits returns all commits in the repository
func (r *Repository) GetAllCommits() ([]*Commit, error) {
	return r.GetCommitsSinceTag("")
}

// CreateTag creates a new git tag
func (r *Repository) CreateTag(name, message string) error {
	// Check if tag already exists
	_, err := r.runGitCommand("tag", "-l", name)
	if err == nil {
		// Tag exists, check if it's actually there
		if output, _ := r.runGitCommand("tag", "-l", name); output != "" {
			return fmt.Errorf("tag %s already exists", name)
		}
	}

	// Create the tag
	if message == "" {
		_, err = r.runGitCommand("tag", name)
	} else {
		_, err = r.runGitCommand("tag", "-a", name, "-m", message)
	}

	if err != nil {
		return fmt.Errorf("failed to create tag: %w", err)
	}

	return nil
}

// GetTags returns all tags in the repository
func (r *Repository) GetTags() ([]*Tag, error) {
	output, err := r.runGitCommand("tag", "-l", "--sort=-creatordate", "--format=%(refname:short)|%(creatordate:iso)|%(objectname)|%(contents:subject)")
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}

	var tags []*Tag
	if output == "" {
		return tags, nil
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) < 4 {
			continue
		}

		date, err := time.Parse("2006-01-02 15:04:05 -0700", parts[1])
		if err != nil {
			date = time.Now()
		}

		tag := &Tag{
			Name:    parts[0],
			Hash:    parts[2],
			Date:    date,
			Message: parts[3],
		}

		tags = append(tags, tag)
	}

	// Sort by date (newest first)
	sort.Slice(tags, func(i, j int) bool {
		return tags[i].Date.After(tags[j].Date)
	})

	return tags, nil
}

// IsClean returns true if the working directory is clean
func (r *Repository) IsClean() (bool, error) {
	output, err := r.runGitCommand("status", "--porcelain")
	if err != nil {
		return false, fmt.Errorf("failed to get status: %w", err)
	}

	return output == "", nil
}

// GetCurrentBranch returns the name of the current branch
func (r *Repository) GetCurrentBranch() (string, error) {
	branch, err := r.runGitCommand("branch", "--show-current")
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}

	if branch == "" {
		return "", fmt.Errorf("HEAD is not on a branch")
	}

	return branch, nil
} 