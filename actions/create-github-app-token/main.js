"use strict";

const {
  setSecret,
  setOutput,
  saveState,
  info,
  setFailed,
  sha256Hex,
  retry,
  fetchIdToken,
} = require("./lib.js");

const VALID_VAULT_INSTANCES = new Set(["dev", "ops"]);

const parseInputs = () => {
  const permissionSet = (process.env.INPUT_PERMISSION_SET || "default").trim();
  const vaultInstance = (process.env.INPUT_VAULT_INSTANCE || "ops").trim();
  const githubAppInput = process.env.INPUT_GITHUB_APP || "";

  if (!VALID_VAULT_INSTANCES.has(vaultInstance)) {
    throw new Error(
      `Invalid value for vault_instance input: '${vaultInstance}'. Must be 'dev' or 'ops'.`,
    );
  }

  const apps = githubAppInput
    .split(",")
    .map((app) => app.trim())
    .filter(Boolean);
  if (apps.length === 0) {
    throw new Error("github_app input is required and must not be empty.");
  }

  const repository = process.env.GITHUB_REPOSITORY || "";
  const repositoryName = repository.split("/")[1] || "";
  if (!repositoryName) {
    throw new Error("GITHUB_REPOSITORY is not set or is malformed.");
  }

  return { permissionSet, vaultInstance, apps, repositoryName };
};

// `github.workflow_ref` looks like
// `owner/repo/.github/workflows/file.yml@refs/heads/main`. Strip the
// `owner/repo/` prefix and the `@ref` suffix, then sha256-hex the result.
// Matches the legacy bash implementation byte-for-byte.
const normalizeWorkflowRefSha = () => {
  const workflowRef = process.env.GITHUB_WORKFLOW_REF || "";
  const normalized = workflowRef
    .replace(/^[^/]+\/[^/]+\//, "")
    .replace(/@.*$/, "");
  return sha256Hex(normalized);
};

const authenticateWithVault = async ({
  vaultUrl,
  proxyJwt,
  vaultJwt,
  role,
}) => {
  const res = await fetch(`${vaultUrl}/v1/auth/github-actions-oidc/login`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "Proxy-Authorization-Token": `Bearer ${proxyJwt}`,
    },
    body: JSON.stringify({ role, jwt: vaultJwt }),
  });
  const body = await res.text();
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

const requestGithubAppToken = async ({
  vaultUrl,
  vaultToken,
  proxyJwt,
  app,
  role,
}) => {
  const res = await fetch(
    `${vaultUrl}/v1/github-app-${encodeURIComponent(app)}/token/${encodeURIComponent(role)}`,
    {
      method: "GET",
      headers: {
        "X-Vault-Token": vaultToken,
        "Proxy-Authorization-Token": `Bearer ${proxyJwt}`,
      },
    },
  );
  const body = await res.text();
  if (!res.ok) {
    throw new Error(
      `Vault GitHub App token request failed (HTTP ${res.status}): ${body}`,
    );
  }
  let parsed;
  try {
    parsed = JSON.parse(body);
  } catch (err) {
    throw new Error(`Failed to parse Vault token response: ${err.message}`, {
      cause: err,
    });
  }
  const token = parsed && parsed.data && parsed.data.token;
  if (!token) {
    throw new Error("Vault token response did not contain `data.token`.");
  }
  return token;
};

const pickRandom = (items) => items[Math.floor(Math.random() * items.length)];

const main = async () => {
  const { permissionSet, vaultInstance, apps, repositoryName } = parseInputs();

  const vaultUrl = `https://vault-github-actions.grafana-${vaultInstance}.net`;
  const proxyAudience = `vault-github-actions-grafana-${vaultInstance}`;
  // The Vault auth audience is the same as the Vault URL (Vault's
  // github-actions-oidc method validates the audience against its configured
  // value).
  const vaultAudience = vaultUrl;

  const refSha = normalizeWorkflowRefSha();
  const role = `${repositoryName}-${refSha}-${permissionSet}`;
  info(`Vault role: ${role}`);

  // 1) Mint OIDC tokens with retry. The proxy JWT is used on every Vault
  //    request; the Vault JWT is only used for the auth/login call.
  const proxyJwt = await retry({ label: "Mint proxy OIDC token" }, () =>
    fetchIdToken(proxyAudience),
  );
  setSecret(proxyJwt);

  const vaultJwt = await retry({ label: "Mint Vault OIDC token" }, () =>
    fetchIdToken(vaultAudience),
  );
  setSecret(vaultJwt);

  // 2) Authenticate with Vault.
  const vaultToken = await retry({ label: "Vault auth" }, () =>
    authenticateWithVault({ vaultUrl, proxyJwt, vaultJwt, role }),
  );
  setSecret(vaultToken);
  info("Vault auth done.");

  // 3) Save state for the post-job revocation BEFORE doing anything that can
  //    fail. If the GitHub App token request below fails, the post step will
  //    still revoke the Vault token (and any lease it might have created on a
  //    partial response).
  saveState("vault_url", vaultUrl);
  saveState("vault_token", vaultToken);
  saveState("proxy_audience", proxyAudience);

  // 4) Request the GitHub App installation token. The bash version selected a
  //    random app from the comma-separated list on every retry attempt; we do
  //    the same to spread load and to give a transient failure on one app a
  //    chance to land on another.
  const githubToken = await retry({ label: "Create GitHub App token" }, () => {
    const app = pickRandom(apps);
    info(`Selected GitHub App: ${app}`);
    return requestGithubAppToken({
      vaultUrl,
      vaultToken,
      proxyJwt,
      app,
      role,
    });
  });
  setSecret(githubToken);
  setOutput("token", githubToken);
  info("GitHub App token created.");
};

main().catch((err) => {
  setFailed(err && err.message ? err.message : String(err));
});
