# dockerhub-login

This is a composite GitHub Action that logs in to Docker Hub as the `grafana` organization.
It uses a [Docker Hub OIDC connection][oidc]: `docker/login-action` exchanges the job's GitHub OIDC token for a short-lived Docker Hub token. No secrets are stored or read.

The token:

- can only read public repositories (it cannot push or read private images)
- expires after 30 minutes

The job needs `id-token: write`. The action fails on `pull_request_target` events.

Example of how to use this action in a repository:

<!-- x-release-please-start-version -->

```yaml
name: Pull from DockerHub
on:
  pull_request:

jobs:
  build:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      id-token: write
    steps:
      - name: Login to DockerHub
        uses: grafana/shared-workflows/actions/dockerhub-login@dockerhub-login/v2.0.0
      - name: Build
        run: make build
```

<!-- x-release-please-end-version -->

[oidc]: https://docs.docker.com/security/authentication/oidc-connections/
