# push-to-gcs

> [!NOTE]
> If you are at Grafana Labs, follow these steps in the [internal documentation](https://enghub.grafana-ops.net/docs/default/component/deployment-tools/platform/continuous-integration/google-artifact-registry/) to set up a GCS bucket before using this action.

This is a composite GitHub Action, used to push objects to a bucket in Google Cloud Storage (GCS).
It uses [OIDC authentication](https://docs.github.com/en/actions/deployment/security-hardening-your-deployments/about-security-hardening-with-openid-connect)
which means that only workflows which get triggered based on certain rules can
trigger these composite workflows.

<!-- x-release-please-start-version -->

```yaml
name: Upload Files to GCS

on:
  push:
    branches:
      - main

jobs:
  upload-to-gcs:
    name: upload
    runs-on: ubuntu-x64-small
    permissions:
      contents: read
      id-token: write
    steps:
      - uses: actions/checkout@v4
        with:
          persist-credentials: false

      - uses: grafana/shared-workflows/actions/login-to-gcs@main
        id: login-to-gcs

        # Upload a single file to the bucket root
      - uses: grafana/shared-workflows/actions/push-to-gcs@push-to-gcs/v0.3.0
        with:
          bucket: ${{ steps.login-to-gcs.outputs.bucket }}
          path: file.txt
          environment: "dev" # Can be dev/prod (defaults to dev)

        # Upload a single file and apply a predefined ACL. See `predefinedAcl` for options.
      - uses: grafana/shared-workflows/actions/push-to-gcs@push-to-gcs/v0.3.0
        with:
          bucket: ${{ steps.login-to-gcs.outputs.bucket }}
          path: file.txt
          predefinedAcl: projectPrivate
          environment: "dev"

        # Here are 3 equivalent statements to upload a single file and its parent directory to the bucket root
      - uses: grafana/shared-workflows/actions/push-to-gcs@push-to-gcs/v0.3.0
        with:
          bucket: ${{ steps.login-to-gcs.outputs.bucket }}
          path: folder/file.txt
      - uses: grafana/shared-workflows/actions/push-to-gcs@push-to-gcs/v0.3.0
        with:
          bucket: ${{ steps.login-to-gcs.outputs.bucket }}
          path: .
          glob: "folder/file.txt"
      - uses: grafana/shared-workflows/actions/push-to-gcs@push-to-gcs/v0.3.0
        with:
          bucket: ${{ steps.login-to-gcs.outputs.bucket }}
          path: folder
          glob: "file.txt"

        # Here are 2 equivalent statements to upload a single file WITHOUT its parent directory to the bucket root
      - uses: grafana/shared-workflows/actions/push-to-gcs@push-to-gcs/v0.3.0
        with:
          bucket: ${{ steps.login-to-gcs.outputs.bucket }}
          path: folder/file.txt
          parent: false
      - uses: grafana/shared-workflows/actions/push-to-gcs@push-to-gcs/v0.3.0
        with:
          bucket: ${{ steps.login-to-gcs.outputs.bucket }}
          path: folder
          glob: "file.txt"
          parent: false

        # Here are 2 equivalent statements to upload a directory with all subdirectories
      - uses: grafana/shared-workflows/actions/push-to-gcs@push-to-gcs/v0.3.0
        with:
          bucket: ${{ steps.login-to-gcs.outputs.bucket }}
          path: folder/
      - uses: grafana/shared-workflows/actions/push-to-gcs@push-to-gcs/v0.3.0
        with:
          bucket: ${{ steps.login-to-gcs.outputs.bucket }}
          path: .
          glob: "folder/**/*"

        # Specify a bucket prefix with `bucket_path`
      - uses: grafana/shared-workflows/actions/push-to-gcs@push-to-gcs/v0.3.0
        name: upload-yaml-to-some-path
        with:
          bucket: ${{ steps.login-to-gcs.outputs.bucket }}
          path: file.txt
          bucket_path: some-path/

        # Upload all files of a type
      - uses: grafana/shared-workflows/actions/push-to-gcs@push-to-gcs/v0.3.0
        with:
          bucket: ${{ steps.login-to-gcs.outputs.bucket }}
          path: folder/
          glob: "*.txt"

        # upload all files of a type recursively
      - uses: grafana/shared-workflows/actions/push-to-gcs@push-to-gcs/v0.3.0
        with:
          bucket: ${{ steps.login-to-gcs.outputs.bucket }}
          path: folder/
          glob: "**/*.txt"
```

<!-- x-release-please-end-version -->

## Inputs

| Name                      | Type    | Description                                                                                                                                                                                  |
| ------------------------- | ------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `bucket`                  | String  | (Required) Name of bucket to upload to. Can be gathered from `login-to-gcs` action.                                                                                                          |
| `path`                    | String  | (Required) The path to a file or folder inside the action's filesystem that should be uploaded to the bucket. You can specify either the absolute path or the relative path from the action. |
| `bucket_path`             | String  | Bucket path where objects will be uploaded. Default is the bucket root.                                                                                                                      |
| `environment`             | String  | Environment for pushing artifacts (can be either dev or prod).                                                                                                                               |
| `service_account`         | String  | Service account to use for authentication, different than the default one. Used only when bucket input is not empty (i.e. when the bucket is not the default one).                           |
| `glob`                    | String  | Glob pattern.                                                                                                                                                                                |
| `parent`                  | String  | Whether parent dir should be included in GCS destination. Dirs included in the `glob` statement are unaffected by this setting.                                                              |
| `predefinedAcl`           | String  | Predefined ACL applied to the uploaded objects. Default is `projectPrivate`. See [Google Documentation][gcs-docs-upload-options] for a list of available options.                            |
| `delete_credentials_file` | Boolean | Delete the credentials file after the action is finished. If you want to keep the credentials file for a later step, set this to false. (Default: `true`)                                    |
| `use_wif_auth`            | Boolean | Use WIF authentication. Overrides the `service_account` input.                                                                                                                               |

> [!TIP]
> To use WIF authentication you must enable `uniform_bucket_level_access` on the destination bucket. If you are at Grafana Labs, instructions can be found [here](https://enghub.grafana-ops.net/docs/default/component/deployment-tools/platform/continuous-integration/google-artifact-registry/). More info can be found in [Google's docs](https://cloud.google.com/storage/docs/uniform-bucket-level-access).

## Outputs

| Name       | Type   | Description                                        |
| ---------- | ------ | -------------------------------------------------- |
| `uploaded` | String | The list of files that were successfully uploaded. |

[gcs-docs-upload-options]: https://googleapis.dev/nodejs/storage/latest/global.html#UploadOptions
