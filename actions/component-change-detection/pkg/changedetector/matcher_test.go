package changedetector

import (
	"strings"
	"sync"
	"testing"
)

func TestMatchPaths(t *testing.T) {
	tests := []struct {
		name     string
		file     string
		includes []string
		excludes []string
		want     bool
		wantErr  bool
	}{
		{
			name:     "matches single path",
			file:     "pkg/app/main.go",
			includes: []string{"pkg/**"},
			excludes: []string{},
			want:     true,
		},
		{
			name:     "matches nested path",
			file:     "pkg/grafanacloud/stack/api/handler.go",
			includes: []string{"pkg/**"},
			excludes: []string{},
			want:     true,
		},
		{
			name:     "excluded by pattern",
			file:     "pkg/app/main_test.go",
			includes: []string{"pkg/**"},
			excludes: []string{"**/*_test.go"},
			want:     false,
		},
		{
			name:     "multiple includes - first matches",
			file:     "cmd/apiserver/main.go",
			includes: []string{"cmd/apiserver/**", "cmd/controller/**"},
			excludes: []string{},
			want:     true,
		},
		{
			name:     "multiple includes - second matches",
			file:     "cmd/controller/config.go",
			includes: []string{"cmd/apiserver/**", "cmd/controller/**"},
			excludes: []string{},
			want:     true,
		},
		{
			name:     "no match",
			file:     "docs/README.md",
			includes: []string{"pkg/**", "cmd/**"},
			excludes: []string{},
			want:     false,
		},
		{
			name:     "exact file match",
			file:     "go.mod",
			includes: []string{"go.mod", "go.sum"},
			excludes: []string{},
			want:     true,
		},
		{
			name:     "exclude takes precedence",
			file:     "pkg/testutil/helper.go",
			includes: []string{"pkg/**"},
			excludes: []string{"pkg/testutil/**"},
			want:     false,
		},
		{
			name:     "double-star matches root-level file",
			file:     "Dockerfile",
			includes: []string{"**"},
			excludes: []string{},
			want:     true,
		},
		{
			name:     "double-star matches nested file",
			file:     "cmd/apiserver/main.go",
			includes: []string{"**"},
			excludes: []string{},
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MatchPaths(tt.file, tt.includes, tt.excludes)
			if (err != nil) != tt.wantErr {
				t.Errorf("MatchPaths() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("MatchPaths() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestMatchPaths_EdgeCases tests edge cases that might occur in real-world scenarios
func TestMatchPaths_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		file     string
		includes []string
		excludes []string
		want     bool
		wantErr  bool
	}{
		// Files with spaces
		{
			name:     "file with single space",
			file:     "pkg/my file.go",
			includes: []string{"pkg/**"},
			excludes: []string{},
			want:     true,
		},
		{
			name:     "file with multiple spaces",
			file:     "pkg/my   multiple   spaces.go",
			includes: []string{"pkg/**"},
			excludes: []string{},
			want:     true,
		},
		{
			name:     "directory with spaces",
			file:     "my directory/src/main.go",
			includes: []string{"my directory/**"},
			excludes: []string{},
			want:     true,
		},
		{
			name:     "spaces in exclude pattern",
			file:     "my directory/test file.go",
			includes: []string{"**"},
			excludes: []string{"my directory/**"},
			want:     false,
		},

		// Unicode characters
		{
			name:     "unicode characters in filename - emoji",
			file:     "pkg/🚀rocket.go",
			includes: []string{"pkg/**"},
			excludes: []string{},
			want:     true,
		},
		{
			name:     "unicode characters - chinese",
			file:     "pkg/文件.go",
			includes: []string{"pkg/**"},
			excludes: []string{},
			want:     true,
		},
		{
			name:     "unicode characters - arabic",
			file:     "pkg/ملف.go",
			includes: []string{"pkg/**"},
			excludes: []string{},
			want:     true,
		},
		{
			name:     "unicode characters - accents",
			file:     "pkg/café/naïve.go",
			includes: []string{"pkg/**"},
			excludes: []string{},
			want:     true,
		},
		{
			name:     "unicode in pattern",
			file:     "pkg/café/naïve.go",
			includes: []string{"pkg/café/**"},
			excludes: []string{},
			want:     true,
		},

		// Special characters
		{
			name:     "file with parentheses",
			file:     "pkg/file(1).go",
			includes: []string{"pkg/**"},
			excludes: []string{},
			want:     true,
		},
		{
			name:     "file with brackets",
			file:     "pkg/file[copy].go",
			includes: []string{"pkg/**"},
			excludes: []string{},
			want:     true,
		},
		{
			name:     "file with dots",
			file:     "pkg/my.file.name.go",
			includes: []string{"pkg/**"},
			excludes: []string{},
			want:     true,
		},
		{
			name:     "file with dashes and underscores",
			file:     "pkg/my-file_name.go",
			includes: []string{"pkg/**"},
			excludes: []string{},
			want:     true,
		},
		{
			name:     "file with plus sign",
			file:     "pkg/file+version.go",
			includes: []string{"pkg/**"},
			excludes: []string{},
			want:     true,
		},
		{
			name:     "file with ampersand",
			file:     "pkg/file&copy.go",
			includes: []string{"pkg/**"},
			excludes: []string{},
			want:     true,
		},

		// Very long paths
		{
			name: "long path under 4096 chars",
			file: "pkg/" + strings.Repeat("a/", 500) + "file.go", // ~1500 chars
			includes: []string{"pkg/**"},
			excludes: []string{},
			want:     true,
		},
		{
			name: "very long filename",
			file: "pkg/" + strings.Repeat("a", 255) + ".go",
			includes: []string{"pkg/**"},
			excludes: []string{},
			want:     true,
		},
		{
			name: "exclude pattern with very long path",
			file: "test/" + strings.Repeat("nested/", 100) + "file.go",
			includes: []string{"**"},
			excludes: []string{"test/**"},
			want:     false,
		},

		// Edge cases with pattern matching
		{
			name:     "empty filename components",
			file:     "pkg//file.go",
			includes: []string{"pkg/**"},
			excludes: []string{},
			want:     true,
		},
		{
			name:     "trailing slash in path",
			file:     "pkg/dir/",
			includes: []string{"pkg/**"},
			excludes: []string{},
			want:     true,
		},
		{
			name:     "leading slash in path",
			file:     "/pkg/file.go",
			includes: []string{"/pkg/**"},
			excludes: []string{},
			want:     true,
		},
		{
			name:     "relative path with dot",
			file:     "./pkg/file.go",
			includes: []string{"./pkg/**"},
			excludes: []string{},
			want:     true,
		},
		{
			name:     "path with parent directory reference",
			file:     "pkg/../cmd/file.go",
			includes: []string{"cmd/**"},
			excludes: []string{},
			want:     false, // Pattern doesn't normalize paths
		},

		// Empty cases
		{
			name:     "empty file path with double-star pattern",
			file:     "",
			includes: []string{"**"},
			excludes: []string{},
			want:     true, // ** matches zero or more segments, including empty
		},
		{
			name:     "empty file path with specific pattern",
			file:     "",
			includes: []string{"pkg/**"},
			excludes: []string{},
			want:     false, // Empty path doesn't match specific directory
		},
		{
			name:     "empty includes",
			file:     "pkg/file.go",
			includes: []string{},
			excludes: []string{},
			want:     false,
		},
		{
			name:     "empty excludes is fine",
			file:     "pkg/file.go",
			includes: []string{"pkg/**"},
			excludes: []string{},
			want:     true,
		},

		// Case sensitivity
		{
			name:     "case sensitive match - lowercase",
			file:     "pkg/file.go",
			includes: []string{"PKG/**"},
			excludes: []string{},
			want:     false, // doublestar is case-sensitive by default
		},
		{
			name:     "case sensitive match - uppercase",
			file:     "PKG/FILE.go",
			includes: []string{"PKG/**"},
			excludes: []string{},
			want:     true,
		},

		// Complex exclude scenarios
		{
			name:     "multiple excludes - both match",
			file:     "pkg/test/helper_test.go",
			includes: []string{"pkg/**"},
			excludes: []string{"**/test/**", "**/*_test.go"},
			want:     false, // Excluded by first pattern
		},
		{
			name:     "exclude pattern more specific than include",
			file:     "pkg/internal/secret.go",
			includes: []string{"pkg/**"},
			excludes: []string{"**/internal/**"},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MatchPaths(tt.file, tt.includes, tt.excludes)
			if (err != nil) != tt.wantErr {
				t.Errorf("MatchPaths() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("MatchPaths() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestMatchPaths_ConcurrentAccess tests that MatchPaths is safe for concurrent use
func TestMatchPaths_ConcurrentAccess(t *testing.T) {
	// Test that multiple goroutines can safely call MatchPaths simultaneously
	// This is important in CI/CD where parallel builds might occur
	
	includes := []string{"pkg/**", "cmd/**", "internal/**"}
	excludes := []string{"**/*_test.go", "**/testdata/**"}
	
	testFiles := []string{
		"pkg/app/main.go",
		"pkg/app/main_test.go",
		"cmd/server/main.go",
		"internal/helper.go",
		"docs/README.md",
	}

	expectedResults := []bool{true, false, true, true, false}

	var wg sync.WaitGroup
	errChan := make(chan error, len(testFiles)*100)

	// Run 100 iterations with all test files concurrently
	for iteration := 0; iteration < 100; iteration++ {
		for i, file := range testFiles {
			wg.Add(1)
			go func(f string, expectedResult bool) {
				defer wg.Done()
				got, err := MatchPaths(f, includes, excludes)
				if err != nil {
					errChan <- err
					return
				}
				if got != expectedResult {
					errChan <- &testError{
						file:     f,
						got:      got,
						expected: expectedResult,
					}
				}
			}(file, expectedResults[i])
		}
	}

	wg.Wait()
	close(errChan)

	// Check for any errors
	for err := range errChan {
		t.Error(err)
	}
}

// testError is a custom error type for concurrent test failures
type testError struct {
	file     string
	got      bool
	expected bool
}

func (e *testError) Error() string {
	return "concurrent test failed for " + e.file + ": got " + 
		boolToString(e.got) + ", want " + boolToString(e.expected)
}

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// TestMatchPaths_PathLengthLimits tests behavior with extremely long paths
func TestMatchPaths_PathLengthLimits(t *testing.T) {
	tests := []struct {
		name        string
		pathBuilder func() string
		includes    []string
		excludes    []string
		want        bool
	}{
		{
			name: "path near typical filesystem limit (4096)",
			pathBuilder: func() string {
				// Build a path close to 4096 chars
				segment := "very_long_directory_name/"
				repeats := 4096 / len(segment)
				return "pkg/" + strings.Repeat(segment, repeats) + "file.go"
			},
			includes: []string{"pkg/**"},
			excludes: []string{},
			want:     true,
		},
		{
			name: "many nested directories",
			pathBuilder: func() string {
				// 500 levels of nesting
				return "pkg/" + strings.Repeat("a/", 500) + "file.go"
			},
			includes: []string{"pkg/**"},
			excludes: []string{},
			want:     true,
		},
		{
			name: "exclude with very long path",
			pathBuilder: func() string {
				return "test/" + strings.Repeat("nested/", 200) + "file.go"
			},
			includes: []string{"**"},
			excludes: []string{"test/**"},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.pathBuilder()
			got, err := MatchPaths(path, tt.includes, tt.excludes)
			if err != nil {
				t.Errorf("MatchPaths() unexpected error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("MatchPaths() = %v, want %v (path length: %d)", got, tt.want, len(path))
			}
		})
	}
}

// TestMatchPaths_SymlinkPaths tests that symlink-like paths are handled correctly
// Note: MatchPaths works on path strings, not actual filesystem operations,
// so we test that symlink-like path patterns are handled correctly
func TestMatchPaths_SymlinkPaths(t *testing.T) {
	tests := []struct {
		name     string
		file     string
		includes []string
		excludes []string
		want     bool
	}{
		{
			name:     "path that looks like symlink target",
			file:     "vendor/github.com/pkg/errors/errors.go",
			includes: []string{"vendor/**"},
			excludes: []string{},
			want:     true,
		},
		{
			name:     "path with symlink-like notation",
			file:     "pkg/link -> target/file.go",
			includes: []string{"pkg/**"},
			excludes: []string{},
			want:     true, // Treated as literal filename with arrow chars
		},
		{
			name:     "exclude vendor directories (common symlink target)",
			file:     "vendor/dependencies/lib.go",
			includes: []string{"**"},
			excludes: []string{"vendor/**"},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MatchPaths(tt.file, tt.includes, tt.excludes)
			if err != nil {
				t.Errorf("MatchPaths() unexpected error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("MatchPaths() = %v, want %v", got, tt.want)
			}
		})
	}
}

