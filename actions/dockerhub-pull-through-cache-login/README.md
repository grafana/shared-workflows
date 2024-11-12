# dockerhub-pull-through-cache-login

This is a composite GitHub Action, used to login to Grafana's Docker Hub pull
through cache.

It uses [OIDC
authentication](https://docs.github.com/en/actions/deployment/security-hardening-your-deployments/about-security-hardening-with-openid-connect)
which means that only workflows which run in the `grafana` GitHub organization
can use this action.

```yaml
name: CI
on:
  pull_request:

# These permissions are needed to assume roles from Github's OIDC.
permissions:
  contents: read
  id-token: write

jobs:
  login:
    runs-on: ubuntu-latest

    steps:
      - uses: grafana/shared-workflows/actions/dockerhub-pull-through-cache-login@main
      # Prefix your image with `us-docker.pkg.dev/ops-tools-1203/dockerhub-proxy`
      - run: docker pull us-docker.pkg.dev/ops-tools-1203/dockerhub-proxy/grafana/grafana:latest
```
