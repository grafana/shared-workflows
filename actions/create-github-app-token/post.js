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
// After revoking the Vault token we verify that the GitHub App installation
// token is actually dead by calling an endpoint that requires it
// (`GET /installation/repositories`) and asserting it returns HTTP 401. Vault
// revokes the token's lease as a side-effect of revoking the Vault token, but
// that can take a moment to propagate, so we retry the check a few times.
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

// Confirming the GitHub App token was revoked can race with Vault propagating
// the lease revocation to GitHub, so retry the check a few times before giving
// up.
const CONFIRM_MAX_ATTEMPTS = 3;
const CONFIRM_BASE_DELAY_MS = 2000;

// GitHub Actions always sets GITHUB_API_URL; fall back to the public API host
// for safety (e.g. when running this file outside of Actions).
const GITHUB_API_URL = process.env.GITHUB_API_URL || "https://api.github.com";

const revokeOnce = async ({ vaultUrl, vaultToken, proxyJwt }) => {
  const endpoint = `${vaultUrl.replace(/\/+$/, "")}/v1/auth/token/revoke-self`;
  try {
    const res = await fetch(endpoint, {
      method: "POST",
      headers: {
        "X-Vault-Token": vaultToken,
        "Proxy-Authorization-Token": `Bearer ${proxyJwt}`,
      },
    });
    const body = await res.text();
    return { status: res.status, body };
  } catch (err) {
    return { status: 0, body: err.message };
  }
};

const revokeVaultToken = async ({ vaultUrl, vaultToken, proxyJwt }) => {
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

// Probe an endpoint that requires the installation token. A revoked token
// returns HTTP 401 ("Bad credentials"); a still-valid token returns 2xx.
const probeGithubToken = async (githubToken) => {
  const endpoint = `${GITHUB_API_URL.replace(/\/+$/, "")}/rate_limit`;
  try {
    const res = await fetch(endpoint, {
      method: "GET",
      headers: {
        Authorization: `Bearer ${githubToken}`,
        Accept: "application/vnd.github+json",
        "X-GitHub-Api-Version": "2022-11-28",
      },
    });
    // Drain the body so the connection can be released cleanly.
    await res.text();
    return { status: res.status };
  } catch (err) {
    return { status: 0, error: err.message };
  }
};

// Confirm the GitHub App token was actually invalidated by the Vault
// revocation. Best-effort: we only warn if we cannot confirm, because the
// token expires on its own when its Vault lease TTL elapses.
const confirmGithubTokenRevoked = async (githubToken) => {
  for (let attempt = 1; attempt <= CONFIRM_MAX_ATTEMPTS; attempt++) {
    const { status, error } = await probeGithubToken(githubToken);

    if (status === 401) {
      info("Confirmed: the GitHub App token is revoked (HTTP 401).");
      return;
    }

    if (status >= 200 && status < 300) {
      warning(
        `GitHub App token still valid (HTTP ${status}) on check ` +
        `${attempt}/${CONFIRM_MAX_ATTEMPTS}; revocation may not have ` +
        "propagated yet.",
      );
    } else {
      warning(
        `GitHub App token revocation check inconclusive on check ` +
        `${attempt}/${CONFIRM_MAX_ATTEMPTS} ` +
        `(HTTP ${status}${error ? `: ${error}` : ""}).`,
      );
    }

    if (attempt < CONFIRM_MAX_ATTEMPTS) {
      await sleep(CONFIRM_BASE_DELAY_MS * attempt);
    }
  }

  warning(
    "Could not confirm the GitHub App token was revoked. It will still " +
    "expire naturally when its Vault lease TTL elapses.",
  );
};

const main = async () => {
  const vaultUrl = getState("vault_url");
  const vaultToken = getState("vault_token");
  const proxyAudience = getState("proxy_audience");
  const githubToken = getState("github_token");

  if (!vaultUrl || !vaultToken || !proxyAudience) {
    info(
      "No cleanup state present (token creation likely failed); " +
      "skipping Vault token revocation.",
    );
    return;
  }

  // Re-mask the secrets defensively in case any subsequent log line echoes
  // them. The main step already masked them, but `::add-mask::` is
  // per-job-step state and re-asserting it here is cheap.
  setSecret(vaultToken);
  if (githubToken) {
    setSecret(githubToken);
  }

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

  await revokeVaultToken({ vaultUrl, vaultToken, proxyJwt });

  // Verify the revocation actually invalidated the GitHub App token.
  if (githubToken) {
    await confirmGithubTokenRevoked(githubToken);
  } else {
    info(
      "No GitHub App token in state; skipping revocation verification " +
      "(token was likely never created).",
    );
  }
};

main().catch((err) => {
  // Never fail the post-step on cleanup errors — the token will expire on its
  // own and surfacing an error here would mask the real job result.
  warning(`Vault token revoke errored: ${err.message}`);
});
