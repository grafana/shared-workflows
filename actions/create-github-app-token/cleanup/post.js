// Revoke the Vault lease that backs the GitHub App installation token. This
// runs as a post-job step, so it executes after the user's workflow has
// finished using the token regardless of whether earlier steps succeeded or
// failed (post-if: always()).
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
const leaseId = process.env.STATE_lease_id || "";

if (!vaultUrl || !vaultToken || !leaseId) {
  console.log(
    "No cleanup state present (token creation likely failed); " +
      "skipping Vault lease revocation.",
  );
  process.exit(0);
}

// Re-mask the vault token defensively in case any log line echoes it.
console.log(`::add-mask::${vaultToken}`);

const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms));

const revokeLeaseOnce = () => {
  const endpoint = new URL(
    "/v1/sys/leases/revoke",
    vaultUrl.replace(/\/+$/, ""),
  );
  const payload = JSON.stringify({ lease_id: leaseId });

  return new Promise((resolve) => {
    const req = https.request(
      {
        method: "PUT",
        hostname: endpoint.hostname,
        port: endpoint.port || 443,
        path: endpoint.pathname,
        headers: {
          "X-Vault-Token": vaultToken,
          "Content-Type": "application/json",
          "Content-Length": Buffer.byteLength(payload),
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

    req.write(payload);
    req.end();
  });
};

const revokeLease = async () => {
  for (let attempt = 1; attempt <= MAX_ATTEMPTS; attempt++) {
    const { status, body } = await revokeLeaseOnce();

    if (status >= 200 && status < 300) {
      console.log(`Vault lease revoked (HTTP ${status}, lease_id=${leaseId}).`);
      return;
    }

    // 403 / 404 typically mean the token or lease is already gone; no point
    // retrying.
    if (status === 403 || status === 404) {
      console.log(
        `::warning::Vault lease revoke skipped (HTTP ${status}): ${body}`,
      );
      return;
    }

    console.log(
      `::warning::Vault lease revoke attempt ${attempt}/${MAX_ATTEMPTS} ` +
        `failed (HTTP ${status}): ${body}`,
    );

    if (attempt < MAX_ATTEMPTS) {
      await sleep(RETRY_BASE_DELAY_MS * attempt);
    }
  }

  console.log(
    `::warning::Failed to revoke Vault lease after ${MAX_ATTEMPTS} attempts. ` +
      "The token will still expire naturally when its TTL elapses.",
  );
};

revokeLease().catch((err) => {
  // Never fail the post-step on cleanup errors — the token will expire on its
  // own and surfacing an error here would mask the real job result.
  console.log(`::warning::Vault lease revoke errored: ${err.message}`);
});
