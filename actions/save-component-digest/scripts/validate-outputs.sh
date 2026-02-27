#!/usr/bin/env bash
set -euo pipefail

# Validate save-component-digest action outputs.
#
# Environment variables:
#   EXPECTED_DIGEST        Full digest string expected in the component file (tag@sha256:...)
#   EXPECTED_DOCKERTAG     Docker tag expected in dockertag.txt
#   EXPECTED_FILENAME      (Optional) Exact filename (without dir) to verify exists, e.g. "grafana_com_api.txt"
#                          Use this to assert that _digest suffix stripping worked correctly.

: "${EXPECTED_DIGEST:?EXPECTED_DIGEST is required}"
: "${EXPECTED_DOCKERTAG:?EXPECTED_DOCKERTAG is required}"

DIGEST_DIR="component-digests"

echo "Validating component-digests directory exists..."
test -d "$DIGEST_DIR" || {
  echo "Error: Directory '$DIGEST_DIR' not found"
  exit 1
}

echo "Validating dockertag.txt..."
test -f "$DIGEST_DIR/dockertag.txt" || {
  echo "Error: $DIGEST_DIR/dockertag.txt not found"
  exit 1
}
ACTUAL_DOCKERTAG=$(< "$DIGEST_DIR/dockertag.txt")
[[ "$ACTUAL_DOCKERTAG" == "$EXPECTED_DOCKERTAG" ]] || {
  echo "Error: dockertag mismatch"
  echo "  Expected: $EXPECTED_DOCKERTAG"
  echo "  Actual:   $ACTUAL_DOCKERTAG"
  exit 1
}
echo "  dockertag: $ACTUAL_DOCKERTAG ✓"

echo "Validating digest content in component file..."
FOUND=false
for f in "$DIGEST_DIR"/*.txt; do
  [[ "$(basename "$f")" == "dockertag.txt" ]] && continue
  CONTENT=$(< "$f")
  if [[ "$CONTENT" == "$EXPECTED_DIGEST" ]]; then
    echo "  $(basename "$f"): $CONTENT ✓"
    FOUND=true
    break
  fi
done
[[ "$FOUND" == "true" ]] || {
  echo "Error: No component file contains expected digest '$EXPECTED_DIGEST'"
  echo "Files in $DIGEST_DIR:"
  ls -la "$DIGEST_DIR/"
  exit 1
}

if [[ -n "${EXPECTED_FILENAME:-}" ]]; then
  echo "Validating expected filename '$EXPECTED_FILENAME' (suffix stripping)..."
  test -f "$DIGEST_DIR/$EXPECTED_FILENAME" || {
    echo "Error: Expected file '$DIGEST_DIR/$EXPECTED_FILENAME' not found"
    echo "Files in $DIGEST_DIR:"
    ls -la "$DIGEST_DIR/"
    exit 1
  }
  echo "  $EXPECTED_FILENAME ✓"
fi

echo "All validations passed!"
