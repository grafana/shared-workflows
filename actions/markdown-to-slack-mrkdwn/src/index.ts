import * as core from "@actions/core";
import { slackifyMarkdown } from "slackify-markdown";

export function transform(markdown: string | undefined | null): string {
  return slackifyMarkdown(markdown ?? "");
}

if (import.meta.main) {
  // trimWhitespace: false so leading/trailing whitespace in the input is
  // preserved verbatim — the conversion should not silently reshape the text.
  core.setOutput(
    "text",
    transform(core.getInput("markdown", { trimWhitespace: false })),
  );
}
