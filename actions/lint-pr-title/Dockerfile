# This file must be built from the root of the repository.
# Example: docker build . -f actions/lint-pr-title/Dockerfile
FROM oven/bun:1.2.18@sha256:2cdd9c93006af1b433c214016d72a3c60d7aa2c75691cb44dfd5250aa379986b AS base

# Create a non-root user with a home directory
# Run tests as non-root user because one test expects EACCES when writing to a
# read-only file, and this will not fail as root.
RUN useradd -ms /bin/bash -d /home/newuser newuser
USER newuser
WORKDIR /home/newuser

FROM base AS install
RUN mkdir -p temp/dev
WORKDIR /home/newuser/temp/dev
COPY --chown=newuser:newuser package.json bun.lock ./
COPY --chown=newuser:newuser actions/dependabot-auto-triage/package.json ./actions/dependabot-auto-triage/
COPY --chown=newuser:newuser actions/get-latest-workflow-artifact/package.json ./actions/get-latest-workflow-artifact/
COPY --chown=newuser:newuser actions/lint-pr-title/package.json ./actions/lint-pr-title/
RUN bun install --frozen-lockfile --filter lint-pr-title

# install with --production (exclude devDependencies)
RUN mkdir -p temp/prod
WORKDIR /home/newuser/temp/prod
COPY --chown=newuser:newuser package.json bun.lock ./
COPY --chown=newuser:newuser actions/dependabot-auto-triage/package.json ./actions/dependabot-auto-triage/
COPY --chown=newuser:newuser actions/get-latest-workflow-artifact/package.json ./actions/get-latest-workflow-artifact/
COPY --chown=newuser:newuser actions/lint-pr-title/package.json ./actions/lint-pr-title/
RUN bun install --frozen-lockfile --production --filter lint-pr-title

FROM base AS prerelease
WORKDIR /home/newuser/app
COPY --from=install /home/newuser/temp/dev/node_modules ./node_modules
COPY --chown=newuser:newuser . .

# Install dev dependencies to run the tests
ENV NODE_ENV=development
RUN bun run --filter lint-pr-title test

FROM base AS release
USER root
WORKDIR /usr/src/app
COPY --from=install /home/newuser/temp/prod/node_modules ./node_modules
COPY --from=prerelease /home/newuser/app/actions/lint-pr-title/src ./actions/lint-pr-title/src
COPY --from=prerelease /home/newuser/app/actions/lint-pr-title/package.json ./actions/lint-pr-title/
COPY --from=prerelease /home/newuser/app/actions/lint-pr-title/commitlint.config.js ./actions/lint-pr-title/

WORKDIR /usr/src/app/actions/lint-pr-title
ENTRYPOINT [ "bun", "run", "src/index.ts" ]
