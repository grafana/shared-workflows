package refprotection

// GitHubAPI is the slice of GitHub we depend on, as an interface so collection
// can be unit-tested without network access.
type GitHubAPI interface {
	BranchRules(repo, branch string) ([]Rule, error)
	LegacyProtection(repo, branch string) (*LegacyProtection, error)
	ListRulesets(repo string) ([]RulesetSummary, error)
	GetRuleset(repo string, id int64) (*Ruleset, error)
}

type RulesetSummary struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Target      string `json:"target"`
	Enforcement string `json:"enforcement"`
}

type Ruleset struct {
	Name        string `json:"name"`
	Target      string `json:"target"`
	Enforcement string `json:"enforcement"`
	Conditions  struct {
		RefName struct {
			Include []string `json:"include"`
			Exclude []string `json:"exclude"`
		} `json:"ref_name"`
	} `json:"conditions"`
	Rules []struct {
		Type       string         `json:"type"`
		Parameters map[string]any `json:"parameters"`
	} `json:"rules"`
}

type LegacyProtection struct {
	RequiredPullRequestReviews *struct {
		RequiredApprovingReviewCount int  `json:"required_approving_review_count"`
		DismissStaleReviews          bool `json:"dismiss_stale_reviews"`
		RequireLastPushApproval      bool `json:"require_last_push_approval"`
	} `json:"required_pull_request_reviews"`
	AllowForcePushes *struct {
		Enabled bool `json:"enabled"`
	} `json:"allow_force_pushes"`
	AllowDeletions *struct {
		Enabled bool `json:"enabled"`
	} `json:"allow_deletions"`
	RequiredSignatures *struct {
		Enabled bool `json:"enabled"`
	} `json:"required_signatures"`
	RequiredStatusChecks any `json:"required_status_checks"`
}

// CollectBranchRules merges effective ruleset rules with legacy branch
// protection; a check passes if either source satisfies it.
func CollectBranchRules(api GitHubAPI, repo, branch string) ([]Rule, error) {
	rules, err := api.BranchRules(repo, branch)
	if err != nil {
		return nil, err
	}
	for i := range rules {
		rules[i].Source = "rulesets"
	}

	legacy, err := api.LegacyProtection(repo, branch)
	if err != nil {
		return nil, err
	}
	if legacy != nil {
		rules = append(rules, legacyToRules(legacy)...)
	}
	return rules, nil
}

// legacyToRules maps classic branch protection onto ruleset-shaped Rules
// (e.g. dismiss_stale_reviews -> dismiss_stale_reviews_on_push).
func legacyToRules(l *LegacyProtection) []Rule {
	var out []Rule
	if l.RequiredPullRequestReviews != nil {
		out = append(out, Rule{
			Type:   "pull_request",
			Source: "legacy",
			Parameters: map[string]any{
				"required_approving_review_count": float64(l.RequiredPullRequestReviews.RequiredApprovingReviewCount),
				"dismiss_stale_reviews_on_push":   l.RequiredPullRequestReviews.DismissStaleReviews,
				"require_last_push_approval":      l.RequiredPullRequestReviews.RequireLastPushApproval,
			},
		})
	}
	if l.AllowForcePushes != nil && !l.AllowForcePushes.Enabled {
		out = append(out, Rule{Type: "non_fast_forward", Source: "legacy"})
	}
	if l.AllowDeletions != nil && !l.AllowDeletions.Enabled {
		out = append(out, Rule{Type: "deletion", Source: "legacy"})
	}
	if l.RequiredSignatures != nil && l.RequiredSignatures.Enabled {
		out = append(out, Rule{Type: "required_signatures", Source: "legacy"})
	}
	if l.RequiredStatusChecks != nil {
		out = append(out, Rule{Type: "required_status_checks", Source: "legacy"})
	}
	return out
}

// CollectTagRules keeps the rules of ACTIVE tag rulesets that match this tag,
// and records matching-but-not-enforced rulesets separately (never counted).
func CollectTagRules(api GitHubAPI, repo, tag string) ([]Rule, []NonActiveRuleset, error) {
	summaries, err := api.ListRulesets(repo)
	if err != nil {
		return nil, nil, err
	}

	fullRef := "refs/tags/" + tag
	var rules []Rule
	var nonActive []NonActiveRuleset

	for _, s := range summaries {
		if s.Target != "tag" {
			continue
		}
		rs, err := api.GetRuleset(repo, s.ID)
		if err != nil || rs == nil {
			continue
		}
		if !RefNameMatches(rs.Conditions.RefName.Include, rs.Conditions.RefName.Exclude, fullRef) {
			continue
		}

		if rs.Enforcement == "active" {
			for _, r := range rs.Rules {
				rules = append(rules, Rule{Type: r.Type, Parameters: r.Parameters, Source: "rulesets"})
			}
			continue
		}

		types := make([]string, 0, len(rs.Rules))
		for _, r := range rs.Rules {
			types = append(types, r.Type)
		}
		nonActive = append(nonActive, NonActiveRuleset{
			Name:        rs.Name,
			Enforcement: rs.Enforcement,
			Rules:       types,
		})
	}
	return rules, nonActive, nil
}
