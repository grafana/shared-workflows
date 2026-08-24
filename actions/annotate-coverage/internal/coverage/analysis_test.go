package coverage

import (
	"testing"

	"github.com/grafana/shared-workflows/actions/annotate-coverage/internal/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalyzeCoverage(t *testing.T) {
	tests := []struct {
		name                     string
		profiles                 []*Profile
		addedLinesByFile         map[string][]int
		expectedUncoveredByFile  map[string][]int
		expectedTotalLines       int
		expectedTotalCovered     int
		expectedDiffAddedLines   int
		expectedDiffAddedCovered int
	}{
		{
			name: "all lines covered",
			profiles: []*Profile{
				{
					FileName: "github.com/org/repo/main.go",
					Mode:     "set",
					Blocks: []ProfileBlock{
						{StartLine: 1, EndLine: 5, Count: 1},
						{StartLine: 6, EndLine: 10, Count: 1},
					},
				},
			},
			addedLinesByFile: map[string][]int{
				"main.go": {3, 7},
			},
			expectedUncoveredByFile:  map[string][]int{},
			expectedTotalLines:       10, // 5 + 5 lines
			expectedTotalCovered:     10, // all covered
			expectedDiffAddedLines:   2,  // 2 lines added in diff
			expectedDiffAddedCovered: 2,  // both covered
		},
		{
			name: "some lines uncovered",
			profiles: []*Profile{
				{
					FileName: "github.com/org/repo/main.go",
					Mode:     "set",
					Blocks: []ProfileBlock{
						{StartLine: 1, EndLine: 5, Count: 1},  // covered
						{StartLine: 6, EndLine: 10, Count: 0}, // not covered
					},
				},
			},
			addedLinesByFile: map[string][]int{
				"main.go": {3, 7, 9},
			},
			expectedUncoveredByFile: map[string][]int{
				"main.go": {7, 9},
			},
			expectedTotalLines:       10, // 5 + 5 lines
			expectedTotalCovered:     5,  // only first block
			expectedDiffAddedLines:   3,  // 3 lines added
			expectedDiffAddedCovered: 1,  // only line 3 is covered
		},
		{
			name: "no coverage for file",
			profiles: []*Profile{
				{
					FileName: "github.com/org/repo/other.go",
					Mode:     "set",
					Blocks: []ProfileBlock{
						{StartLine: 1, EndLine: 10, Count: 1},
					},
				},
			},
			addedLinesByFile: map[string][]int{
				"main.go": {1, 2, 3},
			},
			expectedUncoveredByFile:  map[string][]int{},
			expectedTotalLines:       10, // other.go has 10 lines
			expectedTotalCovered:     10, // all covered
			expectedDiffAddedLines:   0,  // no matching file in diff
			expectedDiffAddedCovered: 0,
		},
		{
			name: "multiple files mixed coverage",
			profiles: []*Profile{
				{
					FileName: "github.com/org/repo/server.go",
					Mode:     "set",
					Blocks: []ProfileBlock{
						{StartLine: 10, EndLine: 20, Count: 1},
						{StartLine: 30, EndLine: 40, Count: 0},
					},
				},
				{
					FileName: "github.com/org/repo/handler.go",
					Mode:     "set",
					Blocks: []ProfileBlock{
						{StartLine: 5, EndLine: 15, Count: 1},
					},
				},
			},
			addedLinesByFile: map[string][]int{
				"server.go":  {15, 35},
				"handler.go": {10, 20}, // line 20 not in any block, will be ignored
			},
			expectedUncoveredByFile: map[string][]int{
				"server.go": {35}, // only server.go line 35 is uncovered
			},
			expectedTotalLines:       33, // server: 11+11=22, handler: 11 = 33
			expectedTotalCovered:     22, // server: 11, handler: 11 = 22
			expectedDiffAddedLines:   3,  // only instrumented lines: 15, 35 in server.go + 10 in handler.go (line 20 ignored)
			expectedDiffAddedCovered: 2,  // line 15 (server) and line 10 (handler) are covered
		},
		{
			name:                     "empty diff",
			profiles:                 []*Profile{},
			addedLinesByFile:         map[string][]int{},
			expectedUncoveredByFile:  map[string][]int{},
			expectedTotalLines:       0,
			expectedTotalCovered:     0,
			expectedDiffAddedLines:   0,
			expectedDiffAddedCovered: 0,
		},
		{
			name: "file with coverage not in diff",
			profiles: []*Profile{
				{
					FileName: "github.com/org/repo/other.go",
					Mode:     "set",
					Blocks: []ProfileBlock{
						{StartLine: 1, EndLine: 5, Count: 0},
						{StartLine: 6, EndLine: 10, Count: 1},
					},
				},
			},
			addedLinesByFile: map[string][]int{
				"main.go": {1, 2, 3},
			},
			expectedUncoveredByFile:  map[string][]int{},
			expectedTotalLines:       10, // 5 + 5 lines
			expectedTotalCovered:     5,  // only second block
			expectedDiffAddedLines:   0,  // no matching file
			expectedDiffAddedCovered: 0,
		},
		{
			name:     "no coverage data",
			profiles: []*Profile{},
			addedLinesByFile: map[string][]int{
				"main.go": {1, 2, 3},
			},
			expectedUncoveredByFile:  map[string][]int{},
			expectedTotalLines:       0,
			expectedTotalCovered:     0,
			expectedDiffAddedLines:   0,
			expectedDiffAddedCovered: 0,
		},
		{
			// Regression: a profile path must only match a diff file on a path
			// boundary. "myhandler.go" must NOT match a diff entry for
			// "handler.go", otherwise coverage gets attributed to the wrong file.
			name: "suffix match respects path boundaries",
			profiles: []*Profile{
				{
					FileName: "github.com/org/repo/myhandler.go",
					Mode:     "set",
					Blocks: []ProfileBlock{
						{StartLine: 1, EndLine: 10, Count: 0}, // uncovered
					},
				},
			},
			addedLinesByFile: map[string][]int{
				"handler.go": {5}, // different file; must not be attributed to myhandler.go
			},
			expectedUncoveredByFile:  map[string][]int{},
			expectedTotalLines:       10,
			expectedTotalCovered:     0,
			expectedDiffAddedLines:   0, // no file match => nothing analyzed
			expectedDiffAddedCovered: 0,
		},
		{
			name: "nested file paths",
			profiles: []*Profile{
				{
					FileName: "github.com/org/repo/internal/server/handler.go",
					Mode:     "set",
					Blocks: []ProfileBlock{
						{StartLine: 1, EndLine: 10, Count: 1},
						{StartLine: 11, EndLine: 20, Count: 0},
					},
				},
			},
			addedLinesByFile: map[string][]int{
				"internal/server/handler.go": {5, 15},
			},
			expectedUncoveredByFile: map[string][]int{
				"internal/server/handler.go": {15},
			},
			expectedTotalLines:       20, // 10 + 10 lines
			expectedTotalCovered:     10, // first block
			expectedDiffAddedLines:   2,  // 2 lines added
			expectedDiffAddedCovered: 1,  // line 5 covered, line 15 not
		},
		{
			name: "non-instrumented lines are ignored",
			profiles: []*Profile{
				{
					FileName: "github.com/org/repo/main.go",
					Mode:     "set",
					Blocks: []ProfileBlock{
						{StartLine: 5, EndLine: 7, Count: 1},   // covered: lines 5-7
						{StartLine: 10, EndLine: 12, Count: 0}, // uncovered: lines 10-12
					},
				},
			},
			addedLinesByFile: map[string][]int{
				"main.go": {1, 3, 6, 8, 11, 15}, // lines 1,3,8,15 are not in any block
			},
			expectedUncoveredByFile: map[string][]int{
				"main.go": {11}, // only line 11 is uncovered (in block with Count=0)
			},
			expectedTotalLines:       6, // 3 + 3 lines
			expectedTotalCovered:     3, // first block only
			expectedDiffAddedLines:   2, // only instrumented lines: 6 and 11
			expectedDiffAddedCovered: 1, // only line 6 is covered and instrumented
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AnalyzeCoverage(tt.profiles, tt.addedLinesByFile)

			assert.Equal(t, tt.expectedUncoveredByFile, result.UncoveredByFile)
			assert.Equal(t, tt.expectedTotalLines, result.TotalLines)
			assert.Equal(t, tt.expectedTotalCovered, result.TotalCovered)
			assert.Equal(t, tt.expectedDiffAddedLines, result.DiffAddedLines)
			assert.Equal(t, tt.expectedDiffAddedCovered, result.DiffAddedCovered)
		})
	}
}

