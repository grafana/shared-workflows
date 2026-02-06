#!/usr/bin/env bash

set -euo pipefail

# Extract component list from config file
# Outputs: components (comma-separated) and components_json (JSON array)

# Validate required environment variables
CONFIG_FILE="${CONFIG_FILE:?CONFIG_FILE environment variable is required}"
GITHUB_OUTPUT="${GITHUB_OUTPUT:?GITHUB_OUTPUT environment variable is required}"

# Make variables readonly to prevent modification
readonly CONFIG_FILE GITHUB_OUTPUT

# Validate config file exists and is a regular file
if [[ ! -f "$CONFIG_FILE" ]]; then
  echo "Error: Config file not found: $CONFIG_FILE" >&2
  exit 1
fi

# Extract component names from config file
COMPONENTS=$(yq eval '.components | keys | join(",")' "$CONFIG_FILE")
echo "components=$COMPONENTS" >> "$GITHUB_OUTPUT"
echo "Detected components: $COMPONENTS"

# Create JSON array for easier processing (ensure it's on one line)
COMPONENTS_JSON=$(yq eval '.components | keys' "$CONFIG_FILE" -o json -I=0)
echo "components_json=$COMPONENTS_JSON" >> "$GITHUB_OUTPUT"
