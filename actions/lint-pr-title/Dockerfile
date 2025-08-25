# This file must be built from the root of the repository.
# Example: docker build . -f actions/lint-pr-title/Dockerfile

FROM oven/bun:1.2.21@sha256:5a2011bf09364b9af658ac1e66f60d08092f4291aeefbff448d58b027734fdd0 AS base

WORKDIR /usr/src/app

# Create a non-root user to run the tests. Run tests as non-root user because
# one test expects EACCES when writing to a read-only file, and this will not
# fail as root.
RUN useradd -ms /bin/bash newuser

# Install dependencies
FROM base AS deps

COPY package.json bun.lock ./
COPY actions/dependabot-auto-triage/package.json ./actions/dependabot-auto-triage/

# Because we use bun's workspaces, we need to have all package.json files
# available, even if they're not used. (We filter to one workspace.)
COPY actions/get-latest-workflow-artifact/package.json ./actions/get-latest-workflow-artifact/
COPY actions/lint-pr-title/package.json ./actions/lint-pr-title/

RUN --mount=type=cache,target=/root/.bun/install/cache \
    bun install --frozen-lockfile --filter lint-pr-title

# Run tests
FROM base AS test

COPY --from=deps /usr/src/app/node_modules ./node_modules
COPY --chown=newuser:newuser actions/lint-pr-title/ ./actions/lint-pr-title/
USER newuser

ENV NODE_ENV=development

RUN bun run --filter lint-pr-title test

# Assemble final image from a clean stage
FROM base

USER root

WORKDIR /usr/src/app

COPY --from=deps /usr/src/app/node_modules ./node_modules
COPY actions/lint-pr-title/src ./actions/lint-pr-title/src
COPY actions/lint-pr-title/package.json ./actions/lint-pr-title/
COPY actions/lint-pr-title/commitlint.config.js ./actions/lint-pr-title/

WORKDIR /usr/src/app/actions/lint-pr-title

ENTRYPOINT [ "bun", "run", "src/index.ts" ]
