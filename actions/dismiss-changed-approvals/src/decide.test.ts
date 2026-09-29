import { describe, expect, test } from "bun:test";

import {
  decideForApproval,
  isBaseMergeCommit,
  type ApprovalInfo,
  type CommitInfo,
} from "./decide";

const HASH_A = "a".repeat(64);
const HASH_B = "b".repeat(64);

function approval(overrides: Partial<ApprovalInfo> = {}): ApprovalInfo {
  return {
    reviewId: 1,
    reviewer: "reviewer",
    commitId: "c1".padEnd(40, "0"),
    ...overrides,
  };
}

function commit(overrides: Partial<CommitInfo> = {}): CommitInfo {
  return {
    sha: "c1".padEnd(40, "0"),
    verified: true,
    authorLogin: "author",
    committerLogin: "author",
    parentCount: 1,
    ...overrides,
  };
}

const baseMerge = (sha: string): CommitInfo =>
  commit({
    sha,
    authorLogin: "author",
    committerLogin: "web-flow",
    parentCount: 2,
  });

describe("isBaseMergeCommit", () => {
  test("web-flow committer with two parents is a base merge", () => {
    expect(isBaseMergeCommit(baseMerge("m1".padEnd(40, "0")))).toBe(true);
  });

  test("regular single-parent commit is not", () => {
    expect(isBaseMergeCommit(commit())).toBe(false);
  });

  test("local two-parent merge pushed by a user is not", () => {
    expect(
      isBaseMergeCommit(commit({ parentCount: 2, committerLogin: "author" })),
    ).toBe(false);
  });
});

describe("decideForApproval", () => {
  test("keeps approval when nothing changed", () => {
    const verdict = decideForApproval({
      approval: approval(),
      approvalDiffHash: HASH_A,
      currentDiffHash: HASH_A,
      commits: [commit()],
      prAuthor: "author",
    });
    expect(verdict.action).toBe("keep");
  });

  test("keeps approval after a clean base-branch merge via GitHub UI", () => {
    const verdict = decideForApproval({
      approval: approval(),
      approvalDiffHash: HASH_A,
      currentDiffHash: HASH_A,
      commits: [commit(), baseMerge("m1".padEnd(40, "0"))],
      prAuthor: "author",
    });
    expect(verdict.action).toBe("keep");
  });

  test("dismisses when the diff changed", () => {
    const verdict = decideForApproval({
      approval: approval(),
      approvalDiffHash: HASH_A,
      currentDiffHash: HASH_B,
      commits: [commit(), commit({ sha: "c2".padEnd(40, "0") })],
      prAuthor: "author",
    });
    expect(verdict.action).toBe("dismiss");
    expect(verdict.reason).toContain("content changed");
  });

  test("dismisses when the approved diff is no longer computable", () => {
    const verdict = decideForApproval({
      approval: approval(),
      approvalDiffHash: null,
      currentDiffHash: HASH_A,
      commits: [commit()],
      prAuthor: "author",
    });
    expect(verdict.action).toBe("dismiss");
    expect(verdict.reason).toContain("history was likely rewritten");
  });

  test("dismisses when a commit after the approval is unsigned, even with an identical diff", () => {
    const verdict = decideForApproval({
      approval: approval(),
      approvalDiffHash: HASH_A,
      currentDiffHash: HASH_A,
      commits: [
        commit(),
        commit({ sha: "c2".padEnd(40, "0"), verified: false }),
      ],
      prAuthor: "author",
    });
    expect(verdict.action).toBe("dismiss");
    expect(verdict.reason).toContain("no verified signature");
  });

  test("dismisses when a commit after the approval comes from another user", () => {
    const verdict = decideForApproval({
      approval: approval(),
      approvalDiffHash: HASH_A,
      currentDiffHash: HASH_A,
      commits: [
        commit(),
        commit({ sha: "c2".padEnd(40, "0"), authorLogin: "mallory" }),
      ],
      prAuthor: "author",
    });
    expect(verdict.action).toBe("dismiss");
    expect(verdict.reason).toContain("@mallory");
  });

  test("ignores unsigned or foreign commits that predate the approval", () => {
    // The reviewer saw and approved these commits; only later pushes matter.
    const verdict = decideForApproval({
      approval: approval({ commitId: "c2".padEnd(40, "0") }),
      approvalDiffHash: HASH_A,
      currentDiffHash: HASH_A,
      commits: [
        commit({ verified: false, authorLogin: "colleague" }),
        commit({ sha: "c2".padEnd(40, "0") }),
      ],
      prAuthor: "author",
    });
    expect(verdict.action).toBe("keep");
  });

  test("keeps approval after a clean rebase when all commits are attributable", () => {
    // Approved commit no longer exists; identical diff, every commit checked.
    const verdict = decideForApproval({
      approval: approval({ commitId: "gone".padEnd(40, "0") }),
      approvalDiffHash: HASH_A,
      currentDiffHash: HASH_A,
      commits: [commit({ sha: "r1".padEnd(40, "0") })],
      prAuthor: "author",
    });
    expect(verdict.action).toBe("keep");
    expect(verdict.reason).toContain("rebased");
  });

  test("dismisses after a rebase when any commit is not attributable to the PR author", () => {
    const verdict = decideForApproval({
      approval: approval({ commitId: "gone".padEnd(40, "0") }),
      approvalDiffHash: HASH_A,
      currentDiffHash: HASH_A,
      commits: [
        commit({ sha: "r1".padEnd(40, "0") }),
        commit({ sha: "r2".padEnd(40, "0"), authorLogin: "colleague" }),
      ],
      prAuthor: "author",
    });
    expect(verdict.action).toBe("dismiss");
  });
});
