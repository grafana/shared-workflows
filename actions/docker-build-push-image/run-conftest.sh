#!/bin/bash
set -euo pipefail

: "${DOCKERFILES:=}"
: "${STRICT:=false}"

if [[ $# -lt 1 ]]; then
    printf "ERROR: %s <test-policies|verify-dockerfiles>\n" "$0" >&2

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

        # newline-delimited split preserves paths that contain spaces
        mapfile -t raw_files <<<"${DOCKERFILES}"
        # drop trailing empty entry if input ended with a newline
        if [[ ${#raw_files[@]} -gt 0 && -z "${raw_files[-1]}" ]]; then
            unset 'raw_files[-1]'
        fi
    fi

    # Resolve each path against the workspace and verify it stays within it.
    # Rejects path-traversal (../../etc/passwd) and absolute paths outside
    # the workspace, which would otherwise let a caller exfiltrate arbitrary
    # files through conftest's parse-error messages.
    workspace="${GITHUB_WORKSPACE:-$PWD}"
    workspace_real="$(realpath "${workspace}")"
    files=()
    for raw in "${raw_files[@]}"; do
        if [[ -z "${raw}" ]]; then
            continue
        fi

        if [[ "${raw}" = /* ]]; then
            candidate="${raw}"
        else
            candidate="${workspace}/${raw}"
        fi

        candidate_real="$(realpath -m "${candidate}")"
        if [[ "${candidate_real}" != "${workspace_real}"/* && "${candidate_real}" != "${workspace_real}" ]]; then
            printf "ERROR: %q resolves outside the workspace (%s)\n" "${raw}" "${candidate_real}" >&2
            exit 1
        fi

        files+=("${candidate_real}")
    done

    if conftest test --policy ./conftest/policy --parser dockerfile -- "${files[@]}"; then
        exit 0
    fi

    if [[ "${STRICT}" = "true" ]]; then
        exit 1
    fi

    printf "::warning::Dockerfile hardening policy violations found; not failing because strict mode is disabled (set strict: true to enforce)\n" >&2
    exit 0
fi

printf "ERROR: unknown subcommand %q (expected test-policies or verify-dockerfiles)\n" "${1}" >&2
exit 1
