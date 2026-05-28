import { slackifyMarkdown } from "slackify-markdown";

export function transform(markdown: string | undefined | null): string {
  return slackifyMarkdown(markdown ?? "");
}

process.stdout.write(transform(process.env.INPUT_MARKDOWN));
