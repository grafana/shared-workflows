/**
 * Pure decision logic for content-based approval dismissal.
 *
 * An approval is kept if and only if the pull request's effective diff
 * (three-dot diff against the base branch) is byte-identical to the diff the
 * reviewer approved, AND every commit added after the approval is signed and
 * attributable (PR author or a GitHub-generated merge of the base branch).
 *
 * Anything the logic cannot positively verify results in a dismissal —
 * the module fails closed.
 */

export interface ApprovalInfo {
  /** Review ID, used for the dismissal API call. */
  reviewId: number;
  /** Login of the approving reviewer. */
  reviewer: string;
  /** Head commit SHA of the PR at the time the approval was submitted. */
  commitId: string;
}

export interface CommitInfo {
  sha: string;
  /** GitHub's signature verification result for the commit. */
  verified: boolean;
  /** Login of the commit author (null when unattributable). */
  authorLogin: string | null;
  /** Login of the committer; "web-flow" for GitHub-UI-generated commits. */
  committerLogin: string | null;
  parentCount: number;
}

export type Verdict =
  { action: "keep"; reason: string } | { action: "dismiss"; reason: string };

/**
 * GitHub-generated merge commits of the base branch ("Update branch" button,
 * merge-queue updates) are committed by the web-flow user and have exactly two
 * parents. Their content is already covered by the diff-equality check; this
 * predicate only exempts them from the "authored by the PR author" rule.
 */
export function isBaseMergeCommit(commit: CommitInfo): boolean {
  return commit.parentCount === 2 && commit.committerLogin === "web-flow";
}

export function decideForApproval(input: {
  approval: ApprovalInfo;
  /**
   * SHA-256 of the three-dot diff at the approved commit, or null when it
   * could not be computed (e.g. the commit vanished after a force-push).
   */
  approvalDiffHash: string | null;
  /** SHA-256 of the three-dot diff at the current head. */
  currentDiffHash: string;
  /** All commits currently on the PR, oldest first. */
  commits: CommitInfo[];
  /** Login of the PR author. */
  prAuthor: string;
}): Verdict {
  const { approval, approvalDiffHash, currentDiffHash, commits, prAuthor } =
    input;

  if (approvalDiffHash === null) {
    return {
      action: "dismiss",
      reason: `the diff approved by @${approval.reviewer} (${short(
        approval.commitId,
      )}) can no longer be computed; history was likely rewritten`,
    };
  }

  if (approvalDiffHash !== currentDiffHash) {
    return {
      action: "dismiss",
      reason: `the pull request content changed after @${approval.reviewer}'s approval of ${short(
        approval.commitId,
      )}`,
    };
  }

  // The content is identical. Validate the provenance of everything pushed
  // after the approval. If the approved commit is no longer part of the PR
  // (rebase), attribution of "new" commits is impossible, so every commit is
  // checked instead.
  const approvedIndex = commits.findIndex((c) => c.sha === approval.commitId);
  const toCheck =
    approvedIndex >= 0 ? commits.slice(approvedIndex + 1) : commits;

  for (const commit of toCheck) {
    if (!commit.verified) {
      return {
        action: "dismiss",
        reason: `commit ${short(commit.sha)} added after the approval has no verified signature`,
      };
    }
    if (!isBaseMergeCommit(commit) && commit.authorLogin !== prAuthor) {
      return {
        action: "dismiss",
        reason: `commit ${short(commit.sha)} added after the approval was authored by ${
          commit.authorLogin ? `@${commit.authorLogin}` : "an unknown user"
        }, not the PR author @${prAuthor}`,
      };
    }
  }

  return {
    action: "keep",
    reason:
      approvedIndex >= 0
        ? "diff unchanged since approval; all newer commits are verified and attributable"
        : "diff unchanged since approval (branch was rebased); all commits are verified and authored by the PR author",
  };
}

function short(sha: string): string {
  return sha.slice(0, 8);
}
