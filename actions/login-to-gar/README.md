# login-to-gar

This is a composite GitHub Action, used to login to Google Artifact Registry
(GAR). It uses [OIDC authentication], which means that only workflows which get
triggered based on certain rules can trigger these composite workflows.

> [!WARNING]
> There is a bug with Workload Identity Federation that prevents docker images from pushing successfully if the
> branch name is too long. This is due a max length with Workload Identity Federation claims on Google's side.
>
> Should you use this action and get a `400` error when pushing the image... try shortening the branch name.

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
      - uses: grafana/shared-workflows/actions/login-to-gar@login-to-gar/v1.0.1
        id: login-to-gar
        with:
          registry: "<YOUR-GAR>" # e.g. us-docker.pkg.dev
```

<!-- x-release-please-end-version -->

[OIDC authentication]: https://docs.github.com/en/actions/deployment/security-hardening-your-deployments/about-security-hardening-with-openid-connect

## Inputs

| Name                    | Description                                                                                                              | Default             |
| ----------------------- | ------------------------------------------------------------------------------------------------------------------------ | ------------------- |
| `registry`              | Google Artifact Registry to authenticate against.                                                                        | `us-docker.pkg.dev` |
| `workspace_credentials` | Whether to place the GCP credentials file in the workspace. Off by default. See [Docker Actions Compatibility] for more. | `false`             |

[Docker Actions Compatibility]: #docker-actions-compatibility

## Docker Actions Compatibility

By default, this action stores credentials in a temporary location outside of
your workspace to prevent accidental commits or exposure. This is a
security-first approach that deviates from the upstream
`google-github-actions/auth` action's default behavior.

The upstream action places credentials in the workspace by design to ensure
Docker-based GitHub Actions work transparently. However, this comes with
significant security risks: active credentials in the working directory can be
easily accidentally printed in logs or committed to repositories,

If you need to use Docker-based GitHub Actions that require workspace access to
credentials files, set `workspace_credentials: true`.

> [!CAUTION]
> When using `workspace_credentials: true`:
>
> 1. [Always place the `checkout` action before `login-to-gar` to prevent
>    credentials from being overwritten][checkout-before-login].
> 2. Add `gha-creds-*.json` to your `.gitignore` and `.dockerignore` files to
>    prevent accidentally committing credentials
> 3. Be extra careful with any steps that commit changes, as they may
>    inadvertently include credential files

[checkout-before-login]: https://github.com/google-github-actions/auth/blob/0920706a19e9d22c3d0da43d1db5939c6ad837a8/README.md#prerequisites
