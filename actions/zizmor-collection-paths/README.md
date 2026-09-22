# zizmor-collection-paths

Composite action used by [`reusable-zizmor.yml`](../../.github/workflows/reusable-zizmor.yml) when a repo has [`.github/zizmor-collection-ignore`](../../.github/workflows/reusable-zizmor.md#skipping-vendored-workflow-trees-security-appsec326).

Runs from the Actions cache (`GITHUB_ACTION_PATH`); nothing is checked out into the caller workspace.

```bash
cd actions/zizmor-collection-paths && python3 -m unittest discover -v
```

## Outputs

<!-- BEGIN_OUTPUTS -->

| Name                 | Description                                                                           |
| -------------------- | ------------------------------------------------------------------------------------- |
| `helper_root`        | Directory containing collection_paths.py and run_zizmor.py (GITHUB_ACTION_PATH).      |
| `paths_list`         | Path to the newline-separated explicit inputs file (when use_explicit_paths is true). |
| `use_explicit_paths` | true when ignore prefixes are active; false to scan .                                 |

<!-- END_OUTPUTS -->
