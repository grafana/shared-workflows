package changedetector

import (
	"fmt"
	"os/exec"
	"strings"
)

// GitOperations defines the interface for git operations needed by the detector
type GitOperations interface {
	// GetChangedFiles returns files changed between two git refs
	GetChangedFiles(from, to string) ([]string, error)
	// RefExists checks if a git ref exists
	RefExists(ref string) (bool, error)
	// GetCommitsBetween returns the list of commits between two refs (from..to)
	GetCommitsBetween(from, to string) ([]string, error)
	// GetFilesChangedInCommit returns the files changed in a specific commit
	GetFilesChangedInCommit(commit string) ([]string, error)
}

// GitOps implements GitOperations using the git CLI
type GitOps struct{}

// NewGitOps creates a new GitOps instance
func NewGitOps() *GitOps {
	return &GitOps{}
}

// GetChangedFiles returns files changed between two git refs
func (g *GitOps) GetChangedFiles(from, to string) ([]string, error) {
	// Use -z for null-terminated output to handle filenames with spaces/newlines
	cmd := exec.Command("git", "diff", "--name-only", "-z", from, to)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get changed files: %w", err)
	}

	// Split on null byte and filter empty strings
	files := strings.Split(string(output), "\x00")
	result := make([]string, 0, len(files))
	for _, file := range files {
		if file != "" {
			result = append(result, file)
		}
	}

	return result, nil
}

// RefExists checks if a git ref (commit, tag, branch) exists
func (g *GitOps) RefExists(ref string) (bool, error) {
	cmd := exec.Command("git", "rev-parse", "--verify", ref)
	err := cmd.Run()
	if err != nil {
		// rev-parse returns non-zero if ref doesn't exist
		if exitErr, ok := err.(*exec.ExitError); ok {
			// Exit code 128 or 1 means ref doesn't exist
			if exitErr.ExitCode() == 128 || exitErr.ExitCode() == 1 {
				return false, nil
			}
		}
		return false, fmt.Errorf("failed to check if ref exists: %w", err)
	}
	return true, nil
}

// GetCommitsBetween returns the list of commits between two refs (from..to)
// Returns commits in chronological order (oldest first)
func (g *GitOps) GetCommitsBetween(from, to string) ([]string, error) {
	// Use git rev-list to get commits between from and to
	// The syntax "from..to" means "commits reachable from 'to' but not from 'from'"
	// --reverse returns commits in chronological order (oldest first)
	cmd := exec.Command("git", "rev-list", "--reverse", fmt.Sprintf("%s..%s", from, to))
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get commits between %s and %s: %w", from, to, err)
	}

	commits := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(commits) == 1 && commits[0] == "" {
		return []string{}, nil
	}

	return commits, nil
}

// GetFilesChangedInCommit returns the files changed in a specific commit
func (g *GitOps) GetFilesChangedInCommit(commit string) ([]string, error) {
	// Use git diff-tree to get files changed in this commit
	// --no-commit-id: suppress commit ID output
	// --name-only: show only file names
	// -z: null-terminated output for safe filename handling
	// -r: recurse into sub-trees
	cmd := exec.Command("git", "diff-tree", "--no-commit-id", "--name-only", "-z", "-r", commit)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get files changed in commit %s: %w", commit, err)
	}

	// Split on null byte and filter empty strings
	files := strings.Split(string(output), "\x00")
	result := make([]string, 0, len(files))
	for _, file := range files {
		if file != "" {
			result = append(result, file)
		}
	}

	return result, nil
}

// MockGitOps is a mock implementation of GitOperations for testing
type MockGitOps struct {
	ChangedFiles               map[string]map[string][]string // from -> to -> files
	ExistingRefs               map[string]bool
	CommitsBetween             map[string]map[string][]string // from -> to -> commits
	FilesInCommit              map[string][]string            // commit -> files
	GetChangedFilesErr         error
	RefExistsErr               error
	GetCommitsBetweenErr       error
	GetFilesChangedInCommitErr error
}

func (m *MockGitOps) GetChangedFiles(from, to string) ([]string, error) {
	if m.GetChangedFilesErr != nil {
		return nil, m.GetChangedFilesErr
	}
	if m.ChangedFiles == nil {
		return []string{}, nil
	}
	if files, ok := m.ChangedFiles[from][to]; ok {
		return files, nil
	}
	return []string{}, nil
}

func (m *MockGitOps) RefExists(ref string) (bool, error) {
	if m.RefExistsErr != nil {
		return false, m.RefExistsErr
	}
	if m.ExistingRefs == nil {
		return true, nil
	}
	return m.ExistingRefs[ref], nil
}

func (m *MockGitOps) GetCommitsBetween(from, to string) ([]string, error) {
	if m.GetCommitsBetweenErr != nil {
		return nil, m.GetCommitsBetweenErr
	}
	if m.CommitsBetween == nil {
		return []string{}, nil
	}
	if commits, ok := m.CommitsBetween[from][to]; ok {
		return commits, nil
	}
	return []string{}, nil
}

func (m *MockGitOps) GetFilesChangedInCommit(commit string) ([]string, error) {
	if m.GetFilesChangedInCommitErr != nil {
		return nil, m.GetFilesChangedInCommitErr
	}
	if m.FilesInCommit == nil {
		return []string{}, nil
	}
	if files, ok := m.FilesInCommit[commit]; ok {
		return files, nil
	}
	return []string{}, nil
}
