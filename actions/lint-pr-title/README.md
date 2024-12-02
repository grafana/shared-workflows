# Lint pull request title

This is a [GitHub Action][github-action] that ensures compliance with the
[Commitlint Spec][commitlint-spec] according to the given configuration file.

Despite its name, this action supports validating pull request titles and
commits in merge queues. It also supports validating the _body_ of the commit
message too, not only the title.

[github-action]: https://github.com/features/actions
[commitlint-spec]: https://github.com/conventional-changelog/commitlint

## Pull Requests

This is helpful when using the "Squash and merge" strategy; GitHub will suggest
using the PR title as the commit message. With this action, you can validate
that the PR title will lead to a correct commit message.

## Merge queues

This action can also be used in merge queues to ensure that the commit messages
which will be merged into the main branch are compliant with the commitlint spec
and the project's configuration. When a project is using merge queues, it is the
commits in the merge queue branches which will be merged into the main branch,
so these are the commits that need to be validated.

## Inputs

| Name          | Description                                                                          | Default                  | Required |
| ------------- | ------------------------------------------------------------------------------------ | ------------------------ | -------- |
| `config-path` | Path to the commitlint configuration file, relative to the action's directory.       | `./commitlint.config.js` | No       |
| `title-only`  | Check only the PR/commit title. If false, it will check the whole PR/commit message. | `true`                   | No       |

## Validation

Examples for valid PR titles:

- fix(some-scope): correct typo.
- feat: add support for Node 12.

### The commit config

This action can receive a reference to a `commitlint.config.js` file using the
input parameter `config-path`. See the [commitlint documentation][docs] for
information on all of the options which can be set. This parameter is resolved
relative to the action's directory. It's recommended to start the path
definition with `${{ github.workspace }}/` to enable local configuration to be
found.

See example below:

```yml
---
uses: grafana/shared-workflows/actions/lint-pr-title@main
with:
  config-path: "${{ github.workspace }}/dir1/dir2/commitlint.config.js"
```

but this `config-path` parameter is optional. By default, the action will use
the [`commitlint.config.js` file located in this directory][config].

[config]: ./commitlint.config.js
[docs]: https://commitlint.js.org/reference/configuration.html

## Example workflows

### Example with config path defined:

In this example the `commitlint.config.js` file is located in the root directory
of the project which is being linted.

<!-- x-release-please-start-version -->

```yml
name: Lint PR title

on:
  pull_request:
    types: [opened, edited, synchronize]
  merge_group:
    types: [checks_requested]
jobs:
  lint-pr-title:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - id: lint-pr-title
        uses: grafana/shared-workflows/actions/lint-pr-title@lint-pr-title-v1.1.0
        with:
          config-path: "${{ github.workspace }}/commitlint.config.js"
          title-only: false
```

## Example without config path:

```yml
name: Lint PR title

on:
  pull_request:
    types: [opened, edited, synchronize]
jobs:
  lint-pr-title:
    runs-on: ubuntu-latest
    steps:
      - id: lint-pr-title
        uses: grafana/shared-workflows/actions/lint-pr-title@lint-pr-title-v1.1.0
```

<!-- x-release-please-end-version -->
