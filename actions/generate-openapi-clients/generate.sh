#! /usr/bin/env bash
set -euo pipefail

# Download openapi-generator-cli
wget -nv "https://repo1.maven.org/maven2/org/openapitools/openapi-generator-cli/${OPENAPI_GENERATOR_VERSION}/openapi-generator-cli-${OPENAPI_GENERATOR_VERSION}.jar" -O ./openapi-generator-cli.jar
trap "rm -f ./openapi-generator-cli.jar" EXIT

# Generate Go client (TODO: Add support for other languages)
GO_DIR="${OUTPUT_DIR}/go"
rm -rf "${GO_DIR}"
java -jar openapi-generator-cli.jar generate \
  -i "${SPEC_PATH}" \
  -g go \
  -o "${GO_DIR}" \
  --git-user-id "grafana" \
  --git-repo-id "${REPO_NAME}/go" \
  --package-name "${PACKAGE_NAME}" \
  -p disallowAdditionalPropertiesIfNotPresent=false \
  -t "${GITHUB_ACTION_PATH}/templates/go"

pushd "${GO_DIR}" && go mod tidy && popd
if ! command -v goimports &> /dev/null
then
    go install golang.org/x/tools/cmd/goimports@latest
fi
find "${GO_DIR}" -name \*.go -exec goimports -w {} \;

