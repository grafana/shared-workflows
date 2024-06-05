# push-to-gcs

This is a composite GitHub Action, used to push objects to a bucket in Google Cloud Storage (GCS).
It uses [OIDC authentication](https://docs.github.com/en/actions/deployment/security-hardening-your-deployments/about-security-hardening-with-openid-connect)
which means that only workflows which get triggered based on certain rules can
trigger these composite workflows.

```yaml
name: Upload Files to GCS

on:
  push:
    branches:
      - main

env:
  ENVIRONMENT: 'dev'

permissions:
  contents: read
  id-token: write

jobs:
  upload-to-gcs:
    name: upload
    runs-on: ubuntu-x64-small
    steps:
      - uses: actions/checkout@v4

        # upload a file to the root of a bucket
      - uses: grafana/shared-workflows/actions/push-to-gcs@main
        name: upload-yaml-to-root
        with:
          object: .github/workflows/upload-files-to-gcs.yaml

      - uses: grafana/shared-workflows/actions/push-to-gcs@main
        name: upload-Dockerfile-to-root
        with:
          object: Dockerfile

        # upload a file to a folder in the bucket
      - uses: grafana/shared-workflows/actions/push-to-gcs@main
        name: upload-yaml-to-some-path
        with:
          object: .github/workflows/upload-files-to-gcs.yaml
          bucket_path: some-path/

      - uses: grafana/shared-workflows/actions/push-to-gcs@main
        name: upload-Dockerfile-to-some-path
        with:
          object: Dockerfile
          bucket_path: some-path
```

## Inputs

| Name          | Type   | Description                                                                                                                               |
|---------------|--------|-------------------------------------------------------------------------------------------------------------------------------------------|
| `bucket`      | String | Name of bucket to upload to. (Default: `grafanalabs-${repository.id}-${environment}`)                                                     |
| `object`      | String | Object name and location to upload to. Will create sub-folders in the bucket. Valid examples include `thing.txt` and `path/to/thing.txt`. |
| `bucket_path` | String | The path in the bucket to save the object. (Default: /)                                                                                   |
| `environment` | String | Environment for pushing artifacts (can be either dev or prod).                                                                            |
