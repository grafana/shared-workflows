package format

import (
	"bytes"
	"testing"

	"github.com/grafana/shared-workflows/actions/annotate-coverage/internal/coverage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitHubAnnotationsFormatter_Format(t *testing.T) {
	tests := []struct {
		name           string
		result         *coverage.AnalysisResult
		expectedOutput string
		expectError    bool
	}{
		{
			name: "single file with non-consecutive uncovered lines",
			result: &coverage.AnalysisResult{
				UncoveredByFile: map[string][]int{
					"main.go": {5, 10, 15},
				},
				DiffAddedLines:   20,
				DiffAddedCovered: 17, // 20 - 3 uncovered = 17 covered
			},
			expectedOutput: `::notice file=main.go,line=5,title=Uncovered line::Line 5 is not covered by tests
::notice file=main.go,line=10,title=Uncovered line::Line 10 is not covered by tests
::notice file=main.go,line=15,title=Uncovered line::Line 15 is not covered by tests
`,
		},
		{
			name: "multiple files with uncovered lines",
			result: &coverage.AnalysisResult{
				UncoveredByFile: map[string][]int{
					"handler.go": {10, 15, 20},
					"main.go":    {5, 7},
				},
				DiffAddedLines:   25,
				DiffAddedCovered: 20, // 25 - 5 uncovered = 20 covered
			},
			expectedOutput: `::notice file=handler.go,line=10,title=Uncovered line::Line 10 is not covered by tests
::notice file=handler.go,line=15,title=Uncovered line::Line 15 is not covered by tests
::notice file=handler.go,line=20,title=Uncovered line::Line 20 is not covered by tests
::notice file=main.go,line=5,title=Uncovered line::Line 5 is not covered by tests
::notice file=main.go,line=7,title=Uncovered line::Line 7 is not covered by tests
`,
		},
		{
			name: "all lines covered",
			result: &coverage.AnalysisResult{
				UncoveredByFile:  map[string][]int{},
				DiffAddedLines:   10,
				DiffAddedCovered: 10, // all 10 covered
			},
			expectedOutput: "::notice::All added lines are covered\n",
		},
		{
			name: "no lines added",
			result: &coverage.AnalysisResult{
				UncoveredByFile:  map[string][]int{},
				DiffAddedLines:   0,
				DiffAddedCovered: 0,
			},
			expectedOutput: "::notice::No lines added in diff\n",
		},
		{
			name: "consecutive lines grouped into ranges",
			result: &coverage.AnalysisResult{
				UncoveredByFile: map[string][]int{
					"server.go": {5, 6, 7, 10, 11, 15, 20, 21, 22, 23},
				},
				DiffAddedLines:   15,
				DiffAddedCovered: 5, // 15 - 10 uncovered = 5 covered
			},
			expectedOutput: `::notice file=server.go,line=5,endLine=7,title=Uncovered lines::Lines 5-7 are not covered by tests
::notice file=server.go,line=10,endLine=11,title=Uncovered lines::Lines 10-11 are not covered by tests
::notice file=server.go,line=15,title=Uncovered line::Line 15 is not covered by tests
::notice file=server.go,line=20,endLine=23,title=Uncovered lines::Lines 20-23 are not covered by tests
`,
		},
		{
			name: "single consecutive line range",
			result: &coverage.AnalysisResult{
				UncoveredByFile: map[string][]int{
					"test.go": {1, 2, 3, 4, 5},
				},
				DiffAddedLines:   10,
				DiffAddedCovered: 5, // 10 - 5 uncovered = 5 covered
			},
			expectedOutput: `::notice file=test.go,line=1,endLine=5,title=Uncovered lines::Lines 1-5 are not covered by tests
`,
		},
		{
			name:        "nil result",
			result:      nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := &GitHubAnnotationsFormatter{}
			var buf bytes.Buffer
			err := formatter.Format(tt.result, &buf)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedOutput, buf.String())
		})
	}
}
