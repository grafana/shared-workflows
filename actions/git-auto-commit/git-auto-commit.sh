#!/usr/bin/env bash
#
# git-auto-commit.sh
#
# Commits changes in the working tree and pushes to the current branch.
# Designed as a lightweight replacement for stefanzweifel/git-auto-commit-action.
#
# Optional environment variables:
#   COMMIT_MESSAGE  - commit message (default: "Apply automatic changes")
#   BRANCH          - branch to push to (default: current HEAD)
#   FILE_PATTERN    - space-separated file patterns for git add (default: ".")
#   GIT_USER_NAME   - git user name (default: github-actions[bot])
#   GIT_USER_EMAIL  - git user email (default: 41898282+github-actions[bot]@users.noreply.github.com)
#   COMMIT_OPTIONS  - additional flags for git commit
#   PUSH_OPTIONS    - additional flags for git push
#   SKIP_PUSH       - set to "true" to skip pushing (default: false)
#   GH_TOKEN        - GitHub token for push authentication

set -euo pipefail

# ---------------------------------------------------------------------------
# Defaults
# ---------------------------------------------------------------------------
COMMIT_MESSAGE="${COMMIT_MESSAGE:-Apply automatic changes}"
BRANCH="${BRANCH:-}"
FILE_PATTERN="${FILE_PATTERN:-.}"
GIT_USER_NAME="${GIT_USER_NAME:-github-actions[bot]}"
GIT_USER_EMAIL="${GIT_USER_EMAIL:-41898282+github-actions[bot]@users.noreply.github.com}"
COMMIT_OPTIONS="${COMMIT_OPTIONS:-}"
PUSH_OPTIONS="${PUSH_OPTIONS:-}"
SKIP_PUSH="${SKIP_PUSH:-false}"

# ---------------------------------------------------------------------------
# Validate inputs
# ---------------------------------------------------------------------------
if [ -n "${BRANCH}" ] && [[ "${BRANCH}" == -* ]]; then
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

cleanup() {
  set +f 2>/dev/null || true  # re-enable globbing
  # Restore remote URL (remove embedded token)
  if [ -n "${ORIGINAL_REMOTE_URL}" ]; then
    git remote set-url origin "${ORIGINAL_REMOTE_URL}" 2>/dev/null || true
  fi
  # Restore branch
  if [ -n "${ORIGINAL_BRANCH}" ] && [ "${ORIGINAL_BRANCH}" != "HEAD" ]; then
    git checkout "${ORIGINAL_BRANCH}" 2>/dev/null || true
  fi
}
trap cleanup EXIT

# ---------------------------------------------------------------------------
# Disable globbing so file patterns in inputs aren't expanded by the shell
# ---------------------------------------------------------------------------
set -f

# ---------------------------------------------------------------------------
# Configure credentials from GH_TOKEN
# ---------------------------------------------------------------------------
# Must happen BEFORE branch fetch/checkout because actions/checkout with
# persist-credentials: false removes auth config, leaving git unable to
# fetch or push. We embed the token in the remote URL so that all subsequent
# git operations (fetch, push) can authenticate.
# The original URL is saved for the cleanup trap to restore.
if [ -n "${GH_TOKEN:-}" ] && [ -n "${GITHUB_REPOSITORY:-}" ]; then
  ORIGINAL_REMOTE_URL=$(git remote get-url origin)
  git remote set-url origin "https://x-access-token:${GH_TOKEN}@github.com/${GITHUB_REPOSITORY}.git"
fi

# ---------------------------------------------------------------------------
# Switch to target branch if specified and different from current HEAD
# (Must happen BEFORE staging so changes are staged on the correct branch)
# ---------------------------------------------------------------------------
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "")
ORIGINAL_BRANCH="${CURRENT_BRANCH}"

if [ -n "${BRANCH}" ] && [ "${BRANCH}" != "${CURRENT_BRANCH}" ]; then
  # On pull_request events, actions/checkout checks out the merge commit in
  # detached HEAD. The PR branch may not exist as a local ref, so we fetch
  # it from the remote first if needed.
  if ! git show-ref --verify --quiet "refs/heads/${BRANCH}"; then
    echo "Branch '${BRANCH}' not found locally, fetching from origin."
    git fetch --depth=1 origin "${BRANCH}"
  fi
  echo "Switching to branch '${BRANCH}'."
  git checkout "${BRANCH}"
fi

# ---------------------------------------------------------------------------
# Stage files
# ---------------------------------------------------------------------------
# shellcheck disable=SC2086
read -r -a FILE_PATTERN_ARRAY <<< "${FILE_PATTERN}"

for pattern in "${FILE_PATTERN_ARRAY[@]}"; do
  git add -- "${pattern}"
done

# ---------------------------------------------------------------------------
# Dirty check: are there staged changes?
# ---------------------------------------------------------------------------
# Uses git diff --staged which correctly ignores CRLF-only differences
# (unlike git status -s). This avoids empty commits from line-ending changes.
if git diff --staged --quiet; then
  echo "Working tree clean. Nothing to commit."
  if [ -n "${GITHUB_OUTPUT:-}" ]; then
    echo "changes-detected=false" >> "${GITHUB_OUTPUT}"
  fi
  exit 0
fi

echo "Changes detected."

if [ -n "${GITHUB_OUTPUT:-}" ]; then
  echo "changes-detected=true" >> "${GITHUB_OUTPUT}"
fi

# ---------------------------------------------------------------------------
# Commit
# ---------------------------------------------------------------------------
# Use git -c to set identity on the commit command only (does not modify repo config).
# shellcheck disable=SC2086
git -c user.name="${GIT_USER_NAME}" -c user.email="${GIT_USER_EMAIL}" \
  commit -m "${COMMIT_MESSAGE}" ${COMMIT_OPTIONS}

COMMIT_HASH=$(git rev-parse HEAD)
echo "Committed: ${COMMIT_HASH}"

if [ -n "${GITHUB_OUTPUT:-}" ]; then
  echo "commit-hash=${COMMIT_HASH}" >> "${GITHUB_OUTPUT}"
fi

# ---------------------------------------------------------------------------
# Push
# ---------------------------------------------------------------------------
if [ "${SKIP_PUSH}" = "true" ]; then
  echo "Skipping push (skip-push is true)."
  exit 0
fi

# Build push options array
PUSH_OPTIONS_ARRAY=()
if [ -n "${PUSH_OPTIONS}" ]; then
  # shellcheck disable=SC2206
  PUSH_OPTIONS_ARRAY=( ${PUSH_OPTIONS} )
fi

# Determine push target
if [ -n "${BRANCH}" ]; then
  echo "Pushing to branch '${BRANCH}'."
  git push origin "HEAD:refs/heads/${BRANCH}" --follow-tags --atomic \
    "${PUSH_OPTIONS_ARRAY[@]+"${PUSH_OPTIONS_ARRAY[@]}"}"
else
  echo "Pushing to current branch."
  git push origin \
    "${PUSH_OPTIONS_ARRAY[@]+"${PUSH_OPTIONS_ARRAY[@]}"}"
fi

echo "Done."
