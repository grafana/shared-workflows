# zizmor-collection-paths

Composite action used by [`reusable-zizmor.yml`](../../.github/workflows/reusable-zizmor.yml) when a repo has [`.github/zizmor-collection-ignore`](../../.github/workflows/reusable-zizmor.md#skipping-vendored-workflow-trees-security-appsec326).

Runs from the Actions cache (`GITHUB_ACTION_PATH`); nothing is checked out into the caller workspace.

```bash
cd actions/zizmor-collection-paths && python3 -m unittest discover -v
```
