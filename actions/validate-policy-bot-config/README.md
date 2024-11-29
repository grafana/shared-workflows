# validate-policy-bot-config

Validates the `.policy.yml` configuration file for [Policy Bot](https://github.com/palantir/policy-bot).
See [https://github.com/palantir/policy-bot?tab=readme-ov-file#configuration](Policy Bots' documentation) for more informations.

## Inputs

- `validation_endpoint`: The endpoint to validate the configuration against. Defaults to `https://policy-bot.grafana.net/api/v1/validate`.

Example workflow:

<!-- x-release-please-start-version -->

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
        uses: grafana/shared-workflows/actions/validate-policy-bot-config@validate-policy-bot-config-v1.0.0
```

<!-- x-release-please-end-version -->
