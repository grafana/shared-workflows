#!/usr/bin/env bash

set -euo pipefail

# Extract component list from config file
# Outputs: components (comma-separated) and components_json (JSON array)

CONFIG_FILE="${CONFIG_FILE:?CONFIG_FILE environment variable is required}"

# Extract component names from config file
COMPONENTS=$(yq eval '.components | keys | join(",")' "$CONFIG_FILE")
echo "components=$COMPONENTS" >> "$GITHUB_OUTPUT"
echo "Detected components: $COMPONENTS"

# Create JSON array for easier processing (ensure it's on one line)
COMPONENTS_JSON=$(yq eval '.components | keys' "$CONFIG_FILE" -o json -I=0)
echo "components_json=$COMPONENTS_JSON" >> "$GITHUB_OUTPUT"
