#!/usr/bin/env bash
set -euo pipefail

JS_DIR="${OUTPUT_DIR}/js/${PACKAGE_NAME}"
rm -rf "${JS_DIR}"
java -jar openapi-generator-cli.jar generate \
  -i "${SPEC_PATH}" \
  -g typescript-fetch \
  -o "${JS_DIR}" \
  --git-user-id "grafana" \
  --git-repo-id "${REPO_NAME}" \
  --additional-properties="npmName=${REPO_NAME}-${PACKAGE_NAME},npmVersion=0.0.0,supportsES6=true,hideGenerationTimestamp=true,disallowAdditionalPropertiesIfNotPresent=false,withoutRuntimeChecks=true,modelPropertyNaming=original" \
  -t "${GITHUB_ACTION_PATH}/templates/typescript-fetch"

# There are no runtime dependencies. Pin the compiler in the package template
# and commit its lockfile alongside the source; node_modules and dist are ignored.
pushd "${JS_DIR}"
npm install --ignore-scripts --no-audit --no-fund
npm run build
popd
