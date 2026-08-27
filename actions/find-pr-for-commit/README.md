# Find PR for Commit

This action is used to find the Pull Request associated with a specific commit
reference.

_Note:_ If there are multiple PRs associated with the commit reference, the
action will return the most recently updated PR.

## Inputs

<!-- BEGIN_INPUTS -->

| Name        | Type   | Required | Default                                                                                            | Description                                                                                                                     |
| ----------- | ------ | -------- | -------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------- |
| `commitrev` | String | No       | `${{ github.event_name == 'pull_request' && github.event.pull_request.head.sha \|\| github.sha }}` | The commit SHA, or revision name such as `refs/heads/main`, to find the PR for                                                  |
| `owner`     | String | No       | `${{ github.repository_owner }}`                                                                   | The owner of the repository                                                                                                     |
| `repo`      | String | No       | `${{ github.event.repository.name }}`                                                              | The repository name                                                                                                             |
| `token`     | String | No       | `${{ github.token }}`                                                                              | The GitHub token to use for the query. Must have `contents:read` and `pull-requests:read` permissions on the target repository. |

<!-- END_INPUTS -->

## Outputs

<!-- BEGIN_OUTPUTS -->

| Name        | Description                              |
| ----------- | ---------------------------------------- |
| `pr_number` | The PR number associated with the commit |

<!-- END_OUTPUTS -->

## Usage

Here is an example of how to use the `Find PR for Commit` action:

### Find the PR for the current commit

You might use this if you want to comment on a PR from a workflow triggered by a
push event, e.g. after merging to `main`.

<!-- x-release-please-start-version -->

```yaml
on:
  push:
    branches:
      - main

jobs:
  comment-on-pr-for-commit:
    permissions:
      contents: read
      pull-requests: read
    steps:
      - name: Find PR for current commit
        id: find-pr
        uses: grafana/shared-workflows/actions/find-pr-for-commit@find-pr-for-commit/v1.0.3

      - name: Use PR number
        run: echo "PR Number is ${{ steps.find-pr.outputs.pr_number }}"
```

### Find the PR for a specific commit

```yaml
on:
  push:
    branches:
      - main

jobs:
  comment-on-pr-for-commit:
    permissions:
      contents: read
      pull-requests: read
    steps:
      - name: Find PR for specific commit
        id: find-pr
        uses: grafana/shared-workflows/actions/find-pr-for-commit@find-pr-for-commit/v1.0.3
        with:
          commitrev: "1234567890abcdef1234567890abcdef12345678"

      - name: Use PR number
        run: echo "PR Number is ${{ steps.find-pr.outputs.pr_number }}"
```

### Find the PR for a named revision

```yaml
on:
  push:
    branches:
      - main

jobs:
  comment-on-pr-for-commit:
    permissions:
      contents: read
      pull-requests: read
    steps:
      - name: Find PR for named revision
        id: find-pr
        uses: grafana/shared-workflows/actions/find-pr-for-commit@find-pr-for-commit/v1.0.3
        with:
        commitrev: "HEAD~2"

      - name: Use PR number
        run: echo "PR Number is ${{ steps.find-pr.outputs.pr_number }}"
```

### Find the PR for a commit on another repository

In this case you will need to supply the `owner` and `repo` inputs and a token
in the `token` input, since the default token in Actions runs is scoped to the
current repository only.

```yaml
on:
  push:
    branches:
      - main

jobs:
    comment-on-pr-for-commit:
        steps:
          - name: Find PR for commit in another repository
              id: find-pr
              uses: grafana/shared-workflows/actions/find-pr-for-commit@find-pr-for-commit/v1.0.3
              with:
                owner: "grafana"
                repo: "grafana"
                commitrev: "1234567890abcdef1234567890abcdef12345678"
                token: ${{ secrets.GRAFANA_READ_TOKEN }}

        - name: Use PR number
          run: echo "PR Number is ${{ steps.find-pr.outputs.pr_number }}"
```

Note that `permissions` are not required in this case, as they only affect the
default `${{ github.token }}` and we are supplying our own.

<!-- x-release-please-end-version -->
