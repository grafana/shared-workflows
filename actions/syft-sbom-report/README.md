# syft-sbom-report

Generate an SPDX SBOM Report and attached to Release Artifcats on Release Publish

Example workflow:

<!-- x-release-please-start-version -->

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
        uses: grafana/shared-workflows/actions@syft-sbom-v0.0.1
        with:
          artifact-name: ${{ github.event.repository.name }}-spdx.json
```

<!-- x-release-please-end-version -->
