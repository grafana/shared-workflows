package diff

import (
	"context"
)

// DiffSource defines the interface for obtaining unified diff data.
// Implementations include local git diff, commit-based diff, and GitHub API.
type DiffSource interface {
	// GetDiff returns the unified diff data as bytes.
	// The output should be in unified diff format compatible with coverage.ParseDiff().
	// Returns an empty byte slice if there are no changes.
	// Returns an error if the diff operation fails.
	GetDiff(ctx context.Context) ([]byte, error)
}
