package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestDiscoverTargetsInRepo(t *testing.T) {
	targets, err := DiscoverTargets("../..")
	if err != nil {
		t.Fatalf("DiscoverTargets: %v", err)
	}
	if len(targets) == 0 {
		t.Fatal("no targets discovered")
	}

	var actions, workflows int
	for _, target := range targets {
		switch target.Kind {
		case KindCompositeAction:
			actions++
			// The repo uses both spellings, so discovery has to accept either.
			if base := filepath.Base(target.YAML); base != "action.yml" && base != "action.yaml" {
				t.Errorf("unexpected action file %q", target.YAML)
			}
			if filepath.Base(target.Doc) != "README.md" {
				t.Errorf("action %q should document to README.md, got %q", target.Name(), target.Doc)
			}
		case KindReusableWorkflow:
			workflows++
			// A reusable workflow documents to a sibling .md of the same name.
			wantDoc := strings.TrimSuffix(target.YAML, filepath.Ext(target.YAML)) + ".md"
			if target.Doc != wantDoc {
				t.Errorf("workflow %q should document to %q, got %q", target.Name(), wantDoc, target.Doc)
			}
		}
	}

	if actions == 0 {
		t.Error("no composite actions discovered")
	}
	if workflows == 0 {
		t.Error("no reusable workflows discovered")
	}
}

func TestDiscoverSkipsNonReusableWorkflows(t *testing.T) {
	targets, err := DiscoverTargets("../..")
	if err != nil {
		t.Fatalf("DiscoverTargets: %v", err)
	}

	// This very workflow has no on.workflow_call trigger, so it must not be
	// treated as something to document.
	for _, target := range targets {
		if strings.HasSuffix(target.YAML, "check-action-docs.yaml") {
			t.Errorf("%s has no workflow_call trigger and should have been skipped", target.YAML)
		}
	}
}

func TestResolveKind(t *testing.T) {
	tests := []struct {
		file string
		flag string
		want Kind
	}{
		{"actions/foo/action.yml", "auto", KindCompositeAction},
		{"actions/foo/action.yaml", "auto", KindCompositeAction},
		{".github/workflows/thing.yml", "auto", KindReusableWorkflow},
		{".github/workflows/thing.yml", "composite", KindCompositeAction},
		{"actions/foo/action.yml", "workflow", KindReusableWorkflow},
	}

	for _, tt := range tests {
		got, err := resolveKind(tt.file, tt.flag)
		if err != nil {
			t.Errorf("resolveKind(%q, %q): %v", tt.file, tt.flag, err)
			continue
		}
		if got != tt.want {
			t.Errorf("resolveKind(%q, %q) = %v, want %v", tt.file, tt.flag, got, tt.want)
		}
	}

	if _, err := resolveKind("x", "nonsense"); err == nil {
		t.Error("expected an error for an unknown -kind")
	}
}
