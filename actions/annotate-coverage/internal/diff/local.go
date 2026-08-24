package diff

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// LocalDiffSource implements DiffSource by running git diff against the working tree.
// It marks untracked files as intent-to-add so they appear in the diff output.
type LocalDiffSource struct {
	// WorkDir is the directory to run git commands in.
	// If empty, uses the current working directory.
	WorkDir string
}

// NewLocalDiffSource creates a new LocalDiffSource.
func NewLocalDiffSource(workDir string) *LocalDiffSource {
	return &LocalDiffSource{
		WorkDir: workDir,
	}
}

// GetDiff executes `git add -N . && git diff` and returns the output.
// It first runs `git add -N .` to mark new untracked files as intent-to-add,
// which allows them to appear in the diff output.
func (s *LocalDiffSource) GetDiff(ctx context.Context) ([]byte, error) {
	// Add untracked files as intent-to-add so they show up in diff.
	// Wire stderr to a buffer so failure messages survive in the returned error
	// (exec.Cmd.Run does not populate ExitError.Stderr on its own).
	var addStderr bytes.Buffer
	addCmd := exec.CommandContext(ctx, "git", "add", "-N", ".")
	if s.WorkDir != "" {
		addCmd.Dir = s.WorkDir
	}
	addCmd.Stderr = &addStderr
	if err := addCmd.Run(); err != nil {
		return nil, fmt.Errorf("git add -N failed: %w: %s", err, strings.TrimSpace(addStderr.String()))
	}

	// Now run git diff to get all changes including new files. cmd.Output()
	// captures stdout and surfaces stderr via ExitError.Stderr automatically.
	diffCmd := exec.CommandContext(ctx, "git", "diff")
	if s.WorkDir != "" {
		diffCmd.Dir = s.WorkDir
	}
	output, err := diffCmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("git diff failed: %w: %s", err, strings.TrimSpace(string(exitErr.Stderr)))
		}
		return nil, err
	}
	return output, nil
}
