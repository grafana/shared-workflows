#!/usr/bin/env bash
set -euo pipefail

# Required env vars (provided in workflow yaml)
: "${DOCKERHUB_IMAGE:?IMAGE_NAME env var is required}"

echo "image=${DOCKERHUB_IMAGE}" | tee -a "${GITHUB_OUTPUT}"
