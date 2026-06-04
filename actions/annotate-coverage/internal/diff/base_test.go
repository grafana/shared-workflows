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

func TestGitBaseDiffSource_GetDiff(t *testing.T) {
	// Create a temporary git repo with multiple commits on different branches
	tmpDir := t.TempDir()

	// Initialize and configure
	exec.Command("git", "-C", tmpDir, "init").Run()
	exec.Command("git", "-C", tmpDir, "config", "user.email", "test@test.com").Run()
	exec.Command("git", "-C", tmpDir, "config", "user.name", "Test User").Run()

	// Create initial commit on main branch
	err := os.WriteFile(filepath.Join(tmpDir, "file.go"), []byte("package main\n"), 0644)
	require.NoError(t, err)
	exec.Command("git", "-C", tmpDir, "add", ".").Run()
	exec.Command("git", "-C", tmpDir, "commit", "-m", "initial commit").Run()

	// Get the initial commit SHA (will be our base)
	out, err := exec.Command("git", "-C", tmpDir, "rev-parse", "HEAD").Output()
	require.NoError(t, err)
	baseSHA := strings.TrimSpace(string(out))

	// Create a feature branch and add commits
	exec.Command("git", "-C", tmpDir, "checkout", "-b", "feature").Run()

	// First commit on feature branch
	err = os.WriteFile(filepath.Join(tmpDir, "file.go"), []byte("package main\n\nfunc foo() {}\n"), 0644)
	require.NoError(t, err)
	exec.Command("git", "-C", tmpDir, "add", ".").Run()
	exec.Command("git", "-C", tmpDir, "commit", "-m", "add foo").Run()

	// Second commit on feature branch
	err = os.WriteFile(filepath.Join(tmpDir, "file.go"), []byte("package main\n\nfunc foo() {}\n\nfunc bar() {}\n"), 0644)
	require.NoError(t, err)
	exec.Command("git", "-C", tmpDir, "add", ".").Run()
	exec.Command("git", "-C", tmpDir, "commit", "-m", "add bar").Run()

	tests := []struct {
		name        string
		baseRef     string
		commitRef   string
		expectError bool
		expectEmpty bool
	}{
		{
			name:        "valid base SHA returns cumulative diff to HEAD",
			baseRef:     baseSHA,
			commitRef:   "",
			expectError: false,
			expectEmpty: false,
		},
		{
			name:        "empty base ref returns error",
			baseRef:     "",
			commitRef:   "",
			expectError: true,
		},
		{
			name:        "invalid base ref returns error",
			baseRef:     "nonexistent",
			commitRef:   "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := NewGitBaseDiffSource(tt.baseRef, tt.commitRef, tmpDir)
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
				outputStr := string(output)
				assert.Contains(t, outputStr, "diff --git")
				// Verify we get both functions (cumulative diff, not just last commit)
				assert.Contains(t, outputStr, "+func foo()")
				assert.Contains(t, outputStr, "+func bar()")
			}
		})
	}
}

func TestGitBaseDiffSource_InvalidWorkDir(t *testing.T) {
	source := NewGitBaseDiffSource("main", "", "/nonexistent/path")
	_, err := source.GetDiff(context.Background())
	assert.Error(t, err)
}

func TestGitBaseDiffSource_NotGitRepository(t *testing.T) {
	tmpDir := t.TempDir()
	// Don't initialize git

	source := NewGitBaseDiffSource("main", "", tmpDir)
	_, err := source.GetDiff(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "git")
}

func TestGitBaseDiffSource_ShortSHA(t *testing.T) {
	// Test that short SHA works as base ref
	tmpDir := t.TempDir()

	exec.Command("git", "-C", tmpDir, "init").Run()
	exec.Command("git", "-C", tmpDir, "config", "user.email", "test@test.com").Run()
	exec.Command("git", "-C", tmpDir, "config", "user.name", "Test User").Run()

	// Initial commit
	err := os.WriteFile(filepath.Join(tmpDir, "file.go"), []byte("package main"), 0644)
	require.NoError(t, err)
	exec.Command("git", "-C", tmpDir, "add", ".").Run()
	exec.Command("git", "-C", tmpDir, "commit", "-m", "initial").Run()

	// Get short SHA
	out, err := exec.Command("git", "-C", tmpDir, "rev-parse", "--short", "HEAD").Output()
	require.NoError(t, err)
	shortSHA := strings.TrimSpace(string(out))

	// Add another commit
	err = os.WriteFile(filepath.Join(tmpDir, "file.go"), []byte("package main\n\nfunc main() {}"), 0644)
	require.NoError(t, err)
	exec.Command("git", "-C", tmpDir, "add", ".").Run()
	exec.Command("git", "-C", tmpDir, "commit", "-m", "add main").Run()

	source := NewGitBaseDiffSource(shortSHA, "", tmpDir)
	output, err := source.GetDiff(context.Background())

	require.NoError(t, err)
	assert.NotEmpty(t, output)
	assert.Contains(t, string(output), "+func main()")
}

