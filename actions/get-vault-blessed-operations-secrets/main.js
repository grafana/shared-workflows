"use strict";

const {
  setSecret,
  setOutput,
  saveState,
  info,
  setFailed,
  retry,
  fetchIdToken,
  fetchWithTimeout,
  validateOperation,
  validateVaultInstance,
  parseSecrets,
  roleName,
} = require("./lib.js");

// An error that retrying cannot fix (bad role, claim mismatch, missing secret).
const permanent = (message) => {
  const err = new Error(message);
  err.retryable = false;
  return err;
};

const parseInputs = () => {
  const operation = validateOperation(
    (process.env.INPUT_OPERATION || "").trim(),
  );
  const vaultInstance = validateVaultInstance(
    (process.env.INPUT_VAULT_INSTANCE || "ops").trim(),
  );
  const secrets = parseSecrets(operation, process.env.INPUT_SECRETS);
  const role = roleName({
    repository: process.env.GITHUB_REPOSITORY,
    operation,
  });
  return { operation, vaultInstance, secrets, role };
};

const authenticateWithVault = async ({
  vaultUrl,
  proxyJwt,
  vaultJwt,
  role,
  operation,
}) => {
  const res = await fetchWithTimeout(
    `${vaultUrl}/v1/auth/github-actions-oidc/login`,
    {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Proxy-Authorization-Token": `Bearer ${proxyJwt}`,
      },
      body: JSON.stringify({ role, jwt: vaultJwt }),
    },
  );
  const body = await res.text();
  if (res.status === 400 || res.status === 403) {
    // Vault rejects the login when the role does not exist or the token's
    // claims do not match it. Both are configuration problems, not transient.
    throw permanent(
      `Vault auth failed (HTTP ${res.status}): ${body}\n` +
        `Check that this repository has adopted the '${operation}' operation ` +
        "in CI gates, and that this step runs inside the blessed workflow's " +
        `credential job (environment: ${operation}) on an allowed, protected ref.`,
    );
  }
  if (!res.ok) {
    throw new Error(`Vault auth failed (HTTP ${res.status}): ${body}`);
  }
  let parsed;
  try {
    parsed = JSON.parse(body);
  } catch (err) {
    throw new Error(`Failed to parse Vault auth response: ${err.message}`, {
      cause: err,
    });
  }
  const token = parsed && parsed.auth && parsed.auth.client_token;
  if (!token) {
    throw new Error("Vault auth response did not contain `auth.client_token`.");
  }
  return token;
};

// Read one KV v2 secret and return its key/value map.
const readSecret = async ({ vaultUrl, vaultToken, proxyJwt, path }) => {
  const res = await fetchWithTimeout(`${vaultUrl}/v1/${path}`, {
    method: "GET",
    headers: {
      "X-Vault-Token": vaultToken,
      "Proxy-Authorization-Token": `Bearer ${proxyJwt}`,
    },
  });
  const body = await res.text();
  if (res.status === 403 || res.status === 404) {
    throw permanent(
      `Vault read of '${path}' failed (HTTP ${res.status}). The secret does ` +
        "not exist, or it is outside the operation's read_paths in the CI " +
        "gates catalog.",
    );
  }
  if (!res.ok) {
    throw new Error(`Vault read of '${path}' failed (HTTP ${res.status}).`);
  }
  let parsed;
  try {
    parsed = JSON.parse(body);
  } catch (err) {
    throw new Error(`Failed to parse Vault response for '${path}'.`, {
      cause: err,
    });
  }
  const data = parsed && parsed.data && parsed.data.data;
  if (!data || typeof data !== "object") {
    throw new Error(`Vault response for '${path}' did not contain data.`);
  }
  return data;
};

const main = async () => {
  const { operation, vaultInstance, secrets, role } = parseInputs();

  const vaultUrl = `https://vault-github-actions.grafana-${vaultInstance}.net`;
  const proxyAudience = `vault-github-actions-grafana-${vaultInstance}`;
  // Vault's github-actions-oidc method checks the audience against its URL.
  const vaultAudience = vaultUrl;
  info(`Operation: ${operation}`);
  info(`Vault role: ${role}`);

  // 1) Mint OIDC tokens. The proxy JWT goes on every request; the Vault JWT is
  //    only used to log in.
  const proxyJwt = await retry({ label: "Mint proxy OIDC token" }, () =>
    fetchIdToken(proxyAudience),
  );
  setSecret(proxyJwt);

  const vaultJwt = await retry({ label: "Mint Vault OIDC token" }, () =>
    fetchIdToken(vaultAudience),
  );
  setSecret(vaultJwt);

  // 2) Log in. This is CI gates' second gate (the first is the job's
  //    `environment:`, which GitHub checks before the job starts). Vault's
  //    bound claims on the role only match an adopted repository, from the
  //    pinned blessed workflow, on an allowed protected ref, through the
  //    operation's environment. No match, no token.
  const vaultToken = await retry({ label: "Vault auth" }, () =>
    authenticateWithVault({ vaultUrl, proxyJwt, vaultJwt, role, operation }),
  );
  setSecret(vaultToken);
  info("Vault auth done.");

  // Save state for the post step before anything else can fail, so the Vault
  // token is revoked whatever happens next.
  saveState("vault_url", vaultUrl);
  saveState("vault_token", vaultToken);
  saveState("proxy_audience", proxyAudience);

  // 3) Read each distinct path once.
  const byPath = new Map();
  for (const path of new Set(secrets.map((s) => s.path))) {
    const data = await retry({ label: `Read ${path}` }, () =>
      readSecret({ vaultUrl, vaultToken, proxyJwt, path }),
    );
    byPath.set(path, data);
  }

  const result = {};
  for (const { envName, path, key } of secrets) {
    const value = byPath.get(path)[key];
    if (value === undefined || value === null) {
      throw new Error(`Key '${key}' not found in '${path}'.`);
    }
    setSecret(String(value));
    result[envName] = String(value);
  }

  setOutput("secrets", JSON.stringify(result));
  info(`Read ${secrets.length} secret(s) for operation '${operation}'.`);
};

main().catch((err) => {
  setFailed(err && err.message ? err.message : String(err));
});
