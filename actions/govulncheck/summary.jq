# Renders the markdown report from the model built by model.jq. Run with
# --raw-output and --argjson max <maximum number of rows per table>.

# Text coming from the vulnerability database is stripped of characters that
# would break out of a table cell.
def cell: gsub("[\r\n]+"; " ") | gsub("\\|"; "\\|");
def link: "[\(.id)](https://pkg.go.dev/vuln/\(.id))";
def ids: if (.aliases | length) > 0 then " (\(.aliases | join(", ") | cell))" else "" end;
def fixedin: if .fixed == "" then "no fix available" else .fixed end;
def more($rows): if ($rows | length) > $max then "\n_\(($rows | length) - $max) more not shown, see the job log._\n" else "" end;

[.vulns[] | select(.level == "called")] as $called
| [.vulns[] | select(.level != "called")] as $other
| "## govulncheck\n",
  (if ($called | length) == 0
   then "No vulnerabilities found in code that is actually called.\n"
   else
       "**\($called | length) vulnerabilit\(if ($called | length) == 1 then "y" else "ies" end) found** in code that is called.\n",
       "| Vulnerability | Module | Found | Fixed in | Vulnerable symbol | Called from | Summary |",
       "| --- | --- | --- | --- | --- | --- | --- |",
       ($called[:$max][] |
           "| \(link)\(ids) | `\(.module)` | `\(.version)` | \(fixedin) | `\(.symbol)` | `\(.callsite)` | \(.summary | cell) |"),
       more($called)
   end),
  (if ($other | length) > 0
   then
       "<details><summary>\($other | length) more in dependencies, not called by this code</summary>\n",
       "| Vulnerability | Module | Found | Fixed in | Level |",
       "| --- | --- | --- | --- | --- |",
       ($other[:$max][] |
           "| \(link) | `\(.module)` | `\(.version)` | \(fixedin) | \(.level) |"),
       more($other),
       "\n</details>\n"
   else empty
   end),
  (.config as $c
   | if ($c.scanner_version // "") != ""
     then "<sub>govulncheck \($c.scanner_version), \($c.go_version // "unknown Go"), database updated \($c.db_last_modified // "unknown")</sub>"
     else empty
     end)
