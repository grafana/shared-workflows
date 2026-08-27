# login-to-gcs

This is a composite GitHub Action, used to push objects to a bucket in Google Cloud Storage (GCS).
It uses [OIDC authentication](https://docs.github.com/en/actions/deployment/security-hardening-your-deployments/about-security-hardening-with-openid-connect)
which means that only workflows which get triggered based on certain rules can
trigger these composite workflows.

<!-- x-release-please-start-version -->

```yaml
name: Login-to-gcs

on:
  push:
    branches:
      - main

jobs:
  login-to-gcs:
    name: login-to-gcs
    permissions:
      contents: read
      id-token: write
    steps:
      - uses: grafana/shared-workflows/actions/login-to-gcs@login-to-gcs/v0.3.1
        id: login-to-gcs
```

<!-- x-release-please-end-version -->

You can now use the shared-workflow `push-to-gcs` or gcloud to push objects from your CI pipeline.

Ex:

```
$ gcloud storage cp OBJECT_LOCATION gs://DESTINATION_BUCKET_NAME
```

## Inputs

<!-- BEGIN_INPUTS -->

| Name                      | Type    | Required | Default | Description                                                                                                                             |
| ------------------------- | ------- | -------- | ------- | --------------------------------------------------------------------------------------------------------------------------------------- |
| `bucket`                  | String  | No       |         | Name of bucket to upload to. Will default to grafanalabs-${repository.name}-${environment}                                              |
| `delete_credentials_file` | Boolean | No       | `false` | Delete the credentials file after the action is finished. If you want to keep the credentials file for a later step, set this to false. |
| `environment`             | String  | No       | `dev`   | Environment for uploading objects (can be either dev or prod).                                                                          |
| `service_account`         | String  | No       |         | Custom service account to use for authentication.                                                                                       |
| `use_wif_auth`            | Boolean | No       | `false` | Use WIF for authentication instead of service account.                                                                                  |

<!-- END_INPUTS -->

## Outputs

<!-- BEGIN_OUTPUTS -->

| Name     | Description                                           |
| -------- | ----------------------------------------------------- |
| `bucket` | The name of the bucket that we have authenticated to. |

<!-- END_OUTPUTS -->
