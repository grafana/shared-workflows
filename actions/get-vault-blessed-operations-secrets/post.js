// Post-job step: revoke the Vault token the main step logged in with, so it
// cannot be reused for the rest of its TTL. Registered as the action's `post:`
// step with `post-if: always()`.
//
// `revoke-self` is part of Vault's built-in `default` policy, so the role needs
// no extra capability. A fresh proxy JWT is minted because the one from the
// main step may have expired.
//
// Best-effort: on any failure the token still expires with its TTL, and we
// never fail the job from here, because that would mask the real job result.

"use strict";

const {
  getState,
  info,
  warning,
  setSecret,
  sleep,
  fetchIdToken,
  fetchWithTimeout,
  retry,
} = require("./lib.js");

const MAX_ATTEMPTS = 3;
const RETRY_BASE_DELAY_MS = 2000;

const revokeOnce = async ({ vaultUrl, vaultToken, proxyJwt }) => {
  try {
    const res = await fetchWithTimeout(
      `${vaultUrl.replace(/\/+$/, "")}/v1/auth/token/revoke-self`,
      {
        method: "POST",
        headers: {
          "X-Vault-Token": vaultToken,
          "Proxy-Authorization-Token": `Bearer ${proxyJwt}`,
        },
      },
    );
    return { status: res.status, body: await res.text() };
  } catch (err) {
    return { status: 0, body: err.message };
  }
};

const main = async () => {
  const vaultUrl = getState("vault_url");
  const vaultToken = getState("vault_token");
  const proxyAudience = getState("proxy_audience");

  if (!vaultUrl || !vaultToken || !proxyAudience) {
    info("No Vault token in state (login likely failed); nothing to revoke.");
    return;
  }
  setSecret(vaultToken);

  let proxyJwt;
  try {
    proxyJwt = await retry({ label: "Mint proxy OIDC token for revoke" }, () =>
      fetchIdToken(proxyAudience),
    );
  } catch (err) {
    warning(
      `Failed to mint proxy JWT for revoke-self: ${err.message}. ` +
        "The Vault token will expire with its TTL.",
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
    if (status === 200 || status === 204) {
      info(`Vault token revoked (HTTP ${status}).`);
      return;
    }
    // 403 / 404: the token is already gone.
    if (status === 403 || status === 404) {
      warning(`Vault token revoke skipped (HTTP ${status}): ${body}`);
      return;
    }
    warning(
      `Vault token revoke attempt ${attempt}/${MAX_ATTEMPTS} failed ` +
        `(HTTP ${status}): ${body}`,
    );
    if (attempt < MAX_ATTEMPTS) {
      await sleep(RETRY_BASE_DELAY_MS * attempt);
    }
  }
  warning(
    `Failed to revoke the Vault token after ${MAX_ATTEMPTS} attempts. ` +
      "It will expire with its TTL.",
  );
};

main().catch((err) => {
  warning(`Vault token revoke errored: ${err.message}`);
});
