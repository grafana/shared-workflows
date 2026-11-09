#!/usr/bin/env bash
set -euo pipefail

# Runs report.sh against the fixtures in testdata/report/ and checks the step
# outputs and the rendered Markdown, including the escaping of untrusted fields.

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORK="$(mktemp -d)"
trap 'rm -rf "${WORK}"' EXIT

failures=0

fail() {
	echo "FAIL: $*"
	failures=$((failures + 1))
}

# run <current fixture> <baseline fixture or ""> [VAR=value ...]: runs report.sh
# and leaves the Markdown in ${WORK}/comment and the outputs in ${WORK}/output.
run() {
	local current="$1" baseline="$2"
	shift 2
	: >"${WORK}/output"
	cp "${DIR}/testdata/report/${current}.json" "${WORK}/report.json"
	local baseline_path=""
	if [[ -n "${baseline}" ]]; then
		baseline_path="${DIR}/testdata/report/${baseline}.json"
	fi
	env -i PATH="${PATH}" \
		REPORT_PATH="${WORK}/report.json" \
		BASELINE_REPORT="${baseline_path}" \
		GITHUB_OUTPUT="${WORK}/output" \
		GITHUB_SERVER_URL="https://github.com" \
		GITHUB_REPOSITORY="grafana/example" \
		GITHUB_RUN_ID="42" \
		IMAGE_REF="example:pr-0123456789abcdef0123456789abcdef01234567" \
		IMAGE_NAME="example" \
		SLUG="example" \
		BASELINE_REF="main" \
		TRIVY_BIN_VERSION="v0.0.0" \
		"$@" \
		bash "${DIR}/report.sh" >"${WORK}/comment" 2>"${WORK}/stderr"
}

expect_output() {
	grep -qxF -- "$1" "${WORK}/output" || fail "$CASE: expected output '$1', got: $(tr '\n' ' ' <"${WORK}/output")"
}

expect_comment() {
	grep -qF -- "$1" "${WORK}/comment" || fail "$CASE: expected comment to contain '$1'"
}

refute_comment() {
	if grep -qF -- "$1" "${WORK}/comment"; then fail "$CASE: comment must not contain '$1'"; fi
}

expect_stderr() {
	grep -qF -- "$1" "${WORK}/stderr" || fail "$CASE: expected stderr to contain '$1'"
}

CASE="no baseline"
run current ""
expect_output "critical=1"
expect_output "high=3"
expect_output "medium=1"
expect_output "total=5"
expect_output "fixable=3"
expect_output "baseline-found=false"
expect_output "introduced="
expect_output "resolved="
expect_comment "### Trivy image scan: <code>example</code>"
expect_comment "**1** Critical"
expect_comment "**3** High"
expect_comment "**1** Medium"
expect_comment "> [!NOTE]"
expect_comment "No baseline for <code>main</code> yet"
expect_comment "<details><summary>All findings (5)</summary>"
expect_comment "<code>example:pr-01234567…</code>"
expect_comment "[workflow run](https://github.com/grafana/example/actions/runs/42)"
expect_comment "<!-- trivy-image-scan:example -->"
expect_stderr "os=alpine 3.19.9"

CASE="duplicate findings across result types are counted once"
run current ""
[[ "$(grep -c 'CVE-2026-40200' "${WORK}/comment")" -eq 2 ]] || fail "$CASE: expected musl and musl-utils rows only"

CASE="untrusted fields are escaped"
run current ""
expect_comment '<code>evil&#96;pkg\|x)</code>'
expect_comment '<code>&lt;script&gt;1&lt;/script&gt;</code>'
expect_comment '<code>2&#95;&#42;&#95;</code>'
refute_comment 'javascript:'
expect_comment '| CVE-2024-0003 |'
expect_comment '(https://example.com/a%28b%29c)'

CASE="introduced and resolved against a cached baseline"
run current baseline BASELINE_SOURCE=cache
expect_output "baseline-found=true"
expect_output "introduced=4"
expect_output "resolved=1"
expect_comment "> [!CAUTION]"
expect_comment "**4 vulnerabilities introduced** compared to <code>main</code>, 1 resolved."
expect_comment "<details open><summary>Introduced (4)</summary>"
expect_comment "<details><summary>Resolved (1)</summary>"
expect_comment "CVE-2023-9999"

CASE="baseline source labels"
run current baseline BASELINE_SOURCE=artifact
expect_comment "compared to <code>main</code> (baseline artifact)"
run current baseline BASELINE_SOURCE=image BASELINE_IMAGE_REF=ghcr.io/example/app:latest
expect_comment "compared to image <code>ghcr.io/example/app:latest</code>"

CASE="resolved only"
run empty baseline BASELINE_SOURCE=cache
expect_output "introduced=0"
expect_output "resolved=2"
expect_comment "> [!TIP]"
expect_comment "**2 vulnerabilities resolved** compared to <code>main</code>; none introduced."
expect_comment "<details><summary>Resolved (2)</summary>"
refute_comment "Introduced ("

CASE="identical baseline"
run current current BASELINE_SOURCE=cache
expect_output "introduced=0"
expect_output "resolved=0"
expect_comment "> [!TIP]"
expect_comment "No vulnerabilities introduced or resolved compared to <code>main</code>."

CASE="malformed baseline degrades to no baseline"
printf 'not json\n' >"${WORK}/bad.json"
run current "" BASELINE_REPORT="${WORK}/bad.json"
expect_output "baseline-found=false"
expect_stderr "Baseline report is unreadable"
expect_comment "> [!NOTE]"

CASE="max-rows truncates and invalid max-rows falls back"
run current "" MAX_ROWS=2
expect_comment "and 3 more; see the [workflow run]"
run current "" MAX_ROWS=abc
expect_stderr "max-rows must be a non-negative integer"
refute_comment "and 3 more"

CASE="requested severities always appear in the count line"
run empty "" SEVERITY="CRITICAL,HIGH,MEDIUM,LOW"
expect_comment "**0** Critical"
expect_comment "**0** Low"
expect_output "total=0"
expect_comment "<details><summary>All findings (0)</summary>"
expect_comment "_none_"

CASE="no workflow run link outside Actions"
run current "" GITHUB_RUN_ID=
refute_comment "[workflow run]"

if [[ "${failures}" -gt 0 ]]; then
	echo "${failures} failure(s)"
	exit 1
fi
echo "All report.sh checks passed"
