package diff

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalDiffSource_GetDiff(t *testing.T) {
	// This test requires a real git repository
	// Create a temporary git repo for testing
	tmpDir := t.TempDir()

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	require.NoError(t, cmd.Run())

	// Configure git (needed for commits)
	exec.Command("git", "-C", tmpDir, "config", "user.email", "test@test.com").Run()
	exec.Command("git", "-C", tmpDir, "config", "user.name", "Test User").Run()

	// Create initial commit
	err := os.WriteFile(filepath.Join(tmpDir, "initial.txt"), []byte("initial"), 0644)
	require.NoError(t, err)
	exec.Command("git", "-C", tmpDir, "add", ".").Run()
	exec.Command("git", "-C", tmpDir, "commit", "-m", "initial").Run()

	tests := []struct {
		name        string
		setup       func(t *testing.T) // Setup changes in tmpDir
		expectEmpty bool
		expectError bool
	}{
		{
			name: "no changes returns empty diff",
			setup: func(t *testing.T) {
				// No changes
			},
			expectEmpty: true,
		},
		{
			name: "modified file returns diff",
			setup: func(t *testing.T) {
				err := os.WriteFile(filepath.Join(tmpDir, "initial.txt"), []byte("modified"), 0644)
				require.NoError(t, err)
			},
			expectEmpty: false,
		},
		{
			name: "new file returns diff",
			setup: func(t *testing.T) {
				err := os.WriteFile(filepath.Join(tmpDir, "new.go"), []byte("package main"), 0644)
				require.NoError(t, err)
			},
			expectEmpty: false,
		},
		{
			name: "deleted file returns diff",
			setup: func(t *testing.T) {
				err := os.Remove(filepath.Join(tmpDir, "initial.txt"))
				require.NoError(t, err)
			},
			expectEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset repo state
			exec.Command("git", "-C", tmpDir, "checkout", ".").Run()
			exec.Command("git", "-C", tmpDir, "clean", "-fd").Run()

			tt.setup(t)

			source := NewLocalDiffSource(tmpDir)
			output, err := source.GetDiff(context.Background())

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			if tt.expectEmpty {
				assert.Empty(t, output)
			} else {
				assert.NotEmpty(t, output)
			}
		})
	}
}

func TestLocalDiffSource_InvalidWorkDir(t *testing.T) {
	source := NewLocalDiffSource("/nonexistent/path")
	_, err := source.GetDiff(context.Background())
	assert.Error(t, err)
}

func TestLocalDiffSource_NotGitRepository(t *testing.T) {
	tmpDir := t.TempDir()
	// Don't initialize git

	source := NewLocalDiffSource(tmpDir)
	_, err := source.GetDiff(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "git")
}
