// Persist the Vault credentials to job state so the post step can revoke the
// Vault token after the user's job has finished. Revoking the token
// cascade-revokes every lease it created, including the GitHub App token.
//
// Uses only Node.js built-ins so it can run as a self-contained action without
// a bundled `node_modules`.

"use strict";

const fs = require("node:fs");

const stateFile = process.env.GITHUB_STATE;
const vaultUrl = process.env.INPUT_VAULT_URL || "";
const vaultToken = process.env.INPUT_VAULT_TOKEN || "";
const proxyAudience = process.env.INPUT_PROXY_AUDIENCE || "";

if (!stateFile) {
  console.log(
    "GITHUB_STATE is not set; skipping registration of cleanup state.",
  );
  process.exit(0);
}

if (!vaultUrl || !vaultToken || !proxyAudience) {
  console.log(
    "Missing required input(s) (vault_url / vault_token / proxy_audience); " +
      "post-step will be a no-op.",
  );
  process.exit(0);
}

// Re-mask the vault token in case any later log statement echoes the state
// values. `auth_vault.sh` already masks it for the duration of the job, but
// adding the mask again is harmless and defensive.
console.log(`::add-mask::${vaultToken}`);

// GITHUB_STATE supports the same heredoc format as GITHUB_ENV / GITHUB_OUTPUT.
// The values stored here do not contain newlines, but the heredoc form is
// robust against unexpected characters such as `=`.
const delim = `ghacleanup_${Date.now()}_${Math.random().toString(36).slice(2)}`;
const writeState = (key, value) => {
  fs.appendFileSync(stateFile, `${key}<<${delim}\n${value}\n${delim}\n`);
};

writeState("vault_url", vaultUrl);
writeState("vault_token", vaultToken);
writeState("proxy_audience", proxyAudience);

console.log("Registered Vault token for post-job revocation.");
