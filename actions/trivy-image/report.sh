#!/usr/bin/env bash
# Render a Trivy JSON report as Markdown on stdout and write counts to
# $GITHUB_OUTPUT when set. When COMMENT_PATH is set, also write the PR comment
# there: the same Markdown, or a reduced version without the table when the full
# one exceeds MAX_BYTES (GitHub rejects comments over 65536 characters).
#
# Environment:
#   REPORT_PATH         Trivy JSON report for the image under test (required)
#   IMAGE_REF, IMAGE_NAME, SLUG, SEVERITY, IGNORE_UNFIXED, SBOM_CACHE_HIT,
#   TRIVY_BIN_VERSION, MAX_ROWS, COMMENT_PATH, MAX_BYTES
#   GITHUB_SERVER_URL, GITHUB_REPOSITORY, GITHUB_RUN_ID (runner defaults) for the run link
set -euo pipefail

: "${REPORT_PATH:?REPORT_PATH is required}"
IMAGE_REF="${IMAGE_REF:-unknown}"
IMAGE_NAME="${IMAGE_NAME:-${IMAGE_REF}}"
SLUG="${SLUG:-image}"
SEVERITY="${SEVERITY:-CRITICAL,HIGH}"
IGNORE_UNFIXED="${IGNORE_UNFIXED:-false}"
SBOM_CACHE_HIT="${SBOM_CACHE_HIT:-false}"
TRIVY_BIN_VERSION="${TRIVY_BIN_VERSION:-}"
MAX_ROWS="${MAX_ROWS:-100}"
COMMENT_PATH="${COMMENT_PATH:-}"
MAX_BYTES="${MAX_BYTES:-60000}"

# Tables up to this many rows are expanded in the comment; longer ones are collapsed.
OPEN_ROWS=20

if ! [[ "${MAX_ROWS}" =~ ^[0-9]+$ ]]; then
  echo "::warning::max-rows must be a non-negative integer (got '${MAX_ROWS}'); using 100." >&2
  MAX_ROWS=100
fi
if ! [[ "${MAX_BYTES}" =~ ^[0-9]+$ ]]; then
  MAX_BYTES=60000
fi

run_url=""
if [[ -n "${GITHUB_SERVER_URL:-}" && -n "${GITHUB_REPOSITORY:-}" && -n "${GITHUB_RUN_ID:-}" ]]; then
  run_url="${GITHUB_SERVER_URL}/${GITHUB_REPOSITORY}/actions/runs/${GITHUB_RUN_ID}"
fi

summary_path="$(dirname "${REPORT_PATH}")/summary.json"

