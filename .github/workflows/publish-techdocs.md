# Reusable workflow: Publish techdocs

This workflow helps you build your project's documentation and publish it to [EngHub](https://enghub.grafana-ops.net).
Please keep in mind that for this you also need to first register your repository with EngHub.
You can find details on this [here](https://enghub.grafana-ops.net/docs/default/component/enghub/user-guides/add-gh-repo/).

## Usage example

```yaml
name: Publish TechDocs
on:
  push:
    branches:
      - main
    paths:
      - "docs/**"
      - "mkdocs.yml"
      - "catalog-info.yaml"
      - ".github/workflows/publish-docs.yml"
concurrency:
  group: "${{ github.workflow }}-${{ github.ref }}"
  cancel-in-progress: true
jobs:
  publish-docs:
    uses: grafana/shared-workflows/.github/workflows/publish-techdocs.yaml@main
    secrets: inherit
    with:
      namespace: default
      kind: component
      name: COMPONENT_NAME
```

## Inputs

| Name                        | Type    | Description                                                                   |
| --------------------------- | ------- | ----------------------------------------------------------------------------- |
| `namespace`                 | string  | The entity's namespace within EngHub (usually `default`)                      |
| `kind`                      | string  | The kind of the entity in EngHub (usually `component`)                        |
| `name`                      | string  | The name of the entity in EngHub (usually matches the name of the repository) |
| `default-working-directory` | string  | The directory where the techdocs-cli should be run (default: `.`)             |
| `checkout-repo`             | boolean | Should this workflow also check out the current repository? (default: `true`) |
