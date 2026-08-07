import { buildSummaryTableRow, type PullRequestCommit } from "./summary";

export const COMMENT_MARKER = "<!-- signed-commits-info-action:report -->";
const COMMENT_HEADER = "### Signed commits report";

export function buildCommentBody(
  unverified: PullRequestCommit[],
  total: number,
  baseRef: string,
  headRef: string,
): string {
  const rows = unverified.map((c) => {
    const content = buildSummaryTableRow(c)
      .map(escapeCell)
      .map((cell, idx) => (idx === 0 ? "`" + cell + "`" : cell))
      .join(" | ");
    return `| ${content} |`;
  });

  return [
    COMMENT_MARKER,
    COMMENT_HEADER,
    "",
    `${unverified.length} of ${total} commit${total === 1 ? "" : "s"} between \`${baseRef}\` and \`${headRef}\` could not be fully verified:`,
    "",
    "| Commit | Author | Reason | Message |",
    "| --- | --- | --- | --- |",
    ...rows,
    "",
    "This repository requires all commits to be signed. See [GitHub docs on commit signature verification](https://docs.github.com/authentication/managing-commit-signature-verification/about-commit-signature-verification).",
  ].join("\n");
}

export function buildAllVerifiedCommentBody(
  total: number,
  baseRef: string,
  headRef: string,
): string {
  return [
    COMMENT_MARKER,
    COMMENT_HEADER,
    "",
    `All ${total} commit${total === 1 ? "" : "s"} between \`${baseRef}\` and \`${headRef}\` have verified signatures. ✅`,
  ].join("\n");
}

export function escapeCell(value: string): string {
  return value
    .replace(/\\/g, "\\\\")
    .replace(/\|/g, "\\|")
    .replace(/\r?\n/g, " ");
}
