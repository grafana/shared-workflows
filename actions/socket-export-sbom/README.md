# socket-export-sbom

Composite action (step) to get the latest scan id for a repo enrolled in the socket.dev GitHub App and then fetch the spdx sbom from socket using the latest scan id.

A good use case is including this sbom as part of a public repo's release artifacts when creating a new release

## Inputs

| Name               | Type     | Description                                                                                                                                                                        | Default Value                 | Required |
| ------------------ | -------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------- | -------- |
| `socket_api_token` | `string` | GitHub token used to authenticate with `gh`. Requires permission to query for protected branches and delete branches (`contents: write`) and pull requests (`pull_requests: read`) | `none`                        | true     |
| `socket_base_url`  | `string` | Base URL of the socket api endpoint.                                                                                                                                               | `"https://api.socket.dev/v0"` | false    |
| `socket_org_name`  | `string` | Name of the socket org.                                                                                                                                                            | `"grafana"`                   | true     |
| `output_file`      | `string` | Name of the file to save the socket sbom on the runner.                                                                                                                            | `"spdx.json"`                 | false    |

## Examples

### Runs as a workflow dispatch but typical use case should run on release

<!-- x-release-please-start-version -->

```yaml
name: Get Repo SBOM from Socket API

on:
  workflow_dispatch:
    inputs:
      output_file:
        description: "Output file path for the SBOM"
        required: false
        default: "spdx.json"

jobs:
  export-sbom:
    permissions:
      contents: read
      id-token: write
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@8e8c483db84b4bee98b60c0593521ed34d9990e8 # v0.1.1

      - name: Get vault secrets
        id: vault-secrets
        uses: grafana/shared-workflows/actions/get-vault-secrets@get-vault-secrets/v0.1.1
        with:
          repo_secrets: |
            SOCKET_API_TOKEN=socket:SOCKET_SBOM_API_KEY
          export_env: false

      - name: Export SBOM from Socket
        id: export-sbom
        uses: grafana/shared-workflows/actions/socket-export-sbom@socket-export-sbom/v0.1.1
        with:
          socket_api_token: ${{ fromJSON(steps.vault-secrets.outputs.secrets).SOCKET_API_TOKEN }}
          output_file: ${{ inputs.output_file }}

      - name: Upload SBOM artifact
        uses: actions/upload-artifact@330a01c490aca151604b8cf639adc76d48f6c5d4 # v0.1.1
        with:
          name: "sbom"
          path: ${{ steps.export-sbom.outputs.path }}
          retention-days: 30
```

<!-- x-release-please-end-version -->
