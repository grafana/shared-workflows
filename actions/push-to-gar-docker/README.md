# push-to-gar-docker

This is a composite GitHub Action, used to push docker images to Google Artifact Registry (GAR).
It uses [OIDC authentication](https://docs.github.com/en/actions/deployment/security-hardening-your-deployments/about-security-hardening-with-openid-connect)
which means that only workflows which get triggered based on certain rules can
trigger these composite workflows.

```yaml
name: CI
on:
  pull_request:

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
          image_name: "backstage" # name of the image to be published, required
          environment: "dev" # can be either dev/prod
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
| `platforms`   | List   | List of platforms the image should be built for (e.g. `linux/amd64,linux/arm64`)     |
| `cache-from`   | String   | Where cache should be fetched from ([more about GHA and container caching](https://www.kenmuse.com/blog/implementing-docker-layer-caching-in-github-actions/))    |
| `cache-to`   | String   | Where cache should be stored to ([more about GHA and container caching](https://www.kenmuse.com/blog/implementing-docker-layer-caching-in-github-actions/))   |
