#!/usr/bin/env bash
set -euo pipefail

# Optional env vars
: "${REPO_NAME:=}"   # default, empty

# Required env vars (provided in workflow yaml)
: "${ENVIRONMENT:?ENVIRONMENT env var is required}"
: "${REGISTRY:?REGISTRY env var is required}"
: "${IMAGE_NAME:?IMAGE_NAME env var is required}"
: "${GH_REPO:?GH_REPO env var is required}"

# Outputs
gar_repo_name="${REPO_NAME}"
gar_project=""
gar_image=""

########################################
# Resolve repo_name
########################################
if [ -z "$REPO_NAME" ]; then
  gar_repo_name="$(echo "${GH_REPO}" | awk -F'/' '{print $2}')"
  gar_repo_name="${gar_repo_name//_/-}"
fi

########################################
# Resolve project
########################################
case "$ENVIRONMENT" in
  dev)
    gar_project="grafanalabs-dev"
    ;;
  prod)
    gar_project="grafanalabs-global"
    ;;
  *)
    echo "❌ Invalid ENVIRONMENT: $ENVIRONMENT (must be 'dev' or 'prod')" >&2
    exit 1
    ;;
esac

########################################
# Build image path
########################################
gar_image="${REGISTRY}/${gar_project}/docker-${gar_repo_name}-${ENVIRONMENT}/${IMAGE_NAME}"

#export REPO_NAME=${gar_repo_name}"
#export GAR_PROJECT=${gar_project}"
#export GAR_IMAGE=${gar_image}"

echo "repo_name=${gar_repo_name}" | tee -a "${GITHUB_OUTPUT}"
echo "project=${gar_project}" | tee -a "${GITHUB_OUTPUT}"
echo "image=${gar_image}" | tee -a "${GITHUB_OUTPUT}"

