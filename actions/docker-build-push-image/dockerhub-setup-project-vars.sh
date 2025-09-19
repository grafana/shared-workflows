#!/usr/bin/env bash
set -euo pipefail

# Required env vars (provided in workflow yaml)
: "${DOCKERHUB_IMAGE:?DOCKERHUB_IMAGE env var is required}"

echo "image=${DOCKERHUB_REGISTRY}/${DOCKERHUB_IMAGE}" | tee -a "${GITHUB_OUTPUT}"
