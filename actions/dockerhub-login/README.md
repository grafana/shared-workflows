# dockerhub-login

This is a composite GitHub Action that logs in to Docker Hub as the `grafana` organization.
It uses a [Docker Hub OIDC connection][oidc]: `docker/login-action` exchanges the job's GitHub OIDC token for a short-lived Docker Hub token. No secrets are stored or read.

The token:

- can only read public repositories (it cannot push or read private images)
- expires after 1 hour

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

## Other GitHub orgs

The default `connection-id` only works for repositories in the `grafana` GitHub org.
Repositories in other orgs need their own Docker Hub OIDC connection (ask Platform Productivity) and must pass its ID:

```yaml
- name: Login to DockerHub
  uses: grafana/shared-workflows/actions/dockerhub-login@<version>
  with:
    connection-id: <your org's connection ID>
```

## Inputs

<!-- BEGIN_INPUTS -->

| Name            | Type   | Required | Default                                | Description                                                                                                                                                    |
| --------------- | ------ | -------- | -------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `connection-id` | String | No       | `e14404e0-cb2e-419e-89da-f08af0347345` | Docker Hub OIDC connection ID. Defaults to the connection for the `grafana` GitHub org. Repositories in other GitHub orgs must pass their org's connection ID. |

<!-- END_INPUTS -->

[oidc]: https://docs.docker.com/security/authentication/oidc-connections/
