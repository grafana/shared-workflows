package refprotection

import "testing"

// mockAPI is an in-memory GitHubAPI for deterministic, offline tests.
type mockAPI struct {
	branchRules map[string][]Rule
	legacy      map[string]*LegacyProtection
	summaries   []RulesetSummary
	rulesets    map[int64]*Ruleset
}

func (m *mockAPI) BranchRules(_, branch string) ([]Rule, error) { return m.branchRules[branch], nil }
func (m *mockAPI) LegacyProtection(_, branch string) (*LegacyProtection, error) {
	return m.legacy[branch], nil
}
func (m *mockAPI) ListRulesets(string) ([]RulesetSummary, error)   { return m.summaries, nil }
func (m *mockAPI) GetRuleset(_ string, id int64) (*Ruleset, error) { return m.rulesets[id], nil }

func tagRuleset(name, enforcement string, include []string, ruleTypes ...string) *Ruleset {
	rs := &Ruleset{Name: name, Target: "tag", Enforcement: enforcement}
	rs.Conditions.RefName.Include = include
	for _, t := range ruleTypes {
		rs.Rules = append(rs.Rules, struct {
			Type       string         `json:"type"`
			Parameters map[string]any `json:"parameters"`
		}{Type: t})
	}
	return rs
}

func TestCollectBranchRules_MergesSources(t *testing.T) {
	api := &mockAPI{
		branchRules: map[string][]Rule{
			"main": {{Type: "non_fast_forward"}},
		},
		legacy: map[string]*LegacyProtection{
			"main": {
				RequiredPullRequestReviews: &struct {
					RequiredApprovingReviewCount int  `json:"required_approving_review_count"`
					DismissStaleReviews          bool `json:"dismiss_stale_reviews"`
					RequireLastPushApproval      bool `json:"require_last_push_approval"`
				}{RequiredApprovingReviewCount: 1, DismissStaleReviews: true, RequireLastPushApproval: true},
				AllowDeletions: &struct {
					Enabled bool `json:"enabled"`
				}{Enabled: false},
			},
		},
	}

	rules, err := CollectBranchRules(api, "example/app", "main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// rulesets source: non_fast_forward; legacy source: pull_request + deletion
	assertHasRule(t, rules, "non_fast_forward", "rulesets")
	assertHasRule(t, rules, "pull_request", "legacy")
	assertHasRule(t, rules, "deletion", "legacy")

	pr := findRule(rules, "pull_request")
	if pr == nil {
		t.Fatal("missing pull_request rule")
	}
	if pr.Parameters["dismiss_stale_reviews_on_push"] != true {
		t.Errorf("legacy dismiss_stale not normalized: %v", pr.Parameters)
	}
	if pr.Parameters["required_approving_review_count"] != float64(1) {
		t.Errorf("legacy approval count not normalized: %v", pr.Parameters)
	}
}

func TestCollectTagRules_ActiveVsNonActive(t *testing.T) {
	api := &mockAPI{
		summaries: []RulesetSummary{
			{ID: 1, Target: "tag"},
			{ID: 2, Target: "tag"},
			{ID: 3, Target: "branch"}, // must be ignored
		},
		rulesets: map[int64]*Ruleset{
			1: tagRuleset("active creation", "active", []string{"refs/tags/v*"}, "creation"),
			2: tagRuleset("draft signing", "evaluate", []string{"~ALL"}, "non_fast_forward"),
			3: tagRuleset("branch thing", "active", []string{"~ALL"}, "pull_request"),
		},
	}

	rules, nonActive, err := CollectTagRules(api, "example/app", "v1.9.2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// active tag ruleset 1 matches v* -> creation counts
	if findRule(rules, "creation") == nil {
		t.Error("expected active 'creation' rule to be counted")
	}
	// evaluate-mode ruleset 2 must NOT count, but must be reported
	if findRule(rules, "non_fast_forward") != nil {
		t.Error("evaluate-mode rule must not be counted toward the bar")
	}
	if len(nonActive) != 1 || nonActive[0].Enforcement != "evaluate" {
		t.Fatalf("expected one non-active ruleset reported, got %+v", nonActive)
	}
	// branch-target ruleset 3 must be ignored entirely
	if findRule(rules, "pull_request") != nil {
		t.Error("branch-target ruleset must be ignored on the tag path")
	}
}

func TestCollectTagRules_NoMatchingRuleset(t *testing.T) {
	api := &mockAPI{
		summaries: []RulesetSummary{{ID: 1, Target: "tag"}},
		rulesets: map[int64]*Ruleset{
			1: tagRuleset("only v tags", "active", []string{"refs/tags/v*"}, "creation"),
		},
	}
	// app-2.9.4 does not match refs/tags/v*
	rules, nonActive, err := CollectTagRules(api, "example/app", "app-2.9.4")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rules) != 0 || len(nonActive) != 0 {
		t.Errorf("expected no rules for non-matching tag, got rules=%v nonActive=%v", rules, nonActive)
	}
}

func findRule(rules []Rule, t string) *Rule {
	for i := range rules {
		if rules[i].Type == t {
			return &rules[i]
		}
	}
	return nil
}

func assertHasRule(t *testing.T, rules []Rule, ruleType, source string) {
	t.Helper()
	for _, r := range rules {
		if r.Type == ruleType && r.Source == source {
			return
		}
	}
	t.Errorf("expected rule %q from source %q; got %+v", ruleType, source, rules)
}
