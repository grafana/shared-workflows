#!/usr/bin/env bash

set -euo pipefail

# Validate component change detection outputs

# Validate outputs are not empty and are valid JSON
echo "Validating CHANGES_JSON..."
echo "${CHANGES_JSON}" | jq . || {
  echo "Error: CHANGES_JSON is not valid JSON"
  exit 1
}

echo "Validating COMPONENTS_JSON..."
echo "${COMPONENTS_JSON}" | jq . || {
  echo "Error: COMPONENTS_JSON is not valid JSON"
  exit 1
}

# Validate we have expected components
echo "Validating expected components are present..."
echo "${COMPONENTS_JSON}" | jq -e '. | index("detector")' || {
  echo "Error: Expected component 'detector' not found in COMPONENTS_JSON"
  exit 1
}

echo "${COMPONENTS_JSON}" | jq -e '. | index("cli")' || {
  echo "Error: Expected component 'cli' not found in COMPONENTS_JSON"
  exit 1
}

echo "${COMPONENTS_JSON}" | jq -e '. | index("action")' || {
  echo "Error: Expected component 'action' not found in COMPONENTS_JSON"
  exit 1
}

echo "All validations passed!"
