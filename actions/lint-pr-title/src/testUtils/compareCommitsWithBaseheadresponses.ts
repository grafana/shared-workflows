// Mock responses for the `compareCommitsWithBasehead` function. The `commits`
// field's `commit.parent` and `commit.message` members only, since that's what
// we make use of. GitHub offers multiple merge methods. Here we're testing
// multiple scenarios for them all.

// Tip: use code folding in your editor to make this file more readable.

const mergeSingleCommit = {
  name: "one commit, only branch in queue",
  mergeMethod: "merge",
  commits: [
    {
      sha: "badcommit",
      commit: {
        message: "bad commit message",
      },
      parents: [{}],
    },
    {
      sha: "mergecommit",
      // This is a merge commit created by GitHub. You can tell that as it has
      // two parents. We don't want to lint this commit message, even though it
      // will end up in the history of the target branch.
      commit: {
        message:
          "Merge pull request #44 from someorg/someuser/mq-1\n\nPrint commits for merge queue",
      },
      parents: [{}, {}],
    },
  ],
  valid: false,
  expectedCheckedCommits: ["badcommit"],
};

const mergeSingleCommitSecondPR = {
  name: "one commit, second branch in queue",
  mergeMethod: "merge",
  commits: [
    {
      sha: "goodcommit",
      commit: {
        message: "fix(everything): commit 2\n\nand a nice body too",
      },
      parents: [{}],
    },
    {
      sha: "mergecommit",
      // Functionally the same as the above. This was taken from the second PR
      // in a merge group with two PRs in it, just to verify that we select the
      // right commits to lint.
      commit: {
        message: "Merge pull request #46 from grafana/iainlane/mq-2\n\nmq 2",
      },
      parents: [{}, {}],
    },
  ],
  valid: true,
  expectedCheckedCommits: ["goodcommit"],
};

const mergeTwoCommits = {
  name: "two commits, only branch in queue",
  mergeMethod: "merge",
  commits: [
    {
      sha: "goodcommit",
      commit: {
        message: "ci(blah): this commit is good",
      },
      parents: [{}],
    },
    {
      sha: "goodcommit2",
      commit: {
        message:
          "fix(everything): this should be an okay commit message\n\nand a nice body too",
      },
      parents: [{}],
    },
    {
      sha: "mergecommit",
      commit: {
        message:
          "Merge pull request #44 from grafana/iainlane/mq-1\n\nPrint commits for merge queue",
      },
      parents: [{}, {}],
    },
  ],
  valid: true,
  expectedCheckedCommits: ["goodcommit", "goodcommit2"],
};

const squashSingleCommit = {
  name: "one commit, only branch in queue",
  mergeMethod: "squash",
  commits: [
    {
      sha: "goodcommit",
      commit: {
        message: "fix(all): commit 1 (#47)",
      },
      parents: [{}],
    },
  ],
  valid: true,
  expectedCheckedCommits: ["goodcommit"],
};

const squashBehindMainAndTwoCommits = {
  name: "two commits, behind main by one commit",
  mergeMethod: "squash",
  commits: [
    {
      sha: "goodcommit",
      commit: {
        message:
          "fix(ci): this is the PR's title (#48)\n\n* fix(ci) commit 1\n\nthis commit is fine\n\n* fix(ci): commit 2.\n\nthis commit is not fine. the subject ends with a period!",
      },
      parents: [{}],
    },
  ],
  valid: true,
  expectedCheckedCommits: ["goodcommit"],
};

const squashSingleCommitSecondPR = {
  name: "one commit, second branch in queue",
  mergeMethod: "squash",
  commits: [
    {
      sha: "badcommit",
      commit: {
        message: "bad title, no scope (#46)\n\nand a nice body too",
      },
      parents: [{}],
    },
  ],

  valid: false,
  expectedCheckedCommits: ["badcommit"],
};

const rebaseSingleCommit = {
  name: "one commit, only branch in queue",
  mergeMethod: "rebase",
  commits: [
    {
      sha: "goodcommit",
      commit: {
        message: "fix(all): commit 1",
      },
      parents: [{}],
    },
  ],
  valid: true,
  expectedCheckedCommits: ["goodcommit"],
};

const rebaseBehindMainAndTwoCommits = {
  name: "two commits, behind main by one commit",
  mergeMethod: "rebase",
  commits: [
    {
      sha: "goodcommit",
      commit: {
        message: "fix(ci) commit 1\n\nthis commit is fine",
      },
      parents: [{}],
    },
    {
      sha: "badcommit",
      commit: {
        message:
          "fix(ci): commit 2.\n\nthis commit is not fine. the subject ends with a period!",
      },
      parents: [{}],
    },
  ],
  valid: false,
  expectedCheckedCommits: ["goodcommit", "badcommit"],
};

const rebaseSingleCommitSecondPR = {
  name: "one commit, second branch in queue",
  mergeMethod: "rebase",
  commits: [
    {
      sha: "badcommit",
      commit: {
        message: "fix(everything): Commit 2\n\nand a nice body too",
      },
      parents: [{}],
    },
  ],
  // Capital `C` in `Commit 2` makes this invalid
  valid: false,
  expectedCheckedCommits: ["badcommit"],
};

/**
 * Example responses from the `compareCommitsWithBasehead` function. The
 * structure of these responses has been taken from real workflow runs. They are
 * used to test the linting of commit messages in a merge group, in an
 * `it.each()` table test in `main.test.ts`.
 */
export const compareCommitsWithBaseheadResponses = [
  mergeSingleCommit,
  mergeSingleCommitSecondPR,
  mergeTwoCommits,
  squashSingleCommit,
  squashBehindMainAndTwoCommits,
  squashSingleCommitSecondPR,
  rebaseSingleCommit,
  rebaseBehindMainAndTwoCommits,
  rebaseSingleCommitSecondPR,
];
