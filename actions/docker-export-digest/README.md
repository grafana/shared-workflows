# docker-export-digest

This is a composite GitHub Action used to export a docker digest as a workflow artifact.

This can be used in conjunction with [docker-build-push-image] and [docker-import-digests-push-manifest] to build
native multi-arch Docker images.

[docker/build-push-action]: https://github.com/docker/build-push-action
[docker-build-push-image]: ../docker-build-push-image/README.md
[docker-export-digest]: ../docker-export-digest/README.md
[docker-import-digests-push-manifest]: ../docker-import-digests-push-manifest/README.md

<!-- x-release-please-start-version -->

```yaml
name: Build a Docker Image

on:
  push:
    branches:
      - main

jobs:
  upload-digest-as-artifact:
    permissions:
      contents: read
      id-token: write
    steps:
      - name: Export and upload digest
        uses: grafana/shared-workflows/actions/docker-export-digest@docker-export-digest/v0.1.2
        with:
          digest: ${{ steps.docker-build-push-image.outputs.digest }}
          platform: linux/arm64
```

<!-- x-release-please-end-version -->

## Inputs

| Name       | Type   | Description                                                                                                |
| ---------- | ------ | ---------------------------------------------------------------------------------------------------------- |
| `digest`   | String | Docker digest. This is included as an output for `docker-build-push-image` and `docker/build-push-action`. |
| `platform` | String | Docker platform, ex: linux/arm64.                                                                          |
