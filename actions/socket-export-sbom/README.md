# socket-export-sbom

Composite action (step) that triggers a fresh Socket full scan for a given repo/branch and then fetches the resulting SBOM in SPDX format from socket.dev.

A good use case is including this sbom as part of a public repo's release artifacts when creating a new release. If the release is immutable, then the steps for this workflow need to be incorporated into the release action.

> **Breaking change:** this action now _creates_ a new Socket full scan for the `branch` you specify, rather than reading whatever scan the Socket GitHub App already recorded for the repo's default branch. Because the Socket CLI scans the manifest files on disk, the calling workflow **must check out the target branch's source tree before invoking this action**. Previously no checkout was required. See the updated examples below.

## Inputs

| Name               | Type     | Description                                                                                                                                                                                                                          | Default Value                 | Required |
| ------------------ | -------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ----------------------------- | -------- |
| `socket_api_token` | `string` | API Key used to authenticate to socket.dev, requires repo: list, repo:read, full-scan:list, report:list scopes                                                                                                                       | `none`                        | true     |
| `socket_base_url`  | `string` | Base URL of the socket api endpoint.                                                                                                                                                                                                 | `"https://api.socket.dev/v0"` | false    |
| `socket_org`       | `string` | Name of the socket org.                                                                                                                                                                                                              | `"grafana"`                   | true     |
| `branch`           | `string` | Branch to scan and export the SBOM for. The caller must have already checked out this branch's source tree before invoking this action, since the Socket CLI scans the local manifest files rather than reading a pre-existing scan. | `none`                        | true     |
| `output_file`      | `string` | Name of the file to save the socket sbom on the runner.                                                                                                                                                                              | `"spdx.json"`                 | false    |

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

      - name: "Checkout"
        uses: actions/checkout@9c091bb21b7c1c1d1991bb908d89e4e9dddfe3e0 # v7.0.0
        with:
          ref: ${{ github.ref_name }}

      - name: "Export SPDX SBOM from Socket"
        id: export-sbom
        uses: grafana/shared-workflows/actions/socket-export-sbom@ff9aaa53f25716fcd6dde39f6d4e41c4e16fb5e1 # socket-export-sbom/v0.1.2
        with:
          socket_api_token: ${{ fromJSON(steps.vault-secrets.outputs.secrets).SOCKET_API_TOKEN }}
          socket_org: grafana
          branch: ${{ github.ref_name }}
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
      - name: "Resolve release tag and repo name"
        id: meta
        env:
          DISPATCH_TAG: ${{ inputs.tag }}
          REF_TAG: ${{ github.ref_name }}
          GH_REPO: ${{ github.repository }}
        run: |
          echo "tag=${DISPATCH_TAG:-$REF_TAG}" >> "$GITHUB_OUTPUT"
          echo "repo=${GH_REPO##*/}" >> "$GITHUB_OUTPUT"
      - name: "Get Socket API token from Vault"
        id: vault-secrets
        uses: grafana/shared-workflows/actions/get-vault-secrets@e46fe1e9a2bf9e618bcf8d8d32f3a7381b45c06d # get-vault-secrets/v2.0.0
        with:
          common_secrets: |
            SOCKET_API_TOKEN=socket:SOCKET_API_KEY

      - name: "Checkout"
        uses: actions/checkout@9c091bb21b7c1c1d1991bb908d89e4e9dddfe3e0 # v7.0.0
        with:
          ref: ${{ steps.meta.outputs.tag }}

      - name: "Export SPDX SBOM from Socket"
        id: export-sbom
        uses: grafana/shared-workflows/actions/socket-export-sbom@ff9aaa53f25716fcd6dde39f6d4e41c4e16fb5e1 # socket-export-sbom/v0.1.2
        with:
          socket_api_token: ${{ fromJSON(steps.vault-secrets.outputs.secrets).SOCKET_API_TOKEN }}
          socket_org: grafana
          branch: ${{ steps.meta.outputs.tag }}
          output_file: ${{ steps.meta.outputs.repo }}-${{ steps.meta.outputs.tag }}.spdx.json

      # Immutable releases lock assets at publish time, so the SBOM must be attached while the
      # release is still a draft. Draft creation does not trigger workflows, so we create the
      # draft here (on tag push) and let a human review and publish it afterwards.
      - name: "Create draft release"
        env:
          GH_TOKEN: ${{ github.token }}
          GH_REPO: ${{ github.repository }}
          TAG: ${{ steps.meta.outputs.tag }}
        run: |
          if gh release view "$TAG" >/dev/null 2>&1; then
            echo "Release $TAG already exists; will attach SBOM to it."
          else
            PRERELEASE=""
            # Version convention: v0.x or any -suffix (-alpha/-beta/-rcN) is a pre-release
            if [[ "$TAG" == v0.* || "$TAG" == *-* ]]; then PRERELEASE="--prerelease"; fi
            gh release create "$TAG" --draft --generate-notes --title "$TAG" $PRERELEASE
          fi

      - name: "Upload SBOM to draft release"
        env:
          GH_TOKEN: ${{ github.token }}
          GH_REPO: ${{ github.repository }}
          TAG: ${{ steps.meta.outputs.tag }}
          SBOM_PATH: ${{ steps.export-sbom.outputs.path }}
        run: gh release upload "$TAG" "$SBOM_PATH" --clobber
```

<!-- x-release-please-end-version -->
