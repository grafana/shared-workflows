#!/usr/bin/env bash
#
# check-ref-protection.sh — gate for prod-publish (WIF ref-protection).
#
# Reads the protection that applies to a single ref and checks it against the
# bar in policy.json. It merges TWO sources into one normalized rule set:
#   1. Rulesets        — GET /repos/{repo}/rules/{branches|tags}/{ref}
#   2. Legacy branch   — GET /repos/{repo}/branches/{branch}/protection
#      protection         (branches only; normalized into ruleset-shaped entries)
#
# A check passes if EITHER source satisfies it (GitHub enforces both and the
# most-restrictive wins, so presence in either source means it is enforced).
#
#   exit 0  all required checks pass  ->  safe to mint a prod token / push
#   exit 1  a required check failed   ->  caller must abort before publishing
#   exit 2  usage / setup error
#
# Auth: uses `gh` ($GH_TOKEN / $GITHUB_TOKEN).
#   - /rules/...            needs only `contents: read`.
#   - /branches/.../protection needs Administration: read (admin). When the
#     token can't read it, legacy protection is skipped with a warning.
#
# Usage: ./check-ref-protection.sh <owner/repo> <branch|tag> <ref-name> [policy.json]

set -euo pipefail

US=$'\037' # unit separator: a non-whitespace field delimiter (see read loop)

# --- arguments ---------------------------------------------------------------
REPO="${1:?usage: <owner/repo> <branch|tag> <ref-name> [policy.json]}"
REF_TYPE="${2:?ref type must be 'branch' or 'tag'}"
REF_NAME="${3:?ref name required}"
POLICY="${4:-"$(dirname "$0")/policy.json"}"

case "$REF_TYPE" in
  branch|tag) ;;
  *) echo "::error::ref type must be 'branch' or 'tag', got '$REF_TYPE'"; exit 2 ;;
esac
[ -f "$POLICY" ] || { echo "::error::policy file not found: $POLICY"; exit 2; }

echo "==> $REPO · $REF_TYPE · $REF_NAME"

# Tags are not supported yet: GitHub has no /rules/tags/{tag} endpoint, so tag
# rules can't be read the same way as branches. Supporting tags needs either
# ruleset enumeration (/rulesets + /rulesets/{id}, Administration: read) or a
# commit-ancestry check. Fail closed rather than report a misleading result.
if [ "$REF_TYPE" = tag ]; then
  echo "::error::tag protection checks are not implemented yet (no /rules/tags endpoint)."
  echo "  See README for the planned tag approach (ruleset enumeration / ancestry)."
  exit 2
fi

# --- source 1: ruleset rules -------------------------------------------------
# Tag each rule with its source so the output can show where protection came from.
raw="$(gh api "/repos/$REPO/rules/branches/$REF_NAME" 2>/dev/null)" || raw='[]'
RULES="$(jq -c 'if type == "array" then map(. + {_src: "rulesets"}) else [] end' <<<"$raw")"

