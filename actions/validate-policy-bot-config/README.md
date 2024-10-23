# validate-policy-bot-config

Generate an SPDX SBOM Report and attached to Release Artifcats on Release Publish

Example workflow:

```yaml
name: validate-policy-bot
on:
  pull_request:
    paths:
      - ".policy.yml"
  push:
    paths:
      - ".policy.yml
jobs:
  validate-policy-bot:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
      - name: Validate Policy Bot configuration
        uses: grafana/shared-workflows/actions/validate-policy-bot-config@main
```
