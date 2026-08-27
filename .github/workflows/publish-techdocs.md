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
    permissions:
      contents: read # to clone the repository to read its docs
      id-token: write # to use OIDC to auth with AWS and push the docs to S3
    uses: grafana/shared-workflows/.github/workflows/publish-techdocs.yaml@main
    with:
      namespace: default
      kind: component
      name: COMPONENT_NAME
```

## Inputs

<!-- BEGIN_INPUTS -->

| Name                             | Type    | Required | Default | Description                                                                                                                                                            |
| -------------------------------- | ------- | -------- | ------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `checkout-submodules`            | string  | No       | `false` | Checkout submodules in the repository. Options are `true` (checkout submodules), `false` (don't checkout submodules), or `recursive` (recursively checkout submodules) |
| `default-working-directory`      | string  | No       | `.`     | The working directory to use for doc generation. Useful for cases without an mkdocs.yml file at the project root.                                                      |
| `instance`                       | string  | No       | `ops`   | The instance to use (`dev` or `ops`). Defaults to `ops`.                                                                                                               |
| `kind`                           | string  | Yes      |         | The kind of the entity in EngHub (usually `component`)                                                                                                                 |
| `name`                           | string  | Yes      |         | The name of the entity in EngHub (usually matches the name of the repository)                                                                                          |
| `namespace`                      | string  | Yes      |         | The entity's namespace within EngHub (usually `default`)                                                                                                               |
| `publish`                        | boolean | No       | `true`  | Enable or disable publishing after building the docs                                                                                                                   |
| `rewrite-relative-links`         | boolean | No       | `false` | Execute rewrite-relative-links step to rewrite relative links in the docs to point to the correct location in the GitHub repository                                    |
| `rewrite-relative-links-dry-run` | boolean | No       | `false` | Execute rewrite-relative-links step but only print the diff without modifying the files                                                                                |

<!-- END_INPUTS -->
