package refprotection

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// Identity answers "what ref, in what repo, is publishing?" In the enforcement
// path it must come from a signed OIDC token or the runner's GITHUB_* env,
// never from caller input, so it can't be spoofed onto a different ref.
type Identity struct {
	Repository     string
	RepositoryID   string
	RefType        string // "branch" | "tag"
	RefName        string
	SHA            string
	EventName      string
	JobWorkflowRef string
	RefProtected   bool
	Source         string // "oidc" | "env" | "args"
}

// GitHub sends some claims (including ref_protected) as strings.
type oidcClaims struct {
	Repository     string `json:"repository"`
	RepositoryID   string `json:"repository_id"`
	Ref            string `json:"ref"`
	SHA            string `json:"sha"`
	EventName      string `json:"event_name"`
	JobWorkflowRef string `json:"job_workflow_ref"`
	RefProtected   string `json:"ref_protected"`
}

// IdentityFromJWT decodes (does NOT verify) the claims of a GitHub OIDC JWT.
// Safe only when the JWT was just fetched from GitHub's endpoint via
// FetchOIDCToken — trust comes from the source, not a signature check. Never
// call it on a token received as input.
func IdentityFromJWT(jwt string) (*Identity, error) {
	parts := strings.Split(jwt, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("malformed JWT: expected 3 dot-separated segments, got %d", len(parts))
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("decode JWT payload: %w", err)
	}
	var c oidcClaims
	if err := json.Unmarshal(payload, &c); err != nil {
		return nil, fmt.Errorf("parse JWT claims: %w", err)
	}
	id := &Identity{
		Repository:     c.Repository,
		RepositoryID:   c.RepositoryID,
		SHA:            c.SHA,
		EventName:      c.EventName,
		JobWorkflowRef: c.JobWorkflowRef,
		RefProtected:   c.RefProtected == "true",
		Source:         "oidc",
	}
	if err := id.setRef(c.Ref); err != nil {
		return nil, err
	}
	return id, nil
}

// IdentityFromEnv builds identity from the runner's GITHUB_* variables.
func IdentityFromEnv() (*Identity, error) {
	id := &Identity{
		Repository:     os.Getenv("GITHUB_REPOSITORY"),
		RepositoryID:   os.Getenv("GITHUB_REPOSITORY_ID"),
		SHA:            os.Getenv("GITHUB_SHA"),
		EventName:      os.Getenv("GITHUB_EVENT_NAME"),
		JobWorkflowRef: os.Getenv("GITHUB_WORKFLOW_REF"),
		Source:         "env",
	}
	if err := id.setRef(os.Getenv("GITHUB_REF")); err != nil {
		return nil, err
	}
	return id, nil
}

func (id *Identity) setRef(ref string) error {
	switch {
	case strings.HasPrefix(ref, "refs/heads/"):
		id.RefType, id.RefName = "branch", strings.TrimPrefix(ref, "refs/heads/")
	case strings.HasPrefix(ref, "refs/tags/"):
		id.RefType, id.RefName = "tag", strings.TrimPrefix(ref, "refs/tags/")
	default:
		return fmt.Errorf("cannot determine ref type from %q (need refs/heads/* or refs/tags/*)", ref)
	}
	return nil
}

// FetchOIDCToken requests an OIDC token for the audience from GitHub's runtime
// endpoint. Requires `id-token: write` on the job.
func FetchOIDCToken(audience string) (string, error) {
	reqURL := os.Getenv("ACTIONS_ID_TOKEN_REQUEST_URL")
	reqTok := os.Getenv("ACTIONS_ID_TOKEN_REQUEST_TOKEN")
	if reqURL == "" || reqTok == "" {
		return "", fmt.Errorf("OIDC env not present (the job needs `id-token: write`): " +
			"ACTIONS_ID_TOKEN_REQUEST_URL / ACTIONS_ID_TOKEN_REQUEST_TOKEN")
	}
	if audience != "" {
		reqURL += "&audience=" + url.QueryEscape(audience)
	}
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "bearer "+reqTok)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request OIDC token: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OIDC endpoint returned %d: %s", resp.StatusCode, body)
	}
	var out struct {
		Value string `json:"value"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return "", fmt.Errorf("parse OIDC response: %w", err)
	}
	if out.Value == "" {
		return "", fmt.Errorf("OIDC response had empty token value")
	}
	return out.Value, nil
}
