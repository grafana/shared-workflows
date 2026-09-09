package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Target pairs a YAML file that declares inputs/outputs with the doc file those
// inputs/outputs should be written into.
type Target struct {
	Kind Kind
	// YAML is the action.yml/action.yaml or workflow file to read.
	YAML string
	// Doc is the markdown file to write into.
	Doc string
}

// Name is a short identifier used in log output.
func (t Target) Name() string {
	if t.Kind == KindCompositeAction {
		return filepath.Base(filepath.Dir(t.YAML))
	}
	return strings.TrimSuffix(filepath.Base(t.YAML), filepath.Ext(t.YAML))
}

// DiscoverTargets finds every composite action and reusable workflow in the repo
// rooted at root, paired with its doc file.
//
// Composite actions live at actions/<name>/action.yml or action.yaml (the repo
// uses both spellings) and are documented in actions/<name>/README.md. Reusable
// workflows live in .github/workflows and are documented in a sibling .md file
// of the same name.
func DiscoverTargets(root string) ([]Target, error) {
	actions, err := discoverActions(root)
	if err != nil {
		return nil, err
	}
	workflows, err := discoverWorkflows(root)
	if err != nil {
		return nil, err
	}

	targets := append(actions, workflows...)
	sort.Slice(targets, func(i, j int) bool { return targets[i].YAML < targets[j].YAML })
	return targets, nil
}

func discoverActions(root string) ([]Target, error) {
	entries, err := os.ReadDir(filepath.Join(root, "actions"))
	if err != nil {
		return nil, fmt.Errorf("reading actions directory: %w", err)
	}

	var targets []Target
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dir := filepath.Join(root, "actions", entry.Name())
		yamlPath, err := findActionYAML(dir)
		if err != nil {
			return nil, err
		}
		if yamlPath == "" {
			// Not every directory under actions/ is a composite action; some hold
			// only supporting code.
			continue
		}
		targets = append(targets, Target{
			Kind: KindCompositeAction,
			YAML: yamlPath,
			Doc:  filepath.Join(dir, "README.md"),
		})
	}
	return targets, nil
}

// findActionYAML returns the action definition in dir, accepting either
// extension GitHub allows. Both present is ambiguous and treated as an error
// rather than silently picking one.
func findActionYAML(dir string) (string, error) {
	var found []string
	for _, name := range []string{"action.yml", "action.yaml"} {
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); err == nil {
			found = append(found, path)
		}
	}
	switch len(found) {
	case 0:
		return "", nil
	case 1:
		return found[0], nil
	default:
		return "", fmt.Errorf("%s contains both action.yml and action.yaml", dir)
	}
}

func discoverWorkflows(root string) ([]Target, error) {
	dir := filepath.Join(root, ".github", "workflows")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading workflows directory: %w", err)
	}

	var targets []Target
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := filepath.Ext(entry.Name())
		if ext != ".yml" && ext != ".yaml" {
			continue
		}
		path := filepath.Join(dir, entry.Name())

		// Only reusable workflows have inputs and outputs worth documenting, so
		// parse each candidate and keep the ones exposing an on.workflow_call
		// trigger. A malformed workflow is a real error, not something to skip.
		if _, err := ParseFile(path, KindReusableWorkflow); err != nil {
			if errors.Is(err, ErrNotReusableWorkflow) {
				continue
			}
			return nil, err
		}
		targets = append(targets, Target{
			Kind: KindReusableWorkflow,
			YAML: path,
			Doc:  filepath.Join(dir, strings.TrimSuffix(entry.Name(), ext)+".md"),
		})
	}
	return targets, nil
}
