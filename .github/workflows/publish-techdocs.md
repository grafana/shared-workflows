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
    with:
      namespace: default
      kind: component
      name: COMPONENT_NAME
```

## Inputs

| Name                             | Type    | Description                                                                                                                                                            |
| -------------------------------- | ------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `namespace`                      | string  | The entity's namespace within EngHub (usually `default`)                                                                                                               |
| `kind`                           | string  | The kind of the entity in EngHub (usually `component`)                                                                                                                 |
| `name`                           | string  | The name of the entity in EngHub (usually matches the name of the repository)                                                                                          |
| `default-working-directory`      | string  | The working directory to use for doc generation. Useful for cases without an mkdocs.yml file at the project root.                                                      |
| `rewrite-relative-links`         | boolean | Execute [rewrite-relative-links][rewrite-action] step to rewrite relative links in the docs to point to the correct location in the GitHub repository                  |
| `rewrite-relative-links-dry-run` | boolean | Execute [rewrite-relative-links][rewrite-action] step but only print the diff without modifying the files                                                              |
| `publish`                        | boolean | Enable or disable publishing after building the docs                                                                                                                   |
| `checkout-submodules`            | string  | Checkout submodules in the repository. Options are `true` (checkout submodules), `false` (don't checkout submodules), or `recursive` (recursively checkout submodules) |
| `instance`                       | string  | The name of the instance to which the docs should be published (`ops` (default), `dev`)                                                                                |

[rewrite-action]: ../../actions/techdocs-rewrite-relative-links/README.md
