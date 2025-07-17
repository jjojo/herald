# Herald - CLI Release Management Tool

## Overview

Herald is a CLI tool that automates release management by analyzing git commit history using conventional commits standard to generate release notes and manage semantic versioning.

## Goals

- Analyze git commit history using conventional commits
- Generate changelog.md with release notes
- Automatic semver version bumping
- Create git tags
- Simple configuration via .heraldrc (YAML)
- Optional CI pipeline triggering
- Simpler alternative to semantic-release

## Architecture

### Core Components

#### 1. CLI Interface (`cli.go`)

- Command parsing using `cobra` package
- Main commands:
  - `herald --release` - Full release process
  - `herald --dry-run` - Preview without making changes
  - `herald --init` - Initialize .heraldrc file
  - `herald --changelog` - Generate changelog only
  - `herald --version-bump` - Calculate next version only

#### 2. Configuration (`config.go`)

- Parse `.heraldrc` YAML file
- Default configuration fallbacks
- Validation of configuration values

#### 3. Git Operations (`git.go`)

- Read commit history since last tag
- Find latest tag
- Create new tags
- Validate git repository state

#### 4. Conventional Commits Parser (`commits.go`)

- Parse commit messages according to conventional commits spec
- Categories: feat, fix, docs, style, refactor, test, chore, etc.
- Extract breaking changes
- Extract scope information

#### 5. Version Management (`version.go`)

- Determine version bump type (major, minor, patch)
- Calculate next semantic version
- Handle pre-release versions

#### 6. Changelog Generator (`changelog.go`)

- Format release notes from parsed commits
- Group by commit type
- Include breaking changes section
- Support custom templates

#### 7. CI Integration (`ci.go`)

- Optional pipeline triggering
- Support for common CI systems (GitHub Actions, GitLab CI, etc.)

## Technical Stack

### Language: Go

- Excellent for CLI tools
- Strong standard library
- Great concurrency support
- Cross-platform compilation
- Fast execution and simple deployment
- Functional programming patterns support

### Key Dependencies

- `github.com/spf13/cobra` - Command line interface framework
- `github.com/spf13/viper` - Configuration management
- `github.com/go-git/go-git/v5` - Git operations (pure Go)
- `gopkg.in/yaml.v3` - YAML parsing
- `golang.org/x/mod/semver` - Semantic version handling
- `regexp` - Built-in regex for commit parsing
- `time` - Built-in date/time handling
- `net/http` - Built-in HTTP client for CI integration

## Configuration Format (.heraldrc)

```yaml
# Version configuration
version:
  initial: "0.1.0"
  prefix: "v" # Tag prefix (v1.0.0)

# Conventional commits configuration
commits:
  types:
    feat: "Features"
    fix: "Bug Fixes"
    docs: "Documentation"
    style: "Styles"
    refactor: "Code Refactoring"
    test: "Tests"
    chore: "Chores"
  breaking_change_keywords: ["BREAKING CHANGE", "BREAKING-CHANGE"]

# Changelog configuration
changelog:
  file: "CHANGELOG.md"
  template: "default" # or path to custom template
  include_all: false # Include all commit types or just feat/fix

# Git configuration
git:
  tag_message: "Release {version}"
  commit_changelog: true
  commit_message: "chore: update changelog for {version}"

# CI Integration (optional)
ci:
  enabled: false
  provider: "github" # github, gitlab, etc.
  trigger_on_release: true
  webhook_url: ""
```

## Implementation Plan

### Phase 1: Core Foundation

1. **Project Setup**

   - Initialize Go module with `go mod init`
   - Set up basic CLI structure with cobra
   - Add core dependencies and Go project structure

2. **Configuration System**

   - Implement .heraldrc parsing
   - Create default configuration
   - Add configuration validation

3. **Git Integration**
   - Implement git repository detection
   - Add commit history reading
   - Create tag listing functionality

### Phase 2: Commit Analysis

1. **Conventional Commits Parser**

   - Implement regex-based commit parsing
   - Extract commit type, scope, description
   - Identify breaking changes

2. **Version Management**
   - Implement semver calculation logic
   - Handle different bump types
   - Support for pre-release versions

### Phase 3: Output Generation

1. **Changelog Generator**

   - Create markdown formatting
   - Group commits by type
   - Generate release sections

2. **Git Tagging**
   - Create new git tags
   - Handle tag messages
   - Validate tag creation

### Phase 4: Advanced Features

1. **Dry Run Mode**

   - Preview changes without applying
   - Show calculated version and changelog

2. **CI Integration**

   - Implement webhook triggers
   - Support multiple CI providers

3. **Templates**
   - Custom changelog templates
   - Configurable output formats

## Error Handling Strategy

### Git Errors

- Repository not found
- No commits since last tag
- Tag already exists
- Git operations fail

### Configuration Errors

- Invalid .heraldrc format
- Missing required fields
- Invalid version formats

### Commit Parsing Errors

- Non-conventional commit messages
- Malformed commit formats
- No valid commits found

## Testing Strategy

### Unit Tests

- Commit parser with various conventional commit formats
- Version bump calculation logic
- Configuration parsing and validation

### Integration Tests

- Full release workflow
- Git operations in test repositories
- Changelog generation with real commit data

### CLI Tests

- Command line argument parsing
- Error message formatting
- Output verification

## File Structure

```
cmd/
└── herald/
    └── main.go       # Entry point

internal/
├── cli/              # Command line interface
│   └── cli.go
├── config/           # Configuration management
│   └── config.go
├── git/              # Git operations
│   └── git.go
├── commits/          # Conventional commits parsing
│   └── commits.go
├── version/          # Semantic version management
│   └── version.go
├── changelog/        # Changelog generation
│   └── changelog.go
├── ci/               # CI integration
│   └── ci.go
└── templates/        # Default templates
    └── changelog.md.tmpl

pkg/                  # Public packages (if any)

test/
├── integration/      # Integration tests
├── fixtures/         # Test data
└── unit/            # Unit tests

docs/
├── configuration.md  # Configuration reference
├── examples/        # Usage examples
└── api.md           # API documentation

go.mod               # Go modules file
go.sum               # Go modules checksum
```

## Usage Examples

### Basic Release

```bash
# Initialize configuration
herald --init

# Preview release
herald --dry-run

# Create release
herald --release
```

### Advanced Usage

```bash
# Generate changelog only
herald --changelog

# Calculate next version
herald --version-bump

# Release with specific version
herald --release --version 2.0.0

# Release pre-release version
herald --release --prerelease alpha
```

## Success Criteria

1. **Functional**: Successfully parse conventional commits and generate accurate version bumps
2. **Reliable**: Handle edge cases and provide clear error messages
3. **Configurable**: Support customization via .heraldrc file
4. **User-friendly**: Intuitive CLI interface with helpful documentation
5. **Fast**: Efficient git operations and quick execution
6. **Maintainable**: Clean, well-tested Go code following functional patterns

## Future Enhancements

- Plugin system for custom commit types
- Multiple changelog formats (JSON, XML, etc.)
- Integration with package managers (npm, cargo, etc.)
- Slack/Discord notifications
- Release asset uploading
- Multi-repository support
