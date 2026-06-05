package local

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunner_readAndMergeCoverageFiles(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T) string // Returns coverage directory path
		expectError bool
		errorMsg    string
	}{
		{
			name: "single valid coverage file",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				coveragePath := filepath.Join(tmpDir, "coverage.out")
				err := os.WriteFile(coveragePath, []byte("mode: set\ngithub.com/test/main.go:1.1,2.2 1 1\n"), 0644)
				require.NoError(t, err)
				return tmpDir
			},
			expectError: false,
		},
		{
			name: "multiple coverage files to merge",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				// Create multiple coverage files
				file1 := filepath.Join(tmpDir, "coverage1.out")
				err := os.WriteFile(file1, []byte("mode: set\ngithub.com/test/main.go:1.1,2.2 1 1\n"), 0644)
				require.NoError(t, err)

				file2 := filepath.Join(tmpDir, "coverage2.out")
				err = os.WriteFile(file2, []byte("mode: set\ngithub.com/test/other.go:3.1,4.2 1 0\n"), 0644)
				require.NoError(t, err)

				return tmpDir
			},
			expectError: false,
		},
		{
			name: "directory does not exist",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				return filepath.Join(tmpDir, "nonexistent")
			},
			expectError: true,
			errorMsg:    "coverage directory not found",
		},
		{
			name: "no .out files in directory",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				// Create a non-coverage file
				err := os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("test"), 0644)
				require.NoError(t, err)
				return tmpDir
			},
			expectError: true,
			errorMsg:    "no coverage files",
		},
		{
			name: "path is a file not a directory",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				filePath := filepath.Join(tmpDir, "test.out")
				err := os.WriteFile(filePath, []byte("mode: set\n"), 0644)
				require.NoError(t, err)
				return filePath
			},
			expectError: true,
			errorMsg:    "not a directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coverageDir := tt.setup(t)
			runner := NewRunner(Config{
				CoveragePath: coverageDir,
			})

			profiles, err := runner.readAndMergeCoverageFiles()

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, profiles)
			assert.Greater(t, len(profiles), 0)
		})
	}
}

func TestRunner_readAndMergeCoverageFiles_ErrorMessage(t *testing.T) {
	// Test that the error message includes helpful instructions
	tmpDir := t.TempDir()
	nonexistentDir := filepath.Join(tmpDir, "nonexistent")

	runner := NewRunner(Config{
		CoveragePath: nonexistentDir,
	})

	_, err := runner.readAndMergeCoverageFiles()
	require.Error(t, err)

	errMsg := err.Error()
	assert.Contains(t, errMsg, "coverage directory not found")
	assert.Contains(t, errMsg, "go test")
	assert.Contains(t, errMsg, "-coverprofile")
	assert.Contains(t, errMsg, nonexistentDir)
}

func TestNewRunner(t *testing.T) {
	config := Config{
		CoveragePath: "/path/to/coverage",
	}

	runner := NewRunner(config)

	assert.NotNil(t, runner)
	assert.Equal(t, config.CoveragePath, runner.config.CoveragePath)
	assert.NotNil(t, runner.diffSource) // Should have default LocalDiffSource
}

func TestRunner_Run_Integration(t *testing.T) {
	// Skip if we're not in a git repository. The `.git` directory can live
	// several levels above this package, so probe via `git rev-parse` instead
	// of a fixed relative path.
	if err := exec.Command("git", "rev-parse", "--is-inside-work-tree").Run(); err != nil {
		t.Skip("Not in a git repository, skipping integration test")
	}

	// Create a temporary coverage directory with files
	tmpDir := t.TempDir()
	coverageDir := filepath.Join(tmpDir, "coverage")
	err := os.Mkdir(coverageDir, 0755)
	require.NoError(t, err)

	// Write coverage files
	coverageContent := `mode: set
github.com/grafana/shared-workflows/actions/annotate-coverage/internal/local/runner.go:10.1,12.2 1 1
github.com/grafana/shared-workflows/actions/annotate-coverage/internal/local/runner.go:14.1,16.2 1 0
`
	coveragePath := filepath.Join(coverageDir, "coverage.out")
	err = os.WriteFile(coveragePath, []byte(coverageContent), 0644)
	require.NoError(t, err)

	runner := NewRunner(Config{
		CoveragePath: coverageDir,
		Format:       "Text",
	})

	ctx := context.Background()
	err = runner.Run(ctx)

	// The run should succeed (no error, or an expected error like "No changes detected")
	// We don't assert on specific output since it depends on the current git state
	if err != nil {
		// If there's an error, it should be one of the expected messages
		errMsg := err.Error()
		assert.True(t,
			strings.Contains(errMsg, "No changes detected") ||
				strings.Contains(errMsg, "No Go files changed") ||
				strings.Contains(errMsg, "failed to parse"),
			"Unexpected error: %v", err)
	}
}

func TestRunner_Run_MissingCoverageDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	coverageDir := filepath.Join(tmpDir, "nonexistent")

	runner := NewRunner(Config{
		CoveragePath: coverageDir,
	})

	ctx := context.Background()
	err := runner.Run(ctx)

	// If there are no changes in git diff, the runner returns nil without reading coverage
	// If there are changes, it should error on missing coverage directory
	if err == nil {
		// No changes in git diff, this is acceptable
		t.Log("No changes in git diff, coverage directory not checked")
	} else {
		// There were changes, so we expect the coverage directory error
		assert.Contains(t, err.Error(), "coverage directory not found")
		assert.Contains(t, err.Error(), "go test")
	}
}
