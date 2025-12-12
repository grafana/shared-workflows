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
        uses: grafana/shared-workflows/actions/get-vault-secrets@get-vault-secrets/v1.0.1
        id: get-signing-secrets
        with:
          export_env: false
          repo_secrets: |
            client-id=azure-trusted-signing:client-id
            subscription-id=azure-trusted-signing:subscription-id
            tenant-id=azure-trusted-signing:tenant-id

      - name: Sign artifacts
        uses: grafana/shared-workflows/actions/azure-trusted-signing@azure-trusted-signing/v1.0.1
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

### Required

| **Name**                  | **Description**                                                                                         |
| :------------------------ | :------------------------------------------------------------------------------------------------------ |
| `application-description` | The description of the application to sign the file(s) for.                                             |
| `artifact-to-sign`        | The name of the GitHub Actions workflow artifact from the current workflow run to sign the contents of. |
| `azure-client-id`         | The client ID to use to authenticate with Azure.                                                        |
| `azure-subscription-id`   | The subscription ID to use to authenticate with Azure.                                                  |
| `azure-tenant-id`         | The tenant ID to use to authenticate with Azure.                                                        |
| `signed-artifact-name`    | The name of the GitHub Actions workflow artifact to upload the signed files to.                         |

### Optional

| **Name**                   | **Description**                                                                  | **Default**                                            |
| :------------------------- | :------------------------------------------------------------------------------- | :----------------------------------------------------- |
| `application-url`          | The URL of the application to sign the file(s) for.                              | The URL of the GitHub repository running the workflow. |
| `file-filter`              | The path filter of which files to sign from the artifact.                        | `'**/*'`                                               |
| `file-list`                | The path to a file containing paths of files to sign or to exclude from signing. | -                                                      |
| `publisher-name`           | The name of the publisher of the application the signed file(s) belong to.       | `'Grafana Labs'`                                       |
| `trusted-signing-account`  | The name of the Azure Trusted Signing account to use.                            | -                                                      |
| `trusted-signing-endpoint` | The endpoint URL of the Azure Trusted Signing service to use.                    | -                                                      |
| `trusted-signing-profile`  | The name of the Azure Trusted Signing profile to use.                            | -                                                      |

## Outputs

| **Name**        | **Description**                                                                 |
| :-------------- | :------------------------------------------------------------------------------ |
| `artifact-name` | The name of the GitHub Actions workflow artifact containing the signed file(s). |

[azure-trusted-signing]: https://learn.microsoft.com/azure/trusted-signing/
