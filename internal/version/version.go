package version

import (
	"fmt"
	"strings"

	"herald/internal/commits"
	"herald/internal/config"

	"golang.org/x/mod/semver"
)

// Version represents a semantic version
type Version struct {
	Major      int
	Minor      int
	Patch      int
	Prerelease string
	Build      string
	Prefix     string
	Raw        string
}

// Manager handles version operations
type Manager struct {
	config *config.Config
}

// NewManager creates a new version manager
func NewManager(cfg *config.Config) *Manager {
	return &Manager{
		config: cfg,
	}
}

// ParseVersion parses a version string into a Version struct
func (m *Manager) ParseVersion(versionStr string) (*Version, error) {
	if versionStr == "" {
		return nil, fmt.Errorf("version string cannot be empty")
	}

	// Store the original raw version
	raw := versionStr

	// Extract prefix if present
	prefix := ""
	if strings.HasPrefix(versionStr, "v") {
		prefix = "v"
		versionStr = versionStr[1:]
	}

	// Use semver to validate and parse
	if !semver.IsValid("v" + versionStr) {
		return nil, fmt.Errorf("invalid semantic version: %s", raw)
	}

	// Parse using semver
	canonical := "v" + versionStr
	prerelease := semver.Prerelease(canonical)
	build := semver.Build(canonical)

	// Parse the version parts manually since semver doesn't have Minor/Patch functions
	var majorInt, minorInt, patchInt int
	
	// Remove prerelease and build metadata for parsing core version
	coreVersion := versionStr
	if prerelease != "" {
		coreVersion = strings.Split(coreVersion, "-")[0]
	}
	if build != "" {
		coreVersion = strings.Split(coreVersion, "+")[0]
	}
	
	// Parse major.minor.patch
	parts := strings.Split(coreVersion, ".")
	if len(parts) >= 1 {
		if _, err := fmt.Sscanf(parts[0], "%d", &majorInt); err != nil {
			majorInt = 0 // fallback to 0 if parsing fails
		}
	}
	if len(parts) >= 2 {
		if _, err := fmt.Sscanf(parts[1], "%d", &minorInt); err != nil {
			minorInt = 0 // fallback to 0 if parsing fails
		}
	}
	if len(parts) >= 3 {
		if _, err := fmt.Sscanf(parts[2], "%d", &patchInt); err != nil {
			patchInt = 0 // fallback to 0 if parsing fails
		}
	}

	return &Version{
		Major:      majorInt,
		Minor:      minorInt,
		Patch:      patchInt,
		Prerelease: prerelease,
		Build:      build,
		Prefix:     prefix,
		Raw:        raw,
	}, nil
}

// BumpVersion creates a new version based on the bump type
func (m *Manager) BumpVersion(currentVersion *Version, bumpType commits.BumpType) *Version {
	newVersion := &Version{
		Major:      currentVersion.Major,
		Minor:      currentVersion.Minor,
		Patch:      currentVersion.Patch,
		Prerelease: "", // Clear prerelease on bump
		Build:      "", // Clear build on bump
		Prefix:     currentVersion.Prefix,
	}

	switch bumpType {
	case commits.Major:
		newVersion.Major++
		newVersion.Minor = 0
		newVersion.Patch = 0
	case commits.Minor:
		newVersion.Minor++
		newVersion.Patch = 0
	case commits.Patch:
		newVersion.Patch++
	}

	newVersion.Raw = newVersion.String()
	return newVersion
}

// String returns the string representation of the version
func (v *Version) String() string {
	version := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)

	if v.Prerelease != "" {
		version += v.Prerelease
	}

	if v.Build != "" {
		version += v.Build
	}

	return v.Prefix + version
}

// WithoutPrefix returns the version string without the prefix
func (v *Version) WithoutPrefix() string {
	version := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)

	if v.Prerelease != "" {
		version += v.Prerelease
	}

	if v.Build != "" {
		version += v.Build
	}

	return version
}

// IsPrerelease returns true if this is a prerelease version
func (v *Version) IsPrerelease() bool {
	return v.Prerelease != ""
}

// Compare compares two versions (-1, 0, 1)
func (v *Version) Compare(other *Version) int {
	thisCanonical := "v" + v.WithoutPrefix()
	otherCanonical := "v" + other.WithoutPrefix()
	return semver.Compare(thisCanonical, otherCanonical)
}

