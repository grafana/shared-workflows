// Socket generates the full scan report lazily, so the SPDX export endpoint 404s for a
// while after `socket scan create` returns. This polls until the report is ready.
//
// This replaces a bare `curl` loop that could not tell a 403 from a 404: it retried
// every non-200 for the full timeout and then reported the failure as "not ready".
// curl also sends no User-Agent, which we suspect trips bot protection in front of the
// Socket API. Both are addressed here.

// `actions/github-script` loads this file with require(), so it has to stay CommonJS.
/* eslint-disable @typescript-eslint/no-require-imports */
const fsPromises = require("node:fs/promises");
const path = require("node:path");
/* eslint-enable @typescript-eslint/no-require-imports */

const REQUEST_TIMEOUT_SECONDS = 30;
const INITIAL_DELAY_SECONDS = 5;
const MAX_DELAY_SECONDS = 30;

// Enough of an error body to identify the failure without flooding the log.
const BODY_SNIPPET_LIMIT = 500;

const USER_AGENT =
  "grafana-shared-workflows-socket-export-sbom (+https://github.com/grafana/shared-workflows)";

const REQUIRED_ENV = [
  "SOCKET_API_TOKEN",
  "SOCKET_BASE_URL",
  "SOCKET_ORG",
  "FULL_SCAN_ID",
  "REPO_NAME",
  "BRANCH",
  "ACTION_PATH",
];

// Socket has not finished generating the report (404), is rate limiting us (429), or is
// briefly unhealthy (5xx). Every other status -- 400, 401, 403 -- means the request
// itself is wrong, and retrying only delays the failure until the timeout expires.
const RETRYABLE_STATUSES = new Set([404, 408, 429]);

// Retrying cannot add a missing scope, and a bot challenge will not clear on its own.
const AUTH_HINT =
  "Check that the Socket API token carries the `report:read` scope and that the request is not being blocked by bot protection.";

// Every duration in this file is in seconds, matching the `export_timeout_seconds`
// input. setTimeout is the only place that needs milliseconds.
function delay(seconds) {
  return new Promise((resolve) => setTimeout(resolve, seconds * 1000));
}

function isRetryableStatus(status) {
  return RETRYABLE_STATUSES.has(status) || status >= 500;
}

// exportUrl joins the base URL and the endpoint path with exactly one separator.
// `socket_base_url` ends in a trailing slash -- the Socket CLI needs it, since it
// appends endpoint paths without adding a separator of its own -- so the slash is
// stripped here rather than assumed away: `https://api.socket.dev/v0//orgs/...` is
// not the same path to Socket.
function exportUrl({ baseUrl, org, scanId }) {
  const base = baseUrl.replace(/\/+$/, "");
  const scan = encodeURIComponent(scanId);

  return `${base}/orgs/${encodeURIComponent(org)}/export/spdx/${scan}`;
}

// outputFileName is the caller's file name, or `<repo>-<branch>.spdx.json` with slashes
// in the branch flattened: "grafana" + "feature/foo" -> "grafana-feature-foo.spdx.json".
function outputFileName({ outputFile, repo, branch }) {
  // basename, because the caller may pass a path but the file goes in the action dir.
  if (outputFile) {
    return path.basename(outputFile);
  }

  return `${repo}-${branch.replaceAll("/", "-")}.spdx.json`;
}

function bodySnippet(body) {
  const text = body.trim();
  if (!text) {
    return "";
  }

  const shown =
    text.length > BODY_SNIPPET_LIMIT
      ? `${text.slice(0, BODY_SNIPPET_LIMIT)}...`
      : text;

  return `: ${shown}`;
}

