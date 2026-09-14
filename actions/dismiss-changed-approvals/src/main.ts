import { createHash } from "node:crypto";

import * as core from "@actions/core";
import * as github from "@actions/github";

import {
  decideForApproval,
  type ApprovalInfo,
  type CommitInfo,
  type Verdict,
} from "./decide";

type Octokit = ReturnType<typeof github.getOctokit>;

async function run(): Promise<void> {
  const token = core.getInput("github-token", { required: true });
  const dryRun = core.getBooleanInput("dry-run");
  const octokit = github.getOctokit(token);
  const { context } = github;

  if (context.eventName !== "pull_request") {
    core.info(
      `Nothing to do for event "${context.eventName}"; this action only evaluates pull_request events.`,
    );
    return;
  }

  const pr = context.payload.pull_request;
  if (!pr) {
    core.setFailed("No pull_request payload found on the event.");
    return;
  }

  const { owner, repo } = context.repo;

  // Fork PRs get a read-only token, so approvals cannot be dismissed. Repos
  // enrolled in this workflow must not rely on fork contributions.
  if (pr.head.repo?.full_name !== `${owner}/${repo}`) {
    core.warning(
      "Pull request comes from a fork; the workflow token cannot dismiss reviews. Skipping.",
    );
    return;
  }

  const approvals = await activeApprovals(octokit, owner, repo, pr.number);
  if (approvals.length === 0) {
    core.info("No active approvals; nothing to evaluate.");
    return;
  }

  const commits = await listCommits(octokit, owner, repo, pr.number);
  const prAuthor: string = pr.user.login;
  const headSha: string = pr.head.sha;
  const baseRef: string = pr.base.ref;

  const currentDiffHash = await diffHash(
    octokit,
    owner,
    repo,
    baseRef,
    headSha,
  );
  if (currentDiffHash === null) {
    // Without the current diff nothing can be evaluated. Fail the check
    // WITHOUT dismissing: merging stays blocked, reviews are preserved.
    core.setFailed(
      `Could not compute the diff of ${baseRef}...${headSha}; refusing to evaluate approvals.`,
    );
    return;
  }

  const results: { approval: ApprovalInfo; verdict: Verdict }[] = [];
  for (const approval of approvals) {
    const approvalDiffHash =
      approval.commitId === headSha
        ? currentDiffHash
        : await diffHash(octokit, owner, repo, baseRef, approval.commitId);

    const verdict = decideForApproval({
      approval,
      approvalDiffHash,
      currentDiffHash,
      commits,
      prAuthor,
    });
    results.push({ approval, verdict });

    if (verdict.action === "dismiss") {
      if (dryRun) {
        core.warning(
          `[dry-run] Would dismiss @${approval.reviewer}'s approval: ${verdict.reason}`,
        );
      } else {
        await octokit.rest.pulls.dismissReview({
          owner,
          repo,
          pull_number: pr.number,
          review_id: approval.reviewId,
          message: `Approval dismissed: ${verdict.reason}.`,
        });
        core.notice(
          `Dismissed @${approval.reviewer}'s approval: ${verdict.reason}`,
        );
      }
    } else {
      core.info(`Kept @${approval.reviewer}'s approval: ${verdict.reason}`);
    }
  }

  await writeSummary(results, currentDiffHash, dryRun);
}

/**
 * Resolve the latest meaningful review per reviewer; a reviewer's approval is
 * active until they submit a later APPROVED/CHANGES_REQUESTED review or it is
 * dismissed. COMMENTED reviews do not supersede an approval.
 */
async function activeApprovals(
  octokit: Octokit,
  owner: string,
  repo: string,
  pullNumber: number,
): Promise<ApprovalInfo[]> {
  const reviews = await octokit.paginate(octokit.rest.pulls.listReviews, {
    owner,
    repo,
    pull_number: pullNumber,
    per_page: 100,
  });

  const latest = new Map<string, (typeof reviews)[number]>();
  for (const review of reviews) {
    const login = review.user?.login;
    if (!login || review.state === "COMMENTED" || review.state === "PENDING") {
      continue;
    }
    latest.set(login, review);
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
  octokit: Octokit,
  owner: string,
  repo: string,
  pullNumber: number,
): Promise<CommitInfo[]> {
  const commits = await octokit.paginate(octokit.rest.pulls.listCommits, {
    owner,
    repo,
    pull_number: pullNumber,
    per_page: 100,
  });
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
 * given commit, computed via the compare API with the diff media type. Returns
 * null when the diff cannot be produced (unknown SHA after a force-push, diff
 * too large, ...) so the caller can fail closed.
 */
async function diffHash(
  octokit: Octokit,
  owner: string,
  repo: string,
  baseRef: string,
  sha: string,
): Promise<string | null> {
  try {
    const response = await octokit.request(
      "GET /repos/{owner}/{repo}/compare/{basehead}",
      {
        owner,
        repo,
        basehead: `${baseRef}...${sha}`,
        headers: { accept: "application/vnd.github.diff" },
      },
    );
    return createHash("sha256")
      .update(response.data as unknown as string)
      .digest("hex");
  } catch (err) {
    core.warning(
      `Failed to compute diff for ${baseRef}...${sha}: ${
        err instanceof Error ? err.message : String(err)
      }`,
    );
    return null;
  }
}

async function writeSummary(
  results: { approval: ApprovalInfo; verdict: Verdict }[],
  currentDiffHash: string,
  dryRun: boolean,
): Promise<void> {
  core.summary.addHeading(
    `Approval dismissal report${dryRun ? " (dry-run)" : ""}`,
  );
  core.summary.addRaw(
    `<p>Current diff hash: <code>${currentDiffHash.slice(0, 16)}…</code></p>`,
    true,
  );
  core.summary.addTable([
    [
      { data: "Reviewer", header: true },
      { data: "Approved commit", header: true },
      { data: "Verdict", header: true },
      { data: "Reason", header: true },
    ],
    ...results.map(({ approval, verdict }) => [
      `@${approval.reviewer}`,
      approval.commitId.slice(0, 8),
      verdict.action === "keep" ? "kept ✅" : "dismissed ❌",
      verdict.reason,
    ]),
  ]);
  await core.summary.write();
}

run().catch((err) =>
  core.setFailed(err instanceof Error ? err.message : String(err)),
);
