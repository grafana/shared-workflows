# check-ref-protection

Composite action that gates prod publishing on **ref protection**. It verifies
the ref a workflow publishes from is actually protected (reviewed and
tamper-resistant) by checking the protection that applies to it against [`policy.json`](./policy.json).

- **Branches** merge ruleset rules and legacy branch protection; a check passes
  if either satisfies it.
- **Tags** evaluate the **active** tag rulesets matching the tag (rulesets in
  `evaluate`/`disabled` mode are reported but never counted).

The ref identity comes from a trusted source (runner env or the signed OIDC
token), never caller inputs, so the gate can't be pointed at a different ref.

<!-- x-release-please-start-version -->

```yaml
jobs:
  gate:
    runs-on: ubuntu-latest
    permissions:
      contents: read
    steps:
      - uses: grafana/shared-workflows/actions/check-ref-protection@check-ref-protection/v0.1.0
        with:
          enforce: "true" # false (default) = warn-only, never blocks
```

<!-- x-release-please-end-version -->

## Inputs

| Name           | Description                                                      | Default                |
| -------------- | ---------------------------------------------------------------- | ---------------------- |
| `enforce`      | `true` to fail on insufficient protection; `false` = warn-only. | `"false"`              |
| `identity`     | Ref identity source: `env` or `oidc`.                           | `"env"`                |
| `policy`       | Path to a `policy.json`. Defaults to the bundled policy.        | _(bundled)_            |
| `github-token` | Token with read access to the repo's rules.                     | `${{ github.token }}`  |

## Permissions

- Branch rulesets need only `contents: read`.
- Legacy branch protection and tag ruleset enumeration need **Administration:
  read** — otherwise those sources are skipped (fail-closed) with a warning;
  pass a GATB app token via `github-token` to include them.
- `identity: oidc` also requires `id-token: write` on the job.

## Policy

Edit [`policy.json`](./policy.json) to change the bar — the Go code is generic.
Each entry has: `id`, `severity` (`required`|`optional`), `description`,
`ruleType`, and an optional match — `param` with `min` (`param >= min`) or
`equals` (`param == equals`), or presence-only when omitted.
