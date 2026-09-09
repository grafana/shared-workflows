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
      - uses: grafana/shared-workflows/actions/push-to-gcs@push-to-gcs/v0.3.2
        with:
          bucket: ${{ steps.login-to-gcs.outputs.bucket }}
          path: file.txt
          environment: "dev" # Can be dev/prod (defaults to dev)

        # Upload a single file and apply a predefined ACL. See `predefinedAcl` for options.
      - uses: grafana/shared-workflows/actions/push-to-gcs@push-to-gcs/v0.3.2
        with:
          bucket: ${{ steps.login-to-gcs.outputs.bucket }}
          path: file.txt
          predefinedAcl: projectPrivate
          environment: "dev"

        # Here are 3 equivalent statements to upload a single file and its parent directory to the bucket root
      - uses: grafana/shared-workflows/actions/push-to-gcs@push-to-gcs/v0.3.2
        with:
          bucket: ${{ steps.login-to-gcs.outputs.bucket }}
          path: folder/file.txt
      - uses: grafana/shared-workflows/actions/push-to-gcs@push-to-gcs/v0.3.2
        with:
          bucket: ${{ steps.login-to-gcs.outputs.bucket }}
          path: .
          glob: "folder/file.txt"
      - uses: grafana/shared-workflows/actions/push-to-gcs@push-to-gcs/v0.3.2
        with:
          bucket: ${{ steps.login-to-gcs.outputs.bucket }}
          path: folder
          glob: "file.txt"

        # Here are 2 equivalent statements to upload a single file WITHOUT its parent directory to the bucket root
      - uses: grafana/shared-workflows/actions/push-to-gcs@push-to-gcs/v0.3.2
        with:
          bucket: ${{ steps.login-to-gcs.outputs.bucket }}
          path: folder/file.txt
          parent: false
      - uses: grafana/shared-workflows/actions/push-to-gcs@push-to-gcs/v0.3.2
        with:
          bucket: ${{ steps.login-to-gcs.outputs.bucket }}
          path: folder
          glob: "file.txt"
          parent: false

        # Here are 2 equivalent statements to upload a directory with all subdirectories
      - uses: grafana/shared-workflows/actions/push-to-gcs@push-to-gcs/v0.3.2
        with:
          bucket: ${{ steps.login-to-gcs.outputs.bucket }}
          path: folder/
      - uses: grafana/shared-workflows/actions/push-to-gcs@push-to-gcs/v0.3.2
        with:
          bucket: ${{ steps.login-to-gcs.outputs.bucket }}
          path: .
          glob: "folder/**/*"

        # Specify a bucket prefix with `bucket_path`
      - uses: grafana/shared-workflows/actions/push-to-gcs@push-to-gcs/v0.3.2
        name: upload-yaml-to-some-path
        with:
          bucket: ${{ steps.login-to-gcs.outputs.bucket }}
          path: file.txt
          bucket_path: some-path/

        # Upload all files of a type
      - uses: grafana/shared-workflows/actions/push-to-gcs@push-to-gcs/v0.3.2
        with:
          bucket: ${{ steps.login-to-gcs.outputs.bucket }}
          path: folder/
          glob: "*.txt"

        # upload all files of a type recursively
      - uses: grafana/shared-workflows/actions/push-to-gcs@push-to-gcs/v0.3.2
        with:
          bucket: ${{ steps.login-to-gcs.outputs.bucket }}
          path: folder/
          glob: "**/*.txt"
```

<!-- x-release-please-end-version -->

## Inputs

<!-- BEGIN_INPUTS -->

| Name                      | Type    | Required | Default          | Description                                                                                                                                                                       |
| ------------------------- | ------- | -------- | ---------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `bucket`                  | String  | No       |                  | Name of bucket to upload to. Can be gathered from the `login-to-gcs` action. Will default to grafanalabs-${repository.name}-${environment}                                        |
| `bucket_path`             | String  | No       |                  | Bucket path where objects will be uploaded. Default is the bucket root.                                                                                                           |
| `delete_credentials_file` | Boolean | No       | `true`           | Delete the credentials file after the action is finished. If you want to keep the credentials file for a later step, set this to false.                                           |
| `environment`             | String  | No       | `dev`            | Environment for uploading objects (can be either dev or prod).                                                                                                                    |
| `glob`                    | String  | No       |                  | Glob pattern.                                                                                                                                                                     |
| `gzip`                    | Boolean | No       | `true`           | If true, then upload files with `content-encoding: gzip`                                                                                                                          |
| `parent`                  | Boolean | No       | `true`           | Whether parent dir should be included in GCS destination. Dirs included in the `glob` statement are unaffected by this setting.                                                   |
| `path`                    | String  | Yes      |                  | The path to a file or folder inside the action's filesystem that should be uploaded to the bucket. You can specify either the absolute path or the relative path from the action. |
| `predefinedAcl`           | String  | No       | `projectPrivate` | Apply a predefined set of access controls to the file(s). Default is projectPrivate (See https://googleapis.dev/nodejs/storage/latest/global.html#UploadOptions)                  |
| `service_account`         | String  | No       |                  | Custom service account to use for authentication.                                                                                                                                 |
| `use_wif_auth`            | Boolean | No       | `false`          | Use WIF for authentication instead of service account.                                                                                                                            |

<!-- END_INPUTS -->

> [!TIP]
> To use WIF authentication you must enable `uniform_bucket_level_access` on the destination bucket. If you are at Grafana Labs, instructions can be found [here](https://enghub.grafana-ops.net/docs/default/component/deployment-tools/platform/continuous-integration/google-artifact-registry/). More info can be found in [Google's docs](https://cloud.google.com/storage/docs/uniform-bucket-level-access).

## Outputs

<!-- BEGIN_OUTPUTS -->

| Name       | Description                              |
| ---------- | ---------------------------------------- |
| `uploaded` | The list of successfully uploaded files. |

<!-- END_OUTPUTS -->
