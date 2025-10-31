#!/bin/bash
set -euo pipefail

TEMP_FILE=$(mktemp)
echo "Using temporary file: ${TEMP_FILE}"
trap 'rm -f "${TEMP_FILE}"' EXIT

MAX_ATTEMPTS=3
for attempt in $(seq 1 "${MAX_ATTEMPTS}"); do
    echo "Attempt ${attempt} to authenticate with Vault..."

    RESPONSE=$(curl -sS -w "%{http_code}" -o "${TEMP_FILE}" \
        -X POST "${VAULT_URL}/v1/auth/github-actions-oidc/login" \
        -H "Content-Type: application/json" \
        -H "Proxy-Authorization-Token: Bearer ${GITHUB_JWT_PROXY}" \
        -d "{
            \"role\": \"${REPOSITORY_NAME}-${REF_SHA}-${PERMISSION_SET}\",
            \"jwt\": \"${GITHUB_JWT_VAULT}\"
        }" || true)

    if [[ "${RESPONSE}" -eq 200 ]]; then
        TOKEN=$(jq -r '.auth.client_token' "${TEMP_FILE}")
        echo "::add-mask::$TOKEN"
        echo "vault_token=${TOKEN}" >> "${GITHUB_OUTPUT}"
        echo "Vault auth done!"
        exit 0
    else
        echo "Vault auth failed (HTTP ${RESPONSE})"
        cat "${TEMP_FILE}" || true
        sleep $((attempt * 5))
    fi
done

echo "Failed to authenticate with Vault after ${MAX_ATTEMPTS} attempts"
exit 1
