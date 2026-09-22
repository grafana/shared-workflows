# azure-trusted-signing

This is a composite GitHub Action used to sign files using [Azure Trusted Signing][azure-trusted-signing].

> [!IMPORTANT]
> This GitHub Action is only supported on Windows-based GitHub Actions runners.

## Example

<!-- markdownlint-disable MD013 -->
<!-- x-release-please-start-version -->

```yaml
name: CI
on:
  push:
    branches: ["main"]
    tags: ["v*"]
  pull_request:
  workflow_dispatch:

jobs:
  package:
    runs-on: ubuntu-latest

    steps:
      - name: Setup .NET
        uses: actions/setup-dotnet@v4

      - name: Build NuGet packages
        run: dotnet pack --configuration Release --output ./artifacts

      - name: Upload artifacts
        uses: actions/upload-artifact@v4
        with:
          name: artifacts
          path: ./artifacts

  sign:
    needs: [package]
    runs-on: windows-latest
    if: github.event.repository.fork == false && startsWith(github.ref, 'refs/tags/')

    environment:
      name: azure-trusted-signing

    outputs:
      artifact-name: ${{ steps.sign-artifacts.outputs.artifact-name }}

    permissions:
      contents: read
      id-token: write

    steps:
      - name: Get secrets for Azure Trusted Signing
        uses: grafana/shared-workflows/actions/get-vault-secrets@get-vault-secrets/v1.0.3
        id: get-signing-secrets
        with:
          export_env: false
          repo_secrets: |
            client-id=azure-trusted-signing:client-id
            subscription-id=azure-trusted-signing:subscription-id
            tenant-id=azure-trusted-signing:tenant-id

      - name: Sign artifacts
        uses: grafana/shared-workflows/actions/azure-trusted-signing@azure-trusted-signing/v1.0.3
        id: sign-artifacts
        with:
          application-description: "My Awesome application"
          artifact-to-sign: "artifacts"
          azure-client-id: ${{ fromJSON(steps.get-signing-secrets.outputs.secrets).client-id }}
          azure-subscription-id: ${{ fromJSON(steps.get-signing-secrets.outputs.secrets).subscription-id }}
          azure-tenant-id: ${{ fromJSON(steps.get-signing-secrets.outputs.secrets).tenant-id }}
          signed-artifact-name: "signed-artifacts"

  release:
    needs: [sign]
    runs-on: ubuntu-latest

    steps:
      - name: Download signed packages
        uses: actions/download-artifact@v5
        with:
          name: ${{ needs.sign.outputs.artifact-name }}

      - name: Release
        run: echo "Do something with the signed artifacts"
```

<!-- x-release-please-end-version -->
<!-- markdownlint-enable MD013 -->

## Inputs

<!-- BEGIN_INPUTS -->

| Name                       | Type   | Required | Default                                                          | Description                                                                                                     |
| -------------------------- | ------ | -------- | ---------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------- |
| `application-description`  | String | Yes      |                                                                  | The description of the application to sign the file(s) for.                                                     |
| `application-url`          | String | No       | `${{ format('{0}/{1}', github.server_url, github.repository) }}` | The optional URL of the application to sign the file(s) for. Defaults to the current GitHub repository URL.     |
| `artifact-to-sign`         | String | Yes      |                                                                  | The name of the GitHub Actions workflow artifact from the current workflow run to sign the contents of.         |
| `azure-client-id`          | String | Yes      |                                                                  | The client ID to use to authenticate with Azure.                                                                |
| `azure-subscription-id`    | String | Yes      |                                                                  | The subscription ID to use to authenticate with Azure.                                                          |
| `azure-tenant-id`          | String | Yes      |                                                                  | The tenant ID to use to authenticate with Azure.                                                                |
| `file-filter`              | String | No       | `**/*`                                                           | The optional path filter of which files to sign from the artifact. Defaults to all files.                       |
| `file-list`                | String | No       |                                                                  | The optional path to a file containing paths of files to sign or to exclude from signing.                       |
| `publisher-name`           | String | No       | `Grafana Labs`                                                   | The optional name of the publisher of the application the signed file(s) belong to. Defaults to "Grafana Labs". |
| `signed-artifact-name`     | String | Yes      |                                                                  | The name of the GitHub Actions workflow artifact to upload the signed files to.                                 |
| `trusted-signing-account`  | String | No       | `grafana-premium-eastus`                                         | The optional name of the Azure Trusted Signing account to use.                                                  |
| `trusted-signing-endpoint` | String | No       | `https://eus.codesigning.azure.net/`                             | The optional endpoint URL of the Azure Trusted Signing service to use.                                          |
| `trusted-signing-profile`  | String | No       | `grafana-production`                                             | The optional name of the Azure Trusted Signing profile to use.                                                  |

<!-- END_INPUTS -->

## Outputs

<!-- BEGIN_OUTPUTS -->

| Name            | Description                                                                     |
| --------------- | ------------------------------------------------------------------------------- |
| `artifact-name` | The name of the GitHub Actions workflow artifact containing the signed file(s). |

<!-- END_OUTPUTS -->

[azure-trusted-signing]: https://learn.microsoft.com/azure/trusted-signing/
