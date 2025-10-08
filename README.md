# Herald

<div align="center">
  <img src="docs/assets/herald-logo.jpeg" alt="Herald Logo" width="300">
  <br>
  <em>A herald is a traditional messenger who announces important news, declarations, and proclamations on behalf of their sovereign - just as Herald announces your software releases with authority and precision.</em>
</div>

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

### NPM (Recommended)

Use Herald with `npx` without installing:

```bash
npx @jjojo/herald --help
npx @jjojo/herald init
npx @jjojo/herald release
```

Install globally via npm:

```bash
npm install -g @jjojo/herald
herald --help
```

Or install locally in your project:

```bash
npm install @jjojo/herald
npx herald --help
```

### Direct Download

Download binaries from [GitHub Releases](https://github.com/jjojo/herald/releases/latest):

```bash
# Linux/macOS
curl -L https://github.com/jjojo/herald/releases/latest/download/herald_linux_amd64.tar.gz | tar xz
chmod +x herald

# Windows
# Download herald_windows_amd64.zip from releases page
```

### Docker

```bash
docker run --rm -v $(pwd):/app ghcr.io/jjojo/herald:latest version-bump
```

### From Source

```bash
git clone https://github.com/jjojo/herald.git
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

Initialize a `.heraldrc` configuration file with comprehensive inline documentation:

```bash
herald init
```

This creates a fully documented configuration file with comments explaining each option, valid values, and examples.

## Configuration

Herald uses a `.heraldrc` YAML file for configuration. The file includes comprehensive inline documentation explaining each option:

### Self-Documenting Configuration

When you run `herald init`, Herald generates a fully documented configuration file with:

- Detailed comments for every configuration option
- Valid values and examples
- Usage explanations for semver levels
- CI integration setup instructions

### Configuration Options

```yaml
# Version configuration
version:
  initial: "0.1.0"
  prefix: "v" # Tag prefix (v1.0.0)

# Conventional commits configuration
commits:
  types:
    feat:
      title: "Features"
      semver: "minor" # Version bump level
    fix:
      title: "Bug Fixes"
      semver: "patch"
    docs:
      title: "Documentation"
      semver: "none" # No version bump
    style:
      title: "Styles"
      semver: "none"
    refactor:
      title: "Code Refactoring"
      semver: "none"
    test:
      title: "Tests"
      semver: "none"
    chore:
      title: "Chores"
      semver: "none"
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

Herald allows you to configure which version bump each commit type should trigger:

- **Configurable per commit type** - Set `semver: "major"`, `"minor"`, `"patch"`, or `"none"`
- **Breaking changes override** - `BREAKING CHANGE:` or `!` always triggers major bump
- **Highest bump wins** - If multiple commit types are present, the highest bump level is used

#### Default Configuration

- `feat:` → **Minor version bump**
- `fix:` → **Patch version bump**
- `docs:`, `style:`, `refactor:`, `test:`, `chore:` → **No version bump**

#### Custom Configuration Examples

```yaml
commits:
  types:
    docs:
      title: "Documentation"
      semver: "patch" # Docs now bump patch version
    chore:
      title: "Chores"
      semver: "minor" # Chores now bump minor version
```

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

ℹ️ Herald is in no way a drop in replacement of semantic-release.
Herald is much more lightweight focusing on version number and changelogs only.

| Feature       | Herald               | semantic-release           |
| ------------- | -------------------- | -------------------------- |
| Language      | Go                   | Node.js                    |
| Configuration | Single YAML file     | Multiple config files      |
| Setup         | Simple binary or npx | Requires Node.js ecosystem |
| Customization | Built-in options     | Plugin-based               |
| Performance   | Fast native binary   | Node.js runtime            |
| Dependencies  | Minimal              | Heavy Node.js dependencies |
