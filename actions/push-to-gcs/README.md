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
  ENVIRONMENT: "dev"

permissions:
  contents: read
  id-token: write

jobs:
  upload-to-gcs:
    name: upload
    runs-on: ubuntu-x64-small
    steps:
      - uses: actions/checkout@v4
      - uses: grafana/shared-workflows/actions/login-to-gcs@rwhitaker/push-to-gcs
        id: login-to-gcs

        # Upload a single file to the bucket root
      - uses: grafana/shared-workflows/actions/push-to-gcs@main
        with:
          bucket: ${{ steps.login-to-gcs.outputs.bucket }}
          path: file.txt

        # Here are 3 equivalent statements to upload a single file and its parent directory to the bucket root
      - uses: grafana/shared-workflows/actions/push-to-gcs@main
        with:
          bucket: ${{ steps.login-to-gcs.outputs.bucket }}
          path: folder/file.txt
      - uses: grafana/shared-workflows/actions/push-to-gcs@main
        with:
          bucket: ${{ steps.login-to-gcs.outputs.bucket }}
          path: .
          glob: "folder/file.txt"
      - uses: grafana/shared-workflows/actions/push-to-gcs@main
        with:
          bucket: ${{ steps.login-to-gcs.outputs.bucket }}
          path: folder
          glob: "file.txt"

        # Here are 2 equivalent statements to upload a single file WITHOUT its parent directory to the bucket root
      - uses: grafana/shared-workflows/actions/push-to-gcs@main
        with:
          bucket: ${{ steps.login-to-gcs.outputs.bucket }}
          path: folder/file.txt
          parent: false
      - uses: grafana/shared-workflows/actions/push-to-gcs@main
        with:
          bucket: ${{ steps.login-to-gcs.outputs.bucket }}
          path: folder
          glob: "file.txt"
          parent: false

        # Here are 2 equivalent statements to upload a directory with all subdirectories
      - uses: grafana/shared-workflows/actions/push-to-gcs@main
        with:
          bucket: ${{ steps.login-to-gcs.outputs.bucket }}
          path: folder/
      - uses: grafana/shared-workflows/actions/push-to-gcs@main
        with:
          bucket: ${{ steps.login-to-gcs.outputs.bucket }}
          path: .
          glob: "folder/**/*"

        # Specify a bucket prefix with `bucket_path`
      - uses: grafana/shared-workflows/actions/push-to-gcs@main
        name: upload-yaml-to-some-path
        with:
          bucket: ${{ steps.login-to-gcs.outputs.bucket }}
          path: file.txt
          bucket_path: some-path/

        # Upload all files of a type
      - uses: grafana/shared-workflows/actions/push-to-gcs@main
        with:
          bucket: ${{ steps.login-to-gcs.outputs.bucket }}
          path: folder/
          glob: "*.txt"

        # upload all files of a type recursively
      - uses: grafana/shared-workflows/actions/push-to-gcs@main
        with:
          bucket: ${{ steps.login-to-gcs.outputs.bucket }}
          path: folder/
          glob: "**/*.txt"
```

## Inputs

| Name          | Type   | Description                                                                                                                                                                                  |
| ------------- | ------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `bucket`      | String | (Required) Name of bucket to upload to. Can be gathered from `login-to-gcs` action.                                                                                                          |
| `path`        | String | (Required) The path to a file or folder inside the action's filesystem that should be uploaded to the bucket. You can specify either the absolute path or the relative path from the action. |
| `bucket_path` | String | Bucket path where objects will be uploaded. Default is the bucket root.                                                                                                                      |
| `environment` | String | Environment for pushing artifacts (can be either dev or prod).                                                                                                                               |
| `glob`        | String | Glob pattern.                                                                                                                                                                                |
| `parent`      | String | Whether parent dir should be included in GCS destination. Dirs included in the `glob` statement are unaffected by this setting.                                                              |

## Outputs

| Name       | Type   | Description                                        |
| ---------- | ------ | -------------------------------------------------- |
| `uploaded` | String | The list of files that were successfully uploaded. |
