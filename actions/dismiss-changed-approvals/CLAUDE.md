# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

A GitHub Action that runs on `pull_request` events and dismisses stale
approvals **only when the reviewed content actually changed**, instead of on
every push like the native branch protection. It exists so that high-velocity
repositories can keep approvals across clean base-branch merges and rebases
without weakening the two-person-review guarantee.

Security invariants — do not weaken these:

1. The action only dismisses reviews; it must never create or approve them.
2. Everything fails closed: verdicts that cannot be computed dismiss the
   approval; runtime errors fail the job without dismissing (merge stays
   blocked by the required check).
3. Diff equality is byte-exact (SHA-256 over the three-dot diff from the
   compare API). No fuzzy or semantic comparison.

## Toolchain

This project uses **bun**, not npm. Do not introduce `npm`, `yarn`, `pnpm`, or
`@vercel/ncc` to the workflow.

```sh
bun install
bun run build       # bundles src/main.ts → dist/index.js (node target)
bun run typecheck   # tsc --noEmit (tsconfig is noEmit-only; bun does the bundling)
bun test
```

`dist/index.js` is the action's runtime entry point referenced from
`action.yml` and **must be committed** so GitHub can execute it without an
install step.

## Architecture

- `src/decide.ts` — pure, fully tested decision logic
  (`decideForApproval`). Takes precomputed diff hashes plus commit metadata
  and returns a keep/dismiss verdict per approval.
- `src/main.ts` — orchestration: resolves active approvals from
  `pulls.listReviews` (latest non-COMMENTED review per user), lists PR commits
  with their `verification` state, computes diff hashes via
  `GET /repos/{owner}/{repo}/compare/{basehead}` with the
  `application/vnd.github.diff` media type (no checkout required), dismisses
  reviews via `pulls.dismissReview`, and writes the audit summary.
- `src/decide.test.ts` — bun tests for the decision logic.

There is deliberately no git checkout: all diffs come from the compare API, so
the workflow needs no `actions/checkout` step and works on any runner.
