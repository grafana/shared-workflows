package main

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// ErrNotReusableWorkflow reports that a workflow file has no `on.workflow_call`
// trigger, and so declares no inputs or outputs to document. Callers walking
// .github/workflows use this to skip ordinary workflows without swallowing real
// parse errors.
var ErrNotReusableWorkflow = errors.New("not a reusable workflow")

// Kind identifies which flavour of YAML a Spec was parsed from. The two flavours
// declare their inputs and outputs in different places and only reusable
// workflows carry an explicit type per input.
type Kind int

const (
	// KindCompositeAction is an action.yml/action.yaml describing a composite action.
	KindCompositeAction Kind = iota
	// KindReusableWorkflow is a workflow in .github/workflows with an `on.workflow_call` trigger.
	KindReusableWorkflow
)

func (k Kind) String() string {
	if k == KindReusableWorkflow {
		return "reusable workflow"
	}
	return "composite action"
}

// IOEntry is a single input or output declared by an action or reusable workflow.
type IOEntry struct {
	Name        string
	Description string
	Default     string
	HasDefault  bool
	Required    bool
	// Type is only ever set for reusable workflow inputs; composite actions have
	// no type field in their schema.
	Type string
}

// Spec is the set of inputs and outputs parsed out of a single YAML file.
type Spec struct {
	Kind    Kind
	Path    string
	Inputs  []IOEntry
	Outputs []IOEntry
}

// InputNames returns the declared input names in the order they are rendered.
func (s *Spec) InputNames() []string { return names(s.Inputs) }

// OutputNames returns the declared output names in the order they are rendered.
func (s *Spec) OutputNames() []string { return names(s.Outputs) }

func names(entries []IOEntry) []string {
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		out = append(out, e.Name)
	}
	return out
}

// rawIO mirrors the YAML shape of a single input/output. Every field is a
// yaml.Node because GitHub accepts unquoted scalars of several types here --
// `default: true` and `default: "true"` are both legal -- and Node.Value gives
// us the scalar text without having to enumerate those types.
type rawIO struct {
	Description yaml.Node `yaml:"description"`
	Default     yaml.Node `yaml:"default"`
	Required    yaml.Node `yaml:"required"`
	Type        yaml.Node `yaml:"type"`
	Value       yaml.Node `yaml:"value"`
}

type rawCompositeAction struct {
	Inputs  map[string]rawIO `yaml:"inputs"`
	Outputs map[string]rawIO `yaml:"outputs"`
}

// rawReusableWorkflow decodes `on` as a bare Node because the key is not always
// a mapping. GitHub accepts `on: push` and `on: [push, pull_request]` as well as
// the mapping form, and decoding those straight into a struct fails with a type
// error -- which would abort the whole run over a workflow that simply has no
// inputs to document.
type rawReusableWorkflow struct {
	On yaml.Node `yaml:"on"`
}

type rawWorkflowCall struct {
	WorkflowCall *struct {
		Inputs  map[string]rawIO `yaml:"inputs"`
		Outputs map[string]rawIO `yaml:"outputs"`
	} `yaml:"workflow_call"`
}

// ParseFile reads path and extracts its inputs and outputs. Entries are returned
// sorted by name so that generated docs are stable regardless of the order the
// YAML happens to declare them in.
func ParseFile(path string, kind Kind) (*Spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	spec := &Spec{Kind: kind, Path: path}

	switch kind {
	case KindCompositeAction:
		var raw rawCompositeAction
		if err := yaml.Unmarshal(data, &raw); err != nil {
			return nil, fmt.Errorf("parsing %s: %w", path, err)
		}
		spec.Inputs = convert(raw.Inputs)
		spec.Outputs = convert(raw.Outputs)

	case KindReusableWorkflow:
		var raw rawReusableWorkflow
		if err := yaml.Unmarshal(data, &raw); err != nil {
			return nil, fmt.Errorf("parsing %s: %w", path, err)
		}
		// `on: push` and `on: [push]` are legal and simply cannot carry a
		// workflow_call trigger, so there is nothing here to document.
		if raw.On.Kind != yaml.MappingNode {
			return nil, fmt.Errorf("%s: %w", path, ErrNotReusableWorkflow)
		}

		var call rawWorkflowCall
		if err := raw.On.Decode(&call); err != nil {
			return nil, fmt.Errorf("parsing %s: %w", path, err)
		}
		if call.WorkflowCall == nil {
			return nil, fmt.Errorf("%s: %w", path, ErrNotReusableWorkflow)
		}
		spec.Inputs = convert(call.WorkflowCall.Inputs)
		spec.Outputs = convert(call.WorkflowCall.Outputs)

	default:
		return nil, fmt.Errorf("unknown kind %d", kind)
	}

	return spec, nil
}

func convert(raw map[string]rawIO) []IOEntry {
	entries := make([]IOEntry, 0, len(raw))
	for name, r := range raw {
		entries = append(entries, IOEntry{
			Name:        name,
			Description: r.Description.Value,
			Default:     r.Default.Value,
			// An omitted `default:` unmarshals to a zero Node, which lets us tell
			// `default: ""` (deliberately empty) apart from no default at all.
			HasDefault: !r.Default.IsZero(),
			Required:   r.Required.Value == "true",
			Type:       r.Type.Value,
		})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name < entries[j].Name })
	return entries
}

// DisplayType is the value rendered in the Type column.
//
// Reusable workflows declare a type explicitly and it is used verbatim (GitHub
// spells these lowercase: string, boolean, number). Composite actions have no
// type field at all, so the type is inferred from the default value and
// rendered capitalized, matching the convention already used in this repo's
// action READMEs. An input with no default and no explicit type is reported as
// String -- see the caveat in this directory's README.
func (e IOEntry) DisplayType() string {
	if e.Type != "" {
		return e.Type
	}
	switch strings.ToLower(e.Default) {
	case "true", "false":
		return "Boolean"
	default:
		return "String"
	}
}
