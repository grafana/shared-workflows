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
        uses: grafana/shared-workflows/actions/send-slack-message@send-slack-message/v2.0.4
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
        uses: grafana/shared-workflows/actions/send-slack-message@send-slack-message/v2.0.4
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
        uses: grafana/shared-workflows/actions/send-slack-message@send-slack-message/v2.0.4
        with:
          method: chat.postMessage
          payload: |
            {
              "channel": "Channel ID",
              "text": "Deployment started (In Progress)"
            }
      - name: Respond to Slack Message
        uses: grafana/shared-workflows/actions/send-slack-message@send-slack-message/v2.0.4
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

| Name                | Type   | Description                                                                                                                                        |
| ------------------- | ------ | -------------------------------------------------------------------------------------------------------------------------------------------------- |
| `payload`           | String | JSON payload to send.                                                                                                                              |
| `method`            | String | The Slack API method to call.                                                                                                                      |
| `payload-templated` | String | To replace templated variables provided from the step env or default GitHub event context and payload, set the payload-templated variable to true. |

## Outputs

| Name         | Type   | Description                                        |
| ------------ | ------ | -------------------------------------------------- |
| `time`       | String | The time the message was sent.                     |
| `thread_ts`  | String | Threaded timestamp on the message that was posted. |
| `ts`         | String | Timestamp on the message that was posted           |
| `channel_id` | String | The ID of the Slack channel that was posted to.    |
