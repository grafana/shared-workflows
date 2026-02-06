#!/usr/bin/env bash

set -euo pipefail

# Build the change detector Go binary

# Validate we're in the right directory
if [[ ! -d "./cmd/changed-components" ]]; then
  echo "Error: cmd/changed-components directory not found" >&2
  exit 1
fi

echo "Building change detector..."
go build -o changed-components ./cmd/changed-components
chmod +x changed-components
