# Run e2e tests from plugins against specific stack

This is a [GitHub Action][github-action] that help the execution of e2e tests on any plugin against specific selected stacks.
You need to define in which region the selected stack belong, the plugin from where are executed the tests and optionally which other plugins and datasources you want to provision when starting a Grafana instance.
Also, you need to have the **playwright** configuration and the test specifications in the plugin that run the tests and the action will do the rest.
This action use the following input parameters to run:

| Name                  | Description                                                                                     | Default           | Required |
| --------------------- | ----------------------------------------------------------------------------------------------- | ----------------- | -------- |
| `plugin_id`           | Name of the plugin running the tests                                                            |                   | Yes      |
| `stack_slug`          | Name of the stack where you want to run the tests                                               |                   | Yes      |
| `env`                 | Region of the stack where you want to run the tests                                             |                   | Yes      |
| `other_plugins`       | List of other plugins that you want to enable separated by comma                                |                   | No       |
| `datasource_ids`      | List of data sources that you want to enable separated by comma                                 |                   | No       |
| `upload_report_path ` | Name of the folder where you want to store the test report                                      | playwright-report | No       |
| `upload_videos_path`  | Name of the folder where you want to store the test videos                                      | playwright-videos | No       |
| `plugin-secrets`      | A JSON string containing key-value pairs of specific plugin secrets necessary to run the tests. |                   | No       |

## Example workflows

This is an example of how you could use this action.

```yml
name: Build and Test PR

on:
  pull_request:

jobs:
  e2e-tests:
    permissions:
      contents: write
      id-token: write
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          persist-credentials: false

      - name: Get plugin specific secrets
        id: get-secrets
        uses: grafana/shared-workflows/actions/get-vault-secrets@5d7e361bc7e0a183cde8afe9899fb7b596d2659b # v1.2.0
        with:
          repo_secrets: |
            MY_SECRET1=test:token1
            MY_SECRET2=test:token2

      - name: Run e2e cross app tests
        id: e2e-cross-apps-tests
        uses: grafana/shared-workflows/actions/plugins-e2e-tests@main
        with:
          stack_slug: "awsintegrationrevamp"
          env: "dev-central"
          plugin_id: "grafana-csp-app"
          other_plugins: "grafana-k8s-app,grafana-asserts-app"
          datasource_ids: "grafanacloud-awsintegrationrevamp-prom,grafanacloud-awsintegrationrevamp-logs"
          upload_report_path: "playwright-cross-apps-report"
          upload_videos_path: "playwright-cross-apps-videos"
          plugin-secrets: ${{ ${{ steps.get-secrets.outputs.vault_secrets }} }}
```
