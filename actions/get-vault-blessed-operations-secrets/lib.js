// Shared utilities for the get-vault-blessed-operations-secrets Node action.
// Uses only Node.js built-ins so the action can run from a plain git checkout
// without a bundled `node_modules` or a build step.

"use strict";

const fs = require("node:fs");
const { createHash, randomBytes } = require("node:crypto");

// Default timeout (ms) for outbound HTTP requests, so a half-open socket fails
// and gets retried instead of hanging until the job times out.
const REQUEST_TIMEOUT_MS = 30000;

const VALID_VAULT_INSTANCES = new Set(["dev", "ops"]);

// Operation names are CI gates catalog file names (`publish_winget.yaml`):
// lowercase letters and digits, separated by single `_` or `-`.
const OPERATION_PATTERN = /^[a-z0-9]+(?:[_-][a-z0-9]+)*$/;
const ENV_NAME_PATTERN = /^[A-Za-z_][A-Za-z0-9_]*$/;
const PATH_SEGMENT_PATTERN = /^[A-Za-z0-9_-][A-Za-z0-9._-]*$/;
const KEY_PATTERN = /^[A-Za-z0-9._-]+$/;

const fetchWithTimeout = (input, init = {}, timeoutMs = REQUEST_TIMEOUT_MS) =>
  fetch(input, { ...init, signal: AbortSignal.timeout(timeoutMs) });

// `::add-mask::` registers `value` as a secret so later log lines are
// redacted. Multi-line values are masked line by line, because the runner
// matches each log line on its own.
const setSecret = (value) => {
  if (!value) {
    return;
  }
  console.log(`::add-mask::${value}`);
  for (const line of String(value).split(/\r?\n/)) {
    if (line && line !== value) {
      console.log(`::add-mask::${line}`);
    }
  }
};

const writeKvFile = (file, key, value, prefix) => {
  // Heredoc format with a random delimiter, so values containing `=` or
  // newlines are written safely.
  const delim = `${prefix}_${Date.now()}_${randomBytes(8).toString("hex")}`;
  fs.appendFileSync(file, `${key}<<${delim}\n${value}\n${delim}\n`);
};

const setOutput = (name, value) => {
  const file = process.env.GITHUB_OUTPUT;
  if (!file) {
    throw new Error("GITHUB_OUTPUT is not set; cannot set action output.");
  }
  writeKvFile(file, name, value, "ghaoutput");
};

const saveState = (name, value) => {
  const file = process.env.GITHUB_STATE;
  if (!file) {
    throw new Error("GITHUB_STATE is not set; cannot save state.");
  }
  writeKvFile(file, name, value, "ghastate");
};

const getState = (name) => process.env[`STATE_${name}`] || "";

const info = (message) => console.log(message);
const warning = (message) => console.log(`::warning::${message}`);
const error = (message) => console.log(`::error::${message}`);

const setFailed = (message) => {
  error(message);
  process.exitCode = 1;
};

const sha256Hex = (value) =>
  createHash("sha256").update(value, "utf8").digest("hex");

const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms));

// Retry `fn` up to `attempts` times with a linearly growing delay. An error
// with `retryable === false` is thrown straight away.
const retry = async ({ attempts = 3, baseDelayMs = 5000, label }, fn) => {
  let lastError;
  for (let attempt = 1; attempt <= attempts; attempt++) {
    info(`${label}: attempt ${attempt}/${attempts}`);
    try {
      return await fn(attempt);
    } catch (err) {
      if (err && err.retryable === false) {
        throw err;
      }
      lastError = err;
      warning(`${label} attempt ${attempt} failed: ${err.message}`);
      if (attempt < attempts) {
        await sleep(baseDelayMs * attempt);
      }
    }
  }
  throw lastError;
};