func TestIsLineInstrumented(t *testing.T) {
	profile := &Profile{
		FileName: "test.go",
		Mode:     "set",
		Blocks: []ProfileBlock{
			{StartLine: 5, EndLine: 10, Count: 1},  // covered
			{StartLine: 15, EndLine: 20, Count: 0}, // not covered
			{StartLine: 25, EndLine: 30, Count: 5}, // covered (count mode)
		},
	}

	tests := []struct {
		name     string
		profile  *Profile
		line     int
		expected bool
	}{
		{
			name:     "line in covered block",
			profile:  profile,
			line:     7,
			expected: true,
		},
		{
			name:     "line in uncovered block",
			profile:  profile,
			line:     17,
			expected: true,
		},
		{
			name:     "line not in any block",
			profile:  profile,
			line:     12,
			expected: false,
		},
		{
			name:     "line before all blocks",
			profile:  profile,
			line:     1,
			expected: false,
		},
		{
			name:     "line after all blocks",
			profile:  profile,
			line:     50,
			expected: false,
		},
		{
			name:     "line at block start",
			profile:  profile,
			line:     5,
			expected: true,
		},
		{
			name:     "line at block end",
			profile:  profile,
			line:     10,
			expected: true,
		},
		{
			name:     "nil profile",
			profile:  nil,
			line:     5,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isLineInstrumented(tt.profile, tt.line)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsLineCovered(t *testing.T) {
	profile := &Profile{
		FileName: "test.go",
		Mode:     "set",
		Blocks: []ProfileBlock{
			{StartLine: 5, EndLine: 10, Count: 1},  // covered
			{StartLine: 15, EndLine: 20, Count: 0}, // not covered
			{StartLine: 25, EndLine: 30, Count: 5}, // covered (count mode)
		},
	}

	tests := []struct {
		name     string
		profile  *Profile
		line     int
		expected bool
	}{
		{
			name:     "line in covered block - start",
			profile:  profile,
			line:     5,
			expected: true,
		},
		{
			name:     "line in covered block - middle",
			profile:  profile,
			line:     7,
			expected: true,
		},
		{
			name:     "line in covered block - end",
			profile:  profile,
			line:     10,
			expected: true,
		},
		{
			name:     "line in uncovered block",
			profile:  profile,
			line:     17,
			expected: false,
		},
		{
			name:     "line not in any block",
			profile:  profile,
			line:     12,
			expected: false,
		},
		{
			name:     "line before all blocks",
			profile:  profile,
			line:     1,
			expected: false,
		},
		{
			name:     "line after all blocks",
			profile:  profile,
			line:     50,
			expected: false,
		},
		{
			name:     "line in covered block with high count",
			profile:  profile,
			line:     27,
			expected: true,
		},
		{
			name:     "nil profile",
			profile:  nil,
			line:     5,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isLineCovered(tt.profile, tt.line)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAnalysisResult_HasUncoveredLines(t *testing.T) {
	tests := []struct {
		name     string
		result   *AnalysisResult
		expected bool
	}{
		{
			name: "has uncovered lines",
			result: &AnalysisResult{
				DiffAddedLines:   5,
				DiffAddedCovered: 3, // 2 uncovered
			},
			expected: true,
		},
		{
			name: "no uncovered lines",
			result: &AnalysisResult{
				DiffAddedLines:   5,
				DiffAddedCovered: 5, // all covered
			},
			expected: false,
		},
		{
			name: "no lines added",
			result: &AnalysisResult{
				DiffAddedLines:   0,
				DiffAddedCovered: 0,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.result.HasUncoveredLines())
		})
	}
}

func TestAnalysisResult_GetSortedFiles(t *testing.T) {
	result := &AnalysisResult{
		UncoveredByFile: map[string][]int{
			"zebra.go":   {1, 2},
			"alpha.go":   {3, 4},
			"charlie.go": {5, 6},
		},
	}

	files := result.GetSortedFiles()
	expected := []string{"alpha.go", "charlie.go", "zebra.go"}

	assert.Equal(t, expected, files)
}

func TestAnalyzeCoverage_EdgeCases(t *testing.T) {
	t.Run("line at block boundary", func(t *testing.T) {
		profiles := []*Profile{
			{
				FileName: "test.go",
				Blocks: []ProfileBlock{
					{StartLine: 1, EndLine: 5, Count: 1},
					{StartLine: 6, EndLine: 10, Count: 0},
				},
			},
		}

		addedLines := map[string][]int{
			"test.go": {5, 6}, // boundary lines
		}

		result := AnalyzeCoverage(profiles, addedLines)

		// Line 5 should be covered (end of first block)
		// Line 6 should be uncovered (start of second block with Count=0)
		assert.Equal(t, map[string][]int{"test.go": {6}}, result.UncoveredByFile)
		assert.Equal(t, 10, result.TotalLines)      // 5 + 5 lines
		assert.Equal(t, 5, result.TotalCovered)     // first block only
		assert.Equal(t, 2, result.DiffAddedLines)   // 2 lines added
		assert.Equal(t, 1, result.DiffAddedCovered) // line 5 covered
	})

	t.Run("single line block", func(t *testing.T) {
		profiles := []*Profile{
			{
				FileName: "test.go",
				Blocks: []ProfileBlock{
					{StartLine: 5, EndLine: 5, Count: 1},
				},
			},
		}

		addedLines := map[string][]int{
			"test.go": {5},
		}

		result := AnalyzeCoverage(profiles, addedLines)

		assert.Equal(t, map[string][]int{}, result.UncoveredByFile)
		assert.Equal(t, 1, result.TotalLines)       // 1 line
		assert.Equal(t, 1, result.TotalCovered)     // covered
		assert.Equal(t, 1, result.DiffAddedLines)   // 1 line added
		assert.Equal(t, 1, result.DiffAddedCovered) // covered
	})
}

func TestCalculateCoverageStats(t *testing.T) {
	tests := []struct {
		name                   string
		profiles               []*Profile
		expectedTotal          int
		expectedCovered        int
		expectedPercentage     float64
		expectedFileCount      int
		checkSpecificFile      string
		expectedFileTotal      int
		expectedFileCovered    int
		expectedFilePercentage float64
	}{
		{
			name: "single file fully covered",
			profiles: []*Profile{
				{
					FileName: "main.go",
					Blocks: []ProfileBlock{
						{NumStmt: 5, Count: 10},
						{NumStmt: 3, Count: 5},
					},
				},
			},
			expectedTotal:          8,
			expectedCovered:        8,
			expectedPercentage:     100.0,
			expectedFileCount:      1,
			checkSpecificFile:      "main.go",
			expectedFileTotal:      8,
			expectedFileCovered:    8,
			expectedFilePercentage: 100.0,
		},
		{
			name: "single file partially covered",
			profiles: []*Profile{
				{
					FileName: "handler.go",
					Blocks: []ProfileBlock{
						{NumStmt: 10, Count: 5}, // covered
						{NumStmt: 10, Count: 0}, // not covered
					},
				},
			},
			expectedTotal:          20,
			expectedCovered:        10,
			expectedPercentage:     50.0,
			expectedFileCount:      1,
			checkSpecificFile:      "handler.go",
			expectedFileTotal:      20,
			expectedFileCovered:    10,
			expectedFilePercentage: 50.0,
		},
		{
			name: "multiple files mixed coverage",
			profiles: []*Profile{
				{
					FileName: "server.go",
					Blocks: []ProfileBlock{
						{NumStmt: 10, Count: 1},
						{NumStmt: 10, Count: 0},
					},
				},
				{
					FileName: "client.go",
					Blocks: []ProfileBlock{
						{NumStmt: 5, Count: 2},
						{NumStmt: 5, Count: 3},
					},
				},
			},
			expectedTotal:          30,
			expectedCovered:        20,
			expectedPercentage:     66.66666666666666,
			expectedFileCount:      2,
			checkSpecificFile:      "server.go",
			expectedFileTotal:      20,
			expectedFileCovered:    10,
			expectedFilePercentage: 50.0,
		},
		{
			name:               "no profiles",
			profiles:           []*Profile{},
			expectedTotal:      0,
			expectedCovered:    0,
			expectedPercentage: 0.0,
			expectedFileCount:  0,
		},
		{
			name: "file with no coverage",
			profiles: []*Profile{
				{
					FileName: "unused.go",
					Blocks: []ProfileBlock{
						{NumStmt: 15, Count: 0},
						{NumStmt: 10, Count: 0},
					},
				},
			},
			expectedTotal:          25,
			expectedCovered:        0,
			expectedPercentage:     0.0,
			expectedFileCount:      1,
			checkSpecificFile:      "unused.go",
			expectedFileTotal:      25,
			expectedFileCovered:    0,
			expectedFilePercentage: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := CalculateCoverageStats(tt.profiles)

			assert.Equal(t, tt.expectedTotal, stats.TotalStatements)
			assert.Equal(t, tt.expectedCovered, stats.CoveredStatements)
			assert.Equal(t, tt.expectedPercentage, stats.Percentage)
			assert.Equal(t, tt.expectedFileCount, len(stats.ByFile))

			// Check specific file if specified
			if tt.checkSpecificFile != "" {
				fileStats, ok := stats.ByFile[tt.checkSpecificFile]
				require.True(t, ok, "expected file %s in ByFile map", tt.checkSpecificFile)
				assert.Equal(t, tt.checkSpecificFile, fileStats.FileName)
				assert.Equal(t, tt.expectedFileTotal, fileStats.TotalStatements)
				assert.Equal(t, tt.expectedFileCovered, fileStats.CoveredStatements)
				assert.Equal(t, tt.expectedFilePercentage, fileStats.Percentage)
			}
		})
	}
}

func TestCompareCoverage(t *testing.T) {
	tests := []struct {
		name              string
		base              *CoverageStats
		head              *CoverageStats
		expectedBase      float64
		expectedHead      float64
		expectedDelta     float64
		expectedDecreased bool
	}{
		{
			name: "coverage increased",
			base: &CoverageStats{
				Percentage: 50.0,
			},
			head: &CoverageStats{
				Percentage: 75.0,
			},
			expectedBase:      50.0,
			expectedHead:      75.0,
			expectedDelta:     25.0,
			expectedDecreased: false,
		},
		{
			name: "coverage decreased",
			base: &CoverageStats{
				Percentage: 80.0,
			},
			head: &CoverageStats{
				Percentage: 60.0,
			},
			expectedBase:      80.0,
			expectedHead:      60.0,
			expectedDelta:     -20.0,
			expectedDecreased: true,
		},
		{
			name: "coverage unchanged",
			base: &CoverageStats{
				Percentage: 70.0,
			},
			head: &CoverageStats{
				Percentage: 70.0,
			},
			expectedBase:      70.0,
			expectedHead:      70.0,
			expectedDelta:     0.0,
			expectedDecreased: false,
		},
		{
			name: "nil base (first coverage report)",
			base: nil,
			head: &CoverageStats{
				Percentage: 65.0,
			},
			expectedBase:      0.0,
			expectedHead:      65.0,
			expectedDelta:     65.0,
			expectedDecreased: false,
		},
		{
			name: "nil head",
			base: &CoverageStats{
				Percentage: 50.0,
			},
			head:              nil,
			expectedBase:      50.0,
			expectedHead:      0.0,
			expectedDelta:     -50.0,
			expectedDecreased: true,
		},
		{
			name:              "both nil",
			base:              nil,
			head:              nil,
			expectedBase:      0.0,
			expectedHead:      0.0,
			expectedDelta:     0.0,
			expectedDecreased: false,
		},
		{
			name: "small increase",
			base: &CoverageStats{
				Percentage: 75.5,
			},
			head: &CoverageStats{
				Percentage: 75.6,
			},
			expectedBase:      75.5,
			expectedHead:      75.6,
			expectedDelta:     0.09999999999999432, // float precision
			expectedDecreased: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comparison := CompareCoverage(tt.base, tt.head)

			assert.Equal(t, tt.expectedBase, comparison.BaseCoverage)
			assert.Equal(t, tt.expectedHead, comparison.HeadCoverage)
			assert.InDelta(t, tt.expectedDelta, comparison.Delta, 0.0001) // use InDelta for float comparison
			assert.Equal(t, tt.expectedDecreased, comparison.Decreased)
		})
	}
}

func TestGenerateAnnotations(t *testing.T) {
	tests := []struct {
		name                string
		result              *AnalysisResult
		expectedAnnotations []*github.Annotation
	}{
		{
			name: "single file with non-consecutive lines",
			result: &AnalysisResult{
				UncoveredByFile: map[string][]int{
					"main.go": {5, 10, 15},
				},
			},
			expectedAnnotations: []*github.Annotation{
				{
					Path:      "main.go",
					StartLine: 5,
					EndLine:   5,
					Level:     "notice",
					Title:     "Uncovered line",
					Message:   "Line 5 is not covered by tests",
				},
				{
					Path:      "main.go",
					StartLine: 10,
					EndLine:   10,
					Level:     "notice",
					Title:     "Uncovered line",
					Message:   "Line 10 is not covered by tests",
				},
				{
					Path:      "main.go",
					StartLine: 15,
					EndLine:   15,
					Level:     "notice",
					Title:     "Uncovered line",
					Message:   "Line 15 is not covered by tests",
				},
			},
		},
		{
			name: "consecutive lines grouped into range",
			result: &AnalysisResult{
				UncoveredByFile: map[string][]int{
					"server.go": {5, 6, 7, 10, 11, 15},
				},
			},
			expectedAnnotations: []*github.Annotation{
				{
					Path:      "server.go",
					StartLine: 5,
					EndLine:   7,
					Level:     "notice",
					Title:     "Uncovered lines",
					Message:   "Lines 5-7 are not covered by tests",
				},
				{
					Path:      "server.go",
					StartLine: 10,
					EndLine:   11,
					Level:     "notice",
					Title:     "Uncovered lines",
					Message:   "Lines 10-11 are not covered by tests",
				},
				{
					Path:      "server.go",
					StartLine: 15,
					EndLine:   15,
					Level:     "notice",
					Title:     "Uncovered line",
					Message:   "Line 15 is not covered by tests",
				},
			},
		},
		{
			name: "multiple files sorted alphabetically",
			result: &AnalysisResult{
				UncoveredByFile: map[string][]int{
					"zebra.go": {1, 2, 3},
					"alpha.go": {10},
					"bravo.go": {20, 21},
				},
			},
			expectedAnnotations: []*github.Annotation{
				{
					Path:      "alpha.go",
					StartLine: 10,
					EndLine:   10,
					Level:     "notice",
					Title:     "Uncovered line",
					Message:   "Line 10 is not covered by tests",
				},
				{
					Path:      "bravo.go",
					StartLine: 20,
					EndLine:   21,
					Level:     "notice",
					Title:     "Uncovered lines",
					Message:   "Lines 20-21 are not covered by tests",
				},
				{
					Path:      "zebra.go",
					StartLine: 1,
					EndLine:   3,
					Level:     "notice",
					Title:     "Uncovered lines",
					Message:   "Lines 1-3 are not covered by tests",
				},
			},
		},
		{
			name:                "nil result",
			result:              nil,
			expectedAnnotations: nil,
		},
		{
			name: "no uncovered lines",
			result: &AnalysisResult{
				UncoveredByFile: map[string][]int{},
			},
			expectedAnnotations: nil,
		},
		{
			name: "unsorted lines are sorted and grouped",
			result: &AnalysisResult{
				UncoveredByFile: map[string][]int{
					"test.go": {15, 10, 12, 11, 5},
				},
			},
			expectedAnnotations: []*github.Annotation{
				{
					Path:      "test.go",
					StartLine: 5,
					EndLine:   5,
					Level:     "notice",
					Title:     "Uncovered line",
					Message:   "Line 5 is not covered by tests",
				},
				{
					Path:      "test.go",
					StartLine: 10,
					EndLine:   12,
					Level:     "notice",
					Title:     "Uncovered lines",
					Message:   "Lines 10-12 are not covered by tests",
				},
				{
					Path:      "test.go",
					StartLine: 15,
					EndLine:   15,
					Level:     "notice",
					Title:     "Uncovered line",
					Message:   "Line 15 is not covered by tests",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			annotations := GenerateAnnotations(tt.result)

			assert.Equal(t, len(tt.expectedAnnotations), len(annotations))
			for i, expected := range tt.expectedAnnotations {
				if i < len(annotations) {
					actual := annotations[i]
					assert.Equal(t, expected.Path, actual.Path, "annotation %d path mismatch", i)
					assert.Equal(t, expected.StartLine, actual.StartLine, "annotation %d start line mismatch", i)
					assert.Equal(t, expected.EndLine, actual.EndLine, "annotation %d end line mismatch", i)
					assert.Equal(t, expected.Level, actual.Level, "annotation %d level mismatch", i)
					assert.Equal(t, expected.Title, actual.Title, "annotation %d title mismatch", i)
					assert.Equal(t, expected.Message, actual.Message, "annotation %d message mismatch", i)
				}
			}
		})
	}
}
