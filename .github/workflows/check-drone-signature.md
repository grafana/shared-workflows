# check-drone-signature

This is a reusable workflow that verifies the signature on a repository's Drone CI
configuration file is valid, so that an unsigned or tampered `.drone.yml` cannot
reach the Drone server.

The signature check is skipped for forks, because the secrets needed to validate
it are not available to forked repositories.

```yaml
name: Check Drone signature

on: pull_request

jobs:
  check-drone-signature:
    uses: grafana/shared-workflows/.github/workflows/check-drone-signature.yaml@main
```

## Inputs

<!-- BEGIN_INPUTS -->

| Name                | Type   | Required | Default                     | Description                             |
| ------------------- | ------ | -------- | --------------------------- | --------------------------------------- |
| `drone_config_path` | string | No       | `.drone.yml`                | Path to the Drone CI configuration file |
| `drone_server`      | string | No       | `https://drone.grafana.net` | Drone CI server URL                     |

<!-- END_INPUTS -->
