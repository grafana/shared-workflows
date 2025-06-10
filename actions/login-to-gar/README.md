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
      - uses: grafana/shared-workflows/actions/login-to-gar@login-to-gar/v0.4.3
        id: login-to-gar
        with:
          registry: "<YOUR-GAR>" # e.g. us-docker.pkg.dev
```

<!-- x-release-please-end-version -->

## Inputs

| Name                      | Description                                                                                                                             | Default             |
| ------------------------- | --------------------------------------------------------------------------------------------------------------------------------------- | ------------------- |
| `registry`                | Google Artifact Registry to authenticate against.                                                                                       | `us-docker.pkg.dev` |
| `delete_credentials_file` | Delete the credentials file after the action is finished. If you want to keep the credentials file for a later step, set this to false. | `false`             |

> [!WARNING]
> 1. When using the `login-to-gar` action in GitHub Actions workflows, always place the `checkout` action before it. This is because the `login-to-gar` action stores Docker credentials in the workspace, and these credentials would be lost if the workspace is overwritten by a subsequent `checkout` action. The correct order makes sure that the Docker credentials persist throughout the workflow.
> 2. Add `gha-creds-*.json` to your `.gitignore` and `.dockerignore` files to prevent accidentally committing credentials to your artifacts. ([source](https://github.com/google-github-actions/auth/blob/0920706a19e9d22c3d0da43d1db5939c6ad837a8/README.md#prerequisites))