// IsGreaterThan returns true if this version is greater than other
func (v *Version) IsGreaterThan(other *Version) bool {
	return v.Compare(other) > 0
}

// GetCurrentVersion gets the current version from the latest git tag
func (m *Manager) GetCurrentVersion(latestTag string) (*Version, error) {
	if latestTag == "" {
		// No tags exist, use initial version from config
		return m.ParseVersion(m.config.Version.Initial)
	}

	return m.ParseVersion(latestTag)
}

// CalculateNextVersion calculates the next version based on commits
func (m *Manager) CalculateNextVersion(currentVersion *Version, bumpType commits.BumpType) *Version {
	if bumpType == commits.None {
		return currentVersion
	}

	return m.BumpVersion(currentVersion, bumpType)
}

// FormatTagName formats a version as a git tag name
func (m *Manager) FormatTagName(version *Version) string {
	if m.config.Version.Prefix != "" {
		return m.config.Version.Prefix + version.WithoutPrefix()
	}
	return version.String()
}

// CreatePrereleaseVersion creates a prerelease version
func (m *Manager) CreatePrereleaseVersion(baseVersion *Version, prereleaseType string, iteration int) *Version {
	newVersion := &Version{
		Major:  baseVersion.Major,
		Minor:  baseVersion.Minor,
		Patch:  baseVersion.Patch,
		Prefix: baseVersion.Prefix,
		Build:  baseVersion.Build,
	}

	if iteration > 0 {
		newVersion.Prerelease = fmt.Sprintf("-%s.%d", prereleaseType, iteration)
	} else {
		newVersion.Prerelease = fmt.Sprintf("-%s", prereleaseType)
	}

	newVersion.Raw = newVersion.String()
	return newVersion
}

// ValidateVersion validates a version string
func (m *Manager) ValidateVersion(versionStr string) error {
	_, err := m.ParseVersion(versionStr)
	return err
}

// GetInitialVersion returns the initial version from configuration
func (m *Manager) GetInitialVersion() (*Version, error) {
	return m.ParseVersion(m.config.Version.Initial)
}

// IsValidBumpType checks if a bump type is valid
func IsValidBumpType(bumpType string) bool {
	switch bumpType {
	case "major", "minor", "patch":
		return true
	default:
		return false
	}
}

// ParseBumpType parses a string into a BumpType
func ParseBumpType(bumpType string) (commits.BumpType, error) {
	switch strings.ToLower(bumpType) {
	case "major":
		return commits.Major, nil
	case "minor":
		return commits.Minor, nil
	case "patch":
		return commits.Patch, nil
	case "none":
		return commits.None, nil
	default:
		return commits.None, fmt.Errorf("invalid bump type: %s", bumpType)
	}
}

// FindLatestVersion finds the latest version from a list of version strings
func (m *Manager) FindLatestVersion(versions []string) (*Version, error) {
	if len(versions) == 0 {
		return m.GetInitialVersion()
	}

	var latest *Version
	for _, versionStr := range versions {
		version, err := m.ParseVersion(versionStr)
		if err != nil {
			continue // Skip invalid versions
		}

		if latest == nil || version.IsGreaterThan(latest) {
			latest = version
		}
	}

	if latest == nil {
		return m.GetInitialVersion()
	}

	return latest, nil
}

// GenerateVersionSuggestions generates version suggestions based on commit analysis
func (m *Manager) GenerateVersionSuggestions(currentVersion *Version, conventionalCommits []*commits.ConventionalCommit) map[string]*Version {
	suggestions := make(map[string]*Version)

	// Calculate automatic bump
	parser := commits.NewParser(m.config)
	autoBumpType := parser.CalculateBumpType(conventionalCommits)
	
	if autoBumpType != commits.None {
		autoVersion := m.BumpVersion(currentVersion, autoBumpType)
		suggestions["auto"] = autoVersion
		suggestions[autoBumpType.String()] = autoVersion
	}

	// Generate all possible bumps
	if autoBumpType != commits.Major {
		suggestions["major"] = m.BumpVersion(currentVersion, commits.Major)
	}
	if autoBumpType != commits.Minor {
		suggestions["minor"] = m.BumpVersion(currentVersion, commits.Minor)
	}
	if autoBumpType != commits.Patch {
		suggestions["patch"] = m.BumpVersion(currentVersion, commits.Patch)
	}

	return suggestions
} 