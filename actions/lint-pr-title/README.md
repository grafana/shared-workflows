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

This action uses the following commitlint configuration rules:

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

You can adjust the `commitlint.config.js` if required.

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
      - uses: grafana/shared-workflows/actions/lint-pr-title@main
        id: lint-pr-title
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```
