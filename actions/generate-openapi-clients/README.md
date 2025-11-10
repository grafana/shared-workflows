# Generate OpenAPI clients

This action generates clients from an OpenAPI spec. It's meant to generate clients in an uniform way across our organization

_Note: For now, it only generates Go code. But it's structured in a way that any of the languages supported by the openapi-generator could be supported at the same time._

## Inputs

| Name               | Type    | Description                                                                                   | Default Value                | Required |
| ------------------ | ------- | --------------------------------------------------------------------------------------------- | ---------------------------- | -------- |
| generator-version  | string  | The version of the OpenAPI generator to use                                                   | "7.7.0"                      | false    |
| spec-path          | string  | The path to the OpenAPI spec to generate the client from. Supports JSON or YAML               | N/A                          | true     |
| output-dir         | string  | The directory to output the generated client to                                               | "."                          | false    |
| commit-changes     | boolean | If true, the action will commit and push the changes to the repository, if there's a diff.    | true                         | false    |
| commit-message     | string  | The commit message to use when committing the changes                                         | "Update clients and publish" | false    |
| package-name       | string  | The name of the package to generate                                                           | N/A                          | true     |
| modify-spec-script | string  | The path to an executable script that modifies the OpenAPI spec before generating the client. | ""                           | false    |

## Example workflow

<!-- x-release-please-start-version -->

```yaml
name: Generate Clients

on:
  push:
    branches:
      - main

jobs:
  build-and-publish:
    runs-on: ubuntu-latest
    permissions:
      contents: write # Only needed if `commit-changes` is set to true
    steps:
      - uses: actions/checkout@692973e3d937129bcbf40652eb9f2f61becf3332 # v1.0.3
        with:
          persist-credentials: false

      - uses: actions/setup-go@0a12ed9d6a96ab950c8f026ed9f722fe0da7ef32 # v1.0.3
        with:
          go-version: 1.18
      - name: Generate clients
        uses: grafana/shared-workflows/actions/generate-openapi-clients@generate-openapi-clients/v1.0.3
        with:
          package-name: slo
          spec-path: openapi.yaml
          modify-spec-script: .github/workflows/modify-spec.sh # Optional, see "Spec Modifications" section
```

<!-- x-release-please-end-version -->

### Spec Modifications at Runtime

The `modify-spec-script` attribute is the path to an executable script that modifies the OpenAPI spec before generating the client.
The spec will be piped into the script and the script should output the modified spec to stdout.

_Note: This is used as a workaround for the OpenAPI generator not supporting certain features. By using
this feature, the spec will be modified temporarily, and the changes will not be committed._

Here's an example of a modification script:

```bash
#! /usr/bin/env bash
set -euo pipefail

SCHEMA=`cat` # Read stdin
modify() {
    SCHEMA="$(echo "${SCHEMA}" | jq "${1}")"
}
modify '.components.schemas.FormattedApiApiKey.properties.id = { "anyOf": [ { "type": "string" }, { "type": "number" } ] }'
modify '.components.schemas.FormattedApiApiKeyListResponse.properties.items.items.properties.id = { "anyOf": [ { "type": "string" }, { "type": "number" } ] }'
modify '.components.schemas.FormattedOrgMembership.properties.allowGCloudTrial = { "anyOf": [ { "type": "boolean" }, { "type": "number" } ] }'
modify '.components.schemas.FormattedApiOrgPublic.properties.allowGCloudTrial = { "anyOf": [ { "type": "boolean" }, { "type": "number" } ] }'
modify '.paths["/v1/accesspolicies"].get.responses["200"].content["application/json"].schema = {
  "type": "object",
  "properties": {
    "items": {
      "type": "array",
      "items": {
        "$ref": "#/components/schemas/AuthAccessPolicy"
      }
    }
  }
}'

echo "${SCHEMA}"
```

This script should be saved to a file and its path given in the `modify-spec-script` attribute.
