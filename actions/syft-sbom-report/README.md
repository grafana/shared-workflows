# syft-sbom-report

Generate an SPDX SBOM Report and attached to Release Artifcats on Release Publish

Example workflow:

```yaml
name: syft-sbom-ci
on:
  release:
    types: [published]
  workflow_dispatch:
jobs:
  syft-sbom:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
      - name: Anchore SBOM Action
        uses: anchore/sbom-action@v0.15.10
        with:
          artifact-name: ${{ github.event.repository.name }}-spdx.json
```
