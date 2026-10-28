#!/usr/bin/env bash
set -euo pipefail

# Runs summarize.sh against the fixtures in testdata/ and checks the outputs,
# the job summary and the Slack payload.

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORK="$(mktemp -d)"
trap 'rm -rf "${WORK}"' EXIT

failures=0

fail() {
	echo "FAIL: $*"
	failures=$((failures + 1))
}

# run <fixture> <slack channel>: runs summarize.sh and leaves the results in
# ${WORK}/output and ${WORK}/summary.
run() {
	: >"${WORK}/output"
	: >"${WORK}/summary"
	GOVULNCHECK_JSON="${DIR}/testdata/$1.jsonl" \
		GITHUB_OUTPUT="${WORK}/output" \
		GITHUB_STEP_SUMMARY="${WORK}/summary" \
		GITHUB_REPOSITORY="grafana/example" \
		GITHUB_SERVER_URL="https://github.com" \
		GITHUB_RUN_ID="42" \
		GITHUB_REF_NAME="main" \
		SLACK_CHANNEL="$2" \
		"${DIR}/summarize.sh" >/dev/null
}

expect_output() {
	grep -qxF "$1" "${WORK}/output" || fail "$CASE: expected output '$1', got: $(grep -v '^slack-payload=' "${WORK}/output" | tr '\n' ' ')"
}

expect_summary() {
	grep -qF -- "$1" "${WORK}/summary" || fail "$CASE: expected summary to contain '$1'"
}

reject_summary() {
	! grep -qF -- "$1" "${WORK}/summary" || fail "$CASE: expected summary not to contain '$1'"
}

CASE="called vulnerability"
run called C0123456789
expect_output "vulnerabilities-found=1"
expect_output "has-vulnerabilities=true"
expect_summary "**1 vulnerability found**"
expect_summary "[GO-2021-0113](https://pkg.go.dev/vuln/GO-2021-0113)"
expect_summary "golang.org/x/text/language.Parse"
expect_summary "<details>"
payload=$(sed -n 's/^slack-payload=//p' "${WORK}/output")
[[ "$(jq -r '.channel' <<<"${payload}")" == "C0123456789" ]] || fail "$CASE: wrong Slack channel"
jq -e '.blocks[0].text.text | contains("GO-2021-0113")' <<<"${payload}" >/dev/null || fail "$CASE: Slack message doesn't mention the vulnerability"
jq -e '.blocks[1].elements[0].text | contains("https://github.com/grafana/example/actions/runs/42")' <<<"${payload}" >/dev/null || fail "$CASE: Slack message doesn't link to the run"

CASE="called vulnerability without Slack channel"
run called ""
expect_output "has-vulnerabilities=true"
! grep -q '^slack-payload=' "${WORK}/output" || fail "$CASE: unexpected Slack payload"

CASE="imported vulnerabilities only"
run imported-only C0123456789
expect_output "vulnerabilities-found=0"
expect_output "has-vulnerabilities=false"
expect_summary "No vulnerabilities found"
expect_summary "<details>"
reject_summary "**"
! grep -q '^slack-payload=' "${WORK}/output" || fail "$CASE: unexpected Slack payload"

CASE="clean"
run clean C0123456789
expect_output "vulnerabilities-found=0"
expect_output "has-vulnerabilities=false"
expect_summary "No vulnerabilities found"
reject_summary "<details>"
! grep -q '^slack-payload=' "${WORK}/output" || fail "$CASE: unexpected Slack payload"

CASE="missing input"
if GOVULNCHECK_JSON="${WORK}/does-not-exist.json" GITHUB_OUTPUT="${WORK}/output" GITHUB_STEP_SUMMARY="${WORK}/summary" "${DIR}/summarize.sh" >/dev/null 2>&1; then
	fail "$CASE: expected a failure"
fi

if [[ ${failures} -gt 0 ]]; then
	echo "${failures} check(s) failed"
	exit 1
fi

echo "All checks passed"
