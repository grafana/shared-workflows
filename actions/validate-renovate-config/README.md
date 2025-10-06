# validate-renovate-config

Validates Renovate configuration files using [renovate-config-validator](https://docs.renovatebot.com/config-validation/).

## Inputs

- `path`: Path to the Renovate config file to validate. Defaults to `renovate.json`.

## Example workflow

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
        uses: grafana/shared-workflows/actions/validate-renovate-config@validate-renovate-config/v1.0.0
```

## Validating multiple files

To validate multiple config files, call the action multiple times:

```yaml
- name: Validate main config
  uses: grafana/shared-workflows/actions/validate-renovate-config@validate-renovate-config/v1.0.0
  with:
    path: renovate.json

- name: Validate preset
  uses: grafana/shared-workflows/actions/validate-renovate-config@validate-renovate-config/v1.0.0
  with:
    path: presets/default.json
```
