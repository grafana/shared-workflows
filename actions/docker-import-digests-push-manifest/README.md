# docker-import-digest-push-manifest

This is a composite GitHub Action used to import Docker digests from a shared workflow artifact and merge them into a
tagged manifest.

This can be used in conjunction with [docker-build-push-image] and [docker-export-digest] to build
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
  import-and-merge-digest:
    permissions:
      contents: read
      id-token: write
    steps:
      - name: Download Multi-Arch Digests, Construct and Upload Manifest
        uses: grafana/shared-workflows/actions/docker-import-digests-push-manifest@docker-import-digests-push-manifest/v0.2.2
        with:
          gar-environment: "dev"
          images: ${{ needs.docker-build-push-image.outputs.images }}
          push: true
          tags: |
            latest
            ${{ github.sha }}
```

<!-- x-release-please-end-version -->

## Inputs

<!-- BEGIN_INPUTS -->

| Name               | Type    | Required | Default | Description                                                                                                                                                         |
| ------------------ | ------- | -------- | ------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `gar-environment`  | String  | No       | `dev`   | Environment for pushing artifacts (can be either dev or prod). This sets the GAR Project to either `grafanalabs-dev` or `grafanalabs-global`.                       |
| `generate-summary` | Boolean | No       | `false` | Generates both a markdown summary and outputs the env variable OCI_MANIFEST_OUTPUT_JSON. Only runs when `push` is also `true`.                                      |
| `images`           | String  | Yes      |         | CSV of Docker images to push. These images should not include tags. Ex: us-docker.pkg.dev/grafanalabs-dev/gar-registry/image-name,docker.io/grafana/dockerhub-image |
| `push`             | Boolean | No       | `false` | Whether to push the manifest to the configured registries.                                                                                                          |
| `tags`             | String  | Yes      |         | List of Docker tags to be pushed.                                                                                                                                   |

<!-- END_INPUTS -->

## Outputs

<!-- BEGIN_OUTPUTS -->

| Name                       | Description                                                                                                                                     |
| -------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------- |
| `image-digests`            | Newline-separated list of image digests in the format `<image>:<tag>@<digest>`. Empty unless both `push` and `generate-summary` are `true`.     |
| `oci-manifest-output-json` | JSON array of manifests with tag, indexDigest, and per-platform digest information. Empty unless both `push` and `generate-summary` are `true`. |

<!-- END_OUTPUTS -->

### `OCI_MANIFEST_OUTPUT_JSON` environment variable

When both `push: true` and `generate-summary: true`, the action also sets the `OCI_MANIFEST_OUTPUT_JSON` environment variable in subsequent steps. It contains a JSON array of objects with the following structure:

```json
[
  {
    "tag": "us-docker.pkg.dev/grafanalabs-dev/gar-registry/my-image:latest",
    "indexDigest": "sha256:abc123...",
    "manifests": [
      {
        "kind": "image",
        "platform": "linux/amd64",
        "digest": "sha256:def456..."
      },
      {
        "kind": "image",
        "platform": "linux/arm64",
        "digest": "sha256:ghi789..."
      },
      {
        "kind": "attestation",
        "digest": "sha256:jkl012...",
        "refersTo": "sha256:def456..."
      }
    ]
  }
]
```

You can use this to reference specific platform digests in dependent repositories:

```yaml
- name: Print pushed digests
  run: echo "$OCI_MANIFEST_OUTPUT_JSON" | jq -r '.[] | "\(.tag): \(.indexDigest)"'
```
