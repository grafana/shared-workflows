# Generate OpenAPI clients

This action generates clients from an OpenAPI spec. It's meant to generate clients in an uniform way across our organization

_Note: For now, it only generates Go code. But it's structured in a way that any of the languages supported by the openapi-generator could be supported at the same time._

## Inputs

| Name              | Type    | Description                                                                                | Default Value                | Required |
| ----------------- | ------- | ------------------------------------------------------------------------------------------ | ---------------------------- | -------- |
| generator-version | string  | The version of the OpenAPI generator to use                                                | "7.7.0"                      | false    |
| spec-path         | string  | The path to the OpenAPI spec to generate the client from. Supports JSON or YAML            | N/A                          | true     |
| output-dir        | string  | The directory to output the generated client to                                            | "."                          | false    |
| commit-changes    | boolean | If true, the action will commit and push the changes to the repository, if there's a diff. | true                         | false    |
| commit-message    | string  | The commit message to use when committing the changes                                      | "Update clients and publish" | false    |
| package-name      | string  | The name of the package to generate                                                        | N/A                          | true     |

## Example workflow

```yaml
name: Generate Clients

on:
  push:
    branches:
      - main

permissions:
  contents: write # Only needed if `commit-changes` is set to true

jobs:
  build-and-publish:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@692973e3d937129bcbf40652eb9f2f61becf3332 # v4.1.7
      # Check out head_ref if working from a pull request
      # with:
      #   ref: ${{ github.head_ref }}
      - uses: actions/setup-go@0a12ed9d6a96ab950c8f026ed9f722fe0da7ef32 # v5.0.2
        with:
          go-version: 1.18
      - name: Generate clients
        uses: grafana/shared-workflows/actions/generate-openapi-clients@main
        with:
          package-name: slo
          spec-path: openapi.yaml
```
