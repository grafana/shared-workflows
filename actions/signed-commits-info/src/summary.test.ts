import { beforeEach, describe, expect, test } from 'bun:test'
import * as core from '@actions/core'

import { buildSummary, buildSummaryTableRow, type PullRequestCommit } from './summary'

function makeCommit(overrides: {
  sha?: string
  authorName?: string | null
  authorLogin?: string | null
  message?: string
  verified?: boolean
  reason?: string
}): PullRequestCommit {
  const {
    sha = 'a'.repeat(40),
    authorName = 'Ada Lovelace',
    authorLogin = 'ada',
    message = 'Add feature',
    verified = false,
    reason = 'unsigned',
  } = overrides
  return {
    sha,
    commit: {
      author: authorName === null ? null : { name: authorName, email: 'test@example.org', date: '2026-01-01T00:00:00Z' },
      message,
      verification: { verified, reason, signature: null, payload: null, verified_at: null },
    },
    author: authorLogin === null ? null : { login: authorLogin },
  } as unknown as PullRequestCommit
}

describe('buildSummary', () => {
  beforeEach(() => {
    core.summary.emptyBuffer()
  })

  test('renders a heading', () => {
    buildSummary(0, [], 'main', 'feat')
    expect(core.summary.stringify()).toContain('<h1>Signed commits report</h1>')
  })

  test('shows the success message when everything is verified', () => {
    buildSummary(3, [], 'main', 'feat')
    const out = core.summary.stringify()
    expect(out).toContain('Checked 3 commits between <code>main</code> and <code>feat</code>.')
    expect(out).toContain('All commits have verified signatures. ✅')
    expect(out).not.toContain('<table>')
  })

  test('uses singular form when total is 1', () => {
    buildSummary(1, [], 'main', 'feat')
    expect(core.summary.stringify()).toContain('Checked 1 commit between')
  })

  test('renders the unverified count and a table when commits fail verification', () => {
    const commit = makeCommit({
      sha: '1234567890abcdef1234567890abcdef12345678',
      authorName: 'Ada',
      message: 'broken commit\nbody discarded',
      reason: 'unsigned',
    })
    buildSummary(2, [commit], 'main', 'feat')
    const out = core.summary.stringify()

    expect(out).toContain('1 commit could not be fully verified:')
    expect(out).toContain('<table>')
    expect(out).toContain('<th>Commit</th>')
    expect(out).toContain('<th>Author</th>')
    expect(out).toContain('<th>Reason</th>')
    expect(out).toContain('<th>Message</th>')
    expect(out).toContain('<td>12345678</td>')
    expect(out).toContain('<td>Ada</td>')
    expect(out).toContain('<td>unsigned</td>')
    expect(out).toContain('<td>broken commit</td>')
    expect(out).not.toContain('body discarded')
  })

  test('uses plural form when multiple commits are unverified', () => {
    buildSummary(
      5,
      [makeCommit({ sha: 'a'.repeat(40) }), makeCommit({ sha: 'b'.repeat(40) })],
      'main',
      'feat',
    )
    expect(core.summary.stringify()).toContain('2 commits could not be fully verified:')
  })

  test('includes the docs link only when there are unverified commits', () => {
    buildSummary(1, [makeCommit({})], 'main', 'feat')
    expect(core.summary.stringify()).toContain('docs.github.com')

    core.summary.emptyBuffer()
    buildSummary(1, [], 'main', 'feat')
    expect(core.summary.stringify()).not.toContain('docs.github.com')
  })

  test('falls back to author login then "unknown" when name is missing', () => {
    buildSummary(
      2,
      [
        makeCommit({ sha: '1'.repeat(40), authorName: null, authorLogin: 'gracehopper' }),
        makeCommit({ sha: '2'.repeat(40), authorName: null, authorLogin: null }),
      ],
      'main',
      'feat',
    )
    const out = core.summary.stringify()
    expect(out).toContain('<td>gracehopper</td>')
    expect(out).toContain('<td>unknown</td>')
  })
})

describe('buildSummaryTableRow', () => {
  test('returns four non-empty string cells even when author fields are missing', () => {
    const row = buildSummaryTableRow(makeCommit({ authorName: null, authorLogin: null }))
    expect(row).toHaveLength(4)
    for (const cell of row) {
      expect(typeof cell).toBe('string')
      expect(cell.length).toBeGreaterThan(0)
    }
  })

  test('returns four string cells when verification is missing', () => {
    const commit = makeCommit({});
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    (commit as any).commit.verification = undefined;
    const row = buildSummaryTableRow(commit)
    expect(row).toHaveLength(4)
    for (const cell of row) {
      expect(typeof cell).toBe('string')
    }
  })
})
