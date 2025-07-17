package cli

import (
	"fmt"
	"strings"

	"herald/internal/changelog"
	"herald/internal/commits"
	"herald/internal/config"
	"herald/internal/git"
	"herald/internal/version"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "herald",
	Short: "Herald - Release management tool using conventional commits",
	Long: `
 ,ggg,        gg                                                    
dP""Y8b       88                                  ,dPYb,         8I 
Yb, ,88       88                                  IP' Yb         8I 
 ,"  88       88                                  I8  8I         8I 
     88aaaaaaa88                                  I8  8'         8I 
     88"""""""88   ,ggg,    ,gggggg,    ,gggg,gg  I8 dP    ,gggg,8I 
     88       88  i8" "8i   dP""""8I   dP"  "Y8I  I8dP    dP"  "Y8I 
     88       88  I8, ,8I  ,8'    8I  i8'    ,8I  I8P    i8'    ,8I 
     88       Y8, ,YbadP' ,dP     Y8,,d8,   ,d8b,,d8b,_ ,d8,   ,d8b,
     88       ,Y8888P"Y8888P      ,Y8P"Y8888P",Y88P'"Y88P"Y8888P",Y8
                                                                    
                                                                    
                                                                    

Herald is a CLI tool that automates release management by analyzing 
git commit history using conventional commits standard to generate release notes 
and manage semantic versioning.`,
}

var (
	cfgFile     string
	dryRun      bool
	nextVersion bool
)

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is .heraldrc)")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "preview changes without applying them")
	rootCmd.PersistentFlags().BoolVar(&nextVersion, "next-version", false, "output only the next version number")

	// Add subcommands
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(releaseCmd)
	rootCmd.AddCommand(changelogCmd)
	rootCmd.AddCommand(versionBumpCmd)
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize .heraldrc configuration file",
	RunE: func(cmd *cobra.Command, args []string) error {
		return config.InitializeConfig()
	},
}

var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Create a full release with version bump, changelog, and git tag",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig(cfgFile)
		if err != nil {
			return err
		}
		return runRelease(cfg, dryRun)
	},
}

var changelogCmd = &cobra.Command{
	Use:   "changelog",
	Short: "Generate changelog only",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig(cfgFile)
		if err != nil {
			return err
		}
		return runChangelog(cfg, dryRun)
	},
}

var versionBumpCmd = &cobra.Command{
	Use:   "version-bump",
	Short: "Calculate and show the next version",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig(cfgFile)
		if err != nil {
			return err
		}
		return runVersionBump(cfg)
	},
}

// Placeholder functions for command implementations
func runRelease(cfg *config.Config, dryRun bool) error {
	return executeRelease(cfg, dryRun)
}

func runChangelog(cfg *config.Config, dryRun bool) error {
	return executeChangelog(cfg, dryRun)
}

func runVersionBump(cfg *config.Config) error {
	return executeVersionBump(cfg)
}

