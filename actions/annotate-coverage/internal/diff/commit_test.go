package diff

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

func TestGitCommitDiffSource_GetDiff(t *testing.T) {
	// Create a temporary git repo with commits
	tmpDir := t.TempDir()

	// Initialize and configure
	runGit(t, tmpDir, "init")
	runGit(t, tmpDir, "config", "user.email", "test@test.com")
	runGit(t, tmpDir, "config", "user.name", "Test User")

	// First commit
	err := os.WriteFile(filepath.Join(tmpDir, "file.go"), []byte("package main\n"), 0644)
	require.NoError(t, err)
	runGit(t, tmpDir, "add", ".")
	runGit(t, tmpDir, "commit", "-m", "first commit")

	// Second commit with changes
	err = os.WriteFile(filepath.Join(tmpDir, "file.go"), []byte("package main\n\nfunc main() {}\n"), 0644)
	require.NoError(t, err)
	runGit(t, tmpDir, "add", ".")
	runGit(t, tmpDir, "commit", "-m", "second commit")

	// Get commit SHA
	out, err := exec.Command("git", "-C", tmpDir, "rev-parse", "HEAD").Output()
	require.NoError(t, err)
	commitSHA := strings.TrimSpace(string(out))

	// Get first commit SHA for root commit test
	out, err = exec.Command("git", "-C", tmpDir, "rev-list", "--max-parents=0", "HEAD").Output()
	require.NoError(t, err)
	rootCommitSHA := strings.TrimSpace(string(out))

	tests := []struct {
		name        string
		commitSHA   string
		expectError bool
		expectEmpty bool
	}{
		{
			name:        "valid commit returns diff",
			commitSHA:   commitSHA,
			expectError: false,
			expectEmpty: false,
		},
		{
			name:        "root commit returns diff with --root flag",
			commitSHA:   rootCommitSHA,
			expectError: false,
			expectEmpty: false,
		},
		{
			name:        "empty commit SHA returns error",
			commitSHA:   "",
			expectError: true,
		},
		{
			name:        "invalid commit SHA returns error",
			commitSHA:   "invalidcommit",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := NewGitCommitDiffSource(tt.commitSHA, tmpDir)
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
				// Verify it's valid unified diff format
				assert.Contains(t, string(output), "diff --git")
			}
		})
	}
}

func TestGitCommitDiffSource_InvalidWorkDir(t *testing.T) {
	source := NewGitCommitDiffSource("abc123", "/nonexistent/path")
	_, err := source.GetDiff(context.Background())
	assert.Error(t, err)
}

func TestGitCommitDiffSource_NotGitRepository(t *testing.T) {
	tmpDir := t.TempDir()
	// Don't initialize git

	source := NewGitCommitDiffSource("abc123", tmpDir)
	_, err := source.GetDiff(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "git")
}

func TestGitCommitDiffSource_ShortSHA(t *testing.T) {
	// Test that short SHA works
	tmpDir := t.TempDir()

	runGit(t, tmpDir, "init")
	runGit(t, tmpDir, "config", "user.email", "test@test.com")
	runGit(t, tmpDir, "config", "user.name", "Test User")

	err := os.WriteFile(filepath.Join(tmpDir, "file.go"), []byte("package main"), 0644)
	require.NoError(t, err)
	runGit(t, tmpDir, "add", ".")
	runGit(t, tmpDir, "commit", "-m", "commit")

	// Get short SHA
	out, err := exec.Command("git", "-C", tmpDir, "rev-parse", "--short", "HEAD").Output()
	require.NoError(t, err)
	shortSHA := strings.TrimSpace(string(out))

	source := NewGitCommitDiffSource(shortSHA, tmpDir)
	output, err := source.GetDiff(context.Background())

	require.NoError(t, err)
	assert.NotEmpty(t, output)
}
