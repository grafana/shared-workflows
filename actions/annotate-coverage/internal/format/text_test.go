package format

import (
	"bytes"
	"strings"
	"testing"

	"github.com/grafana/shared-workflows/actions/annotate-coverage/internal/coverage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTextFormatter_Format(t *testing.T) {
	tests := []struct {
		name           string
		result         *coverage.AnalysisResult
		expectedOutput string
		expectError    bool
	}{
		{
			name: "single file with uncovered lines",
			result: &coverage.AnalysisResult{
				UncoveredByFile: map[string][]int{
					"main.go": {5, 10, 15},
				},
				DiffAddedLines:   20,
				DiffAddedCovered: 17, // 20 - 3 uncovered = 17 covered
			},
			expectedOutput: `Uncovered lines in diff:

main.go
  Lines: 5, 10, 15

Summary: 3 uncovered lines out of 20 added lines (85.0% coverage)
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
			expectedOutput: `Uncovered lines in diff:

handler.go
  Lines: 10, 15, 20

main.go
  Lines: 5, 7

Summary: 5 uncovered lines out of 25 added lines (80.0% coverage)
`,
		},
		{
			name: "all lines covered",
			result: &coverage.AnalysisResult{
				UncoveredByFile:  map[string][]int{},
				DiffAddedLines:   10,
				DiffAddedCovered: 10, // all 10 covered
			},
			expectedOutput: "All added lines are covered!\n",
		},
		{
			name: "no lines added",
			result: &coverage.AnalysisResult{
				UncoveredByFile:  map[string][]int{},
				DiffAddedLines:   0,
				DiffAddedCovered: 0,
			},
			expectedOutput: "No lines added in diff\n",
		},
		{
			name: "lines with ranges",
			result: &coverage.AnalysisResult{
				UncoveredByFile: map[string][]int{
					"server.go": {5, 6, 7, 10, 11, 15, 20, 21, 22, 23},
				},
				DiffAddedLines:   15,
				DiffAddedCovered: 5, // 15 - 10 uncovered = 5 covered
			},
			expectedOutput: `Uncovered lines in diff:

server.go
  Lines: 5-7, 10-11, 15, 20-23

Summary: 10 uncovered lines out of 15 added lines (33.3% coverage)
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
			formatter := &TextFormatter{}
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

func TestFormatLineRanges(t *testing.T) {
	tests := []struct {
		name     string
		lines    []int
		expected string
	}{
		{
			name:     "empty list",
			lines:    []int{},
			expected: "",
		},
		{
			name:     "single line",
			lines:    []int{5},
			expected: "5",
		},
		{
			name:     "two consecutive lines",
			lines:    []int{5, 6},
			expected: "5-6",
		},
		{
			name:     "three consecutive lines",
			lines:    []int{5, 6, 7},
			expected: "5-7",
		},
		{
			name:     "non-consecutive lines",
			lines:    []int{5, 10, 15},
			expected: "5, 10, 15",
		},
		{
			name:     "mixed consecutive and non-consecutive",
			lines:    []int{5, 6, 7, 10, 15, 16},
			expected: "5-7, 10, 15-16",
		},
		{
			name:     "all consecutive",
			lines:    []int{1, 2, 3, 4, 5},
			expected: "1-5",
		},
		{
			name:     "complex pattern",
			lines:    []int{1, 3, 4, 5, 8, 9, 12},
			expected: "1, 3-5, 8-9, 12",
		},
		{
			name:     "long consecutive range",
			lines:    []int{10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
			expected: "10-20",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatLineRanges(tt.lines)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAppendRange(t *testing.T) {
	tests := []struct {
		name     string
		result   string
		start    int
		end      int
		expected string
	}{
		{
			name:     "single line to empty result",
			result:   "",
			start:    5,
			end:      5,
			expected: "5",
		},
		{
			name:     "range to empty result",
			result:   "",
			start:    5,
			end:      10,
			expected: "5-10",
		},
		{
			name:     "single line to existing result",
			result:   "1-3",
			start:    5,
			end:      5,
			expected: "1-3, 5",
		},
		{
			name:     "range to existing result",
			result:   "1-3",
			start:    5,
			end:      10,
			expected: "1-3, 5-10",
		},
		{
			name:     "multiple appends",
			result:   "1, 3-5",
			start:    10,
			end:      12,
			expected: "1, 3-5, 10-12",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := appendRange(tt.result, tt.start, tt.end)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTextFormatter_AlphabeticalOrder(t *testing.T) {
	// Test that files are output in alphabetical order
	result := &coverage.AnalysisResult{
		UncoveredByFile: map[string][]int{
			"zebra.go":   {1},
			"alpha.go":   {2},
			"charlie.go": {3},
		},
		DiffAddedLines:   3,
		DiffAddedCovered: 0, // all 3 uncovered
	}

	formatter := &TextFormatter{}
	var buf bytes.Buffer
	err := formatter.Format(result, &buf)
	require.NoError(t, err)

	output := buf.String()

	// Find the positions of the filenames in the output
	alphaPos := strings.Index(output, "alpha.go")
	charliePos := strings.Index(output, "charlie.go")
	zebraPos := strings.Index(output, "zebra.go")

	// Verify alphabetical order
	assert.True(t, alphaPos < charliePos, "alpha.go should come before charlie.go")
	assert.True(t, charliePos < zebraPos, "charlie.go should come before zebra.go")
}
