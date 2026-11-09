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

# run <fixture name or path> [VAR=value ...]: runs report.sh and leaves the
# Markdown in ${WORK}/comment and the outputs in ${WORK}/output.
run() {
	local fixture="$1"
	shift
	: >"${WORK}/output"
	[[ "${fixture}" == */* ]] || fixture="${DIR}/testdata/report/${fixture}.json"
	cp "${fixture}" "${WORK}/report.json"
	env -i PATH="${PATH}" \
		REPORT_PATH="${WORK}/report.json" \
		GITHUB_OUTPUT="${WORK}/output" \
		GITHUB_SERVER_URL="https://github.com" \
		GITHUB_REPOSITORY="grafana/example" \
		GITHUB_RUN_ID="42" \
		IMAGE_REF="example:pr-0123456789abcdef0123456789abcdef01234567" \
		IMAGE_NAME="example" \
		SLUG="example" \
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

CASE="counts and layout"
run current
expect_output "critical=1"
expect_output "high=3"
expect_output "medium=1"
expect_output "low=0"
expect_output "total=5"
expect_output "fixable=3"
expect_comment "### Trivy image scan: <code>example</code>"
expect_comment "🔴 **1** Critical · 🟠 **3** High · 🟡 **1** Medium"
refute_comment "Low"
expect_comment "<details open><summary>Findings (5)</summary>"
expect_comment "<code>example:pr-01234567…</code>"
expect_comment "Trivy v0.0.0"
expect_comment "[workflow run](https://github.com/grafana/example/actions/runs/42)"
expect_comment "Informational only; this check never fails the build."
expect_comment "<!-- trivy-image-scan:example -->"
expect_stderr "image=example:pr-0123456789abcdef0123456789abcdef01234567 os=alpine 3.19.9 trivy=v0.0.0 sbom=generated total=5"

CASE="findings are sorted by severity"
run current
[[ "$(grep -m1 '^| 🔴\|^| 🟠\|^| 🟡\|^| 🟢' "${WORK}/comment")" == *CRITICAL* ]] || fail "$CASE: first row should be CRITICAL"

CASE="duplicate findings across result types are counted once"
run current
[[ "$(grep -c 'CVE-2026-40200' "${WORK}/comment")" -eq 2 ]] || fail "$CASE: expected musl and musl-utils rows only"

CASE="untrusted fields are escaped"
run current
expect_comment '<code>evil&#96;pkg\|x)</code>'
expect_comment '<code>&lt;script&gt;1&lt;/script&gt;</code>'
expect_comment '<code>2&#95;&#42;&#95;</code>'
refute_comment 'javascript:'
expect_comment '| CVE-2024-0003 |'
expect_comment '(https://example.com/a%28b%29c)'

CASE="ignore-unfixed and cache hit are reflected"
run current IGNORE_UNFIXED=true SBOM_CACHE_HIT=true
expect_comment "High · 🟡 **1** Medium (fixed only)"
expect_stderr "sbom=cache"

CASE="max-rows truncates and invalid max-rows falls back"
run current MAX_ROWS=2
expect_comment "and 3 more; see the [workflow run]"
run current MAX_ROWS=abc
expect_stderr "max-rows must be a non-negative integer"
refute_comment "and 3 more"

CASE="long tables are collapsed"
run current MAX_ROWS=100
expect_comment "<details open><summary>Findings (5)</summary>"
# Fabricate a report with more rows than the open threshold.
jq '.Results[0].Vulnerabilities = [range(25) | {VulnerabilityID: "CVE-2024-\(10000 + .)", PkgName: "pkg\(.)", InstalledVersion: "1", Severity: "HIGH"}] | .Results |= .[:1]' \
	"${DIR}/testdata/report/current.json" >"${WORK}/many.json"
run "${WORK}/many.json"
expect_output "total=25"
expect_comment "<details><summary>Findings (25)</summary>"

CASE="requested severities always appear in the count line"
run empty SEVERITY="CRITICAL,HIGH,MEDIUM,LOW"
expect_comment "🔴 **0** Critical · 🟠 **0** High · 🟡 **0** Medium · 🟢 **0** Low"
expect_output "total=0"
expect_comment "<details open><summary>Findings (0)</summary>"
expect_comment "_none_"

CASE="no workflow run link outside Actions"
run current GITHUB_RUN_ID=
refute_comment "[workflow run]"

if [[ "${failures}" -gt 0 ]]; then
	echo "${failures} failure(s)"
	exit 1
fi
echo "All report.sh checks passed"
