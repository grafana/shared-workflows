# Trigger Argo Workflow Action

This GitHub action triggers an Argo workflow in one of the Grafana Labs Argo
Workflows instances. It contains a small wrapper around the `argo` CLI, which is
downloaded as part of the action.

As our Argo Workflows instances require authentication, this workflow will only
work when triggered from a `grafana`-owned repository. We would welcome
contributions to extend this action to work with any instance, or you are free
to fork and modify to run on your own instances. See [#21][issue-21].

[issue-21]: https://github.com/grafana/shared-workflows/issues/21

## How to use

## Inputs

- `instance`: The instance to use (`dev` or `ops`). Defaults to `ops`.
- `namespace`: Required. The namespace to trigger the workflow in.
- `command`: The command to run. Defaults to `submit`.
- `parameters`: The newline-separated parameters to pass to the Argo workflow. Example:
```yaml
  parameters: |
    param1=value1
    param2=value2
```
- `workflow_template`: The workflow template to use. Required if `command` is `submit` (the default).
- `extra_args`: Extra arguments to pass to the Argo CLI. Example: `--generate-name foo-`
- `log_level`: The log level to use. Choose from `debug`, `info`, `warn` or `error`. Defaults to `info`.

## Outputs

- `uri`: The URI of the workflow that was created.

## Usage

Here is an example of how to use this action:

```yaml
steps:
- name: Trigger Argo Workflow
uses: actions/trigger-argo-workflow@main
with:
  instance: 'ops'
  namespace: 'mynamespace'
  workflow_template: 'hello'
  parameters: |
    message=world
  extra_args: '--generate-name hello-world-'
  log_level: 'debug'
```
