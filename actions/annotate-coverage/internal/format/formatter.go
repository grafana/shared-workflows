package format

import (
	"fmt"
	"io"

	"github.com/grafana/shared-workflows/actions/annotate-coverage/internal/coverage"
)

// Formatter formats coverage analysis results for output.
type Formatter interface {
	Format(result *coverage.AnalysisResult, w io.Writer) error
}

// New creates a formatter based on the specified format type.
// Supported formats: "Text", "Markdown", "GitHubAnnotations"
func New(format string) (Formatter, error) {
	switch format {
	case "Text":
		return &TextFormatter{}, nil
	case "Markdown":
		return &MarkdownFormatter{}, nil
	case "GitHubAnnotations":
		return &GitHubAnnotationsFormatter{}, nil
	default:
		return nil, fmt.Errorf("unknown format: %s (supported: Text, Markdown, GitHubAnnotations)", format)
	}
}
