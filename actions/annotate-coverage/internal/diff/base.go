package diff

import (
	"context"
	"fmt"
	"os/exec"
)

// GitBaseDiffSource implements DiffSource by running git diff to compare
// between two references (branches, tags, or commit SHAs).
type GitBaseDiffSource struct {
	// BaseRef is the base reference to compare against (e.g., SHA, branch name, tag).
	BaseRef string
	// CommitRef is the commit reference to compare to.
	// If empty, defaults to HEAD.
	CommitRef string
	// WorkDir is the directory to run git commands in.
	// If empty, uses the current working directory.
	WorkDir string
}

// NewGitBaseDiffSource creates a new GitBaseDiffSource for comparing against the specified base.
// If commitRef is empty, it defaults to HEAD.
func NewGitBaseDiffSource(baseRef, commitRef, workDir string) *GitBaseDiffSource {
	return &GitBaseDiffSource{
		BaseRef:   baseRef,
		CommitRef: commitRef,
		WorkDir:   workDir,
	}
}

// GetDiff executes `git diff <base>..<commit>` and returns the unified diff output.
// The two-dot syntax shows all changes between base and commit.
// If CommitRef is empty, defaults to HEAD.
func (s *GitBaseDiffSource) GetDiff(ctx context.Context) ([]byte, error) {
	if s.BaseRef == "" {
		return nil, fmt.Errorf("base ref is required")
	}

	// Default to HEAD if no commit specified
	commit := s.CommitRef
	if commit == "" {
		commit = "HEAD"
	}

	// git diff <base>..<commit> shows all changes between base and commit
	// This captures all commits between the two references
	cmd := exec.CommandContext(ctx, "git", "diff", s.BaseRef+".."+commit)
	if s.WorkDir != "" {
		cmd.Dir = s.WorkDir
	}

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("git diff failed: %s", string(exitErr.Stderr))
		}
		return nil, err
	}
	return output, nil
}
