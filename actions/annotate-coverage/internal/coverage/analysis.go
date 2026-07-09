package coverage

import (
	"fmt"
	"sort"
	"strings"

	"github.com/grafana/shared-workflows/actions/annotate-coverage/internal/github"
)

// AnalysisResult contains the results of coverage analysis.
type AnalysisResult struct {
	// UncoveredByFile maps filenames to their uncovered line numbers
	UncoveredByFile map[string][]int
	// TotalLines is the total number of instrumented lines across all profiles
	TotalLines int
	// TotalCovered is the total number of covered lines across all profiles
	TotalCovered int
	// DiffAddedLines is the total number of lines added in the diff
	DiffAddedLines int
	// DiffAddedCovered is the total number of covered lines among added lines
	DiffAddedCovered int
}

// findMatchingDiffFile finds the diff file that matches a coverage profile.
// Returns the diff filename, the added lines, and true if found; empty values and false otherwise.
func findMatchingDiffFile(profile *Profile, addedLinesByFile map[string][]int) (string, []int, bool) {
	profileFile := profile.FileName

	// First try exact match
	if addedLines, ok := addedLinesByFile[profileFile]; ok {
		return profileFile, addedLines, true
	}

	// Coverage uses full module paths while the diff uses repo-relative paths,
	// so match when the diff path is a whole trailing path segment of the
	// profile path. The leading "/" enforces a path boundary: a plain suffix
	// check would wrongly match e.g. ".../myhandler.go" against a diff entry for
	// "handler.go". The exact match covers files at the repo root.
	for diffFile, addedLines := range addedLinesByFile {
		if profileFile == diffFile || strings.HasSuffix(profileFile, "/"+diffFile) {
			return diffFile, addedLines, true
		}
	}

	return "", nil, false
}

// AnalyzeCoverage cross-references coverage profiles with diff to find uncovered added lines.
// Coverage profiles are the primary source - we extract uncovered lines from them,
// then filter by the diff to only report lines that were added.
// It takes coverage profiles and a map of added lines by file (from GetAddedLinesByFile).
// Returns an AnalysisResult with uncovered lines grouped by file.
func AnalyzeCoverage(profiles []*Profile, addedLinesByFile map[string][]int) *AnalysisResult {
	result := &AnalysisResult{
		UncoveredByFile:  make(map[string][]int),
		TotalLines:       0,
		TotalCovered:     0,
		DiffAddedLines:   0,
		DiffAddedCovered: 0,
	}

	// First pass: Calculate true total/covered lines from all profiles
	for _, profile := range profiles {
		for _, block := range profile.Blocks {
			// Count lines in this block (EndLine - StartLine + 1)
			linesInBlock := block.EndLine - block.StartLine + 1
			result.TotalLines += linesInBlock
			if block.Count > 0 {
				result.TotalCovered += linesInBlock
			}
		}
	}

	// Second pass: Process files that have coverage data and are in the diff
	// Files without coverage records are excluded from the output
	for _, profile := range profiles {
		// Find the matching diff file
		diffFile, addedLines, found := findMatchingDiffFile(profile, addedLinesByFile)
		if !found {
			// File has coverage but is not in the diff - skip it
			continue
		}

		// Check each added line to see if it's covered
		var uncoveredLines []int
		for _, line := range addedLines {
			// Only consider lines that are instrumented (in a coverage block)
			if !isLineInstrumented(profile, line) {
				continue // Skip non-executable lines (comments, blank lines, etc.)
			}
			if isLineCovered(profile, line) {
				result.DiffAddedCovered++
			} else {
				uncoveredLines = append(uncoveredLines, line)
			}
			result.DiffAddedLines++
		}

		// Only add to result if there are uncovered lines
		if len(uncoveredLines) > 0 {
			result.UncoveredByFile[diffFile] = uncoveredLines
		}
	}

	return result
}

// isLineInstrumented checks if a line falls within any coverage block.
// Returns true if the line is in ANY block (regardless of count).
// Lines not in any block are non-executable (comments, blank lines, etc.) and should be ignored.
func isLineInstrumented(profile *Profile, line int) bool {
	if profile == nil {
		return false
	}

	for _, block := range profile.Blocks {
		if line >= block.StartLine && line <= block.EndLine {
			return true
		}
	}

	return false
}

// isLineCovered checks if a specific line is covered by the profile.
// A line is covered if it falls within a coverage block with Count > 0.
// Returns false if profile is nil or line is not in any covered block.
func isLineCovered(profile *Profile, line int) bool {
	if profile == nil {
		return false
	}

	// Check each block to see if the line falls within a covered block
	for _, block := range profile.Blocks {
		// A line is covered if:
		// 1. It falls within the block's line range (StartLine to EndLine)
		// 2. The block was executed (Count > 0)
		if line >= block.StartLine && line <= block.EndLine && block.Count > 0 {
			return true
		}
	}

	return false
}

