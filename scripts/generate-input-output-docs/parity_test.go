package main

import (
	"strings"
	"testing"
)

func TestCompareSetsInSync(t *testing.T) {
	problems := compareSets(
		"input", "wf.yml", "action.yaml",
		[]string{"a", "b"}, []string{"a", "b"},
		nil, nil,
	)
	if len(problems) != 0 {
		t.Errorf("expected no problems, got %v", problems)
	}
}

func TestCompareSetsReportsMissingFromWorkflow(t *testing.T) {
	problems := compareSets(
		"input", "wf.yml", "action.yaml",
		[]string{"a"}, []string{"a", "b"},
		nil, nil,
	)
	if len(problems) != 1 {
		t.Fatalf("expected 1 problem, got %v", problems)
	}
	if !strings.Contains(problems[0], "`b`") {
		t.Errorf("problem should name the missing input: %q", problems[0])
	}
}

func TestCompareSetsReportsMissingFromAction(t *testing.T) {
	problems := compareSets(
		"input", "wf.yml", "action.yaml",
		[]string{"a", "extra"}, []string{"a"},
		nil, nil,
	)
	if len(problems) != 1 {
		t.Fatalf("expected 1 problem, got %v", problems)
	}
	if !strings.Contains(problems[0], "`extra`") {
		t.Errorf("problem should name the extra input: %q", problems[0])
	}
}

func TestCompareSetsHonoursAllowLists(t *testing.T) {
	problems := compareSets(
		"input", "wf.yml", "action.yaml",
		[]string{"a", "runner-type"}, []string{"a", "builder"},
		[]string{"runner-type"}, []string{"builder"},
	)
	if len(problems) != 0 {
		t.Errorf("declared one-sided names should be allowed, got %v", problems)
	}
}

func TestCompareSetsReportsStaleAllowList(t *testing.T) {
	// A stale exception is drift in its own right: it would mask a future real
	// mismatch on the same name.
	problems := compareSets(
		"input", "wf.yml", "action.yaml",
		[]string{"a"}, []string{"a"},
		[]string{"removed-input"}, nil,
	)
	if len(problems) != 1 {
		t.Fatalf("expected 1 problem, got %v", problems)
	}
	if !strings.Contains(problems[0], "removed-input") {
		t.Errorf("problem should name the stale entry: %q", problems[0])
	}
}

func TestRepoParityConfigIsInSync(t *testing.T) {
	// Guards the checked-in rules against the real files, so `go test` catches
	// drift even without running the parity subcommand.
	cfg, err := LoadParityConfig("parity.yaml")
	if err != nil {
		t.Fatalf("LoadParityConfig: %v", err)
	}
	if len(cfg.Rules) == 0 {
		t.Fatal("parity.yaml declares no rules")
	}

	for _, rule := range cfg.Rules {
		problems, err := rule.Check("../..")
		if err != nil {
			t.Fatalf("checking %s: %v", rule.Workflow, err)
		}
		for _, p := range problems {
			t.Errorf("%s <-> %s: %s", rule.Workflow, rule.Action, p)
		}
	}
}
