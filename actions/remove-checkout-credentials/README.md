# remove-checkout-credentials

This action can be used in combination with `actions/checkout` and removes credentials set by that action.
For `actions/checkout` it is recommended to pass the `persist-credentials: false` setting but that might not be possible in various setups where the credentials _are_ needed for at least another action.
`remove-checkout-credentials` is exactly for those cases, to act as cleanup.

## Example

```yaml
name: CI
on:
  pull_request: {}

jobs:
  build:
    runs-on: ubuntu-latest
    permissions:
      contents: read
    steps:
      - uses: actions/checkout@v4
        with:
          persist-credentials: true

      # Actions that rely on the credentials within .git/config
      # ...

      - uses: grafana/shared-workflows/actions/remove-checkout-credentials@main

      # Actions that do not need the credentials anymore
      # ...
```