// HasUncoveredLines returns true if there are any uncovered lines in the result.
func (r *AnalysisResult) HasUncoveredLines() bool {
	return r.DiffAddedLines > r.DiffAddedCovered
}

// GetSortedFiles returns a sorted list of files with uncovered lines.
// Useful for consistent output ordering.
func (r *AnalysisResult) GetSortedFiles() []string {
	files := make([]string, 0, len(r.UncoveredByFile))
	for file := range r.UncoveredByFile {
		files = append(files, file)
	}
	sort.Strings(files)
	return files
}

// CoverageStats holds coverage statistics.
type CoverageStats struct {
	TotalStatements   int
	CoveredStatements int
	Percentage        float64
	ByFile            map[string]*FileCoverage
}

// FileCoverage holds coverage statistics for a single file.
type FileCoverage struct {
	FileName          string
	TotalStatements   int
	CoveredStatements int
	Percentage        float64
}

// CalculateCoverageStats calculates coverage statistics from profiles.
// It computes overall coverage percentage and per-file breakdown.
// Coverage is calculated as: (covered statements / total statements) * 100
func CalculateCoverageStats(profiles []*Profile) *CoverageStats {
	stats := &CoverageStats{
		ByFile: make(map[string]*FileCoverage),
	}

	if len(profiles) == 0 {
		return stats
	}

	// Process each profile to calculate per-file and overall stats
	for _, profile := range profiles {
		fileStats := &FileCoverage{
			FileName: profile.FileName,
		}

		// Count statements in each block
		for _, block := range profile.Blocks {
			fileStats.TotalStatements += block.NumStmt
			if block.Count > 0 {
				fileStats.CoveredStatements += block.NumStmt
			}
		}

		// Calculate per-file percentage
		if fileStats.TotalStatements > 0 {
			fileStats.Percentage = float64(fileStats.CoveredStatements) / float64(fileStats.TotalStatements) * 100
		}

		// Add to per-file map
		stats.ByFile[profile.FileName] = fileStats

		// Accumulate overall stats
		stats.TotalStatements += fileStats.TotalStatements
		stats.CoveredStatements += fileStats.CoveredStatements
	}

	// Calculate overall percentage
	if stats.TotalStatements > 0 {
		stats.Percentage = float64(stats.CoveredStatements) / float64(stats.TotalStatements) * 100
	}

	return stats
}

// CoverageComparison holds the result of comparing two coverage reports.
type CoverageComparison struct {
	BaseCoverage float64
	HeadCoverage float64
	Delta        float64 // positive = improvement, negative = regression
	Decreased    bool
}

// CompareCoverage compares base and head coverage stats.
// Returns a comparison showing the delta between base and head.
// If base is nil, it's treated as 0% coverage (first coverage report).
func CompareCoverage(base, head *CoverageStats) *CoverageComparison {
	comparison := &CoverageComparison{}

	if head != nil {
		comparison.HeadCoverage = head.Percentage
	}

	if base != nil {
		comparison.BaseCoverage = base.Percentage
	}

	// Calculate delta: positive means improvement, negative means regression
	comparison.Delta = comparison.HeadCoverage - comparison.BaseCoverage
	comparison.Decreased = comparison.Delta < 0

	return comparison
}

// GenerateAnnotations converts analysis result to GitHub Check Run annotations.
// Uses github.GroupIntoRanges to merge consecutive lines into ranges.
// Returns annotations with "notice" level as per GitHub Check Run API format.
func GenerateAnnotations(result *AnalysisResult) []*github.Annotation {
	if result == nil || len(result.UncoveredByFile) == 0 {
		return nil
	}

	var annotations []*github.Annotation

	// Process each file (in sorted order for consistency)
	sortedFiles := result.GetSortedFiles()
	for _, file := range sortedFiles {
		lines := result.UncoveredByFile[file]

		// Group consecutive lines into ranges
		ranges := github.SortAndGroupLines(lines)

		// Create one annotation per range
		for _, r := range ranges {
			var title, message string
			if r.Start == r.End {
				// Single line
				title = "Uncovered line"
				message = fmt.Sprintf("Line %d is not covered by tests", r.Start)
			} else {
				// Range of lines
				title = "Uncovered lines"
				message = fmt.Sprintf("Lines %d-%d are not covered by tests", r.Start, r.End)
			}

			annotations = append(annotations, &github.Annotation{
				Path:      file,
				StartLine: r.Start,
				EndLine:   r.End,
				Level:     "notice",
				Title:     title,
				Message:   message,
			})
		}
	}

	return annotations
}
