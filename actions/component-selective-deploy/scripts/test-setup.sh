#!/usr/bin/env bash
set -euo pipefail

# Set up previous component tags fixture for the component-selective-deploy test.
# Copies the test fixture component-tags.json into the working directory,
# simulating what download-previous-component-tags does in a real deploy workflow.
#
# component-digests/ is populated separately by save-component-digest action calls
# in the test workflow, mirroring real usage.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FIXTURE_DIR="$(cd "$SCRIPT_DIR/../test-fixtures" && pwd)"

cp "$FIXTURE_DIR/component-tags.json" component-tags.json

echo "Previous component tags loaded from fixture:"
jq . component-tags.json
