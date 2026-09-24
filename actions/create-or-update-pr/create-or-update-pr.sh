#!/usr/bin/env bash
#
# create-or-update-pr.sh
#
# Creates (or updates) a pull request from file changes using git + gh CLI.
# Designed as a lightweight replacement for peter-evans/create-pull-request.
#
# Required environment variables:
#   PR_BRANCH       - branch name to push to
#   COMMIT_MSG      - commit message
#   PR_TITLE        - pull request title
#   PR_BODY         - pull request body (markdown)
#   ADD_PATHS       - comma, newline, or space-separated list of file paths/pathspecs to stage
#                     (paths containing spaces are not supported)
#
# Optional environment variables:
#   BASE_BRANCH     - base branch for the PR (default: main)
#   GIT_USER_NAME   - git user name for the commit (default: github-actions[bot])
#   GIT_USER_EMAIL  - git user email for the commit (default: 41898282+github-actions[bot]@users.noreply.github.com)
#   LABELS          - comma or newline-separated list of PR labels
#   REVIEWERS       - comma or newline-separated list of GitHub usernames to request review from
#   DRAFT           - set to "true" to create a draft PR
#   DRY_RUN         - set to "1" to print what would happen without pushing or creating PR
#   GH_TOKEN        - GitHub token (used by gh CLI and for git push authentication)

set -euo pipefail

# ---------------------------------------------------------------------------
# Validate required inputs
# ---------------------------------------------------------------------------
: "${PR_BRANCH:?PR_BRANCH is required}"
: "${COMMIT_MSG:?COMMIT_MSG is required}"
: "${PR_TITLE:?PR_TITLE is required}"
: "${PR_BODY:?PR_BODY is required}"
: "${ADD_PATHS:?ADD_PATHS is required}"

BASE_BRANCH="${BASE_BRANCH:-main}"
DRY_RUN="${DRY_RUN:-0}"
DRAFT="${DRAFT:-false}"
GIT_USER_NAME="${GIT_USER_NAME:-github-actions[bot]}"
GIT_USER_EMAIL="${GIT_USER_EMAIL:-41898282+github-actions[bot]@users.noreply.github.com}"

# ---------------------------------------------------------------------------
# Validate branch names
# ---------------------------------------------------------------------------
if [ "${PR_BRANCH}" = "${BASE_BRANCH}" ]; then
  echo "Error: branch ('${PR_BRANCH}') and base-branch ('${BASE_BRANCH}') must be different" >&2
  exit 1
fi

if [[ "${PR_BRANCH}" == -* ]]; then
  echo "Error: branch name must not start with '-'" >&2
  exit 1
fi

# ---------------------------------------------------------------------------
# Mask the token so it never appears in logs (covers App tokens too)
# ---------------------------------------------------------------------------
if [ -n "${GH_TOKEN:-}" ]; then
  echo "::add-mask::${GH_TOKEN}"
fi

# ---------------------------------------------------------------------------
# Cleanup trap: restore git state on any exit (success or failure)
# ---------------------------------------------------------------------------
ORIGINAL_REMOTE_URL=""
ORIGINAL_BRANCH=""
ORIGINAL_GIT_USER_NAME=""
ORIGINAL_GIT_USER_EMAIL=""

cleanup() {
  # Restore remote URL (remove embedded token)
  if [ -n "${ORIGINAL_REMOTE_URL}" ]; then
    git remote set-url origin "${ORIGINAL_REMOTE_URL}" 2>/dev/null || true
  fi
  # Restore branch
  if [ -n "${ORIGINAL_BRANCH}" ] && [ "${ORIGINAL_BRANCH}" != "HEAD" ]; then
    git checkout "${ORIGINAL_BRANCH}" 2>/dev/null || true
  fi
  # Restore git identity
  if [ -n "${ORIGINAL_GIT_USER_NAME}" ]; then
    git config user.name "${ORIGINAL_GIT_USER_NAME}" 2>/dev/null || true
  fi
  if [ -n "${ORIGINAL_GIT_USER_EMAIL}" ]; then
    git config user.email "${ORIGINAL_GIT_USER_EMAIL}" 2>/dev/null || true
  fi
}
trap cleanup EXIT

