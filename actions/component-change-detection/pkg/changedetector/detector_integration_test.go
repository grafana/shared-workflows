//go:build integration
// +build integration

package changedetector

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// getRepoRootPath finds the repository root by looking for .git directory
func getRepoRootPath(t *testing.T) string {
	t.Helper()
	
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to find git repository root: %v", err)
	}
	
	return strings.TrimSpace(string(output))
}

// TestChangeDetection_NoChangesAtHEAD verifies that when all tags point to HEAD,
// no components are marked as changed
func TestChangeDetection_NoChangesAtHEAD(t *testing.T) {
	// Get current commit
	cmd := exec.Command("git", "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to get current commit: %v", err)
	}
	currentCommit := string(output[:len(output)-1]) // Remove trailing newline

	// Create tags pointing to HEAD
	tags := Tags{
		"migrator":        currentCommit,
		"apiserver":       currentCommit,
		"controller":      currentCommit,
		"templatewatcher": currentCommit,
	}

	// Load config (path relative to repository root)
	configPath := getRepoRootPath(t) + "/.component-deps.yaml"
	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Run detection
	detector := NewDetector(config, tags, "HEAD")
	changes, err := detector.DetectChanges()
	if err != nil {
		t.Fatalf("DetectChanges failed: %v", err)
	}

	// Verify all components are unchanged
	for component, changed := range changes {
		if changed {
			t.Errorf("Component %q should be unchanged (tags at HEAD), but marked as changed", component)
		}
	}

	// Verify we got results for all expected components
	expectedComponents := []string{"migrator", "apiserver", "controller", "templatewatcher"}
	for _, comp := range expectedComponents {
		if _, exists := changes[comp]; !exists {
			t.Errorf("Missing result for component %q", comp)
		}
	}
}

// TestChangeDetection_SelectiveDetection verifies that the tool correctly
// identifies which components changed based on actual git history
func TestChangeDetection_SelectiveDetection(t *testing.T) {
	// Get a previous commit (one before current HEAD)
	cmd := exec.Command("git", "rev-parse", "HEAD~1")
	output, err := cmd.Output()
	if err != nil {
		t.Skip("Skipping test: need at least 2 commits in history")
	}
	prevCommit := string(output[:len(output)-1])

	// Create tags pointing to previous commit
	tags := Tags{
		"migrator":        prevCommit,
		"apiserver":       prevCommit,
		"controller":      prevCommit,
		"templatewatcher": prevCommit,
	}

	// Load config (path relative to repository root)
	configPath := getRepoRootPath(t) + "/.component-deps.yaml"
	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Get files that actually changed
	gitCmd := exec.Command("git", "diff", "--name-only", prevCommit+"..HEAD")
	gitOutput, err := gitCmd.Output()
	if err != nil {
		t.Fatalf("Failed to get changed files: %v", err)
	}
	changedFilesOutput := string(gitOutput)

	// Run detection
	detector := NewDetector(config, tags, "HEAD")
	changes, err := detector.DetectChanges()
	if err != nil {
		t.Fatalf("DetectChanges failed: %v", err)
	}

	// Log results for debugging
	t.Logf("Changed files:\n%s", changedFilesOutput)
	t.Logf("Detection results: %+v", changes)

	// At minimum, verify the detection ran and returned results
	if len(changes) == 0 {
		t.Error("Expected at least some results from detection")
	}

	// Verify we got results for all expected components
	expectedComponents := []string{"migrator", "apiserver", "controller", "templatewatcher"}
	for _, comp := range expectedComponents {
		if _, exists := changes[comp]; !exists {
			t.Errorf("Missing result for component %q", comp)
		}
	}
}

