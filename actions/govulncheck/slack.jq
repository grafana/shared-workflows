# Builds the Slack message payload. Run with --null-input, --compact-output,
# --arg channel/repo/ref/url and --argjson model (the output of model.jq).
[$model.vulns[] | select(.level == "called")] as $called
| ($called | map("<https://pkg.go.dev/vuln/\(.id)|\(.id)> in `\(.module)` (\(if .fixed == "" then "no fix available" else "fixed in \(.fixed)" end))")) as $lines
| "govulncheck found \($called | length) vulnerabilit\(if ($called | length) == 1 then "y" else "ies" end)" as $headline
| ($lines[:10] | map("• " + .) | join("\n")) as $list
| (if ($lines | length) > 10 then "\n_and \(($lines | length) - 10) more_" else "" end) as $rest
| (if $ref != "" then " (`\($ref)`)" else "" end) as $branch
| {
    channel: $channel,
    text: "\($headline) in \($repo)",
    blocks: [
        {
            type: "section",
            text: {
                type: "mrkdwn",
                text: ":rotating_light: *\($headline)* in `\($repo)`\($branch)\n\($list)\($rest)"
            }
        },
        {
            type: "context",
            elements: [{ type: "mrkdwn", text: "<\($url)|View the workflow run>" }]
        }
    ]
}
