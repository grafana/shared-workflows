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
          path: .github/workflows/upload-files-to-gcs.yaml

      - uses: grafana/shared-workflows/actions/push-to-gcs@main
        name: upload-Dockerfile-to-root
        with:
          path: Dockerfile

        # upload a file to a folder in the bucket
      - uses: grafana/shared-workflows/actions/push-to-gcs@main
        name: upload-yaml-to-some-path
        with:
          path: .github/workflows/upload-files-to-gcs.yaml
          bucket_path: some-path/

      - uses: grafana/shared-workflows/actions/push-to-gcs@main
        name: upload-Dockerfile-to-some-path
        with:
          path: Dockerfile
          bucket_path: some-path

        # upload .yml files ./docs to bucket/docs
      - uses: grafana/shared-workflows/actions/push-to-gcs@main
        name: upload-all-yml-docs
        with:
          path: docs
          glob: "*.yml"

        # upload .yml files from docs to bucket/this-folder/docs
      - uses: grafana/shared-workflows/actions/push-to-gcs@main
        name: upload-all-yml-docs
        with:
          path: docs
          glob: "*.yml"
          bucket_path: this-folder
```

## Inputs

| Name          | Type   | Description                                                                                                                                                                                                                                                                                      |
|---------------|--------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `bucket`      | String | Name of bucket to upload to. (Default: `grafanalabs-${repository.id}-${environment}`)                                                                                                                                                                                                            |
| `path`        | String | Path to the object(s) to upload. Can either include a filename to upload 1 file, or can be used in conjunction with the `glob` option to upload matching files from a path. Valid examples include `thing.txt` and `path/to/thing.txt`. Valid examples when also using `glob` include `path/to`. |
| `bucket_path` | String | The path in the bucket to save the object(s). Valid examples include `some-path`, `some-path/`, `some/path`. (Default: root of bucket)                                                                                                                                                           |
| `environment` | String | Environment for pushing artifacts (can be either dev or prod).                                                                                                                                                                                                                                   |
| `glob`        | String | Glob pattern.                                                                                                                                                                                                                                                                                    |

## Outputs

| Name       | Type   | Description                                        |
|------------|--------|----------------------------------------------------|
| `uploaded` | String | The list of files that were successfully uploaded. |