// fetchSBOM makes one attempt at the export. It returns the SBOM body on success, and
// otherwise the reason the attempt failed and whether another attempt could succeed.
async function fetchSBOM({ fetch, url, token }) {
  let response;

  try {
    response = await fetch(url, {
      headers: {
        authorization: `Bearer ${token}`,
        accept: "application/json",
        "user-agent": USER_AGENT,
      },
      signal: AbortSignal.timeout(REQUEST_TIMEOUT_SECONDS * 1000),
    });
  } catch (error) {
    // A stalled, refused or reset connection is worth another attempt.
    return {
      retryable: true,
      reason:
        error.name === "TimeoutError"
          ? `the request did not complete within ${REQUEST_TIMEOUT_SECONDS}s`
          : `the request failed: ${error.message}`,
    };
  }

  const body = await response.text();

  if (!response.ok) {
    return {
      retryable: isRetryableStatus(response.status),
      reason: `Socket returned ${response.status} ${response.statusText}${bodySnippet(body)}`,
    };
  }

  // Socket has served a 200 holding a partial report while the scan is still being
  // generated, so the body has to parse before the export counts as done.
  try {
    JSON.parse(body);
  } catch (error) {
    return {
      retryable: true,
      reason: `Socket returned 200 with a body that is not valid JSON (${error.message})${bodySnippet(body)}`,
    };
  }

  return { body };
}

// exportSBOM writes the SPDX SBOM for FULL_SCAN_ID into the action directory and
// publishes its location as the `path` output. It marks the step as failed when the
// export cannot be retrieved.
//
// env, fetch, sleep, now and writeFile exist so tests can drive the polling loop.
module.exports = async function exportSBOM({
  core,
  env = process.env,
  fetch = globalThis.fetch,
  sleep = delay,
  now = Date.now,
  writeFile = fsPromises.writeFile,
}) {
  const missing = REQUIRED_ENV.filter((name) => !env[name]);
  if (missing.length > 0) {
    core.setFailed(`Missing required environment: ${missing.join(", ")}`);
    return;
  }

  const timeoutSeconds = Number(env.TIMEOUT_SECONDS);
  if (!Number.isFinite(timeoutSeconds) || timeoutSeconds < 0) {
    core.setFailed(
      `export_timeout_seconds must be a non-negative number, got "${env.TIMEOUT_SECONDS}"`,
    );
    return;
  }

  const url = exportUrl({
    baseUrl: env.SOCKET_BASE_URL,
    org: env.SOCKET_ORG,
    scanId: env.FULL_SCAN_ID,
  });

  const dest = path.join(
    env.ACTION_PATH,
    outputFileName({
      outputFile: env.OUTPUT_FILE,
      repo: env.REPO_NAME,
      branch: env.BRANCH,
    }),
  );

  // now() stays a millisecond clock (Date.now); the deadline it feeds is seconds.
  const deadlineSeconds = now() / 1000 + timeoutSeconds;
  let delaySeconds = INITIAL_DELAY_SECONDS;

  for (let attempt = 1; ; attempt++) {
    const { body, retryable, reason } = await fetchSBOM({
      fetch,
      url,
      token: env.SOCKET_API_TOKEN,
    });

    if (body !== undefined) {
      // Only a validated body reaches the destination, so a rejection page can never
      // be mistaken for an SBOM by a later step.
      await writeFile(dest, body);
      core.info(`SBOM ready after ${attempt} attempt(s)`);
      core.setOutput("path", dest);
      return;
    }

    if (!retryable) {
      core.setFailed(`Failed to export the SBOM: ${reason}. ${AUTH_HINT}`);
      return;
    }

    const remainingSeconds = deadlineSeconds - now() / 1000;
    if (remainingSeconds <= 0) {
      core.setFailed(
        `SBOM export not ready after ${attempt} attempt(s) and ${timeoutSeconds}s: ${reason}`,
      );
      return;
    }

    // Never sleep past the deadline; the next check would only report the timeout.
    const waitSeconds = Math.min(delaySeconds, remainingSeconds);
    core.info(`SBOM not ready yet (${reason}); retrying in ${waitSeconds}s`);
    await sleep(waitSeconds);
    delaySeconds = Math.min(delaySeconds * 2, MAX_DELAY_SECONDS);
  }
};
