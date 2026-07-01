import * as core from '@actions/core'
import * as github from '@actions/github'

import { COMMENT_MARKER, buildAllVerifiedCommentBody, buildCommentBody } from './comment'
import { buildSummary } from './summary'

type Octokit = ReturnType<typeof github.getOctokit>

async function run(): Promise<unknown> {
  const token = core.getInput('github-token', { required: true })
  const octokit = github.getOctokit(token)
  const { context } = github

  // For supporting pull requests from forks, we need to support pull_request
  // AND pull_request_target as trigger:
  if (context.eventName !== 'pull_request' && context.eventName !== 'pull_request_target') {
    core.setFailed(`Unsupported event "${context.eventName}"; this action only runs on pull_request events.`)
    return
  }

  const pr = context.payload.pull_request
  if (!pr) {
    core.setFailed('No pull_request payload found on the event.')
    return
  }

  const { owner, repo } = context.repo
  const commits = await octokit.paginate(octokit.rest.pulls.listCommits, {
    owner,
    repo,
    pull_number: pr.number,
    per_page: 100,
  })

  const unverified = commits.filter(c => !c.commit.verification?.verified)

  buildSummary(commits.length, unverified, pr.base.ref, pr.head.ref)

  const existing = await findExistingMarkerComment(octokit, owner, repo, pr.number)
  if (unverified.length > 0) {
    const body = buildCommentBody(unverified, commits.length, pr.base.ref, pr.head.ref)
    await upsertComment(octokit, owner, repo, pr.number, body, existing)
  } else if (existing) {
    const body = buildAllVerifiedCommentBody(commits.length, pr.base.ref, pr.head.ref)
    await upsertComment(octokit, owner, repo, pr.number, body, existing)
  }

  return core.summary.write()
}

async function findExistingMarkerComment(
  octokit: Octokit,
  owner: string,
  repo: string,
  issueNumber: number,
): Promise<{ id: number } | undefined> {
  const comments = await octokit.paginate(octokit.rest.issues.listComments, {
    owner,
    repo,
    issue_number: issueNumber,
    per_page: 100,
  })
  return comments.find(c => c.body?.includes(COMMENT_MARKER))
}

async function upsertComment(
  octokit: Octokit,
  owner: string,
  repo: string,
  issueNumber: number,
  body: string,
  existing: { id: number } | undefined,
): Promise<unknown> {
  if (existing) {
    return octokit.rest.issues.updateComment({ owner, repo, comment_id: existing.id, body })
  }
  return octokit.rest.issues.createComment({ owner, repo, issue_number: issueNumber, body })
}

run().catch(err => core.setFailed(err instanceof Error ? err.message : String(err)))
