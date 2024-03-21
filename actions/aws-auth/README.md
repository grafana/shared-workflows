# aws-auth

This is a composite GitHub Action used to authenticate and access resources in AWS.
It uses the [`cognito-idpool-auth`](https://github.com/catnekaise/cognito-idpool-auth) gaction to perform authentication with an Amazon Cognito Identity Pool using the GitHub Actions OIDC access token. 

Example usage in a repository:

```yaml
name: Authenticate to AWS
on:
  pull_request:

permissions:
  id_token: write

jobs:
  build:
    runs-on: ubuntu-latest

    steps:
      - id: aws-auth
        uses: grafana/shared-workflows/actions/aws-auth@main
        with:
          chain-role-arn: "arn:aws:iam::<ACCOUNT_ID>:role/github-actions/<WORKLOAD_ROLE>"
          chain-pass-claims: "repository_owner, repository_name, job_workflow_ref"
          chain-set-in-environment: true
```

## Inputs

| Name                          | Type   | Description                                                                                                                                                    |
|-------------------------------|--------|----------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `role-arn`              | String | Specify custom workload role. Role ARN must be prefixed with `github-actions` e.g. `arn:aws:iam::366620023056:role/github-actions/s3-test-access`        |
| `pass-claims`           | String | List of GitHub Actions claims (session tags) to pass to the next session when role chaining (default: `"repository_owner, repository_name, job_workflow_ref"`) |
| `set-creds-in-environment`    | Bool   | Set environment variables for AWS CLI and SDKs (default: `true`)                                                                                        |
| `role-duration-seconds` | String | Role duration in seconds (default: `"3600"`)                                                                                                  |
