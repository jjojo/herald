package git

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// Repository wraps git repository operations
type Repository struct {
	repo *git.Repository
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

	// Find the git repository
	repo, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	return &Repository{
		repo: repo,
		path: absPath,
	}, nil
}

// GetLatestTag returns the latest tag in the repository
func (r *Repository) GetLatestTag() (*Tag, error) {
	tagRefs, err := r.repo.Tags()
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}

	var latestTag *Tag
	var latestDate time.Time

	err = tagRefs.ForEach(func(ref *plumbing.Reference) error {
		tagName := ref.Name().Short()
		
		// Get tag object
		tagObj, err := r.repo.TagObject(ref.Hash())
		if err != nil {
			// Might be a lightweight tag, get commit directly
			commit, err := r.repo.CommitObject(ref.Hash())
			if err != nil {
				return err
			}
			
			if commit.Author.When.After(latestDate) {
				latestDate = commit.Author.When
				latestTag = &Tag{
					Name: tagName,
					Hash: ref.Hash().String(),
					Date: commit.Author.When,
					Message: "",
				}
			}
			return nil
		}

		if tagObj.Tagger.When.After(latestDate) {
			latestDate = tagObj.Tagger.When
			latestTag = &Tag{
				Name: tagName,
				Hash: ref.Hash().String(),
				Date: tagObj.Tagger.When,
				Message: tagObj.Message,
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to iterate tags: %w", err)
	}

	return latestTag, nil
}

// GetCommitsSinceTag returns all commits since the specified tag
func (r *Repository) GetCommitsSinceTag(tagName string) ([]*Commit, error) {
	// Get HEAD commit
	head, err := r.repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD: %w", err)
	}

	var sinceHash plumbing.Hash
	if tagName != "" {
		// Find the tag
		tagRef, err := r.repo.Tag(tagName)
		if err != nil {
			return nil, fmt.Errorf("failed to find tag %s: %w", tagName, err)
		}
		sinceHash = tagRef.Hash()
	}

	// Get commit iterator from HEAD
	commits, err := r.repo.Log(&git.LogOptions{
		From: head.Hash(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get commit log: %w", err)
	}

	var result []*Commit
	err = commits.ForEach(func(c *object.Commit) error {
		// Stop if we reach the tag commit
		if !sinceHash.IsZero() && c.Hash == sinceHash {
			return fmt.Errorf("reached tag commit") // Use error to break the loop
		}

		commit := &Commit{
			Hash:    c.Hash.String(),
			Message: c.Message,
			Author:  c.Author.Name,
			Email:   c.Author.Email,
			Date:    c.Author.When,
		}

		// Split message into subject and body
		lines := strings.Split(strings.TrimSpace(c.Message), "\n")
		if len(lines) > 0 {
			commit.Subject = lines[0]
			if len(lines) > 2 {
				commit.Body = strings.Join(lines[2:], "\n")
			}
		}

		result = append(result, commit)
		return nil
	})

	// If we got an error from breaking the loop, that's expected
	if err != nil && !strings.Contains(err.Error(), "reached tag commit") {
		return nil, fmt.Errorf("failed to iterate commits: %w", err)
	}

	return result, nil
}

// GetAllCommits returns all commits in the repository
func (r *Repository) GetAllCommits() ([]*Commit, error) {
	return r.GetCommitsSinceTag("")
}

// CreateTag creates a new git tag
func (r *Repository) CreateTag(name, message string) error {
	// Get HEAD commit
	head, err := r.repo.Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD: %w", err)
	}

	// Check if tag already exists
	_, err = r.repo.Tag(name)
	if err == nil {
		return fmt.Errorf("tag %s already exists", name)
	}

	// Create the tag
	_, err = r.repo.CreateTag(name, head.Hash(), &git.CreateTagOptions{
		Message: message,
	})
	if err != nil {
		return fmt.Errorf("failed to create tag: %w", err)
	}

	return nil
}

// GetTags returns all tags in the repository
func (r *Repository) GetTags() ([]*Tag, error) {
	tagRefs, err := r.repo.Tags()
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}

	var tags []*Tag
	err = tagRefs.ForEach(func(ref *plumbing.Reference) error {
		tagName := ref.Name().Short()
		
		// Try to get tag object first
		tagObj, err := r.repo.TagObject(ref.Hash())
		if err != nil {
			// Lightweight tag, get commit
			commit, err := r.repo.CommitObject(ref.Hash())
			if err != nil {
				return err
			}
			
			tags = append(tags, &Tag{
				Name: tagName,
				Hash: ref.Hash().String(),
				Date: commit.Author.When,
				Message: "",
			})
			return nil
		}

		// Annotated tag
		tags = append(tags, &Tag{
			Name: tagName,
			Hash: ref.Hash().String(),
			Date: tagObj.Tagger.When,
			Message: tagObj.Message,
		})
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to iterate tags: %w", err)
	}

	return tags, nil
}

// IsClean returns true if the working directory is clean
func (r *Repository) IsClean() (bool, error) {
	worktree, err := r.repo.Worktree()
	if err != nil {
		return false, fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := worktree.Status()
	if err != nil {
		return false, fmt.Errorf("failed to get status: %w", err)
	}

	return status.IsClean(), nil
}

// GetCurrentBranch returns the name of the current branch
func (r *Repository) GetCurrentBranch() (string, error) {
	head, err := r.repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD: %w", err)
	}

	if !head.Name().IsBranch() {
		return "", fmt.Errorf("HEAD is not on a branch")
	}

	return head.Name().Short(), nil
} 