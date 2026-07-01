# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

A GitHub Action that runs on `pull_request` events, lists the commits the PR
introduces over its base ref, and reports those without a fully verified
signature to `GITHUB_STEP_SUMMARY` and as a comment to the pull request.
Informational only — never fails the job.

## Toolchain

This project uses **bun**, not npm. Do not introduce `npm`, `yarn`, `pnpm`, or
`@vercel/ncc` to the workflow.

```sh
bun install
bun run build       # bundles src/main.ts → dist/index.js (node target)
bun run typecheck   # tsc --noEmit (tsconfig is noEmit-only; bun does the bundling)
```

`dist/index.js` is the action's runtime entry point referenced from
`action.yml` and **must be committed** so GitHub can execute it without an
install step.

## Architecture

Single-entrypoint Node action. `src/main.ts`:

1. Reads `github-token` input, constructs an Octokit via `@actions/github`.
2. Validates the event is `pull_request` / `pull_request_target`.
3. Paginates `GET /repos/{owner}/{repo}/pulls/{pull_number}/commits` — this
   endpoint already returns the per-commit `verification` object, so there is
   no need to diff refs manually or call the list-commits endpoint separately.
4. Filters commits where `commit.verification.verified !== true` and renders
   them in a table via `core.summary` + generates a pull request comment.