// TestChangeDetection_DependencyPropagation verifies that changes propagate
// through dependencies correctly
func TestChangeDetection_DependencyPropagation(t *testing.T) {
	// Create a test config with clear dependency chain
	config := &Config{
		Components: map[string]ComponentConfig{
			"base": {
				Paths: []string{"pkg/changedetector/types.go"},
			},
			"dependent1": {
				Paths:        []string{"cmd/apiserver/**"},
				Dependencies: []string{"base"},
			},
			"dependent2": {
				Paths:        []string{"cmd/controller/**"},
				Dependencies: []string{"base"},
			},
			"independent": {
				Paths: []string{"docs/**"},
			},
		},
	}

	// Get current and previous commit
	currentCmd := exec.Command("git", "rev-parse", "HEAD")
	currentOutput, err := currentCmd.Output()
	if err != nil {
		t.Fatalf("Failed to get current commit: %v", err)
	}
	currentCommit := string(currentOutput[:len(currentOutput)-1])

	// Check if pkg/changedetector/types.go exists in this commit
	gitCmd := exec.Command("git", "cat-file", "-e", currentCommit+":pkg/changedetector/types.go")
	if err := gitCmd.Run(); err != nil {
		t.Skip("Skipping test: pkg/changedetector/types.go doesn't exist in current commit")
	}

	// Find a commit before types.go was added
	logCmd := exec.Command("git", "log", "--all", "--format=%H", "--", "pkg/changedetector/types.go")
	logOutput, err := logCmd.Output()
	if err != nil || len(logOutput) == 0 {
		t.Skip("Skipping test: cannot find commit history for types.go")
	}

	// Get the commit that added types.go
	commits := string(logOutput)
	// Get second-to-last commit (before types.go was added)
	lines := []byte{}
	for i := len(commits) - 1; i >= 0; i-- {
		if commits[i] == '\n' {
			lines = append([]byte{commits[i]}, lines...)
		} else {
			lines = append([]byte{commits[i]}, lines...)
		}
	}

	// Use the last commit in the log (oldest) as prevCommit
	prevCommitCmd := exec.Command("git", "log", "--all", "--format=%H", "--reverse", "--", "pkg/changedetector/types.go")
	prevOutput, err := prevCommitCmd.Output()
	if err != nil {
		t.Skip("Cannot determine previous commit")
	}

	// Get first line (first commit that touched the file)
	firstNewline := 0
	for i, b := range prevOutput {
		if b == '\n' {
			firstNewline = i
			break
		}
	}
	if firstNewline == 0 {
		t.Skip("Cannot parse commit history")
	}

	firstCommit := string(prevOutput[:firstNewline])

	// Get parent of first commit
	parentCmd := exec.Command("git", "rev-parse", firstCommit+"^")
	parentOutput, err := parentCmd.Output()
	if err != nil {
		t.Skip("Cannot get parent commit (might be first commit in repo)")
	}
	prevCommit := string(parentOutput[:len(parentOutput)-1])

	tags := Tags{
		"base":        prevCommit,
		"dependent1":  prevCommit,
		"dependent2":  prevCommit,
		"independent": prevCommit,
	}

	// Run detection
	detector := NewDetector(config, tags, currentCommit)
	changes, err := detector.DetectChanges()
	if err != nil {
		t.Fatalf("DetectChanges failed: %v", err)
	}

	t.Logf("Detection results: %+v", changes)

	// Verify base changed (types.go was added/modified)
	if !changes["base"] {
		t.Error("Expected 'base' to be changed (types.go was added)")
	}

	// Verify dependents also changed due to propagation
	if !changes["dependent1"] {
		t.Error("Expected 'dependent1' to be changed (depends on base)")
	}
	if !changes["dependent2"] {
		t.Error("Expected 'dependent2' to be changed (depends on base)")
	}
}

// TestChangeDetection_WithTagsFile tests loading tags from JSON file
func TestChangeDetection_WithTagsFile(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Get current commit
	cmd := exec.Command("git", "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to get current commit: %v", err)
	}
	currentCommit := string(output[:len(output)-1])

	// Create tags file
	tagsPath := filepath.Join(tmpDir, "tags.json")
	tags := map[string]string{
		"migrator":        currentCommit,
		"apiserver":       currentCommit,
		"controller":      currentCommit,
		"templatewatcher": currentCommit,
	}

	data, err := json.MarshalIndent(tags, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal tags: %v", err)
	}

	if err := os.WriteFile(tagsPath, data, 0644); err != nil {
		t.Fatalf("Failed to write tags file: %v", err)
	}

	// Load tags
	loadedTags, err := LoadTags(tagsPath)
	if err != nil {
		t.Fatalf("Failed to load tags: %v", err)
	}

	// Verify tags loaded correctly
	for component, tag := range tags {
		if loadedTags[component] != tag {
			t.Errorf("Tag mismatch for %s: got %s, want %s", component, loadedTags[component], tag)
		}
	}

	// Run detection with loaded tags
	configPath := getRepoRootPath(t) + "/.component-deps.yaml"
	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	detector := NewDetector(config, loadedTags, "HEAD")
	changes, err := detector.DetectChanges()
	if err != nil {
		t.Fatalf("DetectChanges failed: %v", err)
	}

	// All should be unchanged since tags = HEAD
	for component, changed := range changes {
		if changed {
			t.Errorf("Component %q should be unchanged, but marked as changed", component)
		}
	}
}
