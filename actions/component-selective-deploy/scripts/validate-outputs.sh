#!/usr/bin/env bash
set -euo pipefail

# Validate component-selective-deploy action outputs.
#
# Environment variables:
#   SELECTED_DIGESTS_JSON  JSON output from the action (selected_digests_json)
#   DOCKERTAG              Docker tag output from the action

: "${SELECTED_DIGESTS_JSON:?SELECTED_DIGESTS_JSON is required}"
: "${DOCKERTAG:?DOCKERTAG is required}"

# ---- Validate action outputs are valid JSON --------------------------------
echo "Validating SELECTED_DIGESTS_JSON is valid JSON..."
echo "$SELECTED_DIGESTS_JSON" | jq . || {
  echo "Error: SELECTED_DIGESTS_JSON is not valid JSON"
  exit 1
}

echo "Validating DOCKERTAG..."
[[ "$DOCKERTAG" == "abc1234" ]] || {
  echo "Error: Expected dockertag 'abc1234', got '$DOCKERTAG'"
  exit 1
}
echo "  dockertag: $DOCKERTAG ✓"

# ---- Validate digest selection logic ---------------------------------------
# apiserver: CHANGED → must use NEW digest
APISERVER_DIGEST=$(echo "$SELECTED_DIGESTS_JSON" | jq -r '.apiserver_digest')
EXPECTED_APISERVER="abc1234@sha256:1111111111111111111111111111111111111111111111111111111111111111"
[[ "$APISERVER_DIGEST" == "$EXPECTED_APISERVER" ]] || {
  echo "Error: apiserver should use NEW digest"
  echo "  Expected: $EXPECTED_APISERVER"
  echo "  Actual:   $APISERVER_DIGEST"
  exit 1
}
echo "  apiserver_digest (changed → new): $APISERVER_DIGEST ✓"

# controller: UNCHANGED → must use OLD digest from component-tags.json
CONTROLLER_DIGEST=$(echo "$SELECTED_DIGESTS_JSON" | jq -r '.controller_digest')
EXPECTED_CONTROLLER="old5678@sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
[[ "$CONTROLLER_DIGEST" == "$EXPECTED_CONTROLLER" ]] || {
  echo "Error: controller should use OLD digest (unchanged)"
  echo "  Expected: $EXPECTED_CONTROLLER"
  echo "  Actual:   $CONTROLLER_DIGEST"
  exit 1
}
echo "  controller_digest (unchanged → old): $CONTROLLER_DIGEST ✓"

# migrator: CHANGED → must use NEW digest
MIGRATOR_DIGEST=$(echo "$SELECTED_DIGESTS_JSON" | jq -r '.migrator_digest')
EXPECTED_MIGRATOR="abc1234@sha256:3333333333333333333333333333333333333333333333333333333333333333"
[[ "$MIGRATOR_DIGEST" == "$EXPECTED_MIGRATOR" ]] || {
  echo "Error: migrator should use NEW digest"
  echo "  Expected: $EXPECTED_MIGRATOR"
  echo "  Actual:   $MIGRATOR_DIGEST"
  exit 1
}
echo "  migrator_digest (changed → new): $MIGRATOR_DIGEST ✓"

# ---- Validate component-tags.json was updated correctly -------------------
echo "Validating component-tags.json was updated..."
test -f component-tags.json || {
  echo "Error: component-tags.json not found after action run"
  exit 1
}

# apiserver: should have new commitSHA
APISERVER_SHA=$(jq -r '.apiserver.commitSHA' component-tags.json)
[[ "$APISERVER_SHA" == "test-commit-sha" ]] || {
  echo "Error: apiserver.commitSHA should be 'test-commit-sha', got '$APISERVER_SHA'"
  exit 1
}
echo "  apiserver.commitSHA updated ✓"

# controller: should retain old commitSHA (unchanged)
CONTROLLER_SHA=$(jq -r '.controller.commitSHA' component-tags.json)
[[ "$CONTROLLER_SHA" == "old5678" ]] || {
  echo "Error: controller.commitSHA should still be 'old5678', got '$CONTROLLER_SHA'"
  exit 1
}
echo "  controller.commitSHA retained ✓"

# migrator: should have new commitSHA
MIGRATOR_SHA=$(jq -r '.migrator.commitSHA' component-tags.json)
[[ "$MIGRATOR_SHA" == "test-commit-sha" ]] || {
  echo "Error: migrator.commitSHA should be 'test-commit-sha', got '$MIGRATOR_SHA'"
  exit 1
}
echo "  migrator.commitSHA updated ✓"

echo ""
echo "All validations passed!"
