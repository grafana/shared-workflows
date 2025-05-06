FROM oven/bun:1.1.27@sha256:0bce8e8f0eb41e131df5529fcce89dc8b97135f4a583fa7f28d48cbcdb26ff4e AS base
WORKDIR /usr/src/app

FROM base AS install
RUN mkdir -p /temp/dev
COPY package.json bun.lockb /temp/dev/
RUN cd /temp/dev && bun install --frozen-lockfile

# install with --production (exclude devDependencies)
RUN mkdir -p /temp/prod
COPY package.json bun.lockb /temp/prod/
RUN cd /temp/prod && bun install --frozen-lockfile --production

FROM base AS prerelease
COPY --from=install /temp/dev/node_modules node_modules
COPY . .

# Install dev dependencies to run the tests
ENV NODE_ENV=development
RUN bun install --frozen-lockfile --dev
RUN bun run test
RUN bun run build

FROM base AS release
COPY --from=install /temp/prod/node_modules node_modules
COPY --from=prerelease /usr/src/app/index.ts .
COPY --from=prerelease /usr/src/app/package.json .

ENTRYPOINT [ "bun", "run", "index.ts" ]
