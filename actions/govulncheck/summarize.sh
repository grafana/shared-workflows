#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Fail with a clear message if required env vars are missing
for var in GOVULNCHECK_JSON GITHUB_OUTPUT GITHUB_STEP_SUMMARY; do
	if [[ -z "${!var:-}" ]]; then
		echo "ERROR: Environment variable ${var} is required but not set"
		exit 1
	fi
done

if ! command -v jq >/dev/null; then
	echo "ERROR: jq is required but was not found in PATH"
	exit 1
fi

if [[ ! -s "${GOVULNCHECK_JSON}" ]]; then
	echo "ERROR: govulncheck output ${GOVULNCHECK_JSON} is missing or empty"
	exit 1
fi

# Maximum number of rows rendered per table, so the report stays well below the
# 1 MiB limit GitHub enforces on step summaries.
MAX_ROWS=50

# Condense the govulncheck message stream into one record per vulnerability.
MODEL=$(jq --slurp --from-file "${SCRIPT_DIR}/model.jq" "${GOVULNCHECK_JSON}")

CALLED=$(jq '[.vulns[] | select(.level == "called")] | length' <<<"${MODEL}")
OTHER=$(jq '[.vulns[] | select(.level != "called")] | length' <<<"${MODEL}")

{
	echo "vulnerabilities-found=${CALLED}"
	echo "has-vulnerabilities=$([[ "${CALLED}" -gt 0 ]] && echo true || echo false)"
} >>"${GITHUB_OUTPUT}"

# Render the markdown report.
jq --raw-output --from-file "${SCRIPT_DIR}/summary.jq" --argjson max "${MAX_ROWS}" <<<"${MODEL}" >>"${GITHUB_STEP_SUMMARY}"

# Build the Slack payload with jq so that nothing from the report is ever
# interpolated into JSON by hand.
if [[ -n "${SLACK_CHANNEL:-}" && "${CALLED}" -gt 0 ]]; then
	RUN_URL="${GITHUB_SERVER_URL:-https://github.com}/${GITHUB_REPOSITORY:-}/actions/runs/${GITHUB_RUN_ID:-}"
	PAYLOAD=$(jq --null-input --compact-output --from-file "${SCRIPT_DIR}/slack.jq" \
		--arg channel "${SLACK_CHANNEL}" \
		--arg repo "${GITHUB_REPOSITORY:-unknown repository}" \
		--arg ref "${GITHUB_REF_NAME:-}" \
		--arg url "${RUN_URL}" \
		--argjson model "${MODEL}")
	echo "slack-payload=${PAYLOAD}" >>"${GITHUB_OUTPUT}"
fi

echo "govulncheck: ${CALLED} called vulnerabilities, ${OTHER} imported or required but not called"
