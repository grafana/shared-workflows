package changedetector

import (
	"fmt"

	"github.com/bmatcuk/doublestar/v4"
)

// MatchPaths checks if a file matches any of the include patterns
// and doesn't match any of the exclude patterns.
// Excludes can include both component-specific and global exclusions.
func MatchPaths(file string, includes, excludes []string) (bool, error) {
	// Check excludes first (more efficient to reject early)
	for _, pattern := range excludes {
		matched, err := doublestar.Match(pattern, file)
		if err != nil {
			return false, fmt.Errorf("invalid exclude pattern %q: %w", pattern, err)
		}
		if matched {
			return false, nil
		}
	}

	// Check if file matches any include pattern
	for _, pattern := range includes {
		matched, err := doublestar.Match(pattern, file)
		if err != nil {
			return false, fmt.Errorf("invalid include pattern %q: %w", pattern, err)
		}
		if matched {
			return true, nil
		}
	}

	return false, nil
}
