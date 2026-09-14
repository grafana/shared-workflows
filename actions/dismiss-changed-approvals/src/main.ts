import { createHash } from "node:crypto";
import { appendFileSync, readFileSync } from "node:fs";

import {
  decideForApproval,
  type ApprovalInfo,
  type CommitInfo,
  type Verdict,
} from "./decide";

const apiBase = process.env.GITHUB_API_URL ?? "https://api.github.com";
const token = process.env["INPUT_GITHUB-TOKEN"] ?? "";
const dryRun = (process.env["INPUT_DRY-RUN"] ?? "false") === "true";

function fail(message: string): void {
  console.log(`::error::${message}`);
  process.exitCode = 1;
}

async function request(
  path: string,
  init: RequestInit = {},
): Promise<Response> {
  const response = await fetch(`${apiBase}${path}`, {
    ...init,
    headers: {
      authorization: `Bearer ${token}`,
      accept: "application/vnd.github+json",
      "x-github-api-version": "2022-11-28",
      ...init.headers,
    },
  });
  if (!response.ok) {
    throw new Error(`${init.method ?? "GET"} ${path}: HTTP ${response.status}`);
  }
  return response;
}

async function paginate<T>(path: string): Promise<T[]> {
  const items: T[] = [];
  for (let page = 1; ; page++) {
    const response = await request(`${path}?per_page=100&page=${page}`);
    const batch = (await response.json()) as T[];
    items.push(...batch);
    if (batch.length < 100) return items;
  }
}

interface Review {
  id: number;
  state: string;
  commit_id: string | null;
  user: { login: string } | null;
}

interface PullCommit {
  sha: string;
  parents: unknown[];
  author: { login: string } | null;
  committer: { login: string } | null;
  commit: { verification?: { verified: boolean } };
}

/**
 * Resolve the latest meaningful review per reviewer; a reviewer's approval is
 * active until they submit a later APPROVED/CHANGES_REQUESTED review or it is
 * dismissed. COMMENTED reviews do not supersede an approval.
 */
async function activeApprovals(
  repo: string,
  pullNumber: number,
): Promise<ApprovalInfo[]> {
  const reviews = await paginate<Review>(
    `/repos/${repo}/pulls/${pullNumber}/reviews`,
  );
  const latest = new Map<string, Review>();
  for (const review of reviews) {
    if (
      review.user &&
      review.state !== "COMMENTED" &&
      review.state !== "PENDING"
    ) {
      latest.set(review.user.login, review);
    }
  }
  return [...latest.values()]
    .filter((r) => r.state === "APPROVED")
    .flatMap((r) =>
      r.user && r.commit_id
        ? [{ reviewId: r.id, reviewer: r.user.login, commitId: r.commit_id }]
        : [],
    );
}

async function listCommits(
  repo: string,
  pullNumber: number,
): Promise<CommitInfo[]> {
  const commits = await paginate<PullCommit>(
    `/repos/${repo}/pulls/${pullNumber}/commits`,
  );
  return commits.map((c) => ({
    sha: c.sha,
    verified: c.commit.verification?.verified === true,
    authorLogin: c.author?.login ?? null,
    committerLogin: c.committer?.login ?? null,
    parentCount: c.parents.length,
  }));
}

/**
 * SHA-256 of the three-dot diff (merge-base diff) between the base ref and the
 * given commit, via the compare API with the diff media type. Returns null
 * when the diff cannot be produced (unknown SHA after a force-push, diff too
 * large, ...) so the caller can fail closed.
 */
async function diffHash(
  repo: string,
  baseRef: string,
  sha: string,
): Promise<string | null> {
  try {
    const response = await request(
      `/repos/${repo}/compare/${encodeURIComponent(baseRef)}...${sha}`,
      { headers: { accept: "application/vnd.github.diff" } },
    );
    return createHash("sha256")
      .update(await response.text())
      .digest("hex");
  } catch (err) {
    console.log(
      `::warning::Failed to compute diff for ${baseRef}...${sha}: ${
        err instanceof Error ? err.message : String(err)
      }`,
    );
    return null;
  }
}

