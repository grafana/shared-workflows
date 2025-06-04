# login-to-gar

This is a composite GitHub Action, used to login to Google Artifact Registry (GAR).
It uses [OIDC authentication](https://docs.github.com/en/actions/deployment/security-hardening-your-deployments/about-security-hardening-with-openid-connect)
which means that only workflows which get triggered based on certain rules can trigger these composite workflows.

<!-- x-release-please-start-version -->

```yaml
name: CI
on:
  pull_request:

jobs:
  login:
    runs-on: ubuntu-latest
    # These permissions are needed to assume roles from Github's OIDC.
    permissions:
      contents: read
      id-token: write
    steps:
      - uses: grafana/shared-workflows/actions/login-to-gar@login-to-gar/v0.4.2
        id: login-to-gar
        with:
          registry: "<YOUR-GAR>" # e.g. us-docker.pkg.dev
          environment: "prod" # can be either dev/prod
```

<!-- x-release-please-end-version -->

## Inputs

| Name                      | Description                                                                                                                             | Default             |
| ------------------------- | --------------------------------------------------------------------------------------------------------------------------------------- | ------------------- |
| `registry`                | Google Artifact Registry to authenticate against.                                                                                       | `us-docker.pkg.dev` |
| `environment`             | Environment for pushing artifacts (can be either dev or prod).                                                                          | `dev`               |
| `delete_credentials_file` | Delete the credentials file after the action is finished. If you want to keep the credentials file for a later step, set this to false. | `false`             |
