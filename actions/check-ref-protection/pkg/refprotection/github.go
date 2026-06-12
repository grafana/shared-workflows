package refprotection

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
)

// HTTPClient is the real GitHubAPI backed by api.github.com. Branch rulesets
// need `contents: read`; legacy protection and ruleset enumeration need
// `Administration: read`.
type HTTPClient struct {
	Token   string
	BaseURL string
	HTTP    *http.Client
}

func NewHTTPClient() *HTTPClient {
	token := os.Getenv("GH_TOKEN")
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}
	return &HTTPClient{Token: token, BaseURL: "https://api.github.com", HTTP: http.DefaultClient}
}

// get returns the status code so callers can treat 403/404 as "no data" (fail
// closed) rather than an error.
func (c *HTTPClient) get(path string, v any) (int, []byte, error) {
	req, err := http.NewRequest(http.MethodGet, c.BaseURL+path, nil)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusOK && v != nil {
		if err := json.Unmarshal(body, v); err != nil {
			return resp.StatusCode, body, fmt.Errorf("decode %s: %w", path, err)
		}
	}
	return resp.StatusCode, body, nil
}

func (c *HTTPClient) BranchRules(repo, branch string) ([]Rule, error) {
	var raw []struct {
		Type       string         `json:"type"`
		Parameters map[string]any `json:"parameters"`
	}
	code, _, err := c.get(fmt.Sprintf("/repos/%s/rules/branches/%s", repo, url.PathEscape(branch)), &raw)
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		return nil, nil // no readable rules -> fail closed
	}
	rules := make([]Rule, 0, len(raw))
	for _, r := range raw {
		rules = append(rules, Rule{Type: r.Type, Parameters: r.Parameters})
	}
	return rules, nil
}

func (c *HTTPClient) LegacyProtection(repo, branch string) (*LegacyProtection, error) {
	var lp LegacyProtection
	code, _, err := c.get(fmt.Sprintf("/repos/%s/branches/%s/protection", repo, url.PathEscape(branch)), &lp)
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		// 404 = not protected; 403 = token lacks Administration: read. Either
		// way we have no legacy data — treat as none (fail closed).
		return nil, nil
	}
	return &lp, nil
}

var linkNextRe = regexp.MustCompile(`<([^>]+)>;\s*rel="next"`)

func (c *HTTPClient) ListRulesets(repo string) ([]RulesetSummary, error) {
	path := fmt.Sprintf("/repos/%s/rulesets?includes_parents=true&per_page=100", repo)
	var all []RulesetSummary
	for path != "" {
		var page []RulesetSummary
		req, err := http.NewRequest(http.MethodGet, c.BaseURL+path, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Accept", "application/vnd.github+json")
		req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
		if c.Token != "" {
			req.Header.Set("Authorization", "Bearer "+c.Token)
		}
		resp, err := c.HTTP.Do(req)
		if err != nil {
			return nil, err
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			// Most likely the token lacks Administration: read — no rulesets
			// visible. Fail closed by returning what we have (possibly none).
			return all, nil
		}
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, fmt.Errorf("decode rulesets: %w", err)
		}
		all = append(all, page...)

		path = ""
		if m := linkNextRe.FindStringSubmatch(resp.Header.Get("Link")); m != nil {
			path = strings.TrimPrefix(m[1], c.BaseURL)
		}
	}
	return all, nil
}

func (c *HTTPClient) GetRuleset(repo string, id int64) (*Ruleset, error) {
	var rs Ruleset
	code, _, err := c.get(fmt.Sprintf("/repos/%s/rulesets/%d", repo, id), &rs)
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		return nil, nil
	}
	return &rs, nil
}
