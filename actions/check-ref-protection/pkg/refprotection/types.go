// Package refprotection checks whether the ref a workflow publishes from meets
// a protection policy. Branch and tag paths both normalize into []Rule, which a
// single policy evaluator then judges.
package refprotection

// Rule is one normalized protection rule. Both ref paths produce these.
type Rule struct {
	Type       string
	Parameters map[string]any
	Source     string // "rulesets" | "legacy"
}

// NonActiveRuleset matches the ref but is not hard-enforced (enforcement !=
// "active"), so it is reported but never counted toward the bar.
type NonActiveRuleset struct {
	Name        string
	Enforcement string
	Rules       []string
}
