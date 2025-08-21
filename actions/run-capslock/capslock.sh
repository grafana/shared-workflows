#!/usr/bin/env bash
set -euo pipefail

# Fail with a clear message if required env vars are missing
for var in GITHUB_ACTION_PATH MAIN_BRANCH GITHUB_HEAD_REF OUTPUT_PLACE GITHUB_SHA; do
    if [[ -z "${!var:-}" ]]; then
        echo "ERROR: Environment variable ${var} is required but not set"
        exit 1
    fi
done

BASE_DIR="${PWD}"
cd "${SCOPE:-.}" || exit 1

if [[ ! -f "go.mod" ]]; then
    echo "No go module found in ${PWD}"
    exit 0
fi

# Helper: run capslock on all packages and save result to file
run_capslock() {
    local output_file=$1
    local packages
    packages=$(go list ./... | paste -sd "," -)
    if [[ -z "${packages}" ]]; then
        echo "No packages found"
        exit 0
    fi
    capslock -granularity intermediate -packages "${packages}" -output json >"${output_file}"
}

# Work with MAIN_BRANCH state
git fetch origin "${MAIN_BRANCH}" --depth 1
git checkout "${MAIN_BRANCH}"
if [[ ! -f "go.mod" ]]; then
    echo "No go module found on branch ${MAIN_BRANCH}"
    exit 0
fi
run_capslock "${BASE_DIR}/capslock.json"

# Switch back to feature branch and rerun
git stash --quiet
git checkout "${GITHUB_SHA}"
run_capslock "${BASE_DIR}/capslock2.json"
git stash pop --quiet || true

# Move results back to action path
mv "${BASE_DIR}/capslock.json" "${GITHUB_ACTION_PATH}/capslock.json"
mv "${BASE_DIR}/capslock2.json" "${GITHUB_ACTION_PATH}/capslock2.json"

# Run comparison
cd "${GITHUB_ACTION_PATH}"
go mod tidy
OUTPUT=$(go run compare.go capslock.json capslock2.json || true)

if [[ "${OUTPUT}" =~ "Between those commits, there were no uses of capabilities via a new package" ]]; then
    echo "No new Capabilities"
    exit 0
fi

# Direct log output
if [[ "${OUTPUT_PLACE}" == "log" ]]; then
    printf "%s\n" "${OUTPUT}"
    exit 0
fi

# Format output
FORMATTED_OUTPUT=$(printf "%s" "${OUTPUT}" | go run "${GITHUB_ACTION_PATH}/formatting/formatting.go" || true)
if [[ "${FORMATTED_OUTPUT}" =~ "No match found" ]]; then
    printf "%s\n" "${OUTPUT}"
    echo "ERROR: formatting the output"
    exit 1
fi

# Send to GitHub Actions outputs
case "${OUTPUT_PLACE}" in
    pr-comment)
        {
            echo 'output<<EOF'
            printf "%s\nEOF" "${FORMATTED_OUTPUT}"
        } >>"${GITHUB_OUTPUT}"
        ;;
    summary)
        {
            echo 'output<<EOF'
            printf "%s\nEOF" "${FORMATTED_OUTPUT}"
        } >>"${GITHUB_STEP_SUMMARY}"
        ;;
    *)
        echo "ERROR: Only 'log', 'pr-comment' or 'summary' are allowed values for output_place input"
        ;;
esac
