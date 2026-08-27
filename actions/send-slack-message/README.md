# send-slack-message

This is a composite GitHub Action used to send Slack messages to the Grafana workspace.
You do not need to set up Slack webhooks in order to use this action.

See the docs for the [slackapi/slack-github-action workflow](https://tools.slack.dev/slack-github-action/sending-techniques/sending-data-slack-api-method/#usage) for more info.

<!-- x-release-please-start-version -->

```yaml
name: Send And Update a Slack message using JSON payload
jobs:
  send-and-update-slack-message:
    name: Send and Update Slack Message
    steps:
      - name: Send Slack Message via Payload
        id: slack
        uses: grafana/shared-workflows/actions/send-slack-message@send-slack-message/v3.0.2
        with:
          method: chat.postMessage
          payload: |
            {
              "channel": "Channel ID",
              "text": "Deployment started (In Progress)",
              "attachments": [
                {
                  "pretext": "Deployment started",
                  "color": "dbab09",
                  "fields": [
                    {
                      "title": "Status",
                      "short": true,
                      "value": "In Progress"
                    }
                  ]
                }
              ]
            }

      - name: Update Slack Message via Payload
        uses: grafana/shared-workflows/actions/send-slack-message@send-slack-message/v3.0.2
        with:
          method: chat.update
          payload-templated: true
          payload: |
            {
              "channel": ${{ steps.slack.outputs.channel_id }},
              "text": "Deployment finished (Completed)",
              "ts": ${{ steps.slack.outputs.ts }},
              "attachments": [
                {
                  "pretext": "Deployment finished",
                  "color": "28a745",
                  "fields": [
                    {
                      "title": "Status",
                      "short": true,
                      "value": "Completed"
                    }
                  ]
                }
              ]
            }
```

```yaml
name: Send And Respond to a Slack message using JSON payload
jobs:
  send-and-respond-to-slack-message:
    name: Send and respond to Slack message
    steps:
      - name: Post to a Slack channel
        id: slack
        uses: grafana/shared-workflows/actions/send-slack-message@send-slack-message/v3.0.2
        with:
          method: chat.postMessage
          payload: |
            {
              "channel": "Channel ID",
              "text": "Deployment started (In Progress)"
            }
      - name: Respond to Slack Message
        uses: grafana/shared-workflows/actions/send-slack-message@send-slack-message/v3.0.2
        with:
          method: chat.postMessage
          payload-templated: true
          payload: |
            {
              "channel": "Channel ID",
              "thread_ts": "${{ steps.slack.outputs.ts }}",
              "text": "Deployment finished (Completed)"
            }
```

<!-- x-release-please-end-version -->

## Inputs

<!-- BEGIN_INPUTS -->

| Name                | Type    | Required | Default | Description                                                                                                                                       |
| ------------------- | ------- | -------- | ------- | ------------------------------------------------------------------------------------------------------------------------------------------------- |
| `method`            | String  | Yes      |         | The Slack API method to call                                                                                                                      |
| `payload`           | String  | No       |         | JSON payload to send                                                                                                                              |
| `payload-templated` | Boolean | No       | `false` | To replace templated variables provided from the step env or default GitHub event context and payload, set the payload-templated variable to true |

<!-- END_INPUTS -->

## Outputs

<!-- BEGIN_OUTPUTS -->

| Name         | Description                                              |
| ------------ | -------------------------------------------------------- |
| `channel_id` | The channel id of the message that was posted into Slack |
| `thread_ts`  | The timestamp on the latest thread posted into Slack     |
| `time`       | The time that the Slack message was sent                 |
| `ts`         | The timestamp on the message that was posted into Slack  |

<!-- END_OUTPUTS -->