# --- source 2: legacy branch protection (branches only) ----------------------
# Normalize the legacy protection object into ruleset-shaped entries so the
# same checks apply to both. Field names are mapped to the ruleset equivalents
# (e.g. legacy `dismiss_stale_reviews` -> `dismiss_stale_reviews_on_push`).
if [ "$REF_TYPE" = branch ]; then
  legacy=""
  if legacy="$(gh api "/repos/$REPO/branches/$REF_NAME/protection" 2>/tmp/crp_err.$$)"; then
    : # read ok
  elif grep -qi 'not protected\|not found' /tmp/crp_err.$$; then
    legacy="" # genuinely no legacy protection on this branch
  else
    echo "::warning::could not read legacy branch protection (token likely needs Administration: read) — skipping legacy source"
    legacy=""
  fi
  rm -f /tmp/crp_err.$$

  if [ -n "$legacy" ]; then
    legacy_rules="$(jq -c '
      [
        (if .required_pull_request_reviews then
          { type: "pull_request", parameters: {
              required_approving_review_count: (.required_pull_request_reviews.required_approving_review_count // 0),
              dismiss_stale_reviews_on_push:   (.required_pull_request_reviews.dismiss_stale_reviews // false),
              require_last_push_approval:      (.required_pull_request_reviews.require_last_push_approval // false)
          }} else empty end),
        (if .allow_force_pushes.enabled == false then { type: "non_fast_forward" } else empty end),
        (if .allow_deletions.enabled  == false then { type: "deletion" } else empty end),
        (if .required_signatures.enabled == true then { type: "required_signatures" } else empty end),
        (if .required_status_checks then { type: "required_status_checks" } else empty end)
      ] | map(. + {_src: "legacy"})' <<<"$legacy")"
    RULES="$(jq -c -n --argjson a "$RULES" --argjson b "$legacy_rules" '$a + $b')"
  fi
fi

# --- report which source(s) protect this ref --------------------------------
if [ "$(jq 'length' <<<"$RULES")" -eq 0 ]; then
  echo "    Protection source(s): none — no protection applies to this ref"
else
  echo "    Protection source(s): $(jq -r '
    [.[]._src] | unique
    | map(if . == "legacy" then "legacy branch protection" else "rulesets" end)
    | join(" + ")' <<<"$RULES")"
fi

# --- evaluate one policy entry against the merged rules ----------------------
# Passes if ANY matching rule satisfies the check. Sets REASON. 0=pass / 1=fail.
REASON=""
check_rule() {
  local rule_type="$1" param="$2" min="$3" equals="$4" best

  if ! jq -e --arg t "$rule_type" 'any(.[]; .type == $t)' <<<"$RULES" >/dev/null; then
    REASON="no '$rule_type' rule applies to this ref"
    return 1
  fi

  # presence-only check
  if [ -z "$param" ]; then
    src="$(jq -r --arg t "$rule_type" \
      'map(select(.type == $t)._src) | unique | join("+")' <<<"$RULES")"
    REASON="'$rule_type' rule is present [$src]"
    return 0
  fi

  # numeric threshold: pass if any matching rule has param >= min
  if [ -n "$min" ]; then
    best="$(jq -r --arg t "$rule_type" --arg p "$param" \
      'map(select(.type == $t) | .parameters[$p] // 0) | max' <<<"$RULES")"
    if [ "$best" -ge "$min" ] 2>/dev/null; then
      src="$(jq -r --arg t "$rule_type" --arg p "$param" \
        'map(select(.type == $t)) | max_by(.parameters[$p] // 0)._src' <<<"$RULES")"
      REASON="$param = $best (>= $min) [$src]"; return 0
    fi
    REASON="$param = $best, need >= $min"; return 1
  fi

  # exact match: pass if any matching rule has param == equals
  if jq -e --arg t "$rule_type" --arg p "$param" --arg v "$equals" \
       'any(.[]; .type == $t and ((.parameters[$p]) | tostring) == $v)' <<<"$RULES" >/dev/null; then
    src="$(jq -r --arg t "$rule_type" --arg p "$param" --arg v "$equals" \
      'map(select(.type == $t and ((.parameters[$p]) | tostring) == $v)) | .[0]._src' <<<"$RULES")"
    REASON="$param = $equals [$src]"; return 0
  fi
  best="$(jq -r --arg t "$rule_type" --arg p "$param" \
    'map(select(.type == $t) | .parameters[$p]) | .[0] | tostring' <<<"$RULES")"
  REASON="$param = $best, need $equals"; return 1
}

# --- run every policy entry for this ref type, render, tally -----------------
declare -i req_total=0 req_passed=0 failed=0
missing=()

printf '\n'
while IFS="$US" read -r id severity description rule_type param min equals; do
  if check_rule "$rule_type" "$param" "$min" "$equals"; then ok=1; else ok=0; fi

  if [ "$severity" = required ]; then
    req_total+=1
    if [ "$ok" -eq 1 ]; then req_passed+=1; else failed=1; missing+=("$description — $REASON"); fi
  fi

  if   [ "$ok" -eq 1 ];            then mark='\033[32m✓\033[0m'
  elif [ "$severity" = required ]; then mark='\033[31m✗\033[0m'
  else                                  mark='\033[33m⚠\033[0m'
  fi
  printf '  %b  %-18s %s\n'       "$mark" "$id" "$description"
  printf '       \033[2m↳ %s\033[0m\n' "$REASON"
done < <(jq -r --arg t "$REF_TYPE" --arg us "$US" '
  .[$t][] | [
    .id, .severity, .description, (.ruleType // ""),
    (.param // ""),
    (if has("min")    then (.min    | tostring) else "" end),
    (if has("equals") then (.equals | tostring) else "" end)
  ] | join($us)
' "$POLICY")

# --- verdict -----------------------------------------------------------------
printf '\n  Required: %d/%d passed\n' "$req_passed" "$req_total"

if [ "$failed" -eq 0 ]; then
  echo "  RESULT: PASS — $REF_TYPE '$REF_NAME' meets the bar. OK to publish."
  exit 0
fi

printf '\n  Missing required protections:\n'
for m in "${missing[@]}"; do echo "    • $m"; done
printf '\n'
echo "::error::$REF_TYPE '$REF_NAME' does NOT meet the prod-publish bar ($req_passed/$req_total required passed)."
echo "  RESULT: FAIL — refusing to publish."
exit 1