jq -n \
  --slurpfile pr "${REPORT_PATH}" \
  '
  def rank: {CRITICAL: 0, HIGH: 1, MEDIUM: 2, LOW: 3}[.] // 4;
  def vulns:
    [ .Results[]? | .Type as $type
      | .Vulnerabilities[]?
      | { id: .VulnerabilityID,
          pkg: .PkgName,
          installed: (.InstalledVersion // ""),
          fixed: (.FixedVersion // ""),
          severity: (.Severity // "UNKNOWN"),
          url: (.PrimaryURL // ""),
          type: ($type // ""),
          key: (.VulnerabilityID + "|" + .PkgName + "|" + (.PkgPath // "")) } ]
    | unique_by(.key)
    | sort_by((.severity | rank), .id, .pkg);
  def count($s): map(select(.severity == $s)) | length;

  ($pr[0] | vulns) as $cur
  | {
      os: ($pr[0].Metadata.OS | if . == null then "" else "\(.Family) \(.Name)" end),
      findings: $cur,
      critical: ($cur | count("CRITICAL")),
      high: ($cur | count("HIGH")),
      medium: ($cur | count("MEDIUM")),
      low: ($cur | count("LOW")),
      total: ($cur | length),
      fixable: ($cur | map(select(.fixed != "")) | length)
    }
  ' > "${summary_path}"

if [[ -n "${GITHUB_OUTPUT:-}" ]]; then
  jq -r '
    "critical=\(.critical)",
    "high=\(.high)",
    "medium=\(.medium)",
    "low=\(.low)",
    "total=\(.total)",
    "fixable=\(.fixable)"
  ' "${summary_path}" >> "${GITHUB_OUTPUT}"
fi

# render <full|reduced>: "reduced" replaces the findings table with a pointer to
# the job summary, for comments that would exceed MAX_BYTES.
render() {
jq -r \
  --arg mode "$1" \
  --arg image_ref "${IMAGE_REF}" \
  --arg image_name "${IMAGE_NAME}" \
  --arg slug "${SLUG}" \
  --arg severity "${SEVERITY}" \
  --arg ignore_unfixed "${IGNORE_UNFIXED}" \
  --arg trivy_version "${TRIVY_BIN_VERSION}" \
  --arg run_url "${run_url}" \
  --argjson max_rows "${MAX_ROWS}" \
  --argjson open_rows "${OPEN_ROWS}" \
  '
  # Report fields come from the image and the vulnerability DB, so treat them as untrusted:
  # escape Markdown/HTML metacharacters and only link to plain https URLs.
  def esc: tostring | gsub("\r|\n"; " ") | gsub("(?<c>[\\\\`*_{}\\[\\]()#+!|<>~])"; "\\\(.c)");
  def html: tostring | gsub("&"; "&amp;") | gsub("<"; "&lt;") | gsub(">"; "&gt;")
    | gsub("`"; "&#96;") | gsub("\\*"; "&#42;") | gsub("_"; "&#95;") | gsub("~"; "&#126;") | gsub("\\["; "&#91;")
    | gsub("\\|"; "\\|") | gsub("\r|\n"; " ");
  def code: if . == "" then "—" else "<code>" + html + "</code>" end;
  def safe_url:
    if test("^https://[A-Za-z0-9._~:/?#@!$&'"'"'*+,;=%()-]+$") then gsub("\\("; "%28") | gsub("\\)"; "%29") else "" end;
  def link: (.url | safe_url) as $u | if $u == "" then (.id | esc) else "[\(.id | esc)](\($u))" end;
  def icon: {CRITICAL: "🔴", HIGH: "🟠", MEDIUM: "🟡", LOW: "🟢"}[.] // "⚪";
  def row: "| \(.severity | icon) \(.severity | esc) | \(link) | \(.pkg | code) | \(.installed | code) | \(.fixed | code) | \(.type | esc) |";
  def table:
    if length == 0 then "_none_"
    else
      ( "| Severity | Vulnerability | Package | Installed | Fixed | Type |",
        "|---|---|---|---|---|---|",
        (.[:$max_rows][] | row),
        (if length > $max_rows then "", "_… and \(length - $max_rows) more; see the [workflow run](\($run_url)) summary or raise `max-rows`._" else empty end)
      )
    end;
  def details($summary; $open):
    ("<details" + (if $open then " open" else "" end) + "><summary>\($summary)</summary>", "", table, "", "</details>");
  def requested($s): ($severity | ascii_upcase | split(",") | map(gsub(" "; "")) | index($s)) != null;
  def count_line:
    . as $s
    | [ (select(requested("CRITICAL") or $s.critical > 0) | "🔴 **\($s.critical)** Critical"),
        (select(requested("HIGH") or $s.high > 0)         | "🟠 **\($s.high)** High"),
        (select(requested("MEDIUM") or $s.medium > 0)     | "🟡 **\($s.medium)** Medium"),
        (select(requested("LOW") or $s.low > 0)           | "🟢 **\($s.low)** Low") ]
    | join(" · ")
    + (if $ignore_unfixed == "true" then " (fixed only)" else "" end);
  # Shorten a trailing hex tag/digest so the image ref does not dominate the footer.
  def short_ref: (capture("^(?<p>.*?)(?<h>[0-9a-f]{12,})$") | "\(.p)\(.h[:8])…") // .;

  "### Trivy image scan: \($image_name | code)",
  "",
  count_line,
  "",
  (if $mode == "reduced" then
      "> [!NOTE]",
      "> The findings table is too large for a comment; see the "
        + (if $run_url != "" then "[workflow run](\($run_url)) summary" else "job summary" end) + "."
   else
      (.total as $t | .findings | details("Findings (\($t))"; $t <= $open_rows))
   end),
  "",
  "<sub>"
    + ($image_ref | short_ref | code)
    + (if $trivy_version != "" then " · Trivy \($trivy_version)" else "" end)
    + (if $run_url != "" then " · [workflow run](\($run_url))" else "" end)
    + "</sub>",
  "<sub>Informational only; this check never fails the build.</sub>",
  "<!-- trivy-image-scan:\($slug) -->"
  ' "${summary_path}"
}

full_path="$(dirname "${REPORT_PATH}")/report.md"
render full > "${full_path}"
cat "${full_path}"

if [[ -n "${COMMENT_PATH}" ]]; then
  if [[ "$(wc -c < "${full_path}")" -le "${MAX_BYTES}" ]]; then
    cp "${full_path}" "${COMMENT_PATH}"
  else
    echo "::notice::Scan report exceeds ${MAX_BYTES} bytes; posting a reduced comment. The full table is in the job summary." >&2
    render reduced > "${COMMENT_PATH}"
  fi
fi

# Scan metadata that is not shown in the comment goes to the step log.
jq -r \
  --arg image_ref "${IMAGE_REF}" \
  --arg trivy_version "${TRIVY_BIN_VERSION}" \
  --arg sbom "$(if [[ "${SBOM_CACHE_HIT}" == "true" ]]; then echo cache; else echo generated; fi)" \
  '"image=\($image_ref) os=\(.os | if . == "" then "unknown" else . end) trivy=\($trivy_version) sbom=\($sbom) total=\(.total)"' \
  "${summary_path}" >&2
