#!/usr/bin/env bash

set -euo pipefail

# Check if component-tags.json artifact was downloaded

# Validate required environment variables
GITHUB_OUTPUT="${GITHUB_OUTPUT:?GITHUB_OUTPUT environment variable is required}"
readonly GITHUB_OUTPUT

if [ -f component-tags.json ]; then
  echo "found=true" >> "$GITHUB_OUTPUT"
  echo "Found component-tags.json"
else
  echo "found=false" >> "$GITHUB_OUTPUT"
  echo "Artifact download failed or not found - will mark all components as changed"
fi
