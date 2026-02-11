#!/usr/bin/env bash

set -euo pipefail

# Build the change detector Go binary

echo "Building change detector..."
go build -o changed-components ./cmd/changed-components
chmod +x changed-components
