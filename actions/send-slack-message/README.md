# send-slack-notification

This is a composite GitHub Action used to send Slack notifications to the Grafana workspace.
You do not need to set up Slack webhooks in order to use this action.

See the docs for the [slackapi/slack-github-action workflow](https://github.com/slackapi/slack-github-action/blob/main/README.md#technique-2-slack-app) for more info. Our installation is via Slack App.

```yaml
name: Send And Update a Slack plain text message
jobs:
  send-and-update-slack-notification:
    name: Send and update Slack notification
    steps:
      - name: Send Slack Message
        id: slack
        uses: grafana/shared-workflows/actions/send-slack-notification@main
        with:
          channel-id: "Channel Name of ID"
          slack-message: "We are testing, testing, testing all day long"

      - name: Update Slack Message
        uses: grafana/shared-workflows/actions/send-slack-notification@main
        with:
          channel-id: ${{ steps.slack.outputs.channel_id }} # Channel ID is required when updating a message
          slack-message: "This is the updated message"
          update-ts: ${{ steps.slack.outputs.ts }}
```

```yaml
name: Send And Update a Slack message using JSON payload
jobs:
  send-and-update-slack-notification:
    name: Send and Update Slack Message
    steps:
      - name: Send Slack Message via Payload
        id: slack
        uses: grafana/shared-workflows/actions/send-slack-notification@main
        with:
          channel-id: "Channel Name or ID"
          payload: |
            {
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
        uses: grafana/shared-workflows/actions/send-slack-notification@main
        with:
          channel-id: ${{ steps.slack.outputs.channel_id }}
          payload: |
            {
              "text": "Deployment finished (Completed)",
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

          update-ts: ${{ steps.slack.outputs.ts }}
```

```yaml
name: Send And Respond to a Slack message using JSON payload
jobs:
  send-and-respond-to-slack-notification:
    name: Send and respond to Slack notification
    steps:
      - name: Post to a Slack channel
        id: slack
        uses: grafana/shared-workflows/actions/send-slack-notification@main
        with:
          channel-id: "Channel Name of ID"
          payload: |
            {
              "text": "Deployment started (In Progress)"
            }
      - name: Respond to Slack Message
        uses: grafana/shared-workflows/actions/send-slack-notification@main
        with:
          channel-id: ${{ steps.slack.outputs.channel_id }}
          payload: |
            {
              "thread_ts": "${{ steps.slack.outputs.ts }}",
              "text": "Deployment finished (Completed)"
            }
```

## Inputs

| Name            | Type   | Description                                                                 |
| --------------- | ------ | --------------------------------------------------------------------------- |
| `channel-id`    | String | Name or ID of the channel to send to.                                       |
| `payload`       | String | JSON payload to send. Use `payload` or `slack-message`, but not both.       |
| `slack-message` | String | Plain text message to send. Use `payload` or `slack-message`, but not both. |
| `update-ts`     | String | The timestamp of a previous message posted. For updating previous messages. |

## Outputs

| Name         | Type   | Description                                        |
| ------------ | ------ | -------------------------------------------------- |
| `time`       | String | The time the message was sent.                     |
| `thread_ts`  | String | Threaded timestamp on the message that was posted. |
| `ts`         | String | Timestamp on the message that was posted           |
| `channel_id` | String | The ID of the Slack channel that was posted to.    |
