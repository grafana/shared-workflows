# markdown-to-slack-mrkdwn

GitHub composite action that converts CommonMark markdown to Slack `mrkdwn` — the syntax accepted inside `blocks[].text` of a `chat.postMessage` payload. Wraps [`slackify-markdown`](https://github.com/jsarafajr/slackify-markdown).

Typical use: take a release-please CHANGELOG diff or a GitHub release body and pipe it into a Slack Block Kit `mrkdwn` section so links, headings, bold, and lists render correctly in Slack.

## Inputs

| Name       | Required | Description               |
| ---------- | -------- | ------------------------- |
| `markdown` | yes      | Markdown text to convert. |

## Outputs

| Name   | Description                |
| ------ | -------------------------- |
| `text` | Slack mrkdwn-encoded text. |

## Usage

<!-- x-release-please-start-version -->

```yaml
- name: Convert markdown to Slack mrkdwn
  id: markdown
  uses: grafana/shared-workflows/actions/markdown-to-slack-mrkdwn@markdown-to-slack-mrkdwn/v0.1.0
  with:
    markdown: ${{ env.CHANGELOG_CONTENT }}

- name: Send Slack message
  uses: grafana/shared-workflows/actions/send-slack-message@send-slack-message/v2.0.5
  with:
    channel-id: my-channel
    payload: |
      {
        "blocks": [
          {
            "type": "section",
            "text": {
              "type": "mrkdwn",
              "text": ${{ toJSON(steps.markdown.outputs.text) }}
            }
          }
        ]
      }
```

<!-- x-release-please-end-version -->

## Notes

- The output `text` value is the _raw_ Slack-mrkdwn string. When embedding it in a JSON payload (as in the example above), pipe it through `toJSON(...)` to JSON-encode quotes, newlines, etc.
- `slackify-markdown` produces a trailing newline; that's preserved here.
- Both `**bold**` and `*bold*` from CommonMark become `*bold*` in mrkdwn. CommonMark italic (`*x*`) becomes `_x_`.
