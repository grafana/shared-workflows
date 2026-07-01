import { describe, expect, test } from 'bun:test'

import {
  COMMENT_MARKER,
  buildAllVerifiedCommentBody,
  buildCommentBody,
  escapeCell,
} from './comment'
import type { PullRequestCommit } from './summary'

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

describe('buildCommentBody', () => {
  test('starts with the hidden marker so upsert can find prior comments', () => {
    const body = buildCommentBody([makeCommit({})], 1, 'main', 'feature')
    expect(body.startsWith(COMMENT_MARKER)).toBe(true)
  })

  test('renders the count, base, and head refs', () => {
    const body = buildCommentBody(
      [makeCommit({}), makeCommit({ sha: 'b'.repeat(40) })],
      5,
      'main',
      'feature/x'
    )
    expect(body).toContain('2 of 5 commits between `main` and `feature/x`')
  })

  test('uses singular "commit" when total is exactly 1', () => {
    const body = buildCommentBody([makeCommit({})], 1, 'main', 'feat')
    expect(body).toContain('1 of 1 commit between')
    expect(body).not.toContain('1 of 1 commits between')
  })

  test('renders one table row per unverified commit with short sha in backticks', () => {
    const body = buildCommentBody(
      [
        makeCommit({ sha: '1234567890abcdef1234567890abcdef12345678', authorName: 'Ada', message: 'first', reason: 'unsigned' }),
        makeCommit({ sha: 'abcdef1234567890abcdef1234567890abcdef12', authorName: 'Grace', message: 'second', reason: 'no_user' }),
      ],
      2,
      'main',
      'feat',
    )
    expect(body).toContain('| `12345678` | Ada | unsigned | first |')
    expect(body).toContain('| `abcdef12` | Grace | no_user | second |')
  })

  test('falls back to author login when commit.author.name is missing', () => {
    const body = buildCommentBody([makeCommit({ authorName: null, authorLogin: 'ada' })], 1, 'main', 'feat')
    expect(body).toContain('| ada |')
  })

  test('falls back to "unknown" when both name and login are missing', () => {
    const body = buildCommentBody(
      [makeCommit({ authorName: null, authorLogin: null })],
      1,
      'main',
      'feat',
    )
    expect(body).toContain('| unknown |')
  })

  test('only includes the first line of multi-line commit messages', () => {
    const body = buildCommentBody(
      [makeCommit({ message: 'subject line\n\nlong body that should be stripped' })],
      1,
      'main',
      'feat',
    )
    expect(body).toContain('subject line')
    expect(body).not.toContain('long body')
  })

  test('escapes pipe characters in cell content', () => {
    const body = buildCommentBody(
      [makeCommit({ authorName: 'a | b', message: 'fix | broken' })],
      1,
      'main',
      'feat',
    )
    expect(body).toContain('a \\| b')
    expect(body).toContain('fix \\| broken')
  })

  test('uses "unknown" reason when verification has no reason field', () => {
    const commit = makeCommit({})
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    ;(commit as any).commit.verification = undefined
    const body = buildCommentBody([commit], 1, 'main', 'feat')
    expect(body).toContain('| unknown |')
  })

  test('links to GitHub docs on signature verification', () => {
    const body = buildCommentBody([makeCommit({})], 1, 'main', 'feat')
    expect(body).toContain('https://docs.github.com/en/authentication/managing-commit-signature-verification/about-commit-signature-verification')
  })
})

describe('buildAllVerifiedCommentBody', () => {
  test('starts with the same marker so it replaces the failure comment', () => {
    const body = buildAllVerifiedCommentBody(3, 'main', 'feat')
    expect(body.startsWith(COMMENT_MARKER)).toBe(true)
  })

  test('renders the total, base, and head refs', () => {
    const body = buildAllVerifiedCommentBody(3, 'main', 'feature/x')
    expect(body).toContain('All 3 commits between `main` and `feature/x` have verified signatures.')
  })

  test('uses singular form when total is 1', () => {
    const body = buildAllVerifiedCommentBody(1, 'main', 'feat')
    expect(body).toContain('All 1 commit between')
    expect(body).not.toContain('All 1 commits between')
  })
})

describe('escapeCell', () => {
  test('escapes pipes', () => {
    expect(escapeCell('a | b | c')).toBe('a \\| b \\| c')
  })

  test('collapses newlines to spaces', () => {
    expect(escapeCell('line1\nline2\r\nline3')).toBe('line1 line2 line3')
  })

  test('leaves plain text untouched', () => {
    expect(escapeCell('nothing special')).toBe('nothing special')
  })
})
