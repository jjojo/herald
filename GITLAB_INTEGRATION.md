# Herald GitLab CI/CD Integration

Herald has been specifically tuned for GitLab CI/CD workflows with a two-stage release process:

1. **Build Stage**: Get version number and build artifacts
2. **Release Stage**: Create git tags, changelog, and GitLab releases

## New Features

### `--next-version` Flag

Get only the next version number for CI/CD scripts:

```bash
# Output: v1.3.0 (no newlines, perfect for CI variables)
herald version-bump --next-version
```

**Usage in GitLab CI:**

```yaml
script:
  - export VERSION=$(herald version-bump --next-version)
  - echo "Building version: $VERSION"
  - go build -ldflags "-X main.Version=$VERSION" -o myapp
```

### GitLab Integration

Herald provides CI/CD integration for GitLab with:

- ✅ Version detection for CI pipelines
- ✅ Changelog generation
- ✅ Git tag creation

## Configuration

Herald uses a simple `.heraldrc` configuration file without any CI-specific settings. The GitLab CI pipeline handles the release process automatically.

## GitLab CI/CD Workflow

### Complete Example

See `.gitlab-ci.yml` for a full working example with:

1. **Build Stage**:

   - Get version with `herald version-bump --next-version`
   - Build artifacts with version info
   - Pass version to subsequent stages

2. **Test Stage**:

   - Test built artifacts
   - Use version info from build stage

3. **Release Stage**:
   - Create git tag and changelog
   - Create GitLab release with API
   - Push changes back to repository

### Key Benefits

✅ **Version in Artifacts**: Include version in build artifacts before testing
✅ **Conditional Releases**: Only release if tests pass
✅ **Automatic GitLab Releases**: Create releases with changelog descriptions
✅ **Git History**: Full git history access for conventional commit analysis
✅ **Manual Triggers**: Optional manual release jobs

## Usage Examples

### Build Stage (Get Version)

```bash
# Get next version for build metadata
VERSION=$(herald version-bump --next-version)
echo "Building version: $VERSION"

# Build with version info
go build -ldflags "-X main.Version=$VERSION" -o myapp
docker build -t myapp:$VERSION .
```

### Release Stage (Create Release)

```bash
# Configure Herald
export GITLAB_ACCESS_TOKEN=$CI_JOB_TOKEN

# Create full release
herald release
# This will:
# - Analyze conventional commits
# - Create git tag (e.g., v1.3.0)
# - Generate/update CHANGELOG.md
# - Create GitLab release with changelog
```

## Configuration Details

### GitLab Project ID

Can be either:

- Numeric ID: `"12345"`
- Project path: `"group/project-name"`

### GitLab Access Token

Required permissions:

- `api` scope for creating releases
- `write_repository` for git operations

Can be provided via:

1. Configuration file: `ci.gitlab.access_token`
2. Environment variable: `GITLAB_ACCESS_TOKEN`
3. GitLab CI token: `CI_JOB_TOKEN` (automatic)

### Release Creation

When `ci.gitlab.create_release: true`, Herald will:

1. Create a GitLab release via API
2. Use the tag created by Herald
3. Include the generated changelog as release description
4. Set release date to current timestamp

## Error Handling

Herald gracefully handles GitLab API errors:

- Logs warnings for failed release creation
- Continues with git operations even if API fails
- Validates configuration before attempting API calls

## Migration from Other Tools

### From semantic-release

Herald provides similar functionality with simpler configuration:

```yaml
# Herald equivalent to semantic-release
version:
  prefix: "v"
commits:
  types:
    feat: { title: "Features", semver: "minor" }
    fix: { title: "Bug Fixes", semver: "patch" }
ci:
  enabled: true
  provider: "gitlab"
  gitlab:
    create_release: true
```

### Key Differences

| Feature            | semantic-release | Herald                       |
| ------------------ | ---------------- | ---------------------------- |
| Language           | Node.js          | Go                           |
| Config             | Multiple files   | Single `.heraldrc`           |
| GitLab Integration | Plugin-based     | Built-in                     |
| Version Output     | Complex          | Simple `--next-version` flag |
| Performance        | Slower           | Fast native binary           |

## Troubleshooting

### Common Issues

1. **"No git history"**: Use `git fetch --unshallow` in CI
2. **"Access denied"**: Check GitLab token permissions
3. **"Project not found"**: Verify project ID format
4. **"No conventional commits"**: Check commit message format

### Debug Mode

```bash
# Dry run to see what would happen
herald release --dry-run

# Check version calculation
herald version-bump
```
