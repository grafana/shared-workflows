package refprotection

import (
	"regexp"
	"strings"
)

// RefNameMatches reports whether fullRef matches a ruleset's ref_name
// conditions: at least one include and no exclude.
func RefNameMatches(include, exclude []string, fullRef string) bool {
	matched := false
	for _, p := range include {
		if patternMatches(p, fullRef) {
			matched = true
			break
		}
	}
	if !matched {
		return false
	}
	for _, p := range exclude {
		if patternMatches(p, fullRef) {
			return false
		}
	}
	return true
}

func patternMatches(pattern, ref string) bool {
	switch pattern {
	case "~ALL":
		return true
	case "~DEFAULT_BRANCH":
		return false // branch-only token; never matches a tag
	}
	return globToRegexp(pattern).MatchString(ref)
}

// globToRegexp builds an anchored regexp where '*' crosses '/' (GitHub fnmatch).
func globToRegexp(pattern string) *regexp.Regexp {
	var b strings.Builder
	b.WriteString("^")
	for _, r := range pattern {
		switch r {
		case '*':
			b.WriteString(".*")
		case '?':
			b.WriteString(".")
		default:
			b.WriteString(regexp.QuoteMeta(string(r)))
		}
	}
	b.WriteString("$")
	return regexp.MustCompile(b.String())
}
