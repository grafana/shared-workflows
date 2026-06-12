package refprotection

import "testing"

// a small policy mirroring the real one's shape: presence, numeric min, and
// exact-equals checks.
const testPolicy = `{
  "branch": [
    {"id":"requires-pr","severity":"required","description":"PR required","ruleType":"pull_request"},
    {"id":"min-approvals","severity":"required","description":">=1 approval","ruleType":"pull_request","param":"required_approving_review_count","min":1},
    {"id":"no-self-approve","severity":"required","description":"no self-approve","ruleType":"pull_request","param":"require_last_push_approval","equals":true},
    {"id":"block-force-push","severity":"required","description":"no force-push","ruleType":"non_fast_forward"},
    {"id":"signed","severity":"optional","description":"signed commits","ruleType":"required_signatures"}
  ],
  "tag": [
    {"id":"restrict-creation","severity":"required","description":"restrict creation","ruleType":"creation"}
  ]
}`

func mustPolicy(t *testing.T) *Policy {
	t.Helper()
	p, err := ParsePolicy([]byte(testPolicy))
	if err != nil {
		t.Fatalf("parse policy: %v", err)
	}
	return p
}

func TestEvaluate_AllPass(t *testing.T) {
	p := mustPolicy(t)
	rules := []Rule{
		{Type: "pull_request", Source: "rulesets", Parameters: map[string]any{
			"required_approving_review_count": float64(2),
			"require_last_push_approval":      true,
		}},
		{Type: "non_fast_forward", Source: "rulesets"},
	}
	results := Evaluate(p.Checks("branch"), rules)

	reqFailed := false
	for _, r := range results {
		if r.Check.Severity == "required" && !r.Pass {
			reqFailed = true
			t.Errorf("required check %q failed: %s", r.Check.ID, r.Reason)
		}
	}
	if reqFailed {
		t.Fatal("expected all required checks to pass")
	}
	// optional signed-commits is absent -> should be a non-pass but not required
	for _, r := range results {
		if r.Check.ID == "signed" && r.Pass {
			t.Error("signed-commits should not pass (no rule present)")
		}
	}
}

func TestEvaluate_MinBelowThreshold(t *testing.T) {
	p := mustPolicy(t)
	rules := []Rule{
		{Type: "pull_request", Source: "rulesets", Parameters: map[string]any{
			"required_approving_review_count": float64(0),
			"require_last_push_approval":      true,
		}},
		{Type: "non_fast_forward", Source: "rulesets"},
	}
	results := Evaluate(p.Checks("branch"), rules)
	got := resultByID(results, "min-approvals")
	if got.Pass {
		t.Errorf("min-approvals should fail at count=0, reason=%q", got.Reason)
	}
}

func TestEvaluate_EqualsMismatch(t *testing.T) {
	p := mustPolicy(t)
	rules := []Rule{
		{Type: "pull_request", Source: "rulesets", Parameters: map[string]any{
			"required_approving_review_count": float64(1),
			"require_last_push_approval":      false, // policy wants true
		}},
	}
	results := Evaluate(p.Checks("branch"), rules)
	got := resultByID(results, "no-self-approve")
	if got.Pass {
		t.Errorf("no-self-approve should fail when require_last_push_approval=false")
	}
}

func TestEvaluate_MissingRuleType(t *testing.T) {
	p := mustPolicy(t)
	results := Evaluate(p.Checks("tag"), nil) // no rules at all
	got := resultByID(results, "restrict-creation")
	if got.Pass {
		t.Error("restrict-creation should fail when no creation rule applies")
	}
	if got.Reason == "" {
		t.Error("expected a reason explaining the miss")
	}
}

func TestSourcesLabel(t *testing.T) {
	rules := []Rule{{Type: "x", Source: "rulesets"}, {Type: "y", Source: "legacy"}}
	if got := SourcesLabel(rules); got != "legacy branch protection + rulesets" {
		t.Errorf("SourcesLabel = %q", got)
	}
}

func resultByID(results []Result, id string) Result {
	for _, r := range results {
		if r.Check.ID == id {
			return r
		}
	}
	return Result{}
}
