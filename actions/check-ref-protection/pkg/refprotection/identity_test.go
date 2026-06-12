package refprotection

import (
	"encoding/base64"
	"encoding/json"
	"testing"
)

// makeJWT builds a syntactically-valid JWT (header.payload.signature) whose
// payload is the given claims. The signature is a dummy — IdentityFromJWT
// decodes claims without verifying, which is safe because in production the
// token is fetched directly from GitHub.
func makeJWT(t *testing.T, claims map[string]any) string {
	t.Helper()
	payload, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("marshal claims: %v", err)
	}
	enc := base64.RawURLEncoding.EncodeToString
	return "header." + enc(payload) + ".signature"
}

func TestIdentityFromJWT_Branch(t *testing.T) {
	jwt := makeJWT(t, map[string]any{
		"repository":       "example/app",
		"repository_id":    "123456",
		"ref":              "refs/heads/release-1",
		"sha":              "abc123",
		"event_name":       "push",
		"job_workflow_ref": "example/workflows/.github/workflows/publish.yml@v1",
		"ref_protected":    "true",
	})

	id, err := IdentityFromJWT(jwt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id.Repository != "example/app" || id.RepositoryID != "123456" {
		t.Errorf("repo: got %q/%q", id.Repository, id.RepositoryID)
	}
	if id.RefType != "branch" || id.RefName != "release-1" {
		t.Errorf("ref: got %q/%q, want branch/release-1", id.RefType, id.RefName)
	}
	if id.SHA != "abc123" || id.EventName != "push" {
		t.Errorf("sha/event: got %q/%q", id.SHA, id.EventName)
	}
	if !id.RefProtected {
		t.Errorf("ref_protected: got false, want true")
	}
	if id.Source != "oidc" {
		t.Errorf("source: got %q, want oidc", id.Source)
	}
}

func TestIdentityFromJWT_Tag(t *testing.T) {
	jwt := makeJWT(t, map[string]any{
		"repository":    "example/app",
		"ref":           "refs/tags/app-2.9.4",
		"ref_protected": "false",
	})
	id, err := IdentityFromJWT(jwt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id.RefType != "tag" || id.RefName != "app-2.9.4" {
		t.Errorf("ref: got %q/%q, want tag/app-2.9.4", id.RefType, id.RefName)
	}
	if id.RefProtected {
		t.Errorf("ref_protected: got true, want false")
	}
}

func TestIdentityFromJWT_Errors(t *testing.T) {
	if _, err := IdentityFromJWT("not-a-jwt"); err == nil {
		t.Error("expected error for malformed JWT")
	}
	// valid JWT shape but an unsupported ref type (e.g. a PR merge ref)
	jwt := makeJWT(t, map[string]any{"ref": "refs/pull/42/merge"})
	if _, err := IdentityFromJWT(jwt); err == nil {
		t.Error("expected error for non-branch/tag ref")
	}
}