# ---------------------------------------------------------------------------
# Helper: split a comma/newline/space-separated string into lines
# Used for add-paths. Note: paths containing spaces are not supported.
# ---------------------------------------------------------------------------
split_paths() {
  local input="$1"
  local normalized
  normalized=$(echo "${input}" | tr ',\n' '  ')
  # shellcheck disable=SC2086
  for item in ${normalized}; do
    item=$(echo "${item}" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
    [ -n "${item}" ] && echo "${item}"
  done
}

# ---------------------------------------------------------------------------
# Helper: split a comma/newline-separated string into lines
# Used for labels and reviewers (preserves spaces within items).
# ---------------------------------------------------------------------------
split_csv() {
  local input="$1"
  local normalized
  normalized=$(echo "${input}" | tr '\n' ',')
  IFS=',' read -r -a items <<< "${normalized}"
  for item in "${items[@]}"; do
    item=$(echo "${item}" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
    [ -n "${item}" ] && echo "${item}"
  done
}

# ---------------------------------------------------------------------------
# Parse add-paths into an array
# ---------------------------------------------------------------------------
mapfile -t PATH_ARRAY < <(split_paths "${ADD_PATHS}")

if [ "${#PATH_ARRAY[@]}" -eq 0 ]; then
  echo "Error: ADD_PATHS resolved to an empty list" >&2
  exit 1
fi

# ---------------------------------------------------------------------------
# Stage files and check for changes
# ---------------------------------------------------------------------------
for p in "${PATH_ARRAY[@]}"; do
  git add -- "${p}"
done

has_changes=false
for p in "${PATH_ARRAY[@]}"; do
  if ! git --no-pager diff --cached --quiet -- "${p}" 2>/dev/null; then
    has_changes=true
    break
  fi
done

if [ "${has_changes}" = "false" ]; then
  echo "No changes detected in: ${ADD_PATHS}"
  echo "Nothing to do."
  if [ -n "${GITHUB_OUTPUT:-}" ]; then
    echo "pull-request-operation=none" >> "${GITHUB_OUTPUT}"
  fi
  exit 0
fi

echo "Changes detected:"
for p in "${PATH_ARRAY[@]}"; do
  git --no-pager diff --cached --stat -- "${p}"
done

# ---------------------------------------------------------------------------
# Dry-run mode
# ---------------------------------------------------------------------------
if [ "${DRY_RUN}" = "1" ]; then
  echo ""
  echo "[dry-run] Would create branch: ${PR_BRANCH}"
  echo "[dry-run] Would commit as: ${GIT_USER_NAME} <${GIT_USER_EMAIL}>"
  echo "[dry-run] Would commit with message: ${COMMIT_MSG}"
  echo "[dry-run] Would push to origin/${PR_BRANCH}"
  echo "[dry-run] Would create PR: ${PR_TITLE} -> ${BASE_BRANCH}"
  [ -n "${LABELS:-}" ]    && echo "[dry-run] Labels: ${LABELS}"
  [ -n "${REVIEWERS:-}" ] && echo "[dry-run] Reviewers: ${REVIEWERS}"
  [ "${DRAFT}" = "true" ] && echo "[dry-run] Draft: true"
  echo "[dry-run] PR body:"
  echo "${PR_BODY}"
  # Unstage so dry-run is side-effect free
  for p in "${PATH_ARRAY[@]}"; do
    git reset HEAD -- "${p}" >/dev/null 2>&1 || true
  done
  exit 0
fi

# ---------------------------------------------------------------------------
# Configure git identity (save originals for cleanup trap)
# ---------------------------------------------------------------------------
ORIGINAL_GIT_USER_NAME=$(git config user.name 2>/dev/null || echo "")
ORIGINAL_GIT_USER_EMAIL=$(git config user.email 2>/dev/null || echo "")

git config user.name "${GIT_USER_NAME}"
git config user.email "${GIT_USER_EMAIL}"

# ---------------------------------------------------------------------------
# Configure push credentials from GH_TOKEN
# ---------------------------------------------------------------------------
# actions/checkout with persist-credentials: false does not store credentials,
# so we configure them here using GH_TOKEN for the push step only.
# The original URL is saved for the cleanup trap to restore.
if [ -n "${GH_TOKEN:-}" ] && [ -n "${GITHUB_REPOSITORY:-}" ]; then
  ORIGINAL_REMOTE_URL=$(git remote get-url origin)
  git remote set-url origin "https://x-access-token:${GH_TOKEN}@github.com/${GITHUB_REPOSITORY}.git"
fi

# ---------------------------------------------------------------------------
# Save current branch for the cleanup trap to restore
# ---------------------------------------------------------------------------
ORIGINAL_BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "")

# ---------------------------------------------------------------------------
# Create branch, commit, and push
# ---------------------------------------------------------------------------
git checkout -B "${PR_BRANCH}"

git commit -m "${COMMIT_MSG}"

# Use explicit refspec to prevent branch name from being interpreted as a flag
git push --force origin "refs/heads/${PR_BRANCH}:refs/heads/${PR_BRANCH}"

# ---------------------------------------------------------------------------
# Build gh CLI flags for labels, reviewers, draft
# ---------------------------------------------------------------------------
CREATE_FLAGS=()
EDIT_FLAGS=()

if [ -n "${LABELS:-}" ]; then
  mapfile -t LABEL_ARRAY < <(split_csv "${LABELS}")
  for label in "${LABEL_ARRAY[@]}"; do
    CREATE_FLAGS+=("--label" "${label}")
    EDIT_FLAGS+=("--add-label" "${label}")
  done
fi

if [ -n "${REVIEWERS:-}" ]; then
  mapfile -t REVIEWER_ARRAY < <(split_csv "${REVIEWERS}")
  for reviewer in "${REVIEWER_ARRAY[@]}"; do
    CREATE_FLAGS+=("--reviewer" "${reviewer}")
    EDIT_FLAGS+=("--add-reviewer" "${reviewer}")
  done
fi

if [ "${DRAFT}" = "true" ]; then
  CREATE_FLAGS+=("--draft")
fi

# ---------------------------------------------------------------------------
# Create or update the pull request
# ---------------------------------------------------------------------------
PR_NUM=$(gh pr list --head "${PR_BRANCH}" --base "${BASE_BRANCH}" --state open --json number --jq '.[0].number // empty')

if [ -n "${PR_NUM}" ]; then
  echo "PR #${PR_NUM} already exists for branch ${PR_BRANCH}. Push updated the branch."

  if ! gh pr edit "${PR_NUM}" \
    --title "${PR_TITLE}" \
    --body "${PR_BODY}" \
    "${EDIT_FLAGS[@]+"${EDIT_FLAGS[@]}"}"; then
    echo "Warning: gh pr edit failed (possibly due to invalid reviewers/labels). PR was still updated via push." >&2
  fi

  PR_URL=$(gh pr view "${PR_NUM}" --json url --jq '.url')
  echo "Updated PR #${PR_NUM}: ${PR_URL}"
  OPERATION="updated"
else
  # gh pr create outputs the PR URL to stdout; grab only the last line
  # in case gh prints warnings before the URL.
  PR_URL=$(gh pr create \
    --head "${PR_BRANCH}" \
    --base "${BASE_BRANCH}" \
    --title "${PR_TITLE}" \
    --body "${PR_BODY}" \
    "${CREATE_FLAGS[@]+"${CREATE_FLAGS[@]}"}" | tail -1)

  # Extract PR number from URL (always ends with /pull/<number>)
  PR_NUM="${PR_URL##*/}"
  echo "Pull request #${PR_NUM} created: ${PR_URL}"
  OPERATION="created"
fi

# ---------------------------------------------------------------------------
# Set outputs
# ---------------------------------------------------------------------------
if [ -n "${GITHUB_OUTPUT:-}" ]; then
  {
    echo "pull-request-number=${PR_NUM}"
    echo "pull-request-url<<GHEOF"
    echo "${PR_URL}"
    echo "GHEOF"
    echo "pull-request-operation=${OPERATION}"
  } >> "${GITHUB_OUTPUT}"
fi

echo "Done. Operation: ${OPERATION}"
