# validate-zizmor-config

Validates a repo-local [zizmor](https://docs.zizmor.sh/) configuration file (`zizmor.yml` or `.github/zizmor.yml`) against Grafana policy before running zizmor.

Intended to be called from [`.github/workflows/reusable-zizmor.yml`](../../.github/workflows/reusable-zizmor.yml), but usable as a standalone validation step.

## Inputs

<!-- BEGIN_INPUTS -->

| Name           | Type   | Required | Default | Description                                                                                                                                         |
| -------------- | ------ | -------- | ------- | --------------------------------------------------------------------------------------------------------------------------------------------------- |
| `config_path`  | String | Yes      |         | Path to the zizmor config file relative to the workspace (e.g. zizmor.yml).                                                                         |
| `sarif_output` | String | No       |         | Optional path for a SARIF report written when validation fails (used by reusable-zizmor so Grafana Bench and Loki ingestion still receive results). |

<!-- END_INPUTS -->

## Requirements

The calling job must run [`astral-sh/setup-uv`](https://github.com/astral-sh/setup-uv) (or otherwise provide `uv`) before this action, and the workspace must contain the file at `config_path`.

## Example workflow

<!-- x-release-please-start-version -->

```yaml
name: Validate zizmor config

on:
  pull_request:
    paths:
      - "zizmor.yml"
      - ".github/zizmor.yml"
  push:
    branches:
      - main
    paths:
      - "zizmor.yml"
      - ".github/zizmor.yml"

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          persist-credentials: false
      - name: Set up uv
        uses: astral-sh/setup-uv@v6
      - name: Validate zizmor config
        uses: grafana/shared-workflows/actions/validate-zizmor-config@validate-zizmor-config/v0.2.1
        with:
          config_path: .github/zizmor.yml
```

<!-- x-release-please-end-version -->

## Tests

From the repository root:

```bash
cd actions/validate-zizmor-config && uv run --with pyyaml==6.0.3 python3 -m unittest discover -v
```

CI: [`.github/workflows/test-validate-zizmor-config.yml`](../../.github/workflows/test-validate-zizmor-config.yml).
