#!/usr/bin/env bash
#
# run-shellcheck.sh
#
# Finds and lints shell scripts using shellcheck.
# Designed as a lightweight replacement for ludeeus/action-shellcheck.
#
# Required environment variables: (none -- all have defaults)
#
# Optional environment variables:
#   SCANDIR          - directory to scan (default: .)
#   IGNORE_PATHS     - space-separated paths/directories to exclude
#   IGNORE_NAMES     - space-separated file names to exclude
#   SEVERITY         - minimum severity: error, warning, info, style
#   FORMAT           - output format: gcc, tty, json, checkstyle, diff, quiet (default: gcc)
#   ADDITIONAL_FILES - space-separated extra file names to scan for
#   CHECK_TOGETHER   - set to "true" to run shellcheck on all files at once
#   SHELLCHECK_OPTS  - additional shellcheck flags (read by shellcheck automatically)

set -euo pipefail

SCANDIR="${SCANDIR:-.}"
FORMAT="${FORMAT:-gcc}"
CHECK_TOGETHER="${CHECK_TOGETHER:-false}"

# ---------------------------------------------------------------------------
# Display shellcheck version
# ---------------------------------------------------------------------------
shellcheck --version

# ---------------------------------------------------------------------------
# Build shellcheck options
# ---------------------------------------------------------------------------
sc_options=()
if [ -n "${SEVERITY:-}" ]; then
  sc_options+=("-S" "${SEVERITY}")
fi
sc_options+=("--format=${FORMAT}")

# ---------------------------------------------------------------------------
# Build find exclusion arguments
# ---------------------------------------------------------------------------
set -f # disable globbing so glob patterns in inputs aren't expanded

exclude_args=()

# Always exclude .git, .go files, and mvnw (matching ludeeus behavior)
exclude_args+=("!" "-path" "*/.git/*")
exclude_args+=("!" "-path" "*.go")
exclude_args+=("!" "-path" "*/mvnw")

# shellcheck disable=SC2086
for path in ${IGNORE_PATHS:-}; do
  exclude_args+=("!" "-path" "*./${path}/*")
  exclude_args+=("!" "-path" "*/${path}/*")
  exclude_args+=("!" "-path" "${path}")
done

# shellcheck disable=SC2086
for name in ${IGNORE_NAMES:-}; do
  exclude_args+=("!" "-name" "${name}")
done

# ---------------------------------------------------------------------------
# Build additional file name arguments
# ---------------------------------------------------------------------------
additional_file_args=()
# shellcheck disable=SC2086
for file in ${ADDITIONAL_FILES:-}; do
  additional_file_args+=("-o" "-name" "*${file}")
done

# ---------------------------------------------------------------------------
# Pass 1: Find files by known shell extensions and names
# ---------------------------------------------------------------------------
filepaths=()
shebang_regex="^#! */[^ ]*/(env *)?[abk]*sh"

while IFS= read -r -d '' file; do
  filepaths+=("$file")
done < <(find "${SCANDIR}" \
    "${exclude_args[@]}" \
    -type f \
    '(' \
    -name '*.bash' \
    -o -name '.bashrc' \
    -o -name 'bashrc' \
    -o -name '.bash_aliases' \
    -o -name '.bash_completion' \
    -o -name '.bash_login' \
    -o -name '.bash_logout' \
    -o -name '.bash_profile' \
    -o -name 'bash_profile' \
    -o -name '*.ksh' \
    -o -name 'suid_profile' \
    -o -name '*.zsh' \
    -o -name '.zlogin' \
    -o -name 'zlogin' \
    -o -name '.zlogout' \
    -o -name 'zlogout' \
    -o -name '.zprofile' \
    -o -name 'zprofile' \
    -o -name '.zsenv' \
    -o -name 'zsenv' \
    -o -name '.zshrc' \
    -o -name 'zshrc' \
    -o -name '*.sh' \
    -o -path '*/.profile' \
    -o -path '*/profile' \
    -o -name '*.shlib' \
    "${additional_file_args[@]+"${additional_file_args[@]}"}" \
    ')' \
    -print0)

# ---------------------------------------------------------------------------
# Pass 2: Extensionless executables with shell shebangs
# ---------------------------------------------------------------------------
while IFS= read -r -d '' file; do
  head -n1 "$file" | grep -Eqs "$shebang_regex" || continue
  filepaths+=("$file")
done < <(find "${SCANDIR}" \
    "${exclude_args[@]}" \
    -type f ! -name '*.*' -perm /111 \
    -print0)

set +f # re-enable globbing

# ---------------------------------------------------------------------------
# Check if any files were found
# ---------------------------------------------------------------------------
if [ "${#filepaths[@]}" -eq 0 ]; then
  echo "No shell scripts found in '${SCANDIR}'."
  exit 0
fi

echo "Found ${#filepaths[@]} shell script(s) to check."

# ---------------------------------------------------------------------------
# Run shellcheck
# ---------------------------------------------------------------------------
statuscode=0

if [ "${CHECK_TOGETHER}" = "true" ]; then
  shellcheck "${sc_options[@]}" "${filepaths[@]}" || statuscode=$?
else
  for file in "${filepaths[@]}"; do
    shellcheck "${sc_options[@]}" "$file" || statuscode=$?
  done
fi

# ---------------------------------------------------------------------------
# Set outputs
# ---------------------------------------------------------------------------
if [ -n "${GITHUB_OUTPUT:-}" ]; then
  echo "files=${filepaths[*]}" >> "${GITHUB_OUTPUT}"
fi

if [ "${statuscode}" -eq 0 ]; then
  echo "All ${#filepaths[@]} file(s) passed shellcheck."
else
  echo "shellcheck found issues in one or more files."
fi

exit "${statuscode}"
