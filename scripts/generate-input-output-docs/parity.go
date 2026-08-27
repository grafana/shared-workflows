package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// ParityRule declares that a reusable workflow is expected to forward every
// input and output of a composite action, plus a set of extras that belong to
// the workflow alone.
//
// docker-build-push-multiarch wraps docker-build-push-image and re-declares its
// inputs so callers get the same surface either way. Keeping the two in sync by
// hand is what motivated this tool; this rule is what enforces it.
type ParityRule struct {
	Workflow string `yaml:"workflow"`
	Action   string `yaml:"action"`
	// WorkflowOnlyInputs are inputs the workflow declares that the action does
	// not, because they configure the workflow itself rather than the build.
	WorkflowOnlyInputs []string `yaml:"workflow-only-inputs"`
	// ActionOnlyInputs are inputs the action declares that the workflow
	// deliberately does not forward.
	ActionOnlyInputs []string `yaml:"action-only-inputs"`
	// WorkflowOnlyOutputs and ActionOnlyOutputs work the same way for outputs.
	WorkflowOnlyOutputs []string `yaml:"workflow-only-outputs"`
	ActionOnlyOutputs   []string `yaml:"action-only-outputs"`
}

// ParityConfig is the on-disk list of rules.
type ParityConfig struct {
	Rules []ParityRule `yaml:"rules"`
}

// LoadParityConfig reads the rules file.
func LoadParityConfig(path string) (*ParityConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg ParityConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	return &cfg, nil
}

// Check compares the workflow and action named by the rule and returns one
// human-readable problem per mismatch. An empty result means they are in sync.
func (r ParityRule) Check(root string) ([]string, error) {
	workflow, err := ParseFile(joinRoot(root, r.Workflow), KindReusableWorkflow)
	if err != nil {
		return nil, err
	}
	action, err := ParseFile(joinRoot(root, r.Action), KindCompositeAction)
	if err != nil {
		return nil, err
	}

	var problems []string
	problems = append(problems, compareSets(
		"input", r.Workflow, r.Action,
		workflow.InputNames(), action.InputNames(),
		r.WorkflowOnlyInputs, r.ActionOnlyInputs,
	)...)
	problems = append(problems, compareSets(
		"output", r.Workflow, r.Action,
		workflow.OutputNames(), action.OutputNames(),
		r.WorkflowOnlyOutputs, r.ActionOnlyOutputs,
	)...)
	return problems, nil
}

// compareSets reports names present on one side but not the other, ignoring the
// names each side is allowed to hold exclusively.
func compareSets(label, workflowPath, actionPath string, workflowNames, actionNames, allowedWorkflowOnly, allowedActionOnly []string) []string {
	inWorkflow := toSet(workflowNames)
	inAction := toSet(actionNames)
	allowWorkflow := toSet(allowedWorkflowOnly)
	allowAction := toSet(allowedActionOnly)

	var problems []string

	var missingFromWorkflow []string
	for _, name := range actionNames {
		if !inWorkflow[name] && !allowAction[name] {
			missingFromWorkflow = append(missingFromWorkflow, name)
		}
	}
	if len(missingFromWorkflow) > 0 {
		sort.Strings(missingFromWorkflow)
		problems = append(problems, fmt.Sprintf(
			"%s declares %s %s which %s does not; forward them or add them to action-only-%ss",
			actionPath, label, quoteList(missingFromWorkflow), workflowPath, label,
		))
	}

	var missingFromAction []string
	for _, name := range workflowNames {
		if !inAction[name] && !allowWorkflow[name] {
			missingFromAction = append(missingFromAction, name)
		}
	}
	if len(missingFromAction) > 0 {
		sort.Strings(missingFromAction)
		problems = append(problems, fmt.Sprintf(
			"%s declares %s %s which %s does not; remove them or add them to workflow-only-%ss",
			workflowPath, label, quoteList(missingFromAction), actionPath, label,
		))
	}

	// A stale allow-list entry is itself drift: it hides a name that no longer
	// exists, so a future real mismatch could slip through unnoticed.
	problems = append(problems, staleAllowances(allowedWorkflowOnly, inWorkflow, "workflow-only-"+label+"s", workflowPath)...)
	problems = append(problems, staleAllowances(allowedActionOnly, inAction, "action-only-"+label+"s", actionPath)...)

	return problems
}

func staleAllowances(allowed []string, present map[string]bool, field, path string) []string {
	var stale []string
	for _, name := range allowed {
		if !present[name] {
			stale = append(stale, name)
		}
	}
	if len(stale) == 0 {
		return nil
	}
	sort.Strings(stale)
	return []string{fmt.Sprintf(
		"%s lists %s which %s no longer declares; drop them from the rule",
		field, quoteList(stale), path,
	)}
}

// joinRoot resolves a rule's repo-relative path against the configured root.
func joinRoot(root, path string) string {
	return filepath.Join(root, filepath.FromSlash(path))
}

func toSet(names []string) map[string]bool {
	set := make(map[string]bool, len(names))
	for _, n := range names {
		set[n] = true
	}
	return set
}

func quoteList(names []string) string {
	quoted := make([]string, 0, len(names))
	for _, n := range names {
		quoted = append(quoted, "`"+n+"`")
	}
	return strings.Join(quoted, ", ")
}
