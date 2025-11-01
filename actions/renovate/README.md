# renovate

Run Renovate in Grafana repositories in order to update dependencies.

A configuration file in the repository is required to run Renovate. The
configuration file should be named `renovate.json` and should be placed in the
`.github` directory.

A second configuration file is needed for the Renovate App. The default
location for this file is `.github/renovate-app.json`. This is an example:

```json
{
  "customEnvVariables": {
    "GOPRIVATE": "github.com/grafana",
    "GONOSUMDB": "github.com/grafana",
    "GONOPROXY": "github.com/grafana"
  }
}
```

This action won't work for non-grafana repositories because it requires access to the Grafana Labs vault instance.

# Example workflow:

The following example runs Renovate every 8 hours.

```yaml
name: Renovate

on:
  schedule:
    - cron:  '47 */8 * * *'

jobs:
  renovate:
    permissions:
      contents: read
      id-token: write
    runs-on: ubuntu-latest
    timeout-minutes: 5
    steps:
      - id: run-renovate
        uses: grafana/shared-workflows/actions/renovate
```
