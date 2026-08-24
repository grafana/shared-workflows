#!/usr/bin/env bash

set -euo pipefail

# Build the change detector Go binary

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ACTION_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "Building change detector..."
cd "$ACTION_DIR"
go build -o changed-components ./cmd/changed-components
chmod +x changed-components