// executeRelease implements the main release functionality
func executeRelease(cfg *config.Config, dryRun bool) error {
	// Open git repository
	repo, err := git.OpenRepository(".")
	if err != nil {
		return fmt.Errorf("failed to open git repository: %w", err)
	}

	// Check if working directory is clean
	isClean, err := repo.IsClean()
	if err != nil {
		return fmt.Errorf("failed to check git status: %w", err)
	}
	if !isClean && !dryRun {
		return fmt.Errorf("working directory is not clean, please commit or stash your changes")
	}

	// Get latest tag
	latestTag, err := repo.GetLatestTag()
	if err != nil {
		fmt.Println("No previous tags found, this will be the first release")
	}

	// Get current version
	versionManager := version.NewManager(cfg)
	var currentVersion *version.Version
	if latestTag != nil {
		currentVersion, err = versionManager.GetCurrentVersion(latestTag.Name)
		if err != nil {
			return fmt.Errorf("failed to parse current version: %w", err)
		}
		fmt.Printf("Current version: %s\n", currentVersion.String())
	} else {
		currentVersion, err = versionManager.GetInitialVersion()
		if err != nil {
			return fmt.Errorf("failed to get initial version: %w", err)
		}
		fmt.Printf("Starting from initial version: %s\n", currentVersion.String())
	}

	// Get commits since last tag
	var gitCommits []*git.Commit
	if latestTag != nil {
		gitCommits, err = repo.GetCommitsSinceTag(latestTag.Name)
	} else {
		gitCommits, err = repo.GetAllCommits()
	}
	if err != nil {
		return fmt.Errorf("failed to get commits: %w", err)
	}

	if len(gitCommits) == 0 {
		fmt.Println("No new commits since last release")
		return nil
	}

	fmt.Printf("Found %d commits since last release\n", len(gitCommits))

	// Parse conventional commits
	parser := commits.NewParser(cfg)
	conventionalCommits, err := parser.ParseCommits(gitCommits)
	if err != nil {
		return fmt.Errorf("failed to parse commits: %w", err)
	}

	// Calculate version bump
	bumpType := parser.CalculateBumpType(conventionalCommits)
	if bumpType == commits.None {
		fmt.Println("No significant changes found, no release needed")
		return nil
	}

	nextVersion := versionManager.CalculateNextVersion(currentVersion, bumpType)
	fmt.Printf("Next version: %s (bump type: %s)\n", nextVersion.String(), bumpType.String())

	// Generate changelog
	changelogGenerator := changelog.NewGenerator(cfg)
	release := changelogGenerator.GenerateRelease(nextVersion, conventionalCommits)

	// Show preview
	stats := changelogGenerator.GetChangelogStats(release)
	fmt.Printf("\nRelease Summary:\n")
	fmt.Printf("- Total commits: %d\n", stats["total"])
	fmt.Printf("- Breaking changes: %d\n", stats["breaking_changes"])
	for commitType, count := range stats {
		if commitType != "total" && commitType != "breaking_changes" && count > 0 {
			fmt.Printf("- %s: %d\n", parser.GetCommitTypeTitle(commitType), count)
		}
	}

	if dryRun {
		fmt.Printf("\n=== DRY RUN MODE ===\n")
		fmt.Printf("Would create tag: %s\n", versionManager.FormatTagName(nextVersion))
		fmt.Printf("Would update changelog: %s\n", cfg.Changelog.File)
		fmt.Printf("\nChangelog preview:\n")
		fmt.Print(changelogGenerator.PreviewRelease(release))
		return nil
	}

	// Create git tag
	tagName := versionManager.FormatTagName(nextVersion)
	tagMessage := strings.ReplaceAll(cfg.Git.TagMessage, "{version}", nextVersion.String())
	
	fmt.Printf("\nCreating git tag: %s\n", tagName)
	err = repo.CreateTag(tagName, tagMessage)
	if err != nil {
		return fmt.Errorf("failed to create git tag: %w", err)
	}

	// Update changelog
	fmt.Printf("Updating changelog: %s\n", cfg.Changelog.File)
	err = changelogGenerator.PrependRelease(release)
	if err != nil {
		return fmt.Errorf("failed to update changelog: %w", err)
	}

	fmt.Printf("\n✅ Release %s completed successfully!\n", nextVersion.String())
	return nil
}

// executeChangelog generates changelog only
func executeChangelog(cfg *config.Config, dryRun bool) error {
	// Open git repository
	repo, err := git.OpenRepository(".")
	if err != nil {
		return fmt.Errorf("failed to open git repository: %w", err)
	}

	// Get latest tag
	latestTag, err := repo.GetLatestTag()
	if err != nil {
		fmt.Println("No previous tags found")
	}

	// Get current version
	versionManager := version.NewManager(cfg)
	var currentVersion *version.Version
	if latestTag != nil {
		currentVersion, err = versionManager.GetCurrentVersion(latestTag.Name)
		if err != nil {
			return fmt.Errorf("failed to parse current version: %w", err)
		}
	} else {
		currentVersion, err = versionManager.GetInitialVersion()
		if err != nil {
			return fmt.Errorf("failed to get initial version: %w", err)
		}
	}

	// Get commits since last tag
	var gitCommits []*git.Commit
	if latestTag != nil {
		gitCommits, err = repo.GetCommitsSinceTag(latestTag.Name)
	} else {
		gitCommits, err = repo.GetAllCommits()
	}
	if err != nil {
		return fmt.Errorf("failed to get commits: %w", err)
	}

	if len(gitCommits) == 0 {
		fmt.Println("No new commits since last release")
		return nil
	}

	// Parse conventional commits
	parser := commits.NewParser(cfg)
	conventionalCommits, err := parser.ParseCommits(gitCommits)
	if err != nil {
		return fmt.Errorf("failed to parse commits: %w", err)
	}

	// Calculate version bump for the changelog
	bumpType := parser.CalculateBumpType(conventionalCommits)
	nextVersion := versionManager.CalculateNextVersion(currentVersion, bumpType)

	// Generate changelog
	changelogGenerator := changelog.NewGenerator(cfg)
	release := changelogGenerator.GenerateRelease(nextVersion, conventionalCommits)

	if dryRun {
		fmt.Print(changelogGenerator.PreviewRelease(release))
		return nil
	}

	// Update changelog
	fmt.Printf("Updating changelog: %s\n", cfg.Changelog.File)
	err = changelogGenerator.PrependRelease(release)
	if err != nil {
		return fmt.Errorf("failed to update changelog: %w", err)
	}

	fmt.Println("✅ Changelog updated successfully!")
	return nil
}

