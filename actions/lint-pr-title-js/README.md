# Lint pull request title Javascript

This is a [Github Action](https://github.com/features/actions) used from `actions/lint-pr-title` action to lint the PR title.

## Example workflow:

```yml
name: Lint pull request title

on:
  pull_request:
    types:
      - opened
      - edited
      - synchronize

jobs:
  main:
    runs-on: ubuntu-latest
    steps:
      - uses: grafana/shared-workflows/actions/lint-pr-title-js@main
        id: lint-pr-title
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```
