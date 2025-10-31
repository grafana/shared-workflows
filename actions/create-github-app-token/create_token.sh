#!/bin/bash
set -euo pipefail

MAX_ATTEMPTS=3
IFS=',' read -ra APPS <<< "${GITHUB_APP}"
setRandomApp() {
    GITHUB_APP=$(printf "%s\n" "${APPS[@]}" | sed 's/^ *//;s/ *$//' | shuf -n1)
    echo "Randomly selected GitHub App: ${GITHUB_APP}"
}

TEMP_FILE=$(mktemp)
echo "Using temporary file: ${TEMP_FILE}"
trap 'rm -f "${TEMP_FILE}"' EXIT

for attempt in $(seq 1 "${MAX_ATTEMPTS}"); do
    echo "Attempt ${attempt} to get GitHub token..."
    setRandomApp
    RESPONSE=$(curl -sS -w "%{http_code}" -o "${TEMP_FILE}" \
        "${VAULT_URL}/v1/github-app-${GITHUB_APP}/token/${REPOSITORY_NAME}-${REF_SHA}-${PERMISSION_SET}" \
        -H "X-Vault-Token: ${VAULT_TOKEN}" \
        -H "Proxy-Authorization-Token: Bearer ${GITHUB_JWT_PROXY}" || true)

    if [[ "${RESPONSE}" -eq 200 ]]; then
        TOKEN=$(jq -r '.data.token' "${TEMP_FILE}")
        echo "github_token=${TOKEN}" >> "${GITHUB_OUTPUT}"
        echo "Create GitHub Token done!"
        exit 0
    else
        echo "Vault request failed (HTTP ${RESPONSE})"
        cat "${TEMP_FILE}" || true
        sleep $((attempt * 5))
    fi
done

echo "Failed to retrieve GitHub token after ${MAX_ATTEMPTS} attempts"
exit 1