// executeVersionBump calculates and displays the next version
func executeVersionBump(cfg *config.Config) error {
	// If --next-version flag is set, just output the version number
	if nextVersion {
		return executeNextVersionOnly(cfg)
	}
	// Open git repository
	repo, err := git.OpenRepository(".")
	if err != nil {
		return fmt.Errorf("failed to open git repository: %w", err)
	}

	// Get latest tag
	latestTag, err := repo.GetLatestTag()
	if err != nil {
		fmt.Println("No previous tags found")
	}

	// Get current version
	versionManager := version.NewManager(cfg)
	var currentVersion *version.Version
	if latestTag != nil {
		currentVersion, err = versionManager.GetCurrentVersion(latestTag.Name)
		if err != nil {
			return fmt.Errorf("failed to parse current version: %w", err)
		}
		fmt.Printf("Current version: %s\n", currentVersion.String())
	} else {
		currentVersion, err = versionManager.GetInitialVersion()
		if err != nil {
			return fmt.Errorf("failed to get initial version: %w", err)
		}
		fmt.Printf("No tags found, starting from: %s\n", currentVersion.String())
	}

	// Get commits since last tag
	var gitCommits []*git.Commit
	if latestTag != nil {
		gitCommits, err = repo.GetCommitsSinceTag(latestTag.Name)
	} else {
		gitCommits, err = repo.GetAllCommits()
	}
	if err != nil {
		return fmt.Errorf("failed to get commits: %w", err)
	}

	if len(gitCommits) == 0 {
		fmt.Println("No new commits since last release")
		return nil
	}

	// Parse conventional commits
	parser := commits.NewParser(cfg)
	conventionalCommits, err := parser.ParseCommits(gitCommits)
	if err != nil {
		return fmt.Errorf("failed to parse commits: %w", err)
	}

	// Calculate version bump
	bumpType := parser.CalculateBumpType(conventionalCommits)
	
	fmt.Printf("Commits since last release: %d\n", len(gitCommits))
	
	// Show commit breakdown
	groups := parser.GroupCommitsByType(conventionalCommits)
	for commitType, commits := range groups {
		if len(commits) > 0 {
			fmt.Printf("- %s: %d\n", parser.GetCommitTypeTitle(commitType), len(commits))
		}
	}
	
	breakingChanges := parser.GetBreakingChanges(conventionalCommits)
	if len(breakingChanges) > 0 {
		fmt.Printf("- Breaking changes: %d\n", len(breakingChanges))
	}
	
	if bumpType == commits.None {
		fmt.Println("\nNo significant changes found, no version bump needed")
		return nil
	}

	nextVersion := versionManager.CalculateNextVersion(currentVersion, bumpType)
	fmt.Printf("\nRecommended version bump: %s\n", bumpType.String())
	fmt.Printf("Next version: %s\n", nextVersion.String())
	
	// Show all possible version suggestions
	suggestions := versionManager.GenerateVersionSuggestions(currentVersion, conventionalCommits)
	if len(suggestions) > 1 {
		fmt.Printf("\nAll possible versions:\n")
		for suggestedType, suggestedVersion := range suggestions {
			fmt.Printf("- %s: %s\n", suggestedType, suggestedVersion.String())
		}
	}

	return nil
}

// executeNextVersionOnly outputs only the next version number (for CI/CD integration)
func executeNextVersionOnly(cfg *config.Config) error {
	// Open git repository
	repo, err := git.OpenRepository(".")
	if err != nil {
		return fmt.Errorf("failed to open git repository: %w", err)
	}

	// Get latest tag
	latestTag, err := repo.GetLatestTag()
	if err != nil {
		// No tags found, use initial version
		latestTag = nil
	}

	// Get current version
	versionManager := version.NewManager(cfg)
	var currentVersion *version.Version
	if latestTag != nil {
		currentVersion, err = versionManager.GetCurrentVersion(latestTag.Name)
		if err != nil {
			return fmt.Errorf("failed to parse current version: %w", err)
		}
	} else {
		currentVersion, err = versionManager.GetInitialVersion()
		if err != nil {
			return fmt.Errorf("failed to get initial version: %w", err)
		}
	}

	// Get commits since last tag
	var gitCommits []*git.Commit
	if latestTag != nil {
		gitCommits, err = repo.GetCommitsSinceTag(latestTag.Name)
	} else {
		gitCommits, err = repo.GetAllCommits()
	}
	if err != nil {
		return fmt.Errorf("failed to get commits: %w", err)
	}

	if len(gitCommits) == 0 {
		// No new commits, output current version
		fmt.Print(currentVersion.String())
		return nil
	}

	// Parse conventional commits
	parser := commits.NewParser(cfg)
	conventionalCommits, err := parser.ParseCommits(gitCommits)
	if err != nil {
		return fmt.Errorf("failed to parse commits: %w", err)
	}

	// Calculate version bump
	bumpType := parser.CalculateBumpType(conventionalCommits)
	nextVersion := versionManager.CalculateNextVersion(currentVersion, bumpType)

	// Output only the version number (no newline for CI/CD piping)
	fmt.Print(nextVersion.String())
	return nil
} 