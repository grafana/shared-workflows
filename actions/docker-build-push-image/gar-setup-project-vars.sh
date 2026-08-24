#!/usr/bin/env bash
set -euo pipefail

# Optional env vars
: "${GAR_REPO:=}"   # default, empty
: "${GAR_IMAGE:=}"  # default, empty

# Required env vars (provided in workflow yaml)
: "${ENVIRONMENT:?ENVIRONMENT env var is required}"
: "${REGISTRY:?REGISTRY env var is required}"
: "${GH_REPO:?GH_REPO env var is required}"

gh_repo_name="$(echo "${GH_REPO}" | awk -F'/' '{print $2}')"  # ex: grafana/repo -> repo


########################################
# Resolve name of Google Artifact Repository
########################################
gar_repo_name="${GAR_REPO}"
if [ -z "${gar_repo_name}" ]; then
  gar_repo_name="docker-${gh_repo_name//_/-}-${ENVIRONMENT}"
fi
echo "gar_repo_name=${gar_repo_name}"

########################################
# Resolve Image Name
########################################
gar_image="${GAR_IMAGE}"
if [ -z "${gar_image}" ]; then
  gar_image="${gh_repo_name}"
fi
echo "gar_image=${gar_image}"

########################################
# Resolve project
########################################
case "${ENVIRONMENT}" in
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
echo "gar_project=${gar_project}"

########################################
# Build image path
########################################
gar_image="${REGISTRY}/${gar_project}/${gar_repo_name}/${gar_image}"

echo "image=${gar_image}" | tee -a "${GITHUB_OUTPUT}"
