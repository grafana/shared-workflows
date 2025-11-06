# Techdocs: Rewrite relative links

**Note:** This action is intended to be primarily used within the publish-techdocs workflow.

This action's job is to scan through all the Markdown files inside the docs
folder (based on the presence of an `mkdocs.yml` file) and rewrite relative
links that point to files _outside_ that docs folder folder to absolute ones.

## Usage

The following example will check out the shared-workflows project and run the action from it.
The action will then check inside the working directory if there is a `mkdocs.yml` file and process the docs mentioned in there.

Assuming that the docs folder for mkdocs is located at `/workspace/docs` and there is a `filename.md` in there with content like this:

```markdown
[outside link](../README.md)
```

Then this link inside the file will be changed to ...

```markdown
[outside link](https://github.com/grafana/reponame/blob/main/README.md)
```

<!-- x-release-please-start-version -->

```yaml
- id: checkout-shared-workflows
  name: Checkout shared workflows
  uses: actions/checkout@b4ffde65f46336ab88eb53be808477a3936bae11 # v1.0.3
  with:
    repository: grafana/shared-workflows
    ref: techdocs-rewrite-relative-links-v1.0.3
    path: _shared-workflows
    persist-credentials: false

- name: Rewrite relative links
  uses: ./_shared-workflows/actions/techdocs-rewrite-relative-links
  with:
    working-directory: ./
    repo-url: "https://github.com/${{ github.repository }}"
    default-branch: "${{ github.event.repository.default_branch }}"
    dry-run: false

    # Since the previous step already checked out the shared-workflows repo, we can use that:
    checkout-action-repository: true
    checkout-action-repository-path: _shared-workflows
```

<!-- x-release-please-end-version -->

Follow that up with the actions that should publish the docs to EngHub. See [the `publish-techdocs.yaml` workflow](https://github.com/grafana/shared-workflows/blob/main/.github/workflows/publish-techdocs.yaml) for details.

## Inputs

| Name                                                   | Type    | Description                                                                                                    |
| ------------------------------------------------------ | ------- | -------------------------------------------------------------------------------------------------------------- |
| `default-branch` (required)                            | string  | Default branch name of the repository                                                                          |
| `repo-url` (required)                                  | string  | Full URL to the GitHub repository                                                                              |
| `working-directory` (required)                         | string  | Directory containing the `mkdocs.yml` file                                                                     |
| `dry-run`                                              | boolean | Do not modify the files but print a diff                                                                       |
| `checkout-action-repository-path` (default: `_action`) | string  | Folder where the repository should be checked out to for running the action or where a checkout already exists |
| `checkout-action-repository` (default: `true`)         | boolean | If the workflow already checks out the shared-workflows repository, you can set this to false                  |
| `verbose` (default `false`)                            | boolean | Log on info level                                                                                              |
| `debug` (default `false`)                              | boolean | Log on debug level                                                                                             |
