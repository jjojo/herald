# Herald GitHub Actions Integration

Herald has been specifically designed for GitHub Actions workflows with GoReleaser integration for comprehensive release management and distribution.

## Features

### GitHub Actions Workflow

Herald provides a complete GitHub Actions workflow that includes:

1. **Build Stage**: Get version and build artifacts with version info
2. **Test Stage**: Run tests and linting on built artifacts
3. **Release Stage**: Create git tags, changelog, and GitHub releases
4. **GoReleaser Stage**: Create cross-platform binaries and packages

### GoReleaser Integration

Herald integrates seamlessly with GoReleaser to provide:

- ✅ Cross-platform binary builds (Linux, macOS, Windows, FreeBSD)
- ✅ Package manager integration (Homebrew, Snap)
- ✅ Docker image creation
- ✅ GitHub Release assets
- ✅ Automatic checksums and signatures

## Configuration

### GitHub-Specific Settings

```yaml
ci:
  enabled: true
  provider: "github"
  trigger_on_release: false # Disabled in CI context
  github:
    repository: "owner/repo"
    access_token: "${{ secrets.GITHUB_TOKEN }}"
    create_release: true
```

### Environment Variables

- `GITHUB_TOKEN`: GitHub access token (auto-provided in Actions)
- `GITHUB_REPOSITORY`: Repository name in "owner/repo" format (auto-provided)
- `GITHUB_REPOSITORY_OWNER`: Repository owner (auto-provided)

## Workflow Stages

### 1. Build Stage

```yaml
- name: Get next version
  id: version
  run: |
    VERSION=$(./herald version-bump --next-version)
    echo "version=$VERSION" >> $GITHUB_OUTPUT

- name: Build application with version
  run: |
    VERSION=${{ steps.version.outputs.version }}
    go build -ldflags "-X main.Version=$VERSION" -o myapp
```

### 2. Test Stage

```yaml
- name: Run tests
  run: |
    echo "Testing version ${{ needs.build.outputs.version }}"
    go test ./...

- name: Run linter
  uses: golangci/golangci-lint-action@v3
```

### 3. Release Stage

```yaml
- name: Create release with Herald
  run: ./herald release

- name: Push changes
  run: |
    git push origin ${{ github.ref_name }}
    git push --tags
```

### 4. GoReleaser Stage

```yaml
- name: Run GoReleaser
  uses: goreleaser/goreleaser-action@v5
  with:
    args: release --clean
  env:
    GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

## Distribution Channels

### Binary Downloads

GoReleaser creates binaries for multiple platforms:

- **Linux**: amd64, arm64, arm (v6, v7)
- **macOS**: amd64, arm64 (Apple Silicon)
- **Windows**: amd64
- **FreeBSD**: amd64

### Package Managers

#### Homebrew (macOS/Linux)

```bash
# Add tap
brew tap owner/tap

# Install Herald
brew install herald

# Use Herald
herald --help
```

#### Snap (Linux)

```bash
# Install from Snap Store
snap install herald

# Use Herald
herald --help
```

### Docker

```bash
# Pull latest image
docker pull ghcr.io/owner/herald:latest

# Run Herald in current directory
docker run --rm -v $(pwd):/app ghcr.io/owner/herald:latest version-bump

# Interactive mode
docker run --rm -it -v $(pwd):/app ghcr.io/owner/herald:latest
```

## GoReleaser Configuration

Herald includes a comprehensive `.goreleaser.yml` that configures:

### Build Settings

```yaml
builds:
  - main: ./cmd/herald
    binary: herald
    goos: [linux, darwin, windows, freebsd]
    goarch: [amd64, arm64, arm]
    ldflags:
      - -X main.Version={{.Version}}
      - -X main.GitCommit={{.ShortCommit}}
      - -X main.BuildDate={{.Date}}
```

### Archives

```yaml
archives:
  - files:
      - README.md
      - LICENSE*
      - CHANGELOG.md
      - .heraldrc
      - GITHUB_INTEGRATION.md
