#!/bin/bash
set -euo pipefail

MAX_ATTEMPTS=3
for attempt in $(seq 1 "${MAX_ATTEMPTS}"); do
    echo "Attempt ${attempt} to authenticate with Vault..."

    RESPONSE=$(curl -sS -w "%{http_code}" -o response.json \
        -X POST "${VAULT_URL}/v1/auth/github-actions-oidc/login" \
        -H "Content-Type: application/json" \
        -H "Proxy-Authorization-Token: Bearer ${GITHUB_JWT_PROXY}" \
        -d "{
            \"role\": \"${REPOSITORY_NAME}-${GITHUB_APP}-${REF_SHA}-${PERMISSION_SET}\",
            \"jwt\": \"${GITHUB_JWT_VAULT}\"
        }" || true)

    if [[ "${RESPONSE}" -eq 200 ]]; then
        TOKEN=$(jq -r '.auth.client_token' response.json)
        echo "::add-mask::$TOKEN"
        echo "vault_token=${TOKEN}" >> "${GITHUB_OUTPUT}"
        echo "Vault auth done!"
        exit 0
    else
        echo "Vault auth failed (HTTP ${RESPONSE})"
        cat response.json || true
        sleep $((attempt * 5))
    fi
done

echo "Failed to authenticate with Vault after ${MAX_ATTEMPTS} attempts"
exit 1
