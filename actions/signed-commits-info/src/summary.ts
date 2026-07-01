import * as core from '@actions/core'
import type * as github from '@actions/github'

type Octokit = ReturnType<typeof github.getOctokit>
export type PullRequestCommit = Awaited<
  ReturnType<Octokit['rest']['pulls']['listCommits']>
>['data'][number]

export function buildSummaryTableRow(commit: PullRequestCommit): [string, string, string, string] {
  const sha = commit.sha.substring(0, 8)
  const author = commit.commit.author?.name ?? commit.author?.login ?? 'unknown'
  const messageLine = commit.commit.message.split('\n')[0] ?? ''
  return [sha, author, commit.commit.verification?.reason ?? 'unknown', messageLine]
}

export function buildSummary(
  total: number,
  unverified: PullRequestCommit[],
  baseRef: string,
  headRef: string,
): void {
  const summary = core.summary.addHeading('Signed commits report')
  summary.addRaw(
    `Checked ${total} commit${total === 1 ? '' : 's'} between <code>${baseRef}</code> and <code>${headRef}</code>.`,
    true,
  )
  summary.addBreak()

  if (unverified.length === 0) {
    summary.addRaw('All commits have verified signatures. ✅', true)
    return
  }

  summary.addRaw(
    `${unverified.length} commit${unverified.length === 1 ? '' : 's'} could not be fully verified:`,
    true,
  )
  summary.addBreak()
  summary.addBreak()
  summary.addTable([
    [
      { data: 'Commit', header: true },
      { data: 'Author', header: true },
      { data: 'Reason', header: true },
      { data: 'Message', header: true },
    ],
    ...unverified.map(c => [...buildSummaryTableRow(c)]),
  ])
  summary.addBreak()
  summary.addRaw(
    'This repository requires all commits to be signed. You can learn more about signing your commits at <a href="https://docs.github.com/en/authentication/managing-commit-signature-verification/about-commit-signature-verification">docs.github.com</a>.',
    true,
  )
}
