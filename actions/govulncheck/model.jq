# Condenses the govulncheck message stream (config, progress, osv, finding,
# ...) into one record per vulnerability. Run with --slurp.
#
# A vulnerability is "called" when at least one of its findings reaches a
# vulnerable symbol, which is the same rule govulncheck uses to decide on its
# exit code in text mode. Otherwise it is only imported (package-level finding)
# or required (module-level finding).
(map(select(.config) | .config) | first // {}) as $config
| (map(select(.osv) | .osv) | INDEX(.id)) as $osv
| (map(select(.finding) | .finding) | group_by(.osv)) as $groups
| {
    config: $config,
    vulns: [
        $groups[] | . as $findings
        | ([$findings[] | select(.trace[0].function != null)] | first) as $called
        | $findings[0].osv as $id
        | {
            id: $id,
            summary: ($osv[$id].summary // ""),
            aliases: ($osv[$id].aliases // []),
            module: $findings[0].trace[0].module,
            version: ($findings[0].trace[0].version // ""),
            fixed: ($findings[0].fixed_version // ""),
            level: (
                if $called != null then "called"
                elif any($findings[]; .trace[0].package != null) then "imported"
                else "required"
                end
            ),
            symbol: (
                if $called != null
                then ($called.trace[0].package + "." + $called.trace[0].function)
                else ""
                end
            ),
            callsite: (
                if $called != null
                then (
                    $called.trace[-1] as $frame
                    | ($frame.function // "") as $fn
                    | if $frame.position != null
                      then "\($fn) (\($frame.position.filename):\($frame.position.line))"
                      else $fn
                      end
                )
                else ""
                end
            )
        }
    ]
}
