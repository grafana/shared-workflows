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
  ROOT_REPO: ${{ github.event.repository.name }} # must be explicitly set like this - composite actions cannot get callers' payload, required
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
          context: "<YOUR_BUILD_PATH>" # e.g. "." - where the Dockerfile is
```