function writeSummary(
  results: { approval: ApprovalInfo; verdict: Verdict }[],
  currentDiffHash: string,
): void {
  const summaryPath = process.env.GITHUB_STEP_SUMMARY;
  if (!summaryPath) return;
  const rows = results
    .map(
      ({ approval, verdict }) =>
        `| @${approval.reviewer} | ${approval.commitId.slice(0, 8)} | ${
          verdict.action === "keep" ? "kept ✅" : "dismissed ❌"
        } | ${verdict.reason} |`,
    )
    .join("\n");
  appendFileSync(
    summaryPath,
    `## Approval dismissal report${dryRun ? " (dry-run)" : ""}\n\n` +
      `Current diff hash: \`${currentDiffHash.slice(0, 16)}…\`\n\n` +
      `| Reviewer | Approved commit | Verdict | Reason |\n` +
      `| --- | --- | --- | --- |\n${rows}\n`,
  );
}

async function run(): Promise<void> {
  if (process.env.GITHUB_EVENT_NAME !== "pull_request") {
    console.log(
      `Nothing to do for event "${process.env.GITHUB_EVENT_NAME}"; this action only evaluates pull_request events.`,
    );
    return;
  }

  const repo = process.env.GITHUB_REPOSITORY ?? "";
  const event = JSON.parse(
    readFileSync(process.env.GITHUB_EVENT_PATH ?? "", "utf8"),
  ) as {
    pull_request?: {
      number: number;
      user: { login: string };
      base: { ref: string };
      head: { sha: string; repo: { full_name: string } | null };
    };
  };
  const pr = event.pull_request;
  if (!pr) {
    fail("No pull_request payload found on the event.");
    return;
  }

  // Fork PRs get a read-only token, so approvals cannot be dismissed. Repos
  // enrolled in this workflow must not rely on fork contributions.
  if (pr.head.repo?.full_name !== repo) {
    console.log(
      "::warning::Pull request comes from a fork; the workflow token cannot dismiss reviews. Skipping.",
    );
    return;
  }

  const approvals = await activeApprovals(repo, pr.number);
  if (approvals.length === 0) {
    console.log("No active approvals; nothing to evaluate.");
    return;
  }

  const currentDiffHash = await diffHash(repo, pr.base.ref, pr.head.sha);
  if (currentDiffHash === null) {
    // Without the current diff nothing can be evaluated. Fail the check
    // WITHOUT dismissing: merging stays blocked, reviews are preserved.
    fail(
      `Could not compute the diff of ${pr.base.ref}...${pr.head.sha}; refusing to evaluate approvals.`,
    );
    return;
  }

  const commits = await listCommits(repo, pr.number);
  const results: { approval: ApprovalInfo; verdict: Verdict }[] = [];
  for (const approval of approvals) {
    const verdict = decideForApproval({
      approval,
      approvalDiffHash:
        approval.commitId === pr.head.sha
          ? currentDiffHash
          : await diffHash(repo, pr.base.ref, approval.commitId),
      currentDiffHash,
      commits,
      prAuthor: pr.user.login,
    });
    results.push({ approval, verdict });

    if (verdict.action === "keep") {
      console.log(`Kept @${approval.reviewer}'s approval: ${verdict.reason}`);
    } else if (dryRun) {
      console.log(
        `::warning::[dry-run] Would dismiss @${approval.reviewer}'s approval: ${verdict.reason}`,
      );
    } else {
      await request(
        `/repos/${repo}/pulls/${pr.number}/reviews/${approval.reviewId}/dismissals`,
        {
          method: "PUT",
          body: JSON.stringify({
            message: `Approval dismissed: ${verdict.reason}.`,
          }),
        },
      );
      console.log(
        `::notice::Dismissed @${approval.reviewer}'s approval: ${verdict.reason}`,
      );
    }
  }

  writeSummary(results, currentDiffHash);
}

run().catch((err) => fail(err instanceof Error ? err.message : String(err)));
