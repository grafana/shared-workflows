# Lint pull request title

This is a [Github Action](https://github.com/features/actions) that ensures that your PR title matches the [Commitlint Spec](https://github.com/conventional-changelog/commitlint) according to the configuration file.

This is helpful when using the "Squash and merge" strategy, Github will suggest to use the PR title as the commit message. With this action you can validate that the PR title will lead to a correct commit message.

## Build

If you make any change on the `index.js` file, you should build the code before to commit the changes using the command:

```
yarn build
```

## Validation

Examples for valid PR titles:

- fix(some-scope): correct typo.
- feat: add support for Node 12.

### The commit config

This action can receive the `commitlint.config.js` file using the input parameter `config-path`.
This parameter should contain the absolute path of the file.
So, it's recommended to start the path definition with `${{ github.workspace }}/`. See example below:

```
...
    uses: grafana/shared-workflows/actions/lint-pr-title@main
    with:
      config-path: '${{ github.workspace }}/dir1/dir2/commitlint.config.js'
...

```
but this `config-path` parameter is optional. If you don't want to define it, the action will use the following rules by default:

```
{
  'body-leading-blank': [1, 'always'],
  'body-max-line-length': [2, 'always', 100],
  'footer-leading-blank': [1, 'always'],
  'footer-max-line-length': [2, 'always', 100],
  'header-max-length': [2, 'always', 100],
  'subject-case': [2, 'never', ['sentence-case', 'start-case', 'pascal-case', 'upper-case']],
  'subject-empty': [2, 'never'],
  'subject-full-stop': [2, 'never', '.'],
  'type-case': [2, 'always', 'lower-case'],
  'type-empty': [2, 'never'],
  'type-enum': [
    2,
    'always',
    ['build', 'chore', 'ci', 'docs', 'feat', 'fix', 'perf', 'refactor', 'revert', 'style', 'test']
  ]
}
```

## Example workflows

### Example with config path defined:
In this example the `commitlint.config.js` file is located in the root directory from where the action is being executed.

```yml
name: Lint PR title

on:
  pull_request:
    types: [opened, edited, synchronize]
jobs:
  lint-pr-title:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - id: lint-pr-title
        uses: grafana/shared-workflows/actions/lint-pr-title@main
        with:
          config-path: '${{ github.workspace }}/commitlint.config.js'
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}


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
        uses: grafana/shared-workflows/actions/lint-pr-title@main
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```
