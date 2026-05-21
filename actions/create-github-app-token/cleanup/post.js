// Revoke the Vault token that issued the GitHub App installation token. Per
// the Vault docs, revoking a token also revokes every dynamic secret created
// with it, so the GitHub App token is invalidated as a side-effect:
// https://developer.hashicorp.com/vault/api-docs/auth/token#revoke-a-token-self
//
// This runs as a post-job step, so it executes after the user's workflow has
// finished using the token regardless of whether earlier steps succeeded or
// failed (post-if: always()).
//
// `revoke-self` is part of Vault's built-in `default` policy, so it requires
// no extra capability on the role.
//
// Uses only Node.js built-ins so it can run as a self-contained action without
// a bundled `node_modules`.

"use strict";

const https = require("node:https");
const { URL } = require("node:url");

const MAX_ATTEMPTS = 3;
const RETRY_BASE_DELAY_MS = 2000;

const vaultUrl = process.env.STATE_vault_url || "";
const vaultToken = process.env.STATE_vault_token || "";

if (!vaultUrl || !vaultToken) {
  console.log(
    "No cleanup state present (token creation likely failed); " +
      "skipping Vault token revocation.",
  );
  process.exit(0);
}

// Re-mask the vault token defensively in case any log line echoes it.
console.log(`::add-mask::${vaultToken}`);

const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms));

const revokeTokenOnce = () => {
  const endpoint = new URL(
    "/v1/auth/token/revoke-self",
    vaultUrl.replace(/\/+$/, ""),
  );

  return new Promise((resolve) => {
    const req = https.request(
      {
        method: "POST",
        hostname: endpoint.hostname,
        port: endpoint.port || 443,
        path: endpoint.pathname,
        headers: {
          "X-Vault-Token": vaultToken,
          "Content-Length": 0,
        },
      },
      (res) => {
        let body = "";
        res.on("data", (chunk) => {
          body += chunk;
        });
        res.on("end", () => {
          resolve({ status: res.statusCode || 0, body });
        });
      },
    );

    req.on("error", (err) => {
      resolve({ status: 0, body: err.message });
    });

    req.end();
  });
};

const revokeToken = async () => {
  for (let attempt = 1; attempt <= MAX_ATTEMPTS; attempt++) {
    const { status, body } = await revokeTokenOnce();

    if (status >= 200 && status < 300) {
      console.log(`Vault token revoked (HTTP ${status}).`);
      return;
    }

    // 403 / 404 typically mean the token is already gone; no point retrying.
    if (status === 403 || status === 404) {
      console.log(
        `::warning::Vault token revoke skipped (HTTP ${status}): ${body}`,
      );
      return;
    }

    console.log(
      `::warning::Vault token revoke attempt ${attempt}/${MAX_ATTEMPTS} ` +
        `failed (HTTP ${status}): ${body}`,
    );

    if (attempt < MAX_ATTEMPTS) {
      await sleep(RETRY_BASE_DELAY_MS * attempt);
    }
  }

  console.log(
    `::warning::Failed to revoke Vault token after ${MAX_ATTEMPTS} attempts. ` +
      "The GitHub App token will still expire naturally when its TTL elapses.",
  );
};

revokeToken().catch((err) => {
  // Never fail the post-step on cleanup errors — the token will expire on its
  // own and surfacing an error here would mask the real job result.
  console.log(`::warning::Vault token revoke errored: ${err.message}`);
});
