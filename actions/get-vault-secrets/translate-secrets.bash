#!/usr/bin/env bash

# Input env:
# - REPO => Repository name
# - COMMON_SECRETS => Common secrets (in the ci/data/common/<path> vault path): {{ Env Variable Name }}={{ Secret Path }}:{{ Secret Key }}
# - REPO_SECRETS => Repo secrets (in the ci/data/repo/${REPO}/<path> vault path): {{ Env Variable Name }}={{ Secret Path }}:{{ Secret Key }}
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

readonly COMMON_SECRETS GITHUB_OUTPUT REPO REPO_SECRETS

RESULT=""

# Function to split a string into parts
split_string() {
  local input_string="$1"
  IFS='=' read -ra parts <<< "$input_string"

  if [ "${#parts[@]}" -eq 2 ]; then
    env_variable_name="${parts[0]}"
    secret_parts="${parts[1]}"

    IFS=':' read -ra secret_parts <<< "$secret_parts"

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
        RESULT="${RESULT}ci/data/common/$secret_path $secret_key | $env_variable_name;\n"
    done
fi

# Translate the repo secrets
if [ -n "$REPO_SECRETS" ]; then
    for repo_secret in $REPO_SECRETS; do
        split_string "$repo_secret"
        RESULT="${RESULT}ci/data/repo/$REPO/$secret_path $secret_key | $env_variable_name;\n"
    done
fi

readonly RESULT

# Print the contents of the output file
echo -e "Secrets that will be queried from Vault:\n$RESULT"
echo -e "secrets<<EOF\n${RESULT}EOF" > "$GITHUB_OUTPUT"
