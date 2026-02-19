package changedetector

import (
	"encoding/json"
	"fmt"
	"os"
)

// Detector performs change detection for components
type Detector struct {
	config        *Config
	git           GitOperations
	tags          Tags
	target        string
	changeReasons map[string]*ChangeReason
}

// NewDetector creates a new Detector instance
func NewDetector(config *Config, tags Tags, target string) *Detector {
	return &Detector{
		config: config,
		git:    NewGitOps(),
		tags:   tags,
		target: target,
	}
}

// DetectChanges performs change detection and returns a map of component -> changed
func (d *Detector) DetectChanges() (map[string]bool, error) {
	// Build dependency graph
	graph, err := BuildDependencyGraph(d.config)
	if err != nil {
		return nil, fmt.Errorf("error building dependency graph: %w", err)
	}

	// Detect direct changes for each component with detailed reasons
	directChanges := make(map[string]bool)
	d.changeReasons = make(map[string]*ChangeReason)

	for name, comp := range d.config.Components {
		reason, err := d.detectComponentChange(name, comp)
		if err != nil {
			return nil, fmt.Errorf("error detecting changes for %q: %w", name, err)
		}
		directChanges[name] = reason.Changed
		d.changeReasons[name] = reason
	}

	// Propagate changes through dependencies
	allChanges := PropagateChanges(graph, directChanges)

	return allChanges, nil
}

// GetChangeReasons returns the detailed change reasons for all components
func (d *Detector) GetChangeReasons() map[string]*ChangeReason {
	return d.changeReasons
}

// GetDetailedChanges returns a structured map of components to their triggering files and commits
func (d *Detector) GetDetailedChanges() DetailedChanges {
	result := make(DetailedChanges)

	for componentName, reason := range d.changeReasons {
		if !reason.Changed || len(reason.TriggeringCommits) == 0 {
			// Component didn't change or no commits to report
			result[componentName] = []FileCommitPair{}
			continue
		}

		pairs := []FileCommitPair{}
		for _, commit := range reason.TriggeringCommits {
			files, hasFiles := reason.CommitToFiles[commit]
			if hasFiles {
				for _, file := range files {
					pairs = append(pairs, FileCommitPair{
						File:   file,
						Commit: commit,
					})
				}
			}
		}
		result[componentName] = pairs
	}

	return result
}

// analyzeCommitsForComponent checks each commit to find which ones trigger changes for a component
func (d *Detector) analyzeCommitsForComponent(
	comp ComponentConfig,
	commits []string,
) ([]string, map[string][]string, error) {
	// Combine global and component-specific excludes
	allExcludes := make([]string, 0, len(d.config.GlobalExcludes)+len(comp.Excludes))
	allExcludes = append(allExcludes, d.config.GlobalExcludes...)
	allExcludes = append(allExcludes, comp.Excludes...)

	triggeringCommits := []string{}
	commitToFiles := make(map[string][]string) // Maps commit -> matching files

	for _, commit := range commits {
		changedFiles, err := d.git.GetFilesChangedInCommit(commit)
		if err != nil {
			// If we can't get files for a commit, treat it as potentially triggering
			triggeringCommits = append(triggeringCommits, commit)
			commitToFiles[commit] = []string{"<error fetching files>"}
			continue
		}

		// Check if any changed file in this commit matches component's paths
		matchingFiles := []string{}
		for _, file := range changedFiles {
			matched, err := MatchPaths(file, comp.Paths, allExcludes)
			if err != nil {
				return nil, nil, fmt.Errorf("error matching paths: %w", err)
			}
			if matched {
				matchingFiles = append(matchingFiles, file)
			}
		}

		if len(matchingFiles) > 0 {
			triggeringCommits = append(triggeringCommits, commit)
			commitToFiles[commit] = matchingFiles
		}
	}

	return triggeringCommits, commitToFiles, nil
}

