package format

import (
	"fmt"
	"io"

	"github.com/grafana/shared-workflows/actions/annotate-coverage/internal/coverage"
)

// TextFormatter formats analysis results as plain text for console output.
// Groups uncovered lines by file and shows line ranges where possible.
type TextFormatter struct{}

// Format formats the analysis result as plain text.
func (f *TextFormatter) Format(result *coverage.AnalysisResult, w io.Writer) error {
	if result == nil {
		return fmt.Errorf("result is nil")
	}

	// Handle case where all lines are covered
	if !result.HasUncoveredLines() {
		if result.DiffAddedLines == 0 {
			fmt.Fprintln(w, "No lines added in diff")
			return nil
		}
		fmt.Fprintln(w, "All added lines are covered!")
		return nil
	}

	// Print header
	fmt.Fprintln(w, "Uncovered lines in diff:")
	fmt.Fprintln(w)

	// Print uncovered lines grouped by file (sorted alphabetically)
	sortedFiles := result.GetSortedFiles()
	for _, file := range sortedFiles {
		lines := result.UncoveredByFile[file]
		fmt.Fprintf(w, "%s\n", file)
		fmt.Fprintf(w, "  Lines: %s\n", formatLineRanges(lines))
		fmt.Fprintln(w)
	}

	// Print summary with percentage
	coveragePercent := 0.0
	if result.DiffAddedLines > 0 {
		coveragePercent = float64(result.DiffAddedCovered) / float64(result.DiffAddedLines) * 100
	}

	uncoveredCount := result.DiffAddedLines - result.DiffAddedCovered
	fmt.Fprintf(w, "Summary: %d uncovered lines out of %d added lines (%.1f%% coverage)\n",
		uncoveredCount, result.DiffAddedLines, coveragePercent)

	return nil
}

// formatLineRanges converts a list of line numbers into a human-readable format
// with ranges where possible (e.g., "1, 3-5, 7, 10-12").
func formatLineRanges(lines []int) string {
	if len(lines) == 0 {
		return ""
	}

	// Lines should already be sorted from the analysis, but let's ensure it
	// by building the output incrementally

	var result string
	rangeStart := lines[0]
	rangeEnd := lines[0]

	for i := 1; i < len(lines); i++ {
		if lines[i] == rangeEnd+1 {
			// Continue the range
			rangeEnd = lines[i]
		} else {
			// End the current range and start a new one
			result = appendRange(result, rangeStart, rangeEnd)
			rangeStart = lines[i]
			rangeEnd = lines[i]
		}
	}

	// Append the final range
	result = appendRange(result, rangeStart, rangeEnd)

	return result
}

// appendRange appends a line range to the result string.
// If start == end, appends a single number. Otherwise, appends "start-end".
func appendRange(result string, start, end int) string {
	if result != "" {
		result += ", "
	}

	if start == end {
		result += fmt.Sprintf("%d", start)
	} else {
		result += fmt.Sprintf("%d-%d", start, end)
	}

	return result
}
