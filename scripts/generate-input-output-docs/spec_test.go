package main

import (
	"errors"
	"path/filepath"
	"testing"
)

func TestParseCompositeAction(t *testing.T) {
	spec, err := ParseFile(filepath.Join("testdata", "action.yaml"), KindCompositeAction)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}

	// Entries must come back sorted by name so generated docs are stable no
	// matter what order the YAML declares them in. The fixture declares `zebra`
	// first on purpose.
	got := spec.InputNames()
	want := []string{"alpha", "backticked", "boolish", "no-default", "pipes", "zebra"}
	if len(got) != len(want) {
		t.Fatalf("input names = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("input %d = %q, want %q", i, got[i], want[i])
		}
	}

	if names := spec.OutputNames(); len(names) != 1 || names[0] != "result" {
		t.Errorf("output names = %v, want [result]", names)
	}
}

func TestParseRequiredAndDefault(t *testing.T) {
	spec, err := ParseFile(filepath.Join("testdata", "action.yaml"), KindCompositeAction)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}

	byName := map[string]IOEntry{}
	for _, e := range spec.Inputs {
		byName[e.Name] = e
	}

	if !byName["alpha"].Required {
		t.Error("alpha should be required")
	}
	if byName["zebra"].Required {
		t.Error("zebra should not be required")
	}

	// An explicitly empty default and an absent default both render blank, but
	// the parser has to tell them apart for callers that care.
	if !byName["zebra"].HasDefault {
		t.Error("zebra declares `default: \"\"` so HasDefault should be true")
	}
	if byName["no-default"].HasDefault {
		t.Error("no-default declares no default so HasDefault should be false")
	}
}

func TestParseReusableWorkflow(t *testing.T) {
	spec, err := ParseFile(filepath.Join("testdata", "reusable-workflow.yml"), KindReusableWorkflow)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}

	if got, want := len(spec.Inputs), 2; got != want {
		t.Fatalf("got %d inputs, want %d", got, want)
	}
	// yaml.v3 must not fold the `on:` key into a boolean, or nothing below it
	// would be found at all.
	if spec.Inputs[0].Name != "flag" {
		t.Errorf("first input = %q, want flag", spec.Inputs[0].Name)
	}
	// Reusable workflows declare a type; it is used verbatim rather than guessed.
	if got := spec.Inputs[0].DisplayType(); got != "boolean" {
		t.Errorf("flag type = %q, want boolean (lowercase, as declared)", got)
	}
	if got := spec.Inputs[1].DisplayType(); got != "string" {
		t.Errorf("name type = %q, want string", got)
	}
}

func TestParsePlainWorkflowIsNotReusable(t *testing.T) {
	// `on:` spellings that carry no workflow_call trigger must all report
	// ErrNotReusableWorkflow, never a parse error. Discovery treats a parse error
	// as fatal, so getting this wrong means one contributor writing `on: push`
	// breaks doc generation for the whole repo.
	for _, name := range []string{
		"plain-workflow.yml",        // on: { pull_request: }
		"shorthand-on-scalar.yml",   // on: push
		"shorthand-on-sequence.yml", // on: [push, pull_request]
	} {
		t.Run(name, func(t *testing.T) {
			_, err := ParseFile(filepath.Join("testdata", name), KindReusableWorkflow)
			if !errors.Is(err, ErrNotReusableWorkflow) {
				t.Fatalf("err = %v, want ErrNotReusableWorkflow so discovery can skip it", err)
			}
		})
	}
}

func TestDisplayTypeInference(t *testing.T) {
	tests := []struct {
		name  string
		entry IOEntry
		want  string
	}{
		{"explicit type wins", IOEntry{Type: "number", Default: "true"}, "number"},
		{"true infers Boolean", IOEntry{Default: "true"}, "Boolean"},
		{"false infers Boolean", IOEntry{Default: "false"}, "Boolean"},
		{"mixed case infers Boolean", IOEntry{Default: "True"}, "Boolean"},
		{"other defaults are String", IOEntry{Default: "ubuntu-latest"}, "String"},
		// A composite action has no type field, so a boolean input with no
		// default is indistinguishable from a string one.
		{"no default is String", IOEntry{}, "String"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.entry.DisplayType(); got != tt.want {
				t.Errorf("DisplayType() = %q, want %q", got, tt.want)
			}
		})
	}
}
