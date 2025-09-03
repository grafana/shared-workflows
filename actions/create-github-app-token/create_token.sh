#!/bin/bash
set -euo pipefail

MAX_ATTEMPTS=3

for attempt in $(seq 1 "${MAX_ATTEMPTS}"); do
    echo "Attempt ${attempt} to get GitHub token..."

    RESPONSE=$(curl -sS -w "%{http_code}" -o response.json \
        "${VAULT_URL}/v1/github-app-${GITHUB_APP}/token/${REPOSITORY_NAME}-${REF_SHA}-${PERMISSION_SET}" \
        -H "X-Vault-Token: ${VAULT_TOKEN}" \
        -H "Proxy-Authorization-Token: Bearer ${GITHUB_JWT}" || true)

    if [[ "${RESPONSE}" -eq 200 ]]; then
        TOKEN=$(jq -r '.data.token' response.json)
        echo "github_token=${TOKEN}" >> "${GITHUB_OUTPUT}"
        exit 0
    else
        echo "Vault request failed (HTTP ${RESPONSE})"
        cat response.json || true
        sleep $((attempt * 5))
    fi
done

echo "Failed to retrieve GitHub token after ${MAX_ATTEMPTS} attempts"
exit 1
