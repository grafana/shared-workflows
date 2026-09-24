#!/usr/bin/env bats
#
# Unit tests for run-shellcheck.sh
#
# Hermetic: operates only on the committed test fixtures and a temp directory.
# Requires `shellcheck` on PATH (pre-installed on GitHub-hosted runners).

setup() {
	SCRIPT="${BATS_TEST_DIRNAME}/run-shellcheck.sh"
	FIXTURES="${BATS_TEST_DIRNAME}/test-fixtures/scripts"
	TMP_DIR="$(mktemp -d)"
	export GITHUB_OUTPUT="${TMP_DIR}/output.txt"
	: >"${GITHUB_OUTPUT}"
}

teardown() {
	rm -rf "${TMP_DIR}"
}

@test "passes on clean scripts and records the checked files" {
	SCANDIR="${FIXTURES}/good" run "${SCRIPT}"
	[ "$status" -eq 0 ]
	[[ "$output" == *"passed shellcheck"* ]]
	grep -q "clean.sh" "${GITHUB_OUTPUT}"
}

@test "detects extensionless executables via their shebang" {
	# clean-no-ext has no extension; it is only found by shebang detection.
	SCANDIR="${FIXTURES}/good" run "${SCRIPT}"
	[ "$status" -eq 0 ]
	grep -q "clean-no-ext" "${GITHUB_OUTPUT}"
}

@test "fails when a script has shellcheck issues" {
	SCANDIR="${FIXTURES}/bad" run "${SCRIPT}"
	[ "$status" -ne 0 ]
	[[ "$output" == *"shellcheck found issues"* ]]
}

@test "ignore-paths excludes directories" {
	SCANDIR="${FIXTURES}" IGNORE_PATHS="bad ignored" run "${SCRIPT}"
	[ "$status" -eq 0 ]
	run grep -q "/bad/" "${GITHUB_OUTPUT}"
	[ "$status" -ne 0 ]
	run grep -q "/ignored/" "${GITHUB_OUTPUT}"
	[ "$status" -ne 0 ]
}

@test "ignore-names excludes files by name" {
	SCANDIR="${FIXTURES}/bad" IGNORE_NAMES="issues.sh" run "${SCRIPT}"
	[ "$status" -eq 0 ]
	[[ "$output" == *"No shell scripts found"* ]]
}

@test "severity=error still catches error-level findings" {
	# bad/issues.sh contains SC2068 (error) and SC2086 (info).
	SCANDIR="${FIXTURES}/bad" SEVERITY="error" run "${SCRIPT}"
	[ "$status" -ne 0 ]
}

@test "check-together runs all files in a single invocation" {
	SCANDIR="${FIXTURES}/good" CHECK_TOGETHER="true" run "${SCRIPT}"
	[ "$status" -eq 0 ]
}

@test "exits cleanly when no shell scripts are found" {
	mkdir -p "${TMP_DIR}/empty"
	SCANDIR="${TMP_DIR}/empty" run "${SCRIPT}"
	[ "$status" -eq 0 ]
	[[ "$output" == *"No shell scripts found"* ]]
}
