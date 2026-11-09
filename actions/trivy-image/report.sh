#!/usr/bin/env bash
# Render a Trivy JSON report (and optional baseline report) as Markdown on stdout
# and write counts to $GITHUB_OUTPUT when set.
#
# Environment:
#   REPORT_PATH         Trivy JSON report for the image under test (required)
#   BASELINE_REPORT     Trivy JSON report for the baseline image (optional)
#   BASELINE_SOURCE     image | cache | artifact | none
#   BASELINE_IMAGE_REF  Baseline image reference when BASELINE_SOURCE=image
#   IMAGE_REF, IMAGE_NAME, SLUG, SEVERITY, IGNORE_UNFIXED, SBOM_CACHE_HIT,
#   BASELINE_REF, TRIVY_BIN_VERSION, MAX_ROWS
#   GITHUB_SERVER_URL, GITHUB_REPOSITORY, GITHUB_RUN_ID (runner defaults) for the run link
set -euo pipefail

: "${REPORT_PATH:?REPORT_PATH is required}"
BASELINE_REPORT="${BASELINE_REPORT:-}"
BASELINE_SOURCE="${BASELINE_SOURCE:-none}"
BASELINE_IMAGE_REF="${BASELINE_IMAGE_REF:-}"
IMAGE_REF="${IMAGE_REF:-unknown}"
IMAGE_NAME="${IMAGE_NAME:-${IMAGE_REF}}"
SLUG="${SLUG:-image}"
SEVERITY="${SEVERITY:-CRITICAL,HIGH}"
IGNORE_UNFIXED="${IGNORE_UNFIXED:-false}"
SBOM_CACHE_HIT="${SBOM_CACHE_HIT:-false}"
BASELINE_REF="${BASELINE_REF:-}"
TRIVY_BIN_VERSION="${TRIVY_BIN_VERSION:-}"
MAX_ROWS="${MAX_ROWS:-100}"

if ! [[ "${MAX_ROWS}" =~ ^[0-9]+$ ]]; then
  echo "::warning::max-rows must be a non-negative integer (got '${MAX_ROWS}'); using 100." >&2
  MAX_ROWS=100
fi

# A baseline report that is missing or malformed degrades to "no baseline".
if [[ -n "${BASELINE_REPORT}" ]] && jq -e . "${BASELINE_REPORT}" >/dev/null 2>&1; then
  baseline_arg=(--slurpfile base "${BASELINE_REPORT}")
else
  if [[ -n "${BASELINE_REPORT}" ]]; then
    echo "::warning::Baseline report is unreadable; showing results without a diff." >&2
  fi
  baseline_arg=(--argjson base '[]')
fi

run_url=""
if [[ -n "${GITHUB_SERVER_URL:-}" && -n "${GITHUB_REPOSITORY:-}" && -n "${GITHUB_RUN_ID:-}" ]]; then
  run_url="${GITHUB_SERVER_URL}/${GITHUB_REPOSITORY}/actions/runs/${GITHUB_RUN_ID}"
fi

summary_path="$(dirname "${REPORT_PATH}")/summary.json"

jq -n \
  --slurpfile pr "${REPORT_PATH}" \
  "${baseline_arg[@]}" \
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
  def key_set: map({key: .key, value: true}) | from_entries;

  ($pr[0] | vulns) as $cur
  | (if ($base | length) > 0 then ($base[0] | vulns) else null end) as $baseVulns
  | ($cur | key_set) as $curSet
  | (if $baseVulns == null then {} else ($baseVulns | key_set) end) as $baseSet
  | {
      os: ($pr[0].Metadata.OS | if . == null then "" else "\(.Family) \(.Name)" end),
      findings: $cur,
      critical: ($cur | count("CRITICAL")),
      high: ($cur | count("HIGH")),
      medium: ($cur | count("MEDIUM")),
      low: ($cur | count("LOW")),
      total: ($cur | length),
      fixable: ($cur | map(select(.fixed != "")) | length),
      baseline_found: ($baseVulns != null),
      introduced: (if $baseVulns == null then null else ($cur | map(select(.key as $k | ($baseSet | has($k)) | not))) end),
      resolved: (if $baseVulns == null then null else ($baseVulns | map(select(.key as $k | ($curSet | has($k)) | not))) end)
    }
  ' > "${summary_path}"

