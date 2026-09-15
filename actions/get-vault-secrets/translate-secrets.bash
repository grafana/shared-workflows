#!/usr/bin/env bash

# Input env:
# - VERSION => vault-secrets schema version (v1 or v2). Selects the Vault mount:
#     v1 uses the shared `ci` mount; v2 uses a per-org `ci-${REPO_OWNER}` mount.
# - REPO => Repository name (owner/repo)
# - REPO_OWNER => Repository owner/org (only required for v2)
# - COMMON_SECRETS => Common secrets (in the <mount>/data/common/<path> vault path): {{ Env Variable Name }}={{ Secret Path }}:{{ Secret Key }}
# - REPO_SECRETS => Repo secrets (in the <mount>/data/repo/${REPO}/<path> vault path): {{ Env Variable Name }}={{ Secret Path }}:{{ Secret Key }}
# Output format: "{{ Secret Path }} {{ Secret Key }} | {{ Env Variable Name }}" in the $GITHUB_OUTPUT file

# Check if the REPO environment variable is set
if [ -z "$REPO" ]; then
	echo "Error: REPO environment variable is not set."
	exit 1
fi

# Check if the GITHUB_OUTPUT environment variable is set. It should be set by Github Actions.
if [ -z "$GITHUB_OUTPUT" ]; then
	echo "Error: GITHUB_OUTPUT environment variable is not set."
	exit 1
fi

# Determine the Vault mount from the schema version. Only the mount changes
# between versions; the rest of the secret path is identical.
# - v1 (default): the shared `ci` mount.
# - v2: a per-org `ci-${REPO_OWNER}` mount.
if [ "$VERSION" = "v2" ]; then
	if [ -z "$REPO_OWNER" ]; then
		echo "Error: REPO_OWNER environment variable is not set."
		exit 1
	fi
	MOUNT="ci-${REPO_OWNER}"
else
	MOUNT="ci"
fi

readonly COMMON_SECRETS GITHUB_OUTPUT MOUNT REPO REPO_OWNER REPO_SECRETS VERSION

RESULT=""

# Function to split a string into parts
split_string() {
	local input_string="$1"
	IFS='=' read -ra parts <<<"$input_string"

	if [ "${#parts[@]}" -eq 2 ]; then
		env_variable_name="${parts[0]}"
		secret_parts="${parts[1]}"

		IFS=':' read -ra secret_parts <<<"$secret_parts"

		if [ "${#secret_parts[@]}" -eq 2 ]; then
			secret_path="${secret_parts[0]}"
			secret_key="${secret_parts[1]}"
		fi
	fi
}

# Translate the common secrets
if [ -n "$COMMON_SECRETS" ]; then
	for common_secret in $COMMON_SECRETS; do
		split_string "$common_secret"
		RESULT="${RESULT}${MOUNT}/data/common/$secret_path $secret_key | $env_variable_name;\n"
	done
fi

# Translate the repo secrets
if [ -n "$REPO_SECRETS" ]; then
	for repo_secret in $REPO_SECRETS; do
		split_string "$repo_secret"
		RESULT="${RESULT}${MOUNT}/data/repo/$REPO/$secret_path $secret_key | $env_variable_name;\n"
	done
fi

readonly RESULT

# Print the contents of the output file
echo -e "Secrets that will be queried from Vault:\n$RESULT"
echo -e "secrets<<EOF\n${RESULT}EOF" >"$GITHUB_OUTPUT"
