package format

import (
	"fmt"
	"io"

	"github.com/grafana/shared-workflows/actions/annotate-coverage/internal/coverage"
)

// MarkdownFormatter formats analysis results as Markdown.
// Outputs a table with uncovered lines by file and a summary.
type MarkdownFormatter struct{}

// Format formats the analysis result as Markdown.
func (f *MarkdownFormatter) Format(result *coverage.AnalysisResult, w io.Writer) error {
	if result == nil {
		return fmt.Errorf("result is nil")
	}

	// Handle case where all lines are covered
	if !result.HasUncoveredLines() {
		if result.DiffAddedLines == 0 {
			_, _ = fmt.Fprintln(w, "No lines added in diff")
			return nil
		}
		_, _ = fmt.Fprintln(w, "All added lines are covered!")
		return nil
	}

	// Print header
	_, _ = fmt.Fprintln(w, "## Uncovered Lines in Diff")
	_, _ = fmt.Fprintln(w)

	// Print table header
	_, _ = fmt.Fprintln(w, "| File | Lines |")
	_, _ = fmt.Fprintln(w, "|------|-------|")

	// Print uncovered lines grouped by file (sorted alphabetically)
	sortedFiles := result.GetSortedFiles()
	for _, file := range sortedFiles {
		lines := result.UncoveredByFile[file]
		_, _ = fmt.Fprintf(w, "| %s | %s |\n", file, formatLineRanges(lines))
	}

	_, _ = fmt.Fprintln(w)

	// Print summary with percentage
	coveragePercent := 0.0
	if result.DiffAddedLines > 0 {
		coveragePercent = float64(result.DiffAddedCovered) / float64(result.DiffAddedLines) * 100
	}

	uncoveredCount := result.DiffAddedLines - result.DiffAddedCovered
	_, _ = fmt.Fprintf(w, "**Summary:** %d uncovered lines out of %d added (%.1f%% coverage)\n",
		uncoveredCount, result.DiffAddedLines, coveragePercent)

	return nil
}
