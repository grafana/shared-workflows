# push-to-gar-docker

This is a composite GitHub Action, used to push docker images to Google Artifact Registry (GAR).
It uses [OIDC authentication](https://docs.github.com/en/actions/deployment/security-hardening-your-deployments/about-security-hardening-with-openid-connect)
which means that only workflows which get triggered based on certain rules can
trigger these composite workflows.

```yaml
name: CI
on:
  pull_request:

env:
  ENVIRONMENT: "dev" # can be either dev/prod
  IMAGE_NAME: "backstage" # name of the image to be published, required

# These permissions are needed to assume roles from Github's OIDC.
permissions:
  contents: read
  id-token: write

jobs:
  build-and-push:
    runs-on: ubuntu-latest

    steps:
      - uses: grafana/shared-workflows/actions/push-to-gar-docker@main
        id: push-to-gar
        with:
          registry: "<YOUR-GAR>" # e.g. us-docker.pkg.dev, optional
          tags: |-
            "<IMAGE_TAG>"
            "latest"
          context: "<YOUR_CONTEXT>" # e.g. "." - where the Dockerfile is
```

## Inputs

| Name          | Type   | Description                                                                          |
|---------------|--------|--------------------------------------------------------------------------------------|
| `registry`    | String | Google Artifact Registry to store docker images in.                                  |
| `tags`        | List   | Tags that should be used for the image (see the [metadata-action][mda] for details)  |
| `context`     | List   | Path to the Docker build context.                                                    |
| `environment` | Bool   | Environment for pushing artifacts (can be either dev or prod).                       |
| `image_name`  | String | Name of the image to be pushed to GAR.                                               |
| `build-args`  | String | List of arguments necessary for the Docker image to be built.                        |
| `file`        | String | Path and filename of the dockerfile to build from. (Default: `{context}/Dockerfile`) |
