# aws-auth

This is a composite GitHub Action used to authenticate and access resources in AWS.

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

| Name                       | Type   | Description                                                                                                                                                                           |
|----------------------------|--------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `role-arn`                 | String | Specify custom workload role. Role ARN must be prefixed with `github-actions` e.g. `arn:aws:iam::366620023056:role/github-actions/s3-test-access`                                     |
| `pass-claims`              | String | `, `-separated list of [GitHub Actions claims](https://docs.github.com/en/actions/deployment/security-hardening-your-deployments/about-security-hardening-with-openid-connect#understanding-the-oidc-token) (session tags) to make available to `role-arn`. Currently supported claims (default): `"repository_owner, repository_name, job_workflow_ref"` [^1] |
| `set-creds-in-environment` | Bool   | Set environment variables for AWS CLI and SDKs (default: `true`)                                                                                                                      |
| `role-duration-seconds`    | String | Role duration in seconds (default: `"3600"`)                                                                                                                                          |

[^1]: GitHub OIDC token claims must be mapped to the Cognito identity pool before they can be used. If you would like to use a claim that is not listed, file an issue in this repo or reach out to `@platform-productivity` in `#platform`.

This uses the [`cognito-idpool-auth`](https://github.com/catnekaise/cognito-idpool-auth) action to perform authentication with an Amazon Cognito Identity Pool using the GitHub Actions OIDC access token.
