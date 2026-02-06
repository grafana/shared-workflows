#!/usr/bin/env bash

set -euo pipefail

# Prepare previous tags for comparison
# Converts component-tags.json to component-tags-previous.json format

# Validate required environment variables
COMPONENTS_JSON="${COMPONENTS_JSON:?COMPONENTS_JSON environment variable is required}"
readonly COMPONENTS_JSON

# Validate JSON format of COMPONENTS_JSON
if ! echo "$COMPONENTS_JSON" | jq empty 2>/dev/null; then
  echo "Error: COMPONENTS_JSON is not valid JSON" >&2
  exit 1
fi

# If artifact was downloaded, extract SHAs from component-tags.json
if [ -f component-tags.json ]; then
  echo "Found previous build state from artifact"
  
  # Expect new format: {"component": {"commitSHA": "...", "digest": "..."}}
  FIRST_VALUE=$(jq -r 'to_entries | .[0].value | type' component-tags.json 2>/dev/null || echo "null")
  
  if [ "$FIRST_VALUE" = "object" ]; then
    echo "New format detected - extracting SHAs"
    jq 'to_entries | map({(.key): .value.commitSHA}) | add' component-tags.json > component-tags-previous.json
  else
    # Old/invalid format - treat as first run to rebuild everything
    echo "Unsupported format - treating as first run"
    echo "$COMPONENTS_JSON" | jq -r 'map({(.): "none"}) | add' > component-tags-previous.json
  fi
else
  echo "No previous build state found - first run will mark all components as changed"
  
  # Dynamically create initial state from component list
  echo "$COMPONENTS_JSON" | jq -r 'map({(.): "none"}) | add' > component-tags-previous.json
fi

echo "Previous SHAs for change detection:"
jq . component-tags-previous.json
