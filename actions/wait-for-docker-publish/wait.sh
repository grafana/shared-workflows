#!/usr/bin/env bash
set -euo pipefail

: "${IMAGE:?IMAGE env var is required}"

validate_image() {
  local image="$1"

  if [[ "$image" == *"@sha256:"* ]]; then
    return 0
  fi

  # Strip up to the last '/' so a registry port (e.g. localhost:5000) isn't
  # mistaken for a tag.
  local after_slash="${image##*/}"

  if [[ "$after_slash" == *":"* ]]; then
    return 0
  fi

  return 1
}

if ! validate_image "$IMAGE"; then
  echo "::error::IMAGE must include a tag (repo:tag) or digest (repo@sha256:...); got '${IMAGE}'"
  exit 1
fi

parse_duration() {
  local v="$1"
  local n="${v%[smh]}"

  if ! [[ "$n" =~ ^[0-9]+$ ]]; then
    echo "::error::duration '${v}' is not a valid number with optional s/m/h suffix" >&2
    return 1
  fi

  case "$v" in
    *h) echo "$(( n * 3600 ))" ;;
    *m) echo "$(( n * 60 ))" ;;
    *s) echo "$n" ;;
    *)  echo "$n" ;;
  esac
}

: "${TIMEOUT:=10m}"
: "${INITIAL_INTERVAL:=5s}"
: "${MAX_INTERVAL:=60s}"

timeout_s=$(parse_duration "$TIMEOUT")
initial_s=$(parse_duration "$INITIAL_INTERVAL")
max_s=$(parse_duration "$MAX_INTERVAL")

start=$SECONDS
deadline=$(( start + timeout_s ))
interval=$initial_s
attempt=1
err_file="$(mktemp)"
trap 'rm -f "$err_file"' EXIT

echo "wait-for-docker-publish: polling for ${IMAGE} (timeout ${timeout_s}s)"

while true; do
  : > "$err_file"
  if docker manifest inspect "$IMAGE" >/dev/null 2>"$err_file"; then
    echo "image found after $(( SECONDS - start ))s on attempt ${attempt}"
    exit 0
  fi

  last_err="$(tr '\n' ' ' < "$err_file")"
  remaining=$(( deadline - SECONDS ))
  if (( remaining <= 0 )); then
    echo "::error::timed out after ${TIMEOUT} waiting for ${IMAGE}; last error: ${last_err}"
    exit 1
  fi

  sleep_for=$interval
  if (( sleep_for > remaining )); then
    sleep_for=$remaining
  fi
  echo "attempt ${attempt}: not yet available, sleeping ${sleep_for}s (elapsed $(( SECONDS - start ))s / ${timeout_s}s); last error: ${last_err}"
  sleep "$sleep_for"

  interval=$(( interval * 2 ))
  if (( interval > max_s )); then
    interval=$max_s
  fi
  attempt=$(( attempt + 1 ))
done
