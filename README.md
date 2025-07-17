# Herald

Herald is a CLI tool that automates release management by analyzing git commit history using conventional commits standard to generate release notes and manage semantic versioning.

## Features

- 🔍 **Conventional Commits Analysis** - Parse commit messages to determine version bumps
- 📈 **Semantic Versioning** - Automatic major, minor, and patch version bumps
- 📝 **Changelog Generation** - Beautiful markdown changelogs with organized sections
- 🏷️ **Git Tagging** - Automatic git tag creation with customizable messages
- 🚀 **CI Integration** - Optional pipeline triggering (GitHub Actions, GitLab CI, webhooks)
- 🧪 **Dry Run Mode** - Preview changes before applying them
- ⚙️ **Configurable** - Customize behavior via `.heraldrc` YAML configuration

## Installation

### From Source

```bash
git clone https://github.com/your-org/herald.git
cd herald
go build -o herald ./cmd/herald
```

### Usage

Herald requires a git repository with conventional commits. First, initialize the configuration:

```bash
herald init
```

This creates a `.heraldrc` file with default settings.

## Commands

### `herald release`

Create a full release with version bump, changelog update, and git tag:

```bash
# Create a release
herald release

# Preview what would happen (dry run)
herald release --dry-run
```

### `herald version-bump`

Calculate and display the next version based on commits:

```bash
herald version-bump
```

### `herald changelog`

Generate changelog only without creating tags:

```bash
# Update changelog
herald changelog

# Preview changelog changes
herald changelog --dry-run
```

### `herald init`

Initialize a `.heraldrc` configuration file:

```bash
herald init
```

## Configuration

Herald uses a `.heraldrc` YAML file for configuration:

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
  template: "default"
  include_all: false # Include all commit types or just feat/fix

# Git configuration
git:
  tag_message: "Release {version}"
  commit_changelog: true
  commit_message: "chore: update changelog for {version}"

# CI Integration (optional)
ci:
  enabled: false
  provider: "github" # github, gitlab, webhook
  trigger_on_release: true
  webhook_url: ""
```

## Conventional Commits

Herald analyzes commits following the [Conventional Commits](https://www.conventionalcommits.org/) standard:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

### Commit Types and Version Bumps

- `feat:` - New features → **Minor version bump**
- `fix:` - Bug fixes → **Patch version bump**
- `BREAKING CHANGE:` or `!` - Breaking changes → **Major version bump**
- `docs:`, `style:`, `refactor:`, `test:`, `chore:` - No version bump (unless breaking)

### Examples

```bash
# Feature (minor bump: 1.0.0 → 1.1.0)
git commit -m "feat: add user authentication"

# Bug fix (patch bump: 1.1.0 → 1.1.1)
git commit -m "fix: resolve login redirect issue"

# Breaking change (major bump: 1.1.1 → 2.0.0)
git commit -m "feat!: redesign API endpoints

BREAKING CHANGE: API endpoints have been restructured"

# With scope
git commit -m "feat(auth): add OAuth2 support"
```

## Workflow Example

1. **Make changes** and commit using conventional commits:

   ```bash
   git commit -m "feat: add new dashboard feature"
   git commit -m "fix: resolve navigation bug"
   ```

2. **Preview the release**:

   ```bash
   herald release --dry-run
   ```

3. **Create the release**:

   ```bash
   herald release
   ```

4. **Result**: Herald will:
   - Analyze commits (feat + fix = minor bump)
   - Calculate next version (e.g., 1.0.0 → 1.1.0)
   - Generate changelog entry
   - Create git tag (v1.1.0)
   - Optionally trigger CI pipeline

## CI Integration

### GitHub Actions

Configure GitHub Actions integration:

```yaml
# .heraldrc
ci:
  enabled: true
  provider: "github"
  trigger_on_release: true
  webhook_url: "https://api.github.com/repos/owner/repo/dispatches"
```

### GitLab CI

Configure GitLab CI integration:

```yaml
# .heraldrc
ci:
  enabled: true
  provider: "gitlab"
  trigger_on_release: true
  webhook_url: "https://gitlab.com/api/v4/projects/ID/trigger/pipeline"
```

### Custom Webhook

Configure custom webhook integration:

```yaml
# .heraldrc
ci:
  enabled: true
  provider: "webhook"
  trigger_on_release: true
  webhook_url: "https://your-ci-system.com/webhook"
```

## Generated Changelog

Herald generates changelogs in [Keep a Changelog](https://keepachangelog.com/) format:

```markdown
# Changelog

## [1.1.0] - 2025-01-15

### ⚠ BREAKING CHANGES

- redesign API endpoints (**auth**)
  Complete restructure of authentication endpoints

### Features

- **auth:** add OAuth2 support ([a1b2c3d])
- add user dashboard ([e4f5g6h])

### Bug Fixes

- resolve login redirect issue ([i7j8k9l])
```

## Architecture

Herald is built with Go and follows functional programming patterns:

- **CLI Interface** - Cobra-based command interface
- **Configuration** - YAML-based configuration management
- **Git Operations** - Pure Go git operations via go-git
- **Commit Parsing** - Regex-based conventional commits parser
- **Version Management** - Semantic versioning with golang.org/x/mod/semver
- **Changelog Generation** - Markdown formatting with templates
- **CI Integration** - HTTP webhooks for pipeline triggering

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feat/new-feature`
3. Make changes using conventional commits
4. Test with Herald: `herald release --dry-run`
5. Submit a pull request

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Comparison with semantic-release

| Feature       | Herald             | semantic-release           |
| ------------- | ------------------ | -------------------------- |
| Language      | Go                 | Node.js                    |
| Configuration | Single YAML file   | Multiple config files      |
| Setup         | Simple binary      | Requires Node.js ecosystem |
| Customization | Built-in options   | Plugin-based               |
| Performance   | Fast native binary | Node.js runtime            |
| Dependencies  | Minimal            | Heavy Node.js dependencies |

Herald provides an alternative to semantic-release. Semantic-release is really good too and I have used it a lot. Herald is just my interpretation and simplification with the essential features most projects need. You choose the tool you want and need!
