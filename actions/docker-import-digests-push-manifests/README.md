# docker-import-digest-push-manifest

This is a composite GitHub Action used to import Docker digests from a shared workflow artifact and merge them into a
tagged manifest.

This can be used in conjunction with [docker-build-push-image] and [docker-export-digest] to build
native multi-arch Docker images.

[docker/build-push-action]: https://github.com/docker/build-push-action
[docker-build-push-image]: ../docker-build-push-image/README.md
[docker-export-digest]: ../docker-export-digest/README.md
[docker-import-digests-push-manifests]: /README.md

<!-- x-release-please-start-version -->

```yaml
name: Build a Docker Image

on:
  push:
    branches:
      - main

jobs:
  import-and-merge-digest:
    permissions:
      contents: read
      id-token: write
    steps:
      - name: Download Multi-Arch Digests, Construct and Upload Manifest
        uses: grafana/shared-workflows/actions/docker-import-digests-push-manifests@docker-import-digests-push-manifest/v0.0.0
        with:
          docker-metadata-json: ${{ needs.docker-build-push-image.outputs.metadatajson }}
          gar-environment: "dev"
          images: ${{ needs.docker-build-push-image.outputs.images }}
          push: true
```

<!-- x-release-please-end-version -->

## Inputs

| Name                   | Type    | Description                                                                                                                                                         |
| ---------------------- | ------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `docker-metadata-json` | String  | Docker metadata JSON, from `docker-build-push-image` or `docker/build-push-action`.                                                                                 |
| `gar-environment`      | String  | Environment for pushing artifacts (can be either dev or prod). This sets the GAR Project to either `grafanalabs-dev` or `grafanalabs-global`.                       |
| `images`               | String  | CSV of Docker images to push. These images should not include tags. Ex: us-docker.pkg.dev/grafanalabs-dev/gar-registry/image-name,docker.io/grafana/dockerhub-image |
| `push`                 | Boolean | Whether to push the manifest to the configured registries.                                                                                                          |
| `tags`                 | String  | List of Docker tags to be pushed.                                                                                                                                   |
