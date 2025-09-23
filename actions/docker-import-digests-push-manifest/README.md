# docker-import-digest-push-manifest

This is a composite GitHub Action used to import Docker digests from a shared workflow artifact and merge them into a
tagged manifest.

This is meant to work in conjunction with [docker-build-push-image] and [docker-export-digest].

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
  build-push-image:
    outputs:
      images: ${{ steps.build.outputs.images }}
    permissions:
      contents: read
      id-token: write
    steps:
      - name: Build Docker Image
        id: build
        uses: grafana/shared-workflows/actions/docker-build-push-image@main # TODO: Fix version once released
        with:
          platforms: linux/arm64
          tags: |
            ${{ github.sha }}
            main
          push: true
          registries: "gar,dockerhub"
          include-tags-in-push: false
          outputs: "type=image,push-by-digest=true,name-canonical=true,push=true"
      - name: Export and upload digest
        uses: grafana/shared-workflows/actions/docker-export-digest@rwhitaker/multi-arch-builds
        with:
          digest: ${{ steps.build.outputs.digest }}
          platform: linux/arm64
  merge-digest:
    if: ${{ inputs.push == 'true' }}
    runs-on: ubuntu-arm64-small
    needs: build-and-push
    permissions:
      contents: read
      id-token: write
    steps:
      - name: Download Multi-Arch Digests, Construct and Upload Manifest
        uses: grafana/shared-workflows/actions/docker-import-digests-push-manifest@main # TODO: Pin sha
        with:
          images: ${{ needs.build-push-image.outputs.images }}
          gar-environment: 'dev'
          registries: "gar,dockerhub"
          docker-metadata-json: ${{ needs.build-and-push.outputs.metadatajson }}
```

<!-- x-release-please-end-version -->

## Inputs

| Name                   | Type   | Description                                                                                                                                                         |
|------------------------|--------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `docker-metadata-json` | String | Docker metadata JSON, from `docker-build-push-image` or `docker/build-push-action`.                                                                                 |
| `gar-environment`      | String | Environment for pushing artifacts (can be either dev or prod). This sets the GAR Project to either `grafanalabs-dev` or `grafanalabs-global`.                       |
| `images`               | String | CSV of Docker images to push. These images should not include tags. Ex: us-docker.pkg.dev/grafanalabs-dev/gar-registry/image-name,docker.io/grafana/dockerhub-image |
