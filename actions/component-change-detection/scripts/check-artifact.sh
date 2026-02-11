#!/usr/bin/env bash

set -euo pipefail

# Check if component-tags.json artifact was downloaded

if [ -f component-tags.json ]; then
  echo "found=true" >> "$GITHUB_OUTPUT"
  echo "Found component-tags.json"
else
  echo "found=false" >> "$GITHUB_OUTPUT"
  echo "Artifact download failed or not found - will mark all components as changed"
fi
