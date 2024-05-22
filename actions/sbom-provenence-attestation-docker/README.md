# sbom-provenence-attestation-docker

Generate an SPDX SBOM Report and build provenance. Attach to the docker image, create an attestation and publish that to the GitHub attestation store.

This action wraps [GitHub's attestation support][gh-attestations].

Example workflow:

```yaml
name: Build Docker with SBOM and provenance

permissions:
  contents: read
  id-token: write
  attestations: write

on:
  release:
    types: [published]
  workflow_dispatch:

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@0ad4b8fadaa221de15dcec353f45205ec38ea70b # v4.1.4

      - name: Set Docker Buildx up
        uses: docker/setup-buildx-action@d70bba72b1f3fd22344832f00baa16ece964efeb # v3.3.0

      # ... also use docker/metadata-action to work out the tags
      # or (recommended) use `build-push-to-dockerhub` from this repo to do all
      # of the below steps

      - name: Build and push Docker image
        id: build-push
        uses: docker/build-push-action@4a13e500e55cf31b7a5d59a38ab2040ab0f42f56 # v5.1.0
        with:
          push: true
          tags: latest

      - name: Generate attestations and SBOM
        uses: grafana/shared-workflows/actions/sbom-provenence-attestation-docker@abc123
        with:
          digest: ${{ steps.build-push.outputs.digest }}
          push: true
          image: index.docker.io/grafana/myimage
```

Check it worked:

```console
laney@florence> gh attestation verify oci://index.docker.io/grafana/wait-for-github:iainlane-attestation-test -R grafana/wait-for-github
Loaded digest sha256:9e02c5c059ff864cb1e2ab343f65668ba431bd6160cfb59f269281164a255104 for oci://index.docker.io/grafana/myimage:latest
Loaded 2 attestations from GitHub API
✓ Verification succeeded!

sha256:9e02c5c059ff864cb1e2ab343f65668ba431bd6160cfb59f269281164a255104 was attested by:
REPO                     PREDICATE_TYPE                  WORKFLOW
grafana/myimage          https://cyclonedx.org/bom       .github/workflows/build.yml@refs/pull/130/merge
grafana/myimage          https://slsa.dev/provenance/v1  .github/workflows/build.yml@refs/pull/130/merge
```

[gh-attestations]: https://docs.github.com/en/actions/security-guides/using-artifact-attestations-to-establish-provenance-for-builds

## Inputs

| Name     | Type   | Description                                                                         |
| -------- | ------ | ----------------------------------------------------------------------------------- |
| `digest` | String | The digest of the image to attest.                                                  |
| `image`  | String | The image name to attest.                                                           |
| `push`   | Bool   | Whether to push the attestation to the GitHub attestation store. (default: `false`) |