if [[ -n "${GITHUB_OUTPUT:-}" ]]; then
  jq -r '
    "critical=\(.critical)",
    "high=\(.high)",
    "medium=\(.medium)",
    "low=\(.low)",
    "total=\(.total)",
    "fixable=\(.fixable)",
    "baseline-found=\(.baseline_found)",
    "introduced=\(.introduced | if . == null then "" else length end)",
    "resolved=\(.resolved | if . == null then "" else length end)"
  ' "${summary_path}" >> "${GITHUB_OUTPUT}"
fi

jq -r \
  --arg image_ref "${IMAGE_REF}" \
  --arg image_name "${IMAGE_NAME}" \
  --arg slug "${SLUG}" \
  --arg severity "${SEVERITY}" \
  --arg ignore_unfixed "${IGNORE_UNFIXED}" \
  --arg baseline_ref "${BASELINE_REF}" \
  --arg baseline_source "${BASELINE_SOURCE}" \
  --arg baseline_image_ref "${BASELINE_IMAGE_REF}" \
  --arg trivy_version "${TRIVY_BIN_VERSION}" \
  --arg run_url "${run_url}" \
  --argjson max_rows "${MAX_ROWS}" \
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
  def vulns($n): if $n == 1 then "1 vulnerability" else "\($n) vulnerabilities" end;
  def requested($s): ($severity | ascii_upcase | split(",") | map(gsub(" "; "")) | index($s)) != null;
  def count_line:
    . as $s
    | [ (select(requested("CRITICAL") or $s.critical > 0) | "🔴 **\($s.critical)** Critical"),
        (select(requested("HIGH") or $s.high > 0)         | "🟠 **\($s.high)** High"),
        (select(requested("MEDIUM") or $s.medium > 0)     | "🟡 **\($s.medium)** Medium"),
        (select(requested("LOW") or $s.low > 0)           | "🟢 **\($s.low)** Low") ]
    | join(" · ")
    + (if $ignore_unfixed == "true" then " (fixed only)" else "" end);
  def baseline_label:
    if $baseline_source == "image" then "image \($baseline_image_ref | code)"
    else ($baseline_ref | code) + (if $baseline_source == "artifact" then " (baseline artifact)" else "" end)
    end;
  # Shorten a trailing hex tag/digest so the image ref does not dominate the footer.
  def short_ref: (capture("^(?<p>.*?)(?<h>[0-9a-f]{12,})$") | "\(.p)\(.h[:8])…") // .;

  "### Trivy image scan: \($image_name | code)",
  "",
  count_line,
  "",
  (if .baseline_found then
      (.introduced | length) as $n | (.resolved | length) as $r
      | (if $n > 0 then
          "> [!CAUTION]",
          "> **\(vulns($n)) introduced** compared to \(baseline_label)" + (if $r > 0 then ", \($r) resolved" else "" end) + "."
        elif $r > 0 then
          "> [!TIP]",
          "> **\(vulns($r)) resolved** compared to \(baseline_label); none introduced."
        else
          "> [!TIP]",
          "> No vulnerabilities introduced or resolved compared to \(baseline_label)."
        end),
        "",
        (if $n > 0 then (.introduced | details("Introduced (\($n))"; true)), "" else empty end),
        (if $r > 0 then (.resolved | details("Resolved (\($r))"; false)), "" else empty end)
   elif $baseline_ref != "" or $baseline_image_ref != "" then
      "> [!NOTE]",
      (if $baseline_ref != "" then
        "> No baseline for \($baseline_ref | code) yet; one is saved on the next push to \($baseline_ref | code)."
       else
        "> No baseline available yet."
       end),
      ""
   else empty
   end),
  (.total as $t | .findings | details("All findings (\($t))"; false)),
  "",
  "<sub>"
    + ($image_ref | short_ref | code)
    + (if $trivy_version != "" then " · Trivy \($trivy_version)" else "" end)
    + (if $run_url != "" then " · [workflow run](\($run_url))" else "" end)
    + "</sub>",
  "<sub>Informational only; this check never fails the build.</sub>",
  "<!-- trivy-image-scan:\($slug) -->"
  ' "${summary_path}"

# Scan metadata that is not shown in the comment goes to the step log.
jq -r \
  --arg image_ref "${IMAGE_REF}" \
  --arg trivy_version "${TRIVY_BIN_VERSION}" \
  --arg sbom "$(if [[ "${SBOM_CACHE_HIT}" == "true" ]]; then echo cache; else echo generated; fi)" \
  --arg baseline_source "${BASELINE_SOURCE}" \
  '"image=\($image_ref) os=\(.os | if . == "" then "unknown" else . end) trivy=\($trivy_version) sbom=\($sbom) baseline=\($baseline_source) total=\(.total)"' \
  "${summary_path}" >&2
