# validate-policy-bot-config

Validates the `.policy.yml` configuration file for [Policy Bot](https://github.com/palantir/policy-bot).

See [Policy Bot's documentation](https://github.com/palantir/policy-bot?tab=readme-ov-file#configuration) for more information.

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
      - ".policy.yml"

jobs:
  validate-policy-bot:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          persist-credentials: false
      - name: Validate Policy Bot configuration
        uses: grafana/shared-workflows/actions/validate-policy-bot-config@validate-policy-bot-config/v1.1.2
```

<!-- x-release-please-end-version -->
