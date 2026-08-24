// The request to GitHub's OIDC endpoint intermittently stalls. core.getIDToken()
// waits three minutes on a dead socket and does not retry, so a single stall costs
// the whole job. Call the endpoint directly to bound each attempt, then retry.

const REQUEST_TIMEOUT_MS = 30000;
const MAX_ATTEMPTS = 3;

const TROUBLESHOOTING = `
🔧 OIDC Token Error - How to Fix:

This error typically occurs when your workflow lacks proper permissions for OIDC token generation.

✅ Solution 1 - Add workflow-level permissions:
Add this to the top of your workflow YAML file:

permissions:
  id-token: write
  contents: read

✅ Solution 2 - Add job-level permissions:
Add this to your specific job:

jobs:
  your-job-name:
    permissions:
      id-token: write
      contents: read

✅ Solution 3 - Verify repository configuration:
- Ensure your repository has OIDC enabled
- Check that the Vault OIDC provider is configured for your repository

📚 More info: https://docs.github.com/en/actions/deployment/security-hardening-your-deployments/about-security-hardening-with-openid-connect
`;

function delay(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

// requestIDToken returns an OIDC token for the audience, and throws if it cannot.
// The request is bounded by REQUEST_TIMEOUT_MS.
async function requestIDToken({ core, env, fetch, audience }) {
  const requestUrl = env.ACTIONS_ID_TOKEN_REQUEST_URL;
  const requestToken = env.ACTIONS_ID_TOKEN_REQUEST_TOKEN;

  if (!requestUrl || !requestToken) {
    throw new Error(
      "Unable to get ACTIONS_ID_TOKEN_REQUEST_URL or ACTIONS_ID_TOKEN_REQUEST_TOKEN env variables",
    );
  }

  // Native fetch ignores the proxy variables that @actions/http-client honours, so
  // proxied runners keep the old unbounded call.
  if (env.https_proxy || env.HTTPS_PROXY) {
    return core.getIDToken(audience);
  }

  let response;
  try {
    response = await fetch(
      `${requestUrl}&audience=${encodeURIComponent(audience)}`,
      {
        headers: { authorization: `Bearer ${requestToken}` },
        signal: AbortSignal.timeout(REQUEST_TIMEOUT_MS),
      },
    );
  } catch (error) {
    if (error.name === "TimeoutError") {
      throw new Error(
        `the OIDC endpoint did not respond within ${REQUEST_TIMEOUT_MS}ms`,
        { cause: error },
      );
    }
    throw error;
  }

  if (!response.ok) {
    throw new Error(
      `the OIDC endpoint returned ${response.status} ${response.statusText}`,
    );
  }

  const { value } = await response.json();
  if (!value) {
    throw new Error("the OIDC endpoint returned no token");
  }

  return value;
}

// getIDToken writes an OIDC token for the Vault instance named by VAULT_INSTANCE to
// the `github-jwt` output. It marks the step as failed when every attempt fails.
//
// env, fetch and sleep exist so tests can drive the retry loop.
module.exports = async function getIDToken({
  core,
  env = process.env,
  fetch = globalThis.fetch,
  sleep = delay,
}) {
  const audience = `vault-github-actions-grafana-${env.VAULT_INSTANCE}`;

  // Without these the job is missing `id-token: write`, which no retry can fix.
  const maxAttempts =
    env.ACTIONS_ID_TOKEN_REQUEST_URL && env.ACTIONS_ID_TOKEN_REQUEST_TOKEN
      ? MAX_ATTEMPTS
      : 1;

  let jwt;
  let lastError;

  for (let attempt = 1; attempt <= maxAttempts; attempt++) {
    try {
      jwt = await requestIDToken({ core, env, fetch, audience });
      break;
    } catch (error) {
      lastError = error;

      if (attempt < maxAttempts) {
        const delayMs = 1000 * 2 ** (attempt - 1);
        core.warning(
          `Attempt ${attempt}/${maxAttempts} to get the OIDC token failed: ${error.message}. Retrying in ${delayMs}ms.`,
        );
        await sleep(delayMs);
      }
    }
  }

  if (jwt === undefined) {
    core.setFailed(
      `❌ Failed to get OIDC token after ${maxAttempts} attempt(s): ${lastError.message}`,
    );
    core.error(TROUBLESHOOTING);
    return;
  }

  core.setSecret(jwt);
  core.setOutput("github-jwt", jwt);
};
