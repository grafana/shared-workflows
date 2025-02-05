# Check Drone Signature

## Overview

This document describes the intention and usage of the `check-drone-signature` GitHub Actions workflow.
You can find [the `check-drone-signature` workflow file here][wf-file].

This reusable workflow checks [the signature of a Drone CI file][drone-sig] against our Drone server and fails if the signature is invalid.
The check can help prevent against merging incorrectly signed files, which often occurs when the Drone CI config is modified without re-signing the file.
When a signature is invalid, Drone CI builds require approval from a user with write or administrative access to proceed.
Only those users with write or administrative access to the repository can create the signature using `drone sign $owner/$repo --save`.

This check is particularly useful in our public OSS repositories, as those require a signed config file.
Beyond public repos, any Drone repositories that are `Protected` require a signed config file.

[wf-file]: ./check-drone-signature.yaml
[drone-sig]: https://docs.drone.io/signature/

## Usage

The following is an example of how to use the `check-drone-signature` workflow in your repository:

```yaml
name: Check Drone CI Signature

on:
  push:
    branches:
      - "main"
    paths:
      - ".drone.yml"
  pull_request:
    paths:
      - ".drone.yml"

permissions:
  id-token: write
  contents: read
  issues: write

jobs:
  drone-signature-check:
    uses: grafana/shared-workflows/.github/workflows/check-drone-signature.yaml@main
    with:
      drone_config_path: .drone.yml
```

This workflow file expects the following:

- `main` is the name of your primary branch
- `.drone.yml` is the filename for your Drone CI configuration file and it's located at the root of your repository

For Grafanistas, you can use [this software template in EngHub][enghub-tmpl] to easily configure and add the workflow file to your repository.

[enghub-tmpl]: https://enghub.grafana-ops.net/create/templates/default/add-drone-signature-check-workflow

### Forks

Because forks do not have access to the secrets required to check the signature, the workflow will not run on pull requests from forks.
A message will be posted on the pull request indicating that the signature check was skipped.

## Inputs

| Name                | Description                             | Type   |
| ------------------- | --------------------------------------- | ------ |
| `drone_config_path` | Path to the Drone CI configuration file | string |
| `drone_server`      | Drone CI server URL                     | string |
