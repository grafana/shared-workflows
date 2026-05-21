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
// The Vault instance is fronted by a proxy that requires a GitHub OIDC JWT
// in the `Proxy-Authorization-Token` header on every request. We mint a fresh
// JWT here rather than reusing the one minted at job start, because GitHub
// OIDC tokens are short-lived (~5 minutes) and may have expired by the time
// the post step runs.
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
const proxyAudience = process.env.STATE_proxy_audience || "";
const oidcRequestUrl = process.env.ACTIONS_ID_TOKEN_REQUEST_URL || "";
const oidcRequestToken = process.env.ACTIONS_ID_TOKEN_REQUEST_TOKEN || "";

if (!vaultUrl || !vaultToken || !proxyAudience) {
  console.log(
    "No cleanup state present (token creation likely failed); " +
      "skipping Vault token revocation.",
  );
  process.exit(0);
}

if (!oidcRequestUrl || !oidcRequestToken) {
  console.log(
    "::warning::ACTIONS_ID_TOKEN_REQUEST_URL/TOKEN not set; cannot mint " +
      "proxy JWT for Vault revoke-self. The Vault token will expire " +
      "naturally when its TTL elapses.",
  );
  process.exit(0);
}

// Re-mask the vault token defensively in case any log line echoes it.
console.log(`::add-mask::${vaultToken}`);

const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms));

const fetchProxyJwt = () => {
  const url = new URL(oidcRequestUrl);
  url.searchParams.set("audience", proxyAudience);

  return new Promise((resolve, reject) => {
    const req = https.request(
      {
        method: "GET",
        hostname: url.hostname,
        port: url.port || 443,
        path: `${url.pathname}${url.search}`,
        headers: {
          Authorization: `Bearer ${oidcRequestToken}`,
          Accept: "application/json",
        },
      },
      (res) => {
        let body = "";
        res.on("data", (chunk) => {
          body += chunk;
        });
        res.on("end", () => {
          if (res.statusCode && res.statusCode >= 200 && res.statusCode < 300) {
            try {
              const parsed = JSON.parse(body);
              if (parsed.value) {
                resolve(parsed.value);
              } else {
                reject(new Error("OIDC response did not contain a token"));
              }
            } catch (err) {
              reject(
                new Error(`Failed to parse OIDC response: ${err.message}`),
              );
            }
          } else {
            reject(
              new Error(
                `OIDC mint failed (HTTP ${res.statusCode || 0}): ${body}`,
              ),
            );
          }
        });
      },
    );

    req.on("error", (err) => {
      reject(err);
    });

    req.end();
  });
};

const revokeTokenOnce = (proxyJwt) => {
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
          "Proxy-Authorization-Token": `Bearer ${proxyJwt}`,
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
  let proxyJwt;
  try {
    proxyJwt = await fetchProxyJwt();
  } catch (err) {
    console.log(
      `::warning::Failed to mint proxy JWT for Vault revoke-self: ${err.message}. ` +
        "The Vault token will expire naturally when its TTL elapses.",
    );
    return;
  }
  // Mask the JWT so it never appears verbatim in subsequent log lines.
  console.log(`::add-mask::${proxyJwt}`);

  for (let attempt = 1; attempt <= MAX_ATTEMPTS; attempt++) {
    const { status, body } = await revokeTokenOnce(proxyJwt);

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