```

### Release Notes

GoReleaser uses Herald's changelog for release notes, providing:

- Installation instructions for all platforms
- What's new section from Herald's changelog
- Links to documentation and source code

## Workflow Examples

### Automatic Releases

Triggered on push to main/master with significant changes:

```yaml
on:
  push:
    branches: [main, master]

jobs:
  build:
    # Get version and build
  test:
    # Run tests
  release:
    # Create Herald release
  goreleaser:
    # Create distribution packages
```

### Manual Releases

Triggered manually via GitHub UI:

```yaml
on:
  workflow_dispatch:

jobs:
  manual-release:
    # Create release and packages
```

### Pull Request Testing

Triggered on pull requests (build and test only):

```yaml
on:
  pull_request:
    branches: [main, master]

jobs:
  build:
    # Build and test without releasing
```

## Security Considerations

### Permissions

The workflow requires specific permissions:

```yaml
permissions:
  contents: write # For creating releases and pushing tags
  packages: write # For publishing Docker images
```

### Secrets

Herald uses standard GitHub secrets:

- `GITHUB_TOKEN`: Automatically provided by GitHub Actions
- No additional secrets required for basic functionality

### Token Scope

The `GITHUB_TOKEN` provides:

- Read access to repository contents
- Write access for creating releases
- Write access for publishing packages

## Advanced Configuration

### Custom Build Flags

```yaml
- name: Build with custom flags
  run: |
    VERSION=$(./herald version-bump --next-version)
    go build -ldflags "
      -X main.Version=$VERSION
      -X main.BuildTime=$(date -u '+%Y-%m-%d_%H:%M:%S')
      -X main.GitHash=$(git rev-parse HEAD)
    " -o myapp
```

### Conditional Releases

```yaml
- name: Check if release needed
  id: check
  run: |
    if ./herald version-bump | grep -q "No significant changes"; then
      echo "should_release=false" >> $GITHUB_OUTPUT
    else
      echo "should_release=true" >> $GITHUB_OUTPUT
    fi

- name: Create release
  if: steps.check.outputs.should_release == 'true'
  run: ./herald release
```

### Multi-Environment Releases

```yaml
strategy:
  matrix:
    environment: [staging, production]

steps:
  - name: Deploy to ${{ matrix.environment }}
    run: |
      VERSION=$(./herald version-bump --next-version)
      deploy.sh ${{ matrix.environment }} $VERSION
```

## Troubleshooting

### Common Issues

1. **"Permission denied"**: Check workflow permissions
2. **"Tag already exists"**: Herald handles duplicate tags gracefully
3. **"GoReleaser failed"**: Check .goreleaser.yml syntax
4. **"No changes to release"**: Normal when no conventional commits found

### Debug Steps

```yaml
- name: Debug Herald
  run: |
    ./herald version-bump --dry-run
    ./herald release --dry-run

- name: Debug GoReleaser
  run: |
    goreleaser check
    goreleaser build --snapshot --clean
```

## Migration from Other Tools

### From semantic-release

Herald provides similar functionality with better performance:

```yaml
# Replace semantic-release step
- name: Semantic Release
  run: npx semantic-release

# With Herald + GoReleaser
- name: Herald Release
  run: ./herald release

- name: GoReleaser
  uses: goreleaser/goreleaser-action@v5
```

### Key Advantages

| Feature          | semantic-release | Herald + GoReleaser                |
| ---------------- | ---------------- | ---------------------------------- |
| Language         | Node.js          | Go (faster)                        |
| Config           | Multiple files   | Single .heraldrc + .goreleaser.yml |
| Distribution     | Manual setup     | Built-in cross-platform            |
| Performance      | Slower           | Fast native binary                 |
| Docker           | Plugin required  | Built-in support                   |
| Package Managers | Limited          | Homebrew, Snap, etc.               |
