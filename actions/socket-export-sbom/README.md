# socket-export-sbom

Composite action (step) that triggers a fresh Socket full scan for a given repo/branch and then fetches the resulting SBOM in SPDX format from socket.dev.

A good use case is including this sbom as part of a public repo's release artifacts when creating a new release. If the release is immutable, then the steps for this workflow need to be incorporated into the release action.

> **Breaking change:** this action now _creates_ a new Socket full scan for the `branch` you specify, rather than reading whatever scan the Socket GitHub App already recorded for the repo's default branch. Because the Socket CLI scans the manifest files on disk, the calling workflow **must check out the target branch's source tree before invoking this action**. Previously no checkout was required. See the updated examples below.

## Inputs

<!-- BEGIN_INPUTS -->

| Name                     | Type   | Required | Default                      | Description                                                                                                                                                                                                                                                                                            |
| ------------------------ | ------ | -------- | ---------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `branch`                 | String | Yes      |                              | Branch to scan and export the SBOM for. The caller must have already checked out this branch before invoking this action.                                                                                                                                                                              |
| `export_timeout_seconds` | String | No       | `180`                        | Max seconds to wait for the SBOM export to become available after scan creation. The full scan report is generated lazily, so the export endpoint may 404 for a while after `socket scan create` returns. 180s comfortably covers grafana/grafana (our largest repo), which took up to 85s in testing. |
| `output_file`            | String | No       |                              | Name of the file to save the sbom. Defaults to '&lt;repo>-&lt;branch>.spdx.json' if not set.                                                                                                                                                                                                           |
| `socket_api_token`       | String | Yes      |                              | Socket API token for authentication. Requires the `full-scans:create` and `report:read` scopes.                                                                                                                                                                                                        |
| `socket_base_url`        | String | No       | `https://api.socket.dev/v0/` | Socket base url. Must end in a trailing slash: the Socket CLI appends endpoint paths to it without adding a separator, so a base of `https://api.socket.dev/v0` requests `/v0report/supported` and 404s.                                                                                               |
| `socket_org`             | String | Yes      | `grafana`                    | Socket org name                                                                                                                                                                                                                                                                                        |

<!-- END_INPUTS -->

## Outputs

<!-- BEGIN_OUTPUTS -->

| Name   | Description                    |
| ------ | ------------------------------ |
| `path` | Path to the exported sbom file |

<!-- END_OUTPUTS -->

## Export behaviour

Socket generates the full scan report lazily, so the SPDX export endpoint returns a 404 for a while after `socket scan create` finishes. The export step ([`export-sbom.js`](./export-sbom.js)) polls with a 5s backoff, doubling to a 30s cap, until the report is ready or `export_timeout_seconds` elapses.

Not every failure is worth retrying, so the step separates the two cases:

| Response                                         | Behaviour                                                             |
| ------------------------------------------------ | --------------------------------------------------------------------- |
| `200` with a parseable JSON body                 | Written to `output_file`, path published as the `path` output         |
| `200` with a body that does not parse            | Retried — Socket can serve a partial report while the scan is running |
| `404`, `408`, `429`, any `5xx`, network failures | Retried until the timeout                                             |
| `400`, `401`, `403`, any other `4xx`             | Fails immediately — no amount of retrying adds a missing scope        |

Only a body that parses as JSON is written to `output_file`, so a rejection page can never be mistaken for an SBOM by a later step.

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
        uses: grafana/shared-workflows/actions/get-vault-secrets@e46fe1e9a2bf9e618bcf8d8d32f3a7381b45c06d # get-vault-secrets/v1.0.0
        with:
          common_secrets: |
            SOCKET_API_TOKEN=socket:SOCKET_API_KEY

      - name: "Checkout"
        uses: actions/checkout@9c091bb21b7c1c1d1991bb908d89e4e9dddfe3e0 # v1.0.0
        with:
          ref: ${{ github.ref_name }}

      - name: "Export SPDX SBOM from Socket"
        id: export-sbom
        uses: grafana/shared-workflows/actions/socket-export-sbom@ff9aaa53f25716fcd6dde39f6d4e41c4e16fb5e1 # socket-export-sbom/v1.0.0
        with:
          socket_api_token: ${{ fromJSON(steps.vault-secrets.outputs.secrets).SOCKET_API_TOKEN }}
          socket_org: grafana
          branch: ${{ github.ref_name }}

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
      - name: "Resolve release tag"
        id: meta
        env:
          DISPATCH_TAG: ${{ inputs.tag }}
          REF_TAG: ${{ github.ref_name }}
        run: echo "tag=${DISPATCH_TAG:-$REF_TAG}" >> "$GITHUB_OUTPUT"
      - name: "Get Socket API token from Vault"
        id: vault-secrets
        uses: grafana/shared-workflows/actions/get-vault-secrets@e46fe1e9a2bf9e618bcf8d8d32f3a7381b45c06d # get-vault-secrets/v1.0.0
        with:
          common_secrets: |
            SOCKET_API_TOKEN=socket:SOCKET_API_KEY

      - name: "Checkout"
        uses: actions/checkout@9c091bb21b7c1c1d1991bb908d89e4e9dddfe3e0 # v1.0.0
        with:
          ref: ${{ steps.meta.outputs.tag }}

      - name: "Export SPDX SBOM from Socket"
        id: export-sbom
        uses: grafana/shared-workflows/actions/socket-export-sbom@ff9aaa53f25716fcd6dde39f6d4e41c4e16fb5e1 # socket-export-sbom/v1.0.0
        with:
          socket_api_token: ${{ fromJSON(steps.vault-secrets.outputs.secrets).SOCKET_API_TOKEN }}
          socket_org: grafana
          branch: ${{ steps.meta.outputs.tag }}

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
