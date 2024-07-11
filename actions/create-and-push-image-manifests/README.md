# create-and-push-image-manifests

This is a composite GitHub Action, used to create and push image manifests to GAR.

```yaml
name: CI
on:
  pull_request:

# These permissions are needed to assume roles from Github's OIDC.
permissions:
  contents: read
  id-token: write

jobs:
  push-manifest:
    runs-on: ubuntu-x64-small
    steps:
      - name: Checkout
        env:
          action_repo: ${{ github.action_repository }}
          action_ref: ${{ github.action_ref }}
        uses: actions/checkout@b4ffde65f46336ab88eb53be808477a3936bae11 # v4.1.1
        with:
          repository: ${{ env.action_repo }}
          ref: ${{ env.action_ref }}
      - name: Create and push image manifests
        uses: grafana/shared-workflows/actions/create-and-push-image-manifests@main
        with:
          full-image-name: <FULL_IMAGE_NAME>
          tag: <TAG>
          environment: <ENVIRONMENT>
```

## Inputs

| Name              | Type   | Description                                                                                                                 |
| ----------------- | ------ | --------------------------------------------------------------------------------------------------------------------------- |
| `full-image-name` | String | Full image name for docker image, e.g. `us-docker.pkg.dev/grafanalabs-dev/docker-grafana-enterprise-dev/grafana-enterprise` |
| `tag`             | String | Tag for the image you want to push                                                                                          |
| `environment`     | String | Environment for pushing artifacts (can be either dev or prod).                                                              |