func TestGitBaseDiffSource_WithExplicitCommit(t *testing.T) {
	// Test that both base and commit can be specified
	tmpDir := t.TempDir()

	// Initialize and configure
	exec.Command("git", "-C", tmpDir, "init").Run()
	exec.Command("git", "-C", tmpDir, "config", "user.email", "test@test.com").Run()
	exec.Command("git", "-C", tmpDir, "config", "user.name", "Test User").Run()

	// Create initial commit
	err := os.WriteFile(filepath.Join(tmpDir, "file.go"), []byte("package main\n"), 0644)
	require.NoError(t, err)
	exec.Command("git", "-C", tmpDir, "add", ".").Run()
	exec.Command("git", "-C", tmpDir, "commit", "-m", "initial commit").Run()

	// Get the initial commit SHA (base)
	out, err := exec.Command("git", "-C", tmpDir, "rev-parse", "HEAD").Output()
	require.NoError(t, err)
	baseSHA := strings.TrimSpace(string(out))

	// Add second commit
	err = os.WriteFile(filepath.Join(tmpDir, "file.go"), []byte("package main\n\nfunc foo() {}\n"), 0644)
	require.NoError(t, err)
	exec.Command("git", "-C", tmpDir, "add", ".").Run()
	exec.Command("git", "-C", tmpDir, "commit", "-m", "add foo").Run()

	// Get the second commit SHA (commit)
	out, err = exec.Command("git", "-C", tmpDir, "rev-parse", "HEAD").Output()
	require.NoError(t, err)
	secondSHA := strings.TrimSpace(string(out))

	// Add third commit
	err = os.WriteFile(filepath.Join(tmpDir, "file.go"), []byte("package main\n\nfunc foo() {}\n\nfunc bar() {}\n"), 0644)
	require.NoError(t, err)
	exec.Command("git", "-C", tmpDir, "add", ".").Run()
	exec.Command("git", "-C", tmpDir, "commit", "-m", "add bar").Run()

	// Test diff between base and second commit (should only show foo, not bar)
	source := NewGitBaseDiffSource(baseSHA, secondSHA, tmpDir)
	output, err := source.GetDiff(context.Background())

	require.NoError(t, err)
	assert.NotEmpty(t, output)
	outputStr := string(output)
	assert.Contains(t, outputStr, "+func foo()")
	assert.NotContains(t, outputStr, "+func bar()") // bar is in third commit, not second
}

func TestGitBaseDiffSource_WithExplicitCommit_BranchNames(t *testing.T) {
	// Test that branch names work for both base and commit
	tmpDir := t.TempDir()

	// Initialize and configure
	exec.Command("git", "-C", tmpDir, "init").Run()
	exec.Command("git", "-C", tmpDir, "config", "user.email", "test@test.com").Run()
	exec.Command("git", "-C", tmpDir, "config", "user.name", "Test User").Run()

	// Create initial commit
	err := os.WriteFile(filepath.Join(tmpDir, "file.go"), []byte("package main\n"), 0644)
	require.NoError(t, err)
	exec.Command("git", "-C", tmpDir, "add", ".").Run()
	exec.Command("git", "-C", tmpDir, "commit", "-m", "initial commit").Run()

	// Get the current branch name (could be main, master, etc.)
	out, err := exec.Command("git", "-C", tmpDir, "rev-parse", "--abbrev-ref", "HEAD").Output()
	require.NoError(t, err)
	baseBranch := strings.TrimSpace(string(out))

	// Create feature branch and add commit
	exec.Command("git", "-C", tmpDir, "checkout", "-b", "feature").Run()
	err = os.WriteFile(filepath.Join(tmpDir, "file.go"), []byte("package main\n\nfunc foo() {}\n"), 0644)
	require.NoError(t, err)
	exec.Command("git", "-C", tmpDir, "add", ".").Run()
	exec.Command("git", "-C", tmpDir, "commit", "-m", "add foo").Run()

	// Test diff between base and feature branches
	source := NewGitBaseDiffSource(baseBranch, "feature", tmpDir)
	output, err := source.GetDiff(context.Background())

	require.NoError(t, err)
	assert.NotEmpty(t, output)
	assert.Contains(t, string(output), "+func foo()")
}