// Mint a GitHub OIDC ID token for `audience`. The job needs
// `permissions: id-token: write`.
const fetchIdToken = async (audience) => {
  const requestUrl = process.env.ACTIONS_ID_TOKEN_REQUEST_URL;
  const requestToken = process.env.ACTIONS_ID_TOKEN_REQUEST_TOKEN;
  if (!requestUrl || !requestToken) {
    throw new Error(
      "ACTIONS_ID_TOKEN_REQUEST_URL/TOKEN not set. Make sure the job " +
        "grants `permissions: id-token: write`.",
    );
  }
  const url = new URL(requestUrl);
  url.searchParams.set("audience", audience);

  const res = await fetchWithTimeout(url, {
    method: "GET",
    headers: {
      Authorization: `Bearer ${requestToken}`,
      Accept: "application/json",
    },
  });
  const body = await res.text();
  if (!res.ok) {
    throw new Error(`OIDC mint failed (HTTP ${res.status}): ${body}`);
  }
  let parsed;
  try {
    parsed = JSON.parse(body);
  } catch (err) {
    throw new Error(`Failed to parse OIDC response: ${err.message}`, {
      cause: err,
    });
  }
  if (!parsed.value) {
    throw new Error("OIDC response did not contain a `value` field.");
  }
  return parsed.value;
};

const validateOperation = (operation) => {
  if (!OPERATION_PATTERN.test(operation)) {
    throw new Error(
      `Invalid value for operation input: '${operation}'. Must be the CI ` +
        "gates operation name: lowercase letters and digits separated by " +
        "single '_' or '-' (e.g. publish_winget).",
    );
  }
  return operation;
};

const validateVaultInstance = (instance) => {
  if (!VALID_VAULT_INSTANCES.has(instance)) {
    throw new Error(
      `Invalid value for vault_instance input: '${instance}'. Must be 'dev' or 'ops'.`,
    );
  }
  return instance;
};

// Parse the `secrets` input into Vault reads. Each non-empty line is
// `ENV_NAME=subpath:key`; the subpath is always relative to
// `ci/data/operations/<operation>/`, so a caller cannot read anything outside
// its operation.
//
// Errors name the line number, never the line's text: the input comes from
// `INPUT_SECRETS`, and nothing derived from it is written to the log.
const parseSecrets = (operation, raw) => {
  const lines = String(raw || "")
    .split(/\r?\n/)
    .map((text, index) => ({ text: text.trim(), line: index + 1 }))
    .filter(({ text }) => text);
  if (lines.length === 0) {
    throw new Error("secrets input is required and must not be empty.");
  }

  const seen = new Set();
  return lines.map(({ text, line }) => {
    const fail = (reason) => {
      throw new Error(`Invalid secrets line ${line}: ${reason}`);
    };

    const match = /^([^=]+)=([^:]+):(.+)$/.exec(text);
    if (!match) {
      fail("expected 'ENV_NAME=subpath:key'.");
    }
    const [, envName, subpath, key] = match;

    if (!ENV_NAME_PATTERN.test(envName)) {
      fail(
        "the name must be letters, digits and '_', not starting with a digit.",
      );
    }
    if (seen.has(envName)) {
      fail("the name is already used on an earlier line.");
    }
    seen.add(envName);

    const segments = subpath.split("/");
    if (segments.some((s) => !PATH_SEGMENT_PATTERN.test(s) || s === "..")) {
      fail(
        "the subpath must be relative to the operation, e.g. 'winget' or " +
          "'registry/token'.",
      );
    }
    if (!KEY_PATTERN.test(key)) {
      fail("the key must be letters, digits and '.', '_' or '-'.");
    }

    return {
      envName,
      path: `ci/data/operations/${operation}/${subpath}`,
      key,
      line,
    };
  });
};

// The Vault JWT role created for one (caller repository, operation) pair. This
// MUST match the role name built by the CI gates Vault binding in
// grafana/deployment_tools (terraform/modules/ci_gates/bindings/vault):
//   blessed-<operation>-<org>-<repo>-<first 8 hex of sha256("<org>/<repo>:<operation>")>
const roleName = ({ repository, operation }) => {
  const [org, repo] = String(repository || "").split("/");
  if (!org || !repo) {
    throw new Error("GITHUB_REPOSITORY is not set or is malformed.");
  }
  const hash = sha256Hex(`${org}/${repo}:${operation}`).slice(0, 8);
  return `blessed-${operation}-${org}-${repo}-${hash}`;
};

module.exports = {
  setSecret,
  setOutput,
  saveState,
  getState,
  info,
  warning,
  error,
  setFailed,
  sleep,
  retry,
  fetchIdToken,
  fetchWithTimeout,
  validateOperation,
  validateVaultInstance,
  parseSecrets,
  roleName,
};
