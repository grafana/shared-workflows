// Shared utilities for the create-github-app-token Node action. Uses only
// Node.js built-ins so the action can run from a plain git checkout without a
// bundled `node_modules` or a build step.

"use strict";

const fs = require("node:fs");
const crypto = require("node:crypto");

// `::add-mask::` registers `value` as a secret so subsequent log lines that
// echo it are redacted. Called defensively before logging anything derived
// from a secret.
const setSecret = (value) => {
  if (value) {
    console.log(`::add-mask::${value}`);
  }
};

const writeKvFile = (file, key, value, prefix) => {
  // Use the heredoc format supported by GITHUB_OUTPUT / GITHUB_STATE so values
  // containing `=` or newlines are handled correctly. The delimiter is
  // randomized to avoid collision with the value contents.
  const delim = `${prefix}_${Date.now()}_${crypto.randomBytes(8).toString("hex")}`;
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
  crypto.createHash("sha256").update(value, "utf8").digest("hex");

const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms));

// Retry `fn` up to `attempts` times. The delay between attempts grows linearly
// (`baseDelayMs * attempt`), matching the bash scripts the action used to ship.
const retry = async ({ attempts = 3, baseDelayMs = 5000, label }, fn) => {
  let lastError;
  for (let attempt = 1; attempt <= attempts; attempt++) {
    info(`${label}: attempt ${attempt}/${attempts}`);
    try {
      return await fn(attempt);
    } catch (err) {
      lastError = err;
      warning(`${label} attempt ${attempt} failed: ${err.message}`);
      if (attempt < attempts) {
        await sleep(baseDelayMs * attempt);
      }
    }
  }
  throw lastError;
};

// Mint a GitHub OIDC ID token for `audience`. Equivalent to
// `core.getIDToken(audience)` but without the @actions/core dependency. The
// caller MUST have `permissions: id-token: write` in the workflow.
const fetchIdToken = async (audience) => {
  const requestUrl = process.env.ACTIONS_ID_TOKEN_REQUEST_URL;
  const requestToken = process.env.ACTIONS_ID_TOKEN_REQUEST_TOKEN;
  if (!requestUrl || !requestToken) {
    throw new Error(
      "ACTIONS_ID_TOKEN_REQUEST_URL/TOKEN not set. Make sure the workflow " +
        "grants `permissions: id-token: write`.",
    );
  }
  const url = new URL(requestUrl);
  url.searchParams.set("audience", audience);

  const res = await fetch(url, {
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

module.exports = {
  setSecret,
  setOutput,
  saveState,
  getState,
  info,
  warning,
  error,
  setFailed,
  sha256Hex,
  sleep,
  retry,
  fetchIdToken,
};
