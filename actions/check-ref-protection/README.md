# check-ref-protection

Composite GitHub Action that gates prod publishing on **ref protection**. It
checks whether the ref a workflow is publishing from (a branch) is actually
protected — reviewed and tamper-resistant — and not merely a ref where the
`ref_protected` boolean happens to be `true`.

It reads the protection that applies to the ref and checks it against a minimum
bar defined in [`policy.json`](./policy.json).

> [!NOTE]
> **Branches only for now.** Tags are not yet supported — GitHub has no
> `/rules/tags/{tag}` endpoint, so tag protection needs a different approach
> (ruleset enumeration or a commit-ancestry check). The action exits with an
> error for tags rather than reporting a misleading result.

## How it works

The action merges **two sources** of protection into one normalized rule set:

1. **Rulesets** — `GET /repos/{repo}/rules/branches/{branch}` (org + repo).
2. **Legacy branch protection** — `GET /repos/{repo}/branches/{branch}/protection`,
   normalized into the same shape (e.g. legacy `dismiss_stale_reviews` →
   `dismiss_stale_reviews_on_push`, `allow_force_pushes: false` →
   `non_fast_forward`).

A check passes if **either** source satisfies it — GitHub enforces both and the
most-restrictive wins, so presence in either source means it is enforced. The
output reports which source(s) protect the ref, both at the top and per check.

## Behaviour

- `enforce: false` (default) — **disabled / warn-only.** Prints the result but
  never fails the step. Safe to add without changing a workflow's outcome.
- `enforce: true` — **blocking.** Fails (exit 1) when the ref does not meet the
  bar, so no token is minted and nothing is published.

<!-- x-release-please-start-version -->

```yaml
name: Publish
on:
  push:
    branches: [main, "release-*"]

jobs:
  gate:
    runs-on: ubuntu-latest
    permissions:
      contents: read # enough for the rulesets source
    steps:
      - uses: grafana/shared-workflows/actions/check-ref-protection@check-ref-protection/v0.1.0
        with:
          enforce: "true"
```

<!-- x-release-please-end-version -->

## Inputs

| Name           | Description                                                                           | Default                    |
| -------------- | ------------------------------------------------------------------------------------- | -------------------------- |
| `repository`   | `owner/repo` to check.                                                                | `${{ github.repository }}` |
| `ref-type`     | `branch` or `tag`. Auto-detected from `github.ref` when empty.                        | _(auto)_                   |
| `ref-name`     | Ref name to check. Auto-detected from `github.ref` when empty.                        | _(auto)_                   |
| `policy`       | Path to a `policy.json`. Defaults to the policy bundled with this action.             | _(bundled)_                |
| `enforce`      | `true` to fail on insufficient protection; `false` to warn-only.                      | `"false"`                  |
| `github-token` | Token with read access to the repo's rules.                                           | `${{ github.token }}`      |

## Tokens / permissions

- **Rulesets source** (`/rules/branches/...`) needs only `contents: read` — the
  default `GITHUB_TOKEN` works.
- **Legacy source** (`/branches/.../protection`) needs **Administration: read**.
  The default `GITHUB_TOKEN` cannot read it, so in CI the legacy source is
  skipped with a warning (a branch protected only by legacy rules would then
  look unprotected). To include the legacy source in CI, pass a GitHub App
  token (GATB) with Administration: read via `github-token`.

## Policy

`policy.json` lists the required rules per ref type. Each entry is a rule check:

| Field         | Meaning                                                  |
| ------------- | -------------------------------------------------------- |
| `id`          | Short identifier shown in the output.                    |
| `severity`    | `required` (fails the bar) or `optional` (warning only). |
| `description` | Human-readable description.                              |
| `ruleType`    | The GitHub ruleset rule `type` (e.g. `pull_request`).    |
| `param`       | Optional parameter inside the rule to inspect.           |
| `min`         | Require `param >= min` (numeric).                        |
| `equals`      | Require `param == equals` (exact match).                 |

To change the bar, edit `policy.json` — the script is generic and is not touched.
