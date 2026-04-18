# validate-zizmor-config

Composite action that enforces Grafana policy on a **repo-local** `zizmor.yml` / `.github/zizmor.yml` before running zizmor.

Intended to be called from [`.github/workflows/reusable-zizmor.yml`](../../.github/workflows/reusable-zizmor.yml) with a **full commit SHA** pin (see `uses: grafana/shared-workflows/actions/validate-zizmor-config@<sha>` there); update that SHA when you change this action so it stays aligned with the revision consumers use for the reusable workflow.

## Inputs

| Name             | Required | Description                                      |
| ---------------- | -------- | ------------------------------------------------ |
| `config_path`    | yes      | Path to the config file relative to the workspace |
| `pyyaml_version` | no       | PyYAML version for `uv run --with` (default 6.0.2) |

## Requirements

The calling job must run **`setup-uv`** (or otherwise provide `uv`) before this action, and the workspace must contain the file at `config_path`.

## Tests

From the repository root:

```bash
cd actions/validate-zizmor-config && uv run --with pyyaml==6.0.2 python3 -m unittest discover -v
```

CI: [`.github/workflows/test-validate-zizmor-config.yml`](../../.github/workflows/test-validate-zizmor-config.yml).
