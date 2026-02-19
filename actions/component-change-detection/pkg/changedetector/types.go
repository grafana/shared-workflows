package changedetector

// ComponentConfig defines the configuration for a single component
type ComponentConfig struct {
	// Paths to include when detecting changes (supports glob patterns)
	Paths []string `yaml:"paths"`
	// Patterns to exclude from change detection
	Excludes []string `yaml:"excludes"`
	// Components that this component depends on
	Dependencies []string `yaml:"dependencies"`
}

// Config represents the full change detection configuration
type Config struct {
	// Global exclusions applied to all components
	GlobalExcludes []string `yaml:"global_excludes,omitempty"`
	// Component-specific configurations
	Components map[string]ComponentConfig `yaml:"components"`
}

// Tags represents the current tags for all components
type Tags map[string]string

// FileCommitPair represents a file and the commit that changed it
type FileCommitPair struct {
	File   string `json:"file"`
	Commit string `json:"commit"`
}

// DetailedChanges maps component names to the files and commits that triggered changes
type DetailedChanges map[string][]FileCommitPair

// ChangeReason describes why a component changed (or didn't change)
type ChangeReason struct {
	// Changed indicates whether the component has changes
	Changed bool
	// Reason provides a human-readable explanation
	Reason string
	// TriggeringCommits lists the commits that caused changes (if any)
	TriggeringCommits []string
	// CommitToFiles maps each triggering commit to the files that matched
	CommitToFiles map[string][]string
	// FromRef is the starting ref used for comparison
	FromRef string
	// ToRef is the target ref used for comparison
	ToRef string
}
