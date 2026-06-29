package refprotection

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

// Check is one policy entry. The match is presence-only (Param == ""), a
// numeric threshold (Min != nil), or an exact value (Equals != nil).
type Check struct {
	ID          string           `json:"id"`
	Severity    string           `json:"severity"`
	Description string           `json:"description"`
	RuleType    string           `json:"ruleType"`
	Param       string           `json:"param,omitempty"`
	Min         *float64         `json:"min,omitempty"`
	Equals      *json.RawMessage `json:"equals,omitempty"`
}

type Policy struct {
	Branch []Check `json:"branch"`
	Tag    []Check `json:"tag"`
}

func (p *Policy) Checks(refType string) []Check {
	if refType == "tag" {
		return p.Tag
	}
	return p.Branch
}

func LoadPolicy(path string) (*Policy, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read policy: %w", err)
	}
	return ParsePolicy(data)
}

func ParsePolicy(data []byte) (*Policy, error) {
	var p Policy
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("parse policy: %w", err)
	}
	return &p, nil
}

type Result struct {
	Check  Check
	Pass   bool
	Reason string
}

func Evaluate(checks []Check, rules []Rule) []Result {
	out := make([]Result, 0, len(checks))
	for _, c := range checks {
		pass, reason := checkRule(c, rules)
		out = append(out, Result{Check: c, Pass: pass, Reason: reason})
	}
	return out
}

// checkRule passes if ANY matching rule satisfies the check.
func checkRule(c Check, rules []Rule) (bool, string) {
	matching := filterByType(rules, c.RuleType)
	if len(matching) == 0 {
		return false, fmt.Sprintf("no '%s' rule applies to this ref", c.RuleType)
	}

	if c.Param == "" {
		return true, fmt.Sprintf("'%s' rule is present [%s]", c.RuleType, sources(matching))
	}

	if c.Min != nil {
		best, bestRule, found := maxParam(matching, c.Param)
		if found && best >= *c.Min {
			return true, fmt.Sprintf("%s = %s (>= %s) [%s]",
				c.Param, num(best), num(*c.Min), bestRule.Source)
		}
		return false, fmt.Sprintf("%s = %s, need >= %s", c.Param, num(best), num(*c.Min))
	}

	if c.Equals != nil {
		want := strings.TrimSpace(string(*c.Equals))
		for _, r := range matching {
			if paramJSON(r, c.Param) == want {
				return true, fmt.Sprintf("%s = %s [%s]", c.Param, want, r.Source)
			}
		}
		return false, fmt.Sprintf("%s = %s, need %s", c.Param, paramJSON(matching[0], c.Param), want)
	}

	return true, fmt.Sprintf("'%s' rule is present [%s]", c.RuleType, sources(matching))
}

// SourcesLabel renders unique sources for display, e.g.
// "legacy branch protection + rulesets".
func SourcesLabel(rules []Rule) string {
	seen := map[string]bool{}
	var s []string
	for _, r := range rules {
		if r.Source == "" || seen[r.Source] {
			continue
		}
		seen[r.Source] = true
		if r.Source == "legacy" {
			s = append(s, "legacy branch protection")
		} else {
			s = append(s, "rulesets")
		}
	}
	sort.Strings(s)
	return strings.Join(s, " + ")
}

func filterByType(rules []Rule, t string) []Rule {
	var out []Rule
	for _, r := range rules {
		if r.Type == t {
			out = append(out, r)
		}
	}
	return out
}

func sources(rules []Rule) string {
	seen := map[string]bool{}
	var s []string
	for _, r := range rules {
		if r.Source != "" && !seen[r.Source] {
			seen[r.Source] = true
			s = append(s, r.Source)
		}
	}
	sort.Strings(s)
	return strings.Join(s, "+")
}

func maxParam(rules []Rule, param string) (float64, Rule, bool) {
	var best float64
	var bestRule Rule
	found := false
	for _, r := range rules {
		v, ok := r.Parameters[param]
		if !ok {
			continue
		}
		f, ok := toFloat(v)
		if !ok {
			continue
		}
		if !found || f > best {
			best, bestRule, found = f, r, true
		}
	}
	return best, bestRule, found
}

func toFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case json.Number:
		f, err := n.Float64()
		return f, err == nil
	}
	return 0, false
}

// paramJSON renders a param as compact JSON so bools/numbers/strings compare
// cleanly against the policy `equals` value.
func paramJSON(r Rule, param string) string {
	v, ok := r.Parameters[param]
	if !ok {
		return "null"
	}
	b, err := json.Marshal(v)
	if err != nil {
		return "null"
	}
	return string(b)
}

func num(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}
