package main

import (
	"context"
	"fmt"
	"os"

	"github.com/grafana/shared-workflows/actions/annotate-coverage/internal/diff"
	"github.com/grafana/shared-workflows/actions/annotate-coverage/internal/local"
	"github.com/spf13/cobra"
)

var (
	// Version information (set via ldflags during build)
	version = "dev"
	commit  = "unknown"
	date    = "unknown"

	// CLI flags
	coveragePath string
	format       string
	baseRef      string
	commitSHA    string
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "annotate-coverage",
	Short: "annotate-coverage - Highlight uncovered lines in a diff",
	Long: `annotate-coverage analyzes Go coverage data against the current git diff
to find uncovered lines in your changes.

Usage:
  annotate-coverage [flags]

This command analyzes coverage files in the specified directory and compares them
against the current git diff to highlight uncovered lines in your changes.

Diff modes:
  - Default: Compares against current working directory changes (git diff)
  - --base <ref>: Compares base ref to HEAD (git diff <base>..HEAD)
  - --base <ref> --commit <ref>: Compares two refs (git diff <base>..<commit>)
  - --commit <sha>: Shows changes for a specific commit (git diff-tree <sha>)`,
	RunE: run,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("annotate-coverage %s\n", version)
		fmt.Printf("  commit: %s\n", commit)
		fmt.Printf("  built:  %s\n", date)
	},
}

func init() {
	// Add subcommands
	rootCmd.AddCommand(versionCmd)

	// Define flags
	rootCmd.Flags().StringVar(&coveragePath, "coverage", ".coverage", "Directory containing coverage files")
	rootCmd.Flags().StringVar(&format, "format", "Text", "Output format (Text, Markdown, GitHubAnnotations)")
	rootCmd.Flags().StringVar(&baseRef, "base", "", "Base reference to compare from (can be combined with --commit)")
	rootCmd.Flags().StringVar(&commitSHA, "commit", "", "Commit reference to compare to (defaults to HEAD when used with --base)")
}

func run(cmd *cobra.Command, args []string) error {
	// Create appropriate DiffSource based on flags
	var diffSource diff.DiffSource

	if baseRef != "" {
		// --base flag: compare base to commit (defaults to HEAD if commit not specified)
		// Supports both: --base <ref> and --base <ref> --commit <ref>
		diffSource = diff.NewGitBaseDiffSource(baseRef, commitSHA, "")
	} else if commitSHA != "" {
		// --commit flag only: single commit analysis
		diffSource = diff.NewGitCommitDiffSource(commitSHA, "")
	} else {
		// Default: use local git diff (working directory changes)
		diffSource = diff.NewLocalDiffSource("")
	}

	runner := local.NewRunner(local.Config{
		CoveragePath: coveragePath,
		Format:       format,
	}, local.WithDiffSource(diffSource))

	return runner.Run(context.Background())
}
