# aws-auth

This is a composite GitHub Action used to authenticate and access resources in AWS.

Example usage in a repository:

<!-- x-release-please-start-version -->

```yaml
name: Authenticate to AWS
on:
  pull_request:

jobs:
  build:
    runs-on: ubuntu-latest
    permissions:
      id-token: write
    steps:
      - id: aws-auth
        uses: grafana/shared-workflows/actions/aws-auth@aws-auth/v1.0.4
        with:
          aws-region: "us-west-1"
          role-arn: "arn:aws:iam::366620023056:role/github-actions/s3-test-access"
          pass-claims: "repository_owner, repository_name, job_workflow_ref, ref, event_name"
          set-creds-in-environment: true

      - id: cat-file-from-s3-bucket
        run: |
          aws s3 cp 's3://grafanalabs-github-actions-test-repo/test.txt' 'test.txt'
          cat 'test.txt'
```

<!-- x-release-please-end-version -->

## Inputs

<!-- markdownlint-disable no-space-in-code -->

<!-- BEGIN_INPUTS -->

| Name                               | Type    | Required | Default                                                                | Description                                                                                                                                                                                                                                                                                                                     |
| ---------------------------------- | ------- | -------- | ---------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `aws-region`                       | String  | Yes      | `us-east-2`                                                            | AWS region that contains the resources you want to use                                                                                                                                                                                                                                                                          |
| `checkout-actions-repository-path` | String  | No       |                                                                        | The path in the filesystem where this repository has been checked out. This is mandatory for setups where executing this action inside a local clone of the repository.                                                                                                                                                         |
| `pass-claims`                      | String  | Yes      | `event_name, repository_owner, repository_name, job_workflow_ref, ref` | `, `-separated [GitHub Actions claims](https://docs.github.com/en/actions/deployment/security-hardening-your-deployments/about-security-hardening-with-openid-connect#understanding-the-oidc-token) (session tags) to make available to `role-arn`. Claims must be mapped to the Cognito Identity Pool before they can be used. |
| `role-arn`                         | String  | Yes      |                                                                        | ARN of the workload role to assume. Must be prefixed with `github-actions`, e.g. `arn:aws:iam::366620023056:role/github-actions/s3-test-access`. See the Setting up Workload Role section below for an example.                                                                                                                 |
| `role-duration-seconds`            | String  | No       | `3600`                                                                 | Role duration in seconds                                                                                                                                                                                                                                                                                                        |
| `set-creds-in-environment`         | Boolean | No       | `true`                                                                 | Set environment variables for AWS CLI and SDKs                                                                                                                                                                                                                                                                                  |

<!-- END_INPUTS -->

<!-- markdownlint-restore -->

See [Setting up Workload Role](#setting-up-workload-role) for an example of a `role-arn`. If you would like to pass a claim that is not in the `pass-claims` default, file an issue in this repo or reach out to `@platform-productivity` in `#platform`.

This uses the [`cognito-idpool-auth`](https://github.com/catnekaise/cognito-idpool-auth) action to perform authentication with an Amazon Cognito Identity Pool using the GitHub Actions OIDC access token.

## Outputs

<!-- BEGIN_OUTPUTS -->

| Name                                 | Description                        |
| ------------------------------------ | ---------------------------------- |
| `aws_access_key_id`                  | AWS Access Key Id                  |
| `aws_region`                         | AWS Region                         |
| `aws_secret_access_key`              | AWS Secret Access Key              |
| `aws_session_token`                  | AWS Session Name                   |
| `cognito_identity_oidc_access_token` | Cognito Identity OIDC Access Token |

<!-- END_OUTPUTS -->

## Setting up Workload Role

IAM workload roles are used to grant permissions to AWS in a secure manner. From a workflow run, once authenticated, the role is granted temporary credentials to access AWS resources permitted by the associated IAM role and attached trust/permission policies. The following steps will guide you through the process of setting up an IAM workload role for read access to a single object in an S3 bucket.

### Create IAM Role

Ensure that the path is prefixed with `github-actions` when creating the role. The Cognito Identity Pool only allows authenticated roles that match the following naming pattern: `"arn:aws:iam::*:role/github-actions/*"`.

The role should only be present in the account that contains the resources it needs to access.

### Trust Policy

This is where you provide additional constraints for when permissions are applied. The condition block can be customized as you see fit with additional [GitHub OIDC token claims](https://docs.github.com/en/actions/deployment/security-hardening-your-deployments/about-security-hardening-with-openid-connect#understanding-the-oidc-token), as long as those claims are passed via `pass-claims`.

As this defines which GitHub Actions runs are allowed to use the role's permissions, it is critical to make these configurations as precise as possible. Furthermore, all runs are limited to be triggered exclusively from repositories under `grafana/`, and it is not possible to exceed this restriction.

In this case, permissions are only granted when the `job_workflow_ref` tag matches the workflow that initiated the action.

Example trust policy:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:aws:iam::590183704419:role/github-actions-oidc-jump-role"
      },
      "Action": ["sts:AssumeRole", "sts:TagSession"],
      "Condition": {
        "StringEquals": {
          "aws:PrincipalTag/job_workflow_ref": "grafana/<REPO>/.github/workflows/<WORKFLOW_FILE>@refs/heads/main"
        }
      }
    }
  ]
}
```

### Permissions Policy

This is where you define the minimum permissions necessary to do a specific operation.

Example permissions policy:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": ["s3:GetObject"],
      "Resource": "arn:aws:s3:::grafanalabs-github-actions-${aws:PrincipalTag/repository_name}/*"
    }
  ]
}
```
