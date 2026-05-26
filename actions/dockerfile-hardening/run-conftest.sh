#!/bin/bash
set -euo pipefail

: "${DOCKERFILES:=}"

if [[ $# -lt 1 ]]; then
    printf "ERROR: $0 <test-policies|verify-dockerfiles>\n" >&2

    exit 1
fi

if [[ "${1}" = "test-policies" ]]; then
    # smoke test
    conftest test --policy ./conftest/policy --parser dockerfile conftest/tests/fixtures/happy-path.Dockerfile

    # individual rego tests
    conftest verify --policy ./conftest

    exit
fi

if [[ "${1}" = "verify-dockerfiles" ]]; then
    shift

    cli_files=()
    while getopts "f:" opt; do
        case "$opt" in
            f) cli_files+=("$OPTARG") ;;
            *)
                printf "usage: %s verify-dockerfiles [-f FILE]...\n" "$0" >&2
                exit 1
                ;;
        esac
    done

    # if no Dockerfile args are passed in, check the \$DOCKERFILES env var (newline-delimited)
    if [[ ${#cli_files[@]} -gt 0 ]]; then
        raw_files=("${cli_files[@]}")
    else
        if [[ -z "${DOCKERFILES//[[:space:]]/}" ]]; then
            printf "ERROR: no Dockerfiles supplied (pass -f FILE or set DOCKERFILES)\n" >&2
            exit 1
        fi

        read -r -a raw_files <<<"$(echo "${DOCKERFILES}" | tr '\n' ' ')"
    fi

    # Ensure any Dockerfiles passed by the caller are rooted at the project's
    # Github workspace (the project's main dir) IF they are not already a full
    # file path from '/' (root)
    workspace="${GITHUB_WORKSPACE:-$PWD}"
    files=()
    for raw in "${raw_files[@]}"; do
        if [[ "${raw}" = /* ]]; then
            files+=("${raw}")
        else
            files+=("${workspace}/${raw}")
        fi
    done

    conftest test --policy ./conftest/policy --parser dockerfile "${files[@]}"
fi

