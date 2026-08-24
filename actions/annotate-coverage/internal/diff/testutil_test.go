package diff

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
)

// runGit runs `git -C dir args...` and fails the test if it errors, surfacing
// combined output in the failure message. Used by tests that set up a temp
// git repo and don't want to repeat error-checking boilerplate.
func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	gitArgs := append([]string{"-C", dir}, args...)
	out, err := exec.Command("git", gitArgs...).CombinedOutput()
	require.NoErrorf(t, err, "git %v: %s", args, out)
}