// detectComponentChange checks if a single component has changed and returns detailed reason
func (d *Detector) detectComponentChange(name string, comp ComponentConfig) (*ChangeReason, error) {
	tag, hasTag := d.tags[name]

	// If no tag exists, component needs to be built
	if !hasTag || tag == "" || tag == "none" {
		return &ChangeReason{
			Changed:           true,
			Reason:            "no_previous_deployment",
			TriggeringCommits: []string{},
			CommitToFiles:     make(map[string][]string),
			FromRef:           tag,
			ToRef:             d.target,
		}, nil
	}

	// Verify tag exists in git
	exists, err := d.git.RefExists(tag)
	if err != nil {
		return nil, fmt.Errorf("error checking if tag exists: %w", err)
	}
	if !exists {
		return &ChangeReason{
			Changed:           true,
			Reason:            "previous_commit_not_found",
			TriggeringCommits: []string{},
			CommitToFiles:     make(map[string][]string),
			FromRef:           tag,
			ToRef:             d.target,
		}, nil
	}

	// Get list of commits between tag and target
	commits, err := d.git.GetCommitsBetween(tag, d.target)
	if err != nil {
		// Git operations failed (e.g., commit not available in shallow clone)
		// Treat as changed to trigger rebuild rather than failing
		return &ChangeReason{
			Changed:           true,
			Reason:            "git_operation_failed",
			TriggeringCommits: []string{},
			CommitToFiles:     make(map[string][]string),
			FromRef:           tag,
			ToRef:             d.target,
		}, nil
	}

	// If no commits between refs, nothing changed
	if len(commits) == 0 {
		return &ChangeReason{
			Changed:           false,
			Reason:            "no_commits_between_refs",
			TriggeringCommits: []string{},
			CommitToFiles:     make(map[string][]string),
			FromRef:           tag,
			ToRef:             d.target,
		}, nil
	}

	// Analyse commits to find which ones trigger changes
	triggeringCommits, commitToFiles, err := d.analyzeCommitsForComponent(comp, commits)
	if err != nil {
		return nil, err
	}

	if len(triggeringCommits) > 0 {
		reason := &ChangeReason{
			Changed:           true,
			Reason:            "matching_files_changed",
			TriggeringCommits: triggeringCommits,
			CommitToFiles:     commitToFiles,
			FromRef:           tag,
			ToRef:             d.target,
		}

		// Log one example file+commit immediately when change detected
		d.logChangeExample(name, reason)

		return reason, nil
	}

	return &ChangeReason{
		Changed:           false,
		Reason:            "no_matching_files",
		TriggeringCommits: []string{},
		CommitToFiles:     make(map[string][]string),
		FromRef:           tag,
		ToRef:             d.target,
	}, nil
}

// logChangeExample logs one example file+commit when a change is detected
func (d *Detector) logChangeExample(componentName string, reason *ChangeReason) {
	if !reason.Changed || len(reason.TriggeringCommits) == 0 {
		return
	}

	// Get the first commit and its first file as an example
	firstCommit := reason.TriggeringCommits[0]
	files, hasFiles := reason.CommitToFiles[firstCommit]

	if hasFiles && len(files) > 0 {
		firstFile := files[0]
		shortCommit := firstCommit
		if len(firstCommit) > 8 {
			shortCommit = firstCommit[:8]
		}
		_, _ = fmt.Fprintf(os.Stderr, "%s: %s (%s)\n", componentName, firstFile, shortCommit)
	}
}

// LoadTags loads tags from a JSON file
func LoadTags(path string) (Tags, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading tags file: %w", err)
	}

	var tags Tags
	if err := json.Unmarshal(data, &tags); err != nil {
		return nil, fmt.Errorf("error parsing tags JSON: %w", err)
	}

	return tags, nil
}

// WriteChanges writes the changes map to a JSON file
func WriteChanges(changes map[string]bool, path string) error {
	data, err := json.MarshalIndent(changes, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling changes: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("error writing changes file: %w", err)
	}

	return nil
}
