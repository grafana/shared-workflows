#!/usr/bin/env bats
#
# Unit tests for create-or-update-pr.sh
#
# Hermetic: runs against a throwaway git repository in a temp directory with
# DRY_RUN=1, so it never pushes or talks to the GitHub API.

setup() {
	SCRIPT="${BATS_TEST_DIRNAME}/create-or-update-pr.sh"
	TMP_DIR="$(mktemp -d)"
	# Keep the repo in a subdirectory so GITHUB_OUTPUT lives outside the working
	# tree and is never staged.
	mkdir -p "${TMP_DIR}/repo"
	cd "${TMP_DIR}/repo" || exit 1
	git init -q
	git config user.email "init@example.com"
	git config user.name "init"
	echo "base" >tracked.txt
	git add tracked.txt
	git commit -q -m "initial commit"
	export GITHUB_OUTPUT="${TMP_DIR}/output.txt"
	: >"${GITHUB_OUTPUT}"
	export PR_BRANCH="updater/test"
	export COMMIT_MSG="test commit"
	export PR_TITLE="test title"
	export PR_BODY="test body"
	export ADD_PATHS="tracked.txt"
	export BASE_BRANCH="main"
	export DRY_RUN="1"
}

teardown() {
	cd / || true
	rm -rf "${TMP_DIR}"
}

@test "requires PR_BRANCH" {
	PR_BRANCH="" run "${SCRIPT}"
	[ "$status" -ne 0 ]
	[[ "$output" == *"PR_BRANCH is required"* ]]
}

@test "requires COMMIT_MSG" {
	COMMIT_MSG="" run "${SCRIPT}"
	[ "$status" -ne 0 ]
	[[ "$output" == *"COMMIT_MSG is required"* ]]
}

@test "requires PR_TITLE" {
	PR_TITLE="" run "${SCRIPT}"
	[ "$status" -ne 0 ]
	[[ "$output" == *"PR_TITLE is required"* ]]
}

@test "requires PR_BODY" {
	PR_BODY="" run "${SCRIPT}"
	[ "$status" -ne 0 ]
	[[ "$output" == *"PR_BODY is required"* ]]
}

@test "requires ADD_PATHS" {
	ADD_PATHS="" run "${SCRIPT}"
	[ "$status" -ne 0 ]
	[[ "$output" == *"ADD_PATHS is required"* ]]
}

@test "rejects a branch equal to the base-branch" {
	PR_BRANCH="main" run "${SCRIPT}"
	[ "$status" -ne 0 ]
	[[ "$output" == *"must be different"* ]]
}

@test "rejects a branch name starting with a dash" {
	PR_BRANCH="-bad" run "${SCRIPT}"
	[ "$status" -ne 0 ]
	[[ "$output" == *"must not start with '-'"* ]]
}

@test "rejects add-paths that resolve to an empty list" {
	ADD_PATHS="   " run "${SCRIPT}"
	[ "$status" -ne 0 ]
	[[ "$output" == *"empty list"* ]]
}

@test "reports operation=none when there are no changes" {
	run "${SCRIPT}"
	[ "$status" -eq 0 ]
	[[ "$output" == *"Nothing to do"* ]]
	grep -q "pull-request-operation=none" "${GITHUB_OUTPUT}"
}

@test "dry-run prints the planned actions and stages nothing" {
	echo "changed" >>tracked.txt
	LABELS="bug,enhancement" REVIEWERS="octocat" DRAFT="true" run "${SCRIPT}"
	[ "$status" -eq 0 ]
	[[ "$output" == *"[dry-run] Would create branch: updater/test"* ]]
	[[ "$output" == *"[dry-run] Labels: bug,enhancement"* ]]
	[[ "$output" == *"[dry-run] Reviewers: octocat"* ]]
	[[ "$output" == *"[dry-run] Draft: true"* ]]
	# Side-effect free: nothing is left staged.
	[ -z "$(git diff --cached --name-only)" ]
}

@test "splits add-paths on commas and spaces" {
	echo "x" >one.txt
	echo "y" >two.txt
	ADD_PATHS="one.txt, two.txt" run "${SCRIPT}"
	[ "$status" -eq 0 ]
	[[ "$output" == *"one.txt"* ]]
	[[ "$output" == *"two.txt"* ]]
}
