#!/usr/bin/env bash
set -euo pipefail

: "${IMAGE:?IMAGE env var is required}"

validate_image() {
  local image="$1"

  # Has @sha256:... digest? Accept.
  if [[ "$image" == *"@sha256:"* ]]; then
    return 0
  fi

  # For refs with a registry path like 'localhost:5000/repo:tag', strip up to
  # the last '/' so we look for the tag colon, not the registry port. Bare
  # refs without a slash (e.g. 'localhost:5000') are left as-is and treated
  # as name:tag — matching Docker CLI's own ref parser, which is intentionally
  # lenient at this layer; bad refs surface as docker errors during polling.
  local after_slash="${image##*/}"

  # Has ':tag' after the last slash? Accept.
  if [[ "$after_slash" == *":"* ]]; then
    return 0
  fi

  return 1
}

if ! validate_image "$IMAGE"; then
  echo "::error::IMAGE must include a tag (repo:tag) or digest (repo@sha256:...); got '${IMAGE}'"
  exit 1
fi

# Loop not yet implemented.
echo "::error::wait-for-docker-publish: polling loop not yet implemented"
exit 1
