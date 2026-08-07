#!/usr/bin/env bats
#
# Unit tests for git-auto-commit.sh
#
# Hermetic: runs against a throwaway git repository in a temp directory with
# SKIP_PUSH=true, so nothing is ever pushed to a remote.

setup() {
	SCRIPT="${BATS_TEST_DIRNAME}/git-auto-commit.sh"
	TMP_DIR="$(mktemp -d)"
	# Keep the repo in a subdirectory so GITHUB_OUTPUT lives outside the working
	# tree and is never picked up by a "." file pattern.
	mkdir -p "${TMP_DIR}/repo"
	cd "${TMP_DIR}/repo" || exit 1
	git init -q
	git config user.email "init@example.com"
	git config user.name "init"
	echo "base" >base.txt
	git add base.txt
	git commit -q -m "initial commit"
	export GITHUB_OUTPUT="${TMP_DIR}/output.txt"
	: >"${GITHUB_OUTPUT}"
	export SKIP_PUSH="true"
}

teardown() {
	cd / || true
	rm -rf "${TMP_DIR}"
}

@test "reports no changes on a clean working tree" {
	run "${SCRIPT}"
	[ "$status" -eq 0 ]
	grep -q "changes-detected=false" "${GITHUB_OUTPUT}"
}

@test "commits a new file and reports the commit hash" {
	echo "new" >new.txt
	COMMIT_MESSAGE="add new file" run "${SCRIPT}"
	[ "$status" -eq 0 ]
	grep -q "changes-detected=true" "${GITHUB_OUTPUT}"
	grep -Eq "commit-hash=[0-9a-f]{40}" "${GITHUB_OUTPUT}"
	[ "$(git log -1 --format=%s)" = "add new file" ]
}

@test "file-pattern stages only matching files" {
	echo "a" >a.txt
	echo "b" >b.txt
	FILE_PATTERN="a.txt" run "${SCRIPT}"
	[ "$status" -eq 0 ]
	committed="$(git diff-tree --no-commit-id --name-only -r HEAD)"
	[[ "$committed" == *"a.txt"* ]]
	[[ "$committed" != *"b.txt"* ]]
	# b.txt remains uncommitted in the working tree.
	[ -n "$(git status --porcelain b.txt)" ]
}

@test "uses a custom git identity for author and committer" {
	echo "x" >x.txt
	GIT_USER_NAME="test-bot" GIT_USER_EMAIL="test-bot@example.com" run "${SCRIPT}"
	[ "$status" -eq 0 ]
	[ "$(git log -1 --format='%an <%ae>')" = "test-bot <test-bot@example.com>" ]
	[ "$(git log -1 --format='%cn <%ce>')" = "test-bot <test-bot@example.com>" ]
}

@test "passes through additional commit-options" {
	echo "y" >y.txt
	COMMIT_OPTIONS="--no-verify" run "${SCRIPT}"
	[ "$status" -eq 0 ]
	grep -q "changes-detected=true" "${GITHUB_OUTPUT}"
}

@test "rejects a branch name starting with a dash" {
	echo "z" >z.txt
	BRANCH="-bad" run "${SCRIPT}"
	[ "$status" -ne 0 ]
	[[ "$output" == *"must not start with '-'"* ]]
}
