# zizmor-collection-paths

Used by [`.github/workflows/reusable-zizmor.yml`](../../.github/workflows/reusable-zizmor.yml) when a repo has [`.github/zizmor-collection-ignore`](../../.github/workflows/reusable-zizmor.md#skipping-vendored-workflow-trees-security-appsec326).

Collects workflow and composite-action paths under the checked-out repo root, skipping directory prefixes listed in the ignore file. Scripts run from the OIDC-pinned `shared-workflows` checkout (`_shared-workflows-zizmor`), not from the repo being scanned.

## Tests

```bash
cd actions/zizmor-collection-paths && python3 -m unittest discover -v
```
