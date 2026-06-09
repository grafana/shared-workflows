package local

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/grafana/shared-workflows/actions/annotate-coverage/internal/coverage"
	"github.com/grafana/shared-workflows/actions/annotate-coverage/internal/diff"
	"github.com/grafana/shared-workflows/actions/annotate-coverage/internal/format"
)

// Config holds configuration for local mode.
type Config struct {
	// CoveragePath is the directory containing coverage files (*.out)
	CoveragePath string
	// Format is the output format (Text, Markdown, GitHubAnnotations)
	Format string
}

// Runner handles local coverage analysis.
type Runner struct {
	config     Config
	diffSource diff.DiffSource
}

// Option is a functional option for configuring Runner.
type Option func(*Runner)

// WithDiffSource sets a custom DiffSource for the Runner.
func WithDiffSource(ds diff.DiffSource) Option {
	return func(r *Runner) {
		r.diffSource = ds
	}
}

// NewRunner creates a new Runner with the given configuration.
func NewRunner(config Config, opts ...Option) *Runner {
	r := &Runner{
		config:     config,
		diffSource: diff.NewLocalDiffSource(""), // Default to local diff
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// Run executes the local coverage analysis workflow. It returns an error if
// any step fails; the CLI propagates that error to a non-zero exit code.
func (r *Runner) Run(ctx context.Context) error {
	// Step 1: Get diff using the configured DiffSource
	diffData, err := r.diffSource.GetDiff(ctx)
	if err != nil {
		return fmt.Errorf("failed to get diff: %w", err)
	}

	// Check if diff is empty
	if len(diffData) == 0 {
		fmt.Println("No changes detected in diff")
		return nil
	}

	// Step 2: Parse the diff
	fileDiffs, err := coverage.ParseDiff(diffData)
	if err != nil {
		return fmt.Errorf("failed to parse diff: %w", err)
	}

	// Get added lines by file
	addedLinesByFile := coverage.GetAddedLinesByFile(fileDiffs)

	// Check if there are any Go files in the diff
	if len(addedLinesByFile) == 0 {
		fmt.Println("No Go files changed in diff")
		return nil
	}

	// Step 3: Read and merge coverage files
	profiles, err := r.readAndMergeCoverageFiles()
	if err != nil {
		return err // Error message already formatted
	}

	// Step 4: Analyze coverage against diff
	result := coverage.AnalyzeCoverage(profiles, addedLinesByFile)

	// Step 5: Output results
	formatter, err := format.New(r.config.Format)
	if err != nil {
		return fmt.Errorf("failed to create formatter: %w", err)
	}

	if err := formatter.Format(result, os.Stdout); err != nil {
		return fmt.Errorf("failed to format results: %w", err)
	}

	return nil
}

// readAndMergeCoverageFiles reads all *.out files from the coverage directory
// and merges them into a single set of profiles.
// Returns a user-friendly error if the directory doesn't exist or no files are found.
func (r *Runner) readAndMergeCoverageFiles() ([]*coverage.Profile, error) {
	// Check if directory exists
	dirInfo, err := os.Stat(r.config.CoveragePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("coverage directory not found: %s\n\nRun tests with coverage first:\n  go test ./... -coverprofile=%s/coverage.out",
				r.config.CoveragePath, r.config.CoveragePath)
		}
		return nil, fmt.Errorf("failed to access coverage directory: %w", err)
	}

	if !dirInfo.IsDir() {
		return nil, fmt.Errorf("coverage path is not a directory: %s", r.config.CoveragePath)
	}

	// Read directory entries
	entries, err := os.ReadDir(r.config.CoveragePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read coverage directory: %w", err)
	}

	// Find all *.out files
	var coverageFiles []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".out") {
			coverageFiles = append(coverageFiles, filepath.Join(r.config.CoveragePath, name))
		}
	}

	if len(coverageFiles) == 0 {
		return nil, fmt.Errorf("no coverage files (*.out) found in directory: %s\n\nRun tests with coverage first:\n  go test ./... -coverprofile=%s/coverage.out",
			r.config.CoveragePath, r.config.CoveragePath)
	}

	fmt.Printf("Found %d coverage file(s) to merge\n", len(coverageFiles))

	// Read and parse all coverage files
	var allProfiles []*coverage.Profile
	for _, file := range coverageFiles {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read coverage file %s: %w", file, err)
		}

		profiles, err := coverage.ParseProfiles(data)
		if err != nil {
			return nil, fmt.Errorf("failed to parse coverage file %s: %w", file, err)
		}

		allProfiles = append(allProfiles, profiles...)
	}

	// Merge all profiles
	mergedProfiles, err := coverage.MergeProfiles(allProfiles)
	if err != nil {
		return nil, fmt.Errorf("failed to merge coverage profiles: %w", err)
	}

	return mergedProfiles, nil
}
