# validate-renovate-config

Validates JSON and JSON5 Renovate configuration files using [renovate-config-validator](https://docs.renovatebot.com/config-validation/) and rejects top-level `enabled: false`, `force.enabled: false`, or direct references to the `:disableRenovate` preset. Scoped disabling through package rules or update-type settings is allowed. The policy checks the supplied file; it does not resolve inherited presets or detect catch-all disabling rules.

## Inputs

<!-- BEGIN_INPUTS -->

| Name   | Type   | Required | Default         | Description                                                          |
| ------ | ------ | -------- | --------------- | -------------------------------------------------------------------- |
| `path` | String | No       | `renovate.json` | Path to the Renovate config file to validate (e.g., 'renovate.json') |

<!-- END_INPUTS -->

## Example workflow

<!-- x-release-please-start-version -->

```yaml
name: Validate Renovate Config

on:
  pull_request:
    paths:
      - "renovate.json"
  push:
    branches:
      - main
    paths:
      - "renovate.json"

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          persist-credentials: false
      - name: Validate Renovate Config
        uses: grafana/shared-workflows/actions/validate-renovate-config@validate-renovate-config/v0.1.4
```

## Validating multiple files

To validate multiple config files, call the action multiple times:

```yaml
- name: Validate main config
  uses: grafana/shared-workflows/actions/validate-renovate-config@validate-renovate-config/v0.1.4
  with:
    path: renovate.json

- name: Validate preset
  uses: grafana/shared-workflows/actions/validate-renovate-config@validate-renovate-config/v0.1.4
  with:
    path: presets/default.json
```

<!-- x-release-please-end-version -->
