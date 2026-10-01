#!/usr/bin/env bats

# Set up test environment
setup() {
	export VERSION="v1"
	export REPO="grafana/myrepo"
	export REPO_OWNER="grafana"
	export GITHUB_OUTPUT="secrets_output.txt"
	export COMMON_SECRETS="SECRET1=secret1:key1
SECRET2=subfolder/secret2:key2"
	export REPO_SECRETS="SECRET3=secret3:key3
SECRET4=subfolder/secret4:key4
"
}

# Clean up temporary files after tests
teardown() {
	rm -f "$GITHUB_OUTPUT"
}

@test "Check if REPO environment variable is set" {
	REPO="" run ./translate-secrets.bash
	[ "$status" -ne 0 ]
	[ "${lines[0]}" = "Error: REPO environment variable is not set." ]
}

@test "Check if GITHUB_OUTPUT environment variable is set" {
	GITHUB_OUTPUT="" run ./translate-secrets.bash
	[ "$status" -ne 0 ]
	[ "${lines[0]}" = "Error: GITHUB_OUTPUT environment variable is not set." ]
}

@test "Translate secrets (v1 schema uses the ci mount)" {
	run ./translate-secrets.bash
	echo "$output" >&3
	[ "$status" -eq 0 ]
	[ "$output" = "Secrets that will be queried from Vault:
ci/data/common/secret1 key1 | SECRET1;
ci/data/common/subfolder/secret2 key2 | SECRET2;
ci/data/repo/grafana/myrepo/secret3 key3 | SECRET3;
ci/data/repo/grafana/myrepo/subfolder/secret4 key4 | SECRET4;" ]

	echo -e "\nGITHUB_OUTPUT:\n$(cat "$GITHUB_OUTPUT")" >&3
	[ "$(cat "$GITHUB_OUTPUT")" = "secrets<<EOF
ci/data/common/secret1 key1 | SECRET1;
ci/data/common/subfolder/secret2 key2 | SECRET2;
ci/data/repo/grafana/myrepo/secret3 key3 | SECRET3;
ci/data/repo/grafana/myrepo/subfolder/secret4 key4 | SECRET4;
EOF" ]
}

@test "Translate secrets (v2 schema uses the ci-<owner> mount)" {
	VERSION="v2" run ./translate-secrets.bash
	echo "$output" >&3
	[ "$status" -eq 0 ]
	[ "$output" = "Secrets that will be queried from Vault:
ci-grafana/data/common/secret1 key1 | SECRET1;
ci-grafana/data/common/subfolder/secret2 key2 | SECRET2;
ci-grafana/data/repo/grafana/myrepo/secret3 key3 | SECRET3;
ci-grafana/data/repo/grafana/myrepo/subfolder/secret4 key4 | SECRET4;" ]

	echo -e "\nGITHUB_OUTPUT:\n$(cat "$GITHUB_OUTPUT")" >&3
	[ "$(cat "$GITHUB_OUTPUT")" = "secrets<<EOF
ci-grafana/data/common/secret1 key1 | SECRET1;
ci-grafana/data/common/subfolder/secret2 key2 | SECRET2;
ci-grafana/data/repo/grafana/myrepo/secret3 key3 | SECRET3;
ci-grafana/data/repo/grafana/myrepo/subfolder/secret4 key4 | SECRET4;
EOF" ]
}

@test "Check if REPO_OWNER environment variable is set for v2" {
	VERSION="v2" REPO_OWNER="" run ./translate-secrets.bash
	[ "$status" -ne 0 ]
	[ "${lines[0]}" = "Error: REPO_OWNER environment variable is not set." ]
}
