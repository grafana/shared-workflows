// src/main.ts
import { createHash } from "node:crypto";
import { appendFileSync, readFileSync } from "node:fs";

// src/decide.ts
function isBaseMergeCommit(commit) {
  return commit.parentCount === 2 && commit.committerLogin === "web-flow";
}
function decideForApproval(input) {
  const { approval, approvalDiffHash, currentDiffHash, commits, prAuthor } = input;
  if (approvalDiffHash === null) {
    return {
      action: "dismiss",
      reason: `the diff approved by @${approval.reviewer} (${short(approval.commitId)}) can no longer be computed; history was likely rewritten`
    };
  }
  if (approvalDiffHash !== currentDiffHash) {
    return {
      action: "dismiss",
      reason: `the pull request content changed after @${approval.reviewer}'s approval of ${short(approval.commitId)}`
    };
  }
  const approvedIndex = commits.findIndex((c) => c.sha === approval.commitId);
  const toCheck = approvedIndex >= 0 ? commits.slice(approvedIndex + 1) : commits;
  for (const commit of toCheck) {
    if (!commit.verified) {
      return {
        action: "dismiss",
        reason: `commit ${short(commit.sha)} added after the approval has no verified signature`
      };
    }
    if (!isBaseMergeCommit(commit) && commit.authorLogin !== prAuthor) {
      return {
        action: "dismiss",
        reason: `commit ${short(commit.sha)} added after the approval was authored by ${commit.authorLogin ? `@${commit.authorLogin}` : "an unknown user"}, not the PR author @${prAuthor}`
      };
    }
  }
  return {
    action: "keep",
    reason: approvedIndex >= 0 ? "diff unchanged since approval; all newer commits are verified and attributable" : "diff unchanged since approval (branch was rebased); all commits are verified and authored by the PR author"
  };
}
function short(sha) {
  return sha.slice(0, 8);
}

// src/main.ts
var apiBase = process.env.GITHUB_API_URL ?? "https://api.github.com";
var token = process.env["INPUT_GITHUB-TOKEN"] ?? "";
var dryRun = (process.env["INPUT_DRY-RUN"] ?? "false") === "true";
function fail(message) {
  console.log(`::error::${message}`);
  process.exitCode = 1;
}
async function request(path, init = {}) {
  const response = await fetch(`${apiBase}${path}`, {
    ...init,
    headers: {
      authorization: `Bearer ${token}`,
      accept: "application/vnd.github+json",
      "x-github-api-version": "2022-11-28",
      ...init.headers
    }
  });
  if (!response.ok) {
    throw new Error(`${init.method ?? "GET"} ${path}: HTTP ${response.status}`);
  }
  return response;
}
async function paginate(path) {
  const items = [];
  for (let page = 1;; page++) {
    const response = await request(`${path}?per_page=100&page=${page}`);
    const batch = await response.json();
    items.push(...batch);
    if (batch.length < 100)
      return items;
  }
}
async function activeApprovals(repo, pullNumber) {
  const reviews = await paginate(`/repos/${repo}/pulls/${pullNumber}/reviews`);
  const latest = new Map;
  for (const review of reviews) {
    if (review.user && review.state !== "COMMENTED" && review.state !== "PENDING") {
      latest.set(review.user.login, review);
    }
  }
  return [...latest.values()].filter((r) => r.state === "APPROVED").flatMap((r) => r.user && r.commit_id ? [{ reviewId: r.id, reviewer: r.user.login, commitId: r.commit_id }] : []);
}
async function listCommits(repo, pullNumber) {
  const commits = await paginate(`/repos/${repo}/pulls/${pullNumber}/commits`);
  return commits.map((c) => ({
    sha: c.sha,
    verified: c.commit.verification?.verified === true,
    authorLogin: c.author?.login ?? null,
    committerLogin: c.committer?.login ?? null,
    parentCount: c.parents.length
  }));
}
async function diffHash(repo, baseRef, sha) {
  try {
    const response = await request(`/repos/${repo}/compare/${encodeURIComponent(baseRef)}...${sha}`, { headers: { accept: "application/vnd.github.diff" } });
    return createHash("sha256").update(await response.text()).digest("hex");
  } catch (err) {
    console.log(`::warning::Failed to compute diff for ${baseRef}...${sha}: ${err instanceof Error ? err.message : String(err)}`);
    return null;
  }
}
function writeSummary(results, currentDiffHash) {
  const summaryPath = process.env.GITHUB_STEP_SUMMARY;
  if (!summaryPath)
    return;
  const rows = results.map(({ approval, verdict }) => `| @${approval.reviewer} | ${approval.commitId.slice(0, 8)} | ${verdict.action === "keep" ? "kept ✅" : "dismissed ❌"} | ${verdict.reason} |`).join(`
`);
  appendFileSync(summaryPath, `## Approval dismissal report${dryRun ? " (dry-run)" : ""}

` + `Current diff hash: \`${currentDiffHash.slice(0, 16)}…\`

` + `| Reviewer | Approved commit | Verdict | Reason |
` + `| --- | --- | --- | --- |
${rows}
`);
}
async function run() {
  if (process.env.GITHUB_EVENT_NAME !== "pull_request") {
    console.log(`Nothing to do for event "${process.env.GITHUB_EVENT_NAME}"; this action only evaluates pull_request events.`);
    return;
  }
  const repo = process.env.GITHUB_REPOSITORY ?? "";
  const event = JSON.parse(readFileSync(process.env.GITHUB_EVENT_PATH ?? "", "utf8"));
  const pr = event.pull_request;
  if (!pr) {
    fail("No pull_request payload found on the event.");
    return;
  }
  const isFork = pr.head.repo?.full_name !== repo;
  const approvals = await activeApprovals(repo, pr.number);
  if (approvals.length === 0) {
    console.log("No active approvals; nothing to evaluate.");
    return;
  }
  const currentDiffHash = await diffHash(repo, pr.base.ref, pr.head.sha);
  if (currentDiffHash === null) {
    fail(`Could not compute the diff of ${pr.base.ref}...${pr.head.sha}; refusing to evaluate approvals.`);
    return;
  }
  const commits = await listCommits(repo, pr.number);
  const results = [];
  for (const approval of approvals) {
    const verdict = decideForApproval({
      approval,
      approvalDiffHash: approval.commitId === pr.head.sha ? currentDiffHash : await diffHash(repo, pr.base.ref, approval.commitId),
      currentDiffHash,
      commits,
      prAuthor: pr.user.login
    });
    results.push({ approval, verdict });
    if (verdict.action === "keep") {
      console.log(`Kept @${approval.reviewer}'s approval: ${verdict.reason}`);
    } else if (dryRun) {
      console.log(`::warning::[dry-run] Would dismiss @${approval.reviewer}'s approval: ${verdict.reason}`);
    } else if (isFork) {
      fail(`Would dismiss @${approval.reviewer}'s approval (${verdict.reason}), but the fork token cannot dismiss reviews; failing the check instead.`);
    } else {
      await request(`/repos/${repo}/pulls/${pr.number}/reviews/${approval.reviewId}/dismissals`, {
        method: "PUT",
        body: JSON.stringify({
          message: `Approval dismissed: ${verdict.reason}.`
        })
      });
      console.log(`::notice::Dismissed @${approval.reviewer}'s approval: ${verdict.reason}`);
    }
  }
  writeSummary(results, currentDiffHash);
}
run().catch((err) => fail(err instanceof Error ? err.message : String(err)));
