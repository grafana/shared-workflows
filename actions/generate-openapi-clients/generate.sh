#! /usr/bin/env bash
set -euo pipefail

case "${CLIENT_LANGUAGE:-go}" in
  go) ;;
  javascript) exec "${GITHUB_ACTION_PATH}/generate-javascript.sh" ;;
  *) echo "Unsupported client language: ${CLIENT_LANGUAGE}. Expected go or javascript." >&2; exit 1 ;;
esac

# Generate Go client
GO_DIR="${OUTPUT_DIR}/go/${PACKAGE_NAME}"
rm -rf "${GO_DIR}"
java -jar openapi-generator-cli.jar generate \
  -i "${SPEC_PATH}" \
  -g go \
  -o "${GO_DIR}" \
  --git-user-id "grafana" \
  --git-repo-id "${REPO_NAME}/go" \
  --package-name "${PACKAGE_NAME}" \
  -p isGoSubmodule=true \
  -p disallowAdditionalPropertiesIfNotPresent=false \
  -t "${GITHUB_ACTION_PATH}/templates/go"

pushd "${GO_DIR}" && go mod tidy && popd
if ! command -v goimports &> /dev/null
then
    go install golang.org/x/tools/cmd/goimports@latest
fi
find "${GO_DIR}" -name \*.go -exec goimports -w {} \;
