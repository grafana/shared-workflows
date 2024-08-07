#! /usr/bin/env bash
set -euo pipefail

# Generate Go client (TODO: Add support for other languages)
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

