package format

import (
	"fmt"
	"io"

	"github.com/grafana/shared-workflows/actions/annotate-coverage/internal/coverage"
)

// GitHubAnnotationsFormatter formats analysis results as GitHub Actions workflow commands.
// Outputs one ::notice annotation per block of consecutive uncovered lines.
type GitHubAnnotationsFormatter struct{}

// Format formats the analysis result as GitHub Actions annotations.
func (f *GitHubAnnotationsFormatter) Format(result *coverage.AnalysisResult, w io.Writer) error {
	if result == nil {
		return fmt.Errorf("result is nil")
	}

	// Handle case where all lines are covered
	if !result.HasUncoveredLines() {
		msg := "::notice::All added lines are covered"
		if result.DiffAddedLines == 0 {
			msg = "::notice::No lines added in diff"
		}
		_, err := fmt.Fprintln(w, msg)
		return err
	}

	// Generate annotations using the coverage package
	annotations := coverage.GenerateAnnotations(result)

	// Format each annotation as a GitHub Actions workflow command
	for _, annotation := range annotations {
		var err error
		if annotation.StartLine == annotation.EndLine {
			_, err = fmt.Fprintf(w, "::notice file=%s,line=%d,title=%s::%s\n",
				annotation.Path, annotation.StartLine, annotation.Title, annotation.Message)
		} else {
			_, err = fmt.Fprintf(w, "::notice file=%s,line=%d,endLine=%d,title=%s::%s\n",
				annotation.Path, annotation.StartLine, annotation.EndLine, annotation.Title, annotation.Message)
		}
		if err != nil {
			return err
		}
	}

	return nil
}
