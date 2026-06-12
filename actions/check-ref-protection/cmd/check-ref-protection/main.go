// Command check-ref-protection gates prod publishing on ref protection.
//
// It derives the publishing ref's identity (preferably from a GitHub-signed
// OIDC token, never from caller input in the enforcement path), reads the
// protection that actually applies to that ref, and checks it against
// policy.json.
//
// Exit codes:
//
//	0  the ref meets the bar (or a failure was downgraded by -enforce=false)
//	1  the ref fails the bar and -enforce is set
//	2  usage / setup error
package main

import (
	"flag"
	"fmt"
	"os"

	rp "github.com/grafana/shared-workflows/actions/check-ref-protection/pkg/refprotection"
)

// oidcAudience labels the OIDC token we request in identity=oidc mode, marking
// it as scoped to this gate (distinct from the token the GCP auth step mints).
const oidcAudience = "wif-ref-protection"

const (
	green  = "\033[32m"
	red    = "\033[31m"
	yellow = "\033[33m"
	dim    = "\033[2m"
	reset  = "\033[0m"
)

func main() {
	os.Exit(run())
}

func run() int {
	var (
		policyPath = flag.String("policy", "policy.json", "path to policy.json")
		enforce    = flag.Bool("enforce", false, "fail (exit 1) when the ref does not meet the bar")
		identity   = flag.String("identity", "oidc", `identity source: "oidc" (default) | "args" (local testing)`)
	)
	flag.Parse()

	id, err := resolveIdentity(*identity, flag.Args())
	if err != nil {
		fmt.Printf("::error::%v\n", err)
		return 2
	}

	policy, err := rp.LoadPolicy(*policyPath)
	if err != nil {
		fmt.Printf("::error::%v\n", err)
		return 2
	}

	fmt.Printf("==> %s · %s · %s  (identity: %s)\n", id.Repository, id.RefType, id.RefName, id.Source)

	api := rp.NewHTTPClient()
	var rules []rp.Rule
	var nonActive []rp.NonActiveRuleset

	switch id.RefType {
	case "branch":
		rules, err = rp.CollectBranchRules(api, id.Repository, id.RefName)
	case "tag":
		rules, nonActive, err = rp.CollectTagRules(api, id.Repository, id.RefName)
	default:
		fmt.Printf("::error::unsupported ref type %q\n", id.RefType)
		return 2
	}
	if err != nil {
		fmt.Printf("::error::reading protection: %v\n", err)
		return 2
	}

	reportSources(rules, nonActive)

	results := rp.Evaluate(policy.Checks(id.RefType), rules)
	reqTotal, reqPassed, failed := render(results)

	fmt.Printf("\n  Required: %d/%d passed\n", reqPassed, reqTotal)
	if !failed {
		fmt.Printf("  RESULT: PASS — %s '%s' meets the bar. OK to publish.\n", id.RefType, id.RefName)
		return 0
	}

	fmt.Printf("\n  Missing required protections:\n")
	for _, r := range results {
		if r.Check.Severity == "required" && !r.Pass {
			fmt.Printf("    • %s — %s\n", r.Check.Description, r.Reason)
		}
	}
	if *enforce {
		fmt.Printf("::error::%s '%s' does NOT meet the prod-publish bar (%d/%d required passed).\n",
			id.RefType, id.RefName, reqPassed, reqTotal)
		fmt.Printf("  RESULT: FAIL — refusing to publish.\n")
		return 1
	}
	fmt.Printf("::warning::ref-protection gate DISABLED (enforce=false): '%s' would have been blocked — continuing.\n", id.RefName)
	return 0
}

// resolveIdentity picks the identity source. "oidc" derives a non-spoofable
// identity from the signed GitHub token; "args" takes positional <repo> <type>
// <name> for local/testing use only (never trust these in enforcement).
func resolveIdentity(source string, args []string) (*rp.Identity, error) {
	switch source {
	case "oidc":
		tok, err := rp.FetchOIDCToken(oidcAudience)
		if err != nil {
			return nil, err
		}
		return rp.IdentityFromJWT(tok)
	case "args":
		if len(args) != 3 {
			return nil, fmt.Errorf("usage: check-ref-protection [-identity oidc] | -identity args <owner/repo> <branch|tag> <ref-name>")
		}
		if args[1] != "branch" && args[1] != "tag" {
			return nil, fmt.Errorf("ref type must be 'branch' or 'tag', got %q", args[1])
		}
		return &rp.Identity{Repository: args[0], RefType: args[1], RefName: args[2], Source: "args"}, nil
	default:
		return nil, fmt.Errorf("unknown identity source %q (use oidc|args)", source)
	}
}

func reportSources(rules []rp.Rule, nonActive []rp.NonActiveRuleset) {
	if len(rules) == 0 {
		fmt.Println("    Protection source(s): none — no protection applies to this ref")
	} else {
		fmt.Printf("    Protection source(s): %s\n", rp.SourcesLabel(rules))
	}
	if len(nonActive) > 0 {
		fmt.Printf("    %sNot hard-enforced%s (matching rulesets with enforcement != active — ignored for the bar):\n", yellow, reset)
		for _, na := range nonActive {
			fmt.Printf("      ⚠ %s [enforcement=%s] → rules: %s\n", na.Name, na.Enforcement, join(na.Rules))
		}
	}
}

func render(results []rp.Result) (reqTotal, reqPassed int, failed bool) {
	fmt.Println()
	for _, r := range results {
		required := r.Check.Severity == "required"
		if required {
			reqTotal++
			if r.Pass {
				reqPassed++
			} else {
				failed = true
			}
		}
		var mark string
		switch {
		case r.Pass:
			mark = green + "✓" + reset
		case required:
			mark = red + "✗" + reset
		default:
			mark = yellow + "⚠" + reset
		}
		fmt.Printf("  %s  %-18s %s\n", mark, r.Check.ID, r.Check.Description)
		fmt.Printf("       %s↳ %s%s\n", dim, r.Reason, reset)
	}
	return reqTotal, reqPassed, failed
}

func join(s []string) string {
	out := ""
	for i, v := range s {
		if i > 0 {
			out += ", "
		}
		out += v
	}
	return out
}
