package diff

import (
	"context"
	"fmt"
	"os/exec"
)

// GitCommitDiffSource implements DiffSource by running git diff-tree to get
// the changes introduced by a specific commit.
type GitCommitDiffSource struct {
	// CommitSHA is the commit hash to get the diff for.
	CommitSHA string
	// WorkDir is the directory to run git commands in.
	// If empty, uses the current working directory.
	WorkDir string
}

// NewGitCommitDiffSource creates a new GitCommitDiffSource for the specified commit.
func NewGitCommitDiffSource(commitSHA string, workDir string) *GitCommitDiffSource {
	return &GitCommitDiffSource{
		CommitSHA: commitSHA,
		WorkDir:   workDir,
	}
}

// GetDiff executes `git diff-tree -p <commit>` and returns the unified diff output.
// The -p flag generates patch output (unified diff format).
// For root commits (commits with no parent), uses --root flag to show all changes.
func (s *GitCommitDiffSource) GetDiff(ctx context.Context) ([]byte, error) {
	if s.CommitSHA == "" {
		return nil, fmt.Errorf("commit SHA is required")
	}

	// git diff-tree -p --root <commit> shows the changes introduced by that commit
	// -p generates patch output (unified diff)
	// --root allows viewing root commits (first commit with no parent)
	cmd := exec.CommandContext(ctx, "git", "diff-tree", "-p", "--root", s.CommitSHA)
	if s.WorkDir != "" {
		cmd.Dir = s.WorkDir
	}

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("git diff-tree failed: %s", string(exitErr.Stderr))
		}
		return nil, err
	}
	return output, nil
}
