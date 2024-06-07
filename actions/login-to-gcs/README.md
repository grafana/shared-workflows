# login-to-gcs

This is a composite GitHub Action, used to push objects to a bucket in Google Cloud Storage (GCS).
It uses [OIDC authentication](https://docs.github.com/en/actions/deployment/security-hardening-your-deployments/about-security-hardening-with-openid-connect)
which means that only workflows which get triggered based on certain rules can
trigger these composite workflows.

```yaml
name: Login-to-gcs

on:
  push:
    branches:
      - main

env:
  ENVIRONMENT: "dev"

permissions:
  contents: read
  id-token: write

jobs:
  login-to-gcs:
    name: login-to-gcs
    steps:
      - uses: grafana/shared-workflows/actions/login-to-gcs@rwhitaker/push-to-gcs
        id: login-to-gcs
```

## Inputs

| Name          | Type   | Description                                                                                |
| ------------- | ------ | ------------------------------------------------------------------------------------------ |
| `bucket`      | String | Name of bucket to upload to. Will default to grafanalabs-${repository.name}-${environment} |
| `environment` | String | Environment for pushing artifacts (can be either dev or prod).                             |

## Outputs

| Name     | Type   | Description                                   |
| -------- | ------ | --------------------------------------------- |
| `bucket` | String | Name of the bucket that was authenticated to. |
