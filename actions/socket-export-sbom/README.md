# socket-export-sbom

Composite action (step) to get the latest scan id for a repo enrolled in the socket.dev GitHub App and then fetch the spdx sbom from socket using the latest scan id.

A good use case is including this sbom as part of a public repo's release artifacts when creating a new release. If the release is immutable, then the steps for this workflow need to be incorporated into the release action.

## Inputs

| Name               | Type     | Description                                                                                                    | Default Value                 | Required |
| ------------------ | -------- | -------------------------------------------------------------------------------------------------------------- | ----------------------------- | -------- |
| `socket_api_token` | `string` | API Key used to authenticate to socket.dev, requires repo: list, repo:read, full-scan:list, report:list scopes | `none`                        | true     |
| `socket_base_url`  | `string` | Base URL of the socket api endpoint.                                                                           | `"https://api.socket.dev/v0"` | false    |
| `socket_org_name`  | `string` | Name of the socket org.                                                                                        | `"grafana"`                   | true     |
| `output_file`      | `string` | Name of the file to save the socket sbom on the runner.                                                        | `"spdx.json"`                 | false    |

## Examples

### SBOM Generation for a repo that does not use immutable releases

<!-- x-release-please-start-version -->

```yaml
name: "SBOM on Release"

on:
  release:
    types:
      - published

permissions: {}

jobs:
  export-sbom:
    name: "Export SBOM and attach to release"
    runs-on: ubuntu-latest
    permissions:
      contents: write # to upload the SBOM as a release asset
      id-token: write # to authenticate to Vault
    steps:
      - name: "Get Socket API token from Vault"
        id: vault-secrets
        uses: grafana/shared-workflows/actions/get-vault-secrets@e46fe1e9a2bf9e618bcf8d8d32f3a7381b45c06d # get-vault-secrets/v2.0.0
        with:
          common_secrets: |
            SOCKET_API_TOKEN=socket:SOCKET_API_KEY

      - name: "Export SPDX SBOM from Socket"
        id: export-sbom
        uses: grafana/shared-workflows/actions/socket-export-sbom@ff9aaa53f25716fcd6dde39f6d4e41c4e16fb5e1 # socket-export-sbom/v0.1.2
        with:
          socket_api_token: ${{ fromJSON(steps.vault-secrets.outputs.secrets).SOCKET_API_TOKEN }}
          socket_org: grafana
          output_file: ${{ github.event.repository.name }}-${{ github.event.release.tag_name }}.spdx.json

      - name: "Upload SBOM to release"
        env:
          GH_TOKEN: ${{ github.token }}
          GH_REPO: ${{ github.repository }}
          TAG: ${{ github.event.release.tag_name }}
          SBOM_PATH: ${{ steps.export-sbom.outputs.path }}
        run: gh release upload "$TAG" "$SBOM_PATH" --clobber
```

### SBOM Generation for Repo with Immutable Releases

Either create draft release and upload SBOM asset before full release or include directly in the release workflow

```yaml
name: "SBOM on Release"

on:
  push:
    tags:
      - "v[0-9]+.[0-9]+.[0-9]+" # vX.Y.Z
      - "v[0-9]+.[0-9]+.[0-9]+-*" # vX.Y.Z-alpha / -rc1 / -beta.1 etc.
  workflow_dispatch:
    inputs:
      tag:
        description: "Existing draft release tag to export the SBOM for and attach it to (for testing)"
        required: true
        type: string

permissions: {}

jobs:
  export-sbom:
    name: "Export SBOM and attach to draft release"
    runs-on: ubuntu-latest
    permissions:
      contents: write # to create the draft release and upload the SBOM as a release asset
      id-token: write # to authenticate to Vault
    steps:
      - name: Checkout
        uses: actions/checkout@8e8c483db84b4bee98b60c0593521ed34d9990e8 # v0.1.2

      - name: "Get Socket API token from Vault"
        id: vault-secrets
        uses: grafana/shared-workflows/actions/get-vault-secrets@get-vault-secrets/v0.1.2
        with:
          common_secrets: |
            SOCKET_API_TOKEN=socket:SOCKET_API_KEY

      - name: "Export SPDX SBOM from Socket"
        id: export-sbom
        uses: grafana/shared-workflows/actions/socket-export-sbom@socket-export-sbom/v0.1.2
        with:
          socket_api_token: ${{ fromJSON(steps.vault-secrets.outputs.secrets).SOCKET_API_TOKEN }}
          socket_org: grafana
          output_file: ${{ steps.meta.outputs.repo }}-${{ steps.meta.outputs.tag }}.spdx.json

      - name: Upload SBOM artifact
        uses: actions/upload-artifact@330a01c490aca151604b8cf639adc76d48f6c5d4 # v0.1.2
        with:
          name: "sbom"
          path: ${{ steps.export-sbom.outputs.path }}
          retention-days: 30
```

<!-- x-release-please-end-version -->
