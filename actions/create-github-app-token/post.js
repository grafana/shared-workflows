// Post-job step: revoke the Vault token that issued the GitHub App
// installation token. Per the Vault docs, revoking a token also revokes every
// dynamic secret created with it, so the GitHub App token is invalidated as a
// side-effect:
// https://developer.hashicorp.com/vault/api-docs/auth/token#revoke-a-token-self
//
// This is registered as the action's `post:` step with `post-if: always()`, so
// it runs after the user's job regardless of whether earlier steps succeeded
// or failed.
//
// `revoke-self` is part of Vault's built-in `default` policy, so it requires
// no extra capability on the role and the action doesn't need the lease_id.
//
// The Vault instance is fronted by a proxy that requires a GitHub OIDC JWT in
// the `Proxy-Authorization-Token` header on every request. We mint a fresh
// JWT here rather than reusing the one minted in the main step, because
// GitHub OIDC tokens are short-lived and may have expired by the time the
// post step runs.
//
// This step is best-effort: if anything fails, the Vault token will still
// expire naturally when its TTL elapses. We never fail the job from here,
// because surfacing an error here would mask the real job result.

"use strict";

const {
  getState,
  info,
  warning,
  setSecret,
  sleep,
  fetchIdToken,
  retry,
} = require("./lib.js");

const MAX_ATTEMPTS = 3;
const RETRY_BASE_DELAY_MS = 2000;

const revokeOnce = async ({ vaultUrl, vaultToken, proxyJwt }) => {
  const endpoint = `${vaultUrl.replace(/\/+$/, "")}/v1/auth/token/revoke-self`;
  try {
    const res = await fetch(endpoint, {
      method: "POST",
      headers: {
        "X-Vault-Token": vaultToken,
        "Proxy-Authorization-Token": `Bearer ${proxyJwt}`,
        "Content-Length": "0",
      },
    });
    const body = await res.text();
    return { status: res.status, body };
  } catch (err) {
    return { status: 0, body: err.message };
  }
};

const main = async () => {
  const vaultUrl = getState("vault_url");
  const vaultToken = getState("vault_token");
  const proxyAudience = getState("proxy_audience");

  if (!vaultUrl || !vaultToken || !proxyAudience) {
    info(
      "No cleanup state present (token creation likely failed); " +
        "skipping Vault token revocation.",
    );
    return;
  }

  // Re-mask the vault token defensively in case any subsequent log line
  // echoes it. The main step already masked it, but `::add-mask::` is
  // per-job-step state and re-asserting it here is cheap.
  setSecret(vaultToken);

  let proxyJwt;
  try {
    proxyJwt = await retry({ label: "Mint proxy OIDC token for revoke" }, () =>
      fetchIdToken(proxyAudience),
    );
  } catch (err) {
    warning(
      `Failed to mint proxy JWT for Vault revoke-self: ${err.message}. ` +
        "The Vault token will expire naturally when its TTL elapses.",
    );
    return;
  }
  setSecret(proxyJwt);

  for (let attempt = 1; attempt <= MAX_ATTEMPTS; attempt++) {
    const { status, body } = await revokeOnce({
      vaultUrl,
      vaultToken,
      proxyJwt,
    });

    if (status >= 200 && status < 300) {
      info(`Vault token revoked (HTTP ${status}).`);
      return;
    }

    // 403 / 404 typically mean the token is already gone — no point retrying.
    if (status === 403 || status === 404) {
      warning(`Vault token revoke skipped (HTTP ${status}): ${body}`);
      return;
    }

    warning(
      `Vault token revoke attempt ${attempt}/${MAX_ATTEMPTS} ` +
        `failed (HTTP ${status}): ${body}`,
    );

    if (attempt < MAX_ATTEMPTS) {
      await sleep(RETRY_BASE_DELAY_MS * attempt);
    }
  }

  warning(
    `Failed to revoke Vault token after ${MAX_ATTEMPTS} attempts. ` +
      "The GitHub App token will still expire naturally when its TTL elapses.",
  );
};

main().catch((err) => {
  // Never fail the post-step on cleanup errors — the token will expire on its
  // own and surfacing an error here would mask the real job result.
  warning(`Vault token revoke errored: ${err.message}`);
});
