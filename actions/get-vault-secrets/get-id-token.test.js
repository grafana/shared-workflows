import { describe, expect, mock, test } from "bun:test";

import getIDToken from "./get-id-token.js";

const TOKEN = "a-github-oidc-token";
const PROXY_TOKEN = "a-proxied-github-oidc-token";
const REQUEST_URL = "https://token.example/222/idtoken/abc?api-version=2.0";

function makeCore() {
  return {
    warning: mock(),
    error: mock(),
    setFailed: mock(),
    setSecret: mock(),
    setOutput: mock(),
    getIDToken: mock(() => Promise.resolve(PROXY_TOKEN)),
  };
}

function makeEnv(overrides = {}) {
  return {
    VAULT_INSTANCE: "ops",
    ACTIONS_ID_TOKEN_REQUEST_URL: REQUEST_URL,
    ACTIONS_ID_TOKEN_REQUEST_TOKEN: "a-request-token",
    ...overrides,
  };
}

function ok(body = { value: TOKEN }) {
  return {
    ok: true,
    status: 200,
    statusText: "OK",
    json: () => Promise.resolve(body),
  };
}

function httpError(status, statusText) {
  return { ok: false, status, statusText };
}

function timeout() {
  const error = new Error("The operation timed out");
  error.name = "TimeoutError";
  return error;
}

// makeFetch answers consecutive calls with the given responses. An Error is thrown
// instead of returned, and one more call than there are responses fails the test.
function makeFetch(...responses) {
  let call = 0;

  return mock(() => {
    if (call >= responses.length) {
      throw new Error(`unexpected fetch call ${call + 1}`);
    }

    const response = responses[call++];
    if (response instanceof Error) {
      return Promise.reject(response);
    }

    return Promise.resolve(response);
  });
}

describe("getIDToken", () => {
  test("publishes the token from the OIDC endpoint", async () => {
    const core = makeCore();
    const fetch = makeFetch(ok());

    await getIDToken({ core, env: makeEnv(), fetch, sleep: mock() });

    expect(fetch).toHaveBeenCalledTimes(1);
    expect(core.setSecret).toHaveBeenCalledWith(TOKEN);
    expect(core.setOutput).toHaveBeenCalledWith("github-jwt", TOKEN);
    expect(core.warning).not.toHaveBeenCalled();
    expect(core.setFailed).not.toHaveBeenCalled();
  });

  test("asks for the audience of the Vault instance", async () => {
    const core = makeCore();
    const fetch = makeFetch(ok(), ok());

    await getIDToken({ core, env: makeEnv(), fetch, sleep: mock() });
    await getIDToken({
      core,
      env: makeEnv({ VAULT_INSTANCE: "dev" }),
      fetch,
      sleep: mock(),
    });

    expect(fetch.mock.calls[0][0]).toBe(
      `${REQUEST_URL}&audience=vault-github-actions-grafana-ops`,
    );
    expect(fetch.mock.calls[1][0]).toBe(
      `${REQUEST_URL}&audience=vault-github-actions-grafana-dev`,
    );
  });

  test("authenticates the request and bounds it with a timeout", async () => {
    const core = makeCore();
    const fetch = makeFetch(ok());

    await getIDToken({ core, env: makeEnv(), fetch, sleep: mock() });

    const [, options] = fetch.mock.calls[0];
    expect(options.headers.authorization).toBe("Bearer a-request-token");
    expect(options.signal).toBeInstanceOf(AbortSignal);
  });

  test("retries a timed out request", async () => {
    const core = makeCore();
    const fetch = makeFetch(timeout(), ok());
    const sleep = mock();

    await getIDToken({ core, env: makeEnv(), fetch, sleep });

    expect(fetch).toHaveBeenCalledTimes(2);
    expect(core.setOutput).toHaveBeenCalledWith("github-jwt", TOKEN);
    expect(core.setFailed).not.toHaveBeenCalled();
    expect(core.warning).toHaveBeenCalledTimes(1);
    expect(core.warning.mock.calls[0][0]).toBe(
      "Attempt 1/3 to get the OIDC token failed: the OIDC endpoint did not respond within 30000ms. Retrying in 1000ms.",
    );
    expect(sleep.mock.calls).toEqual([[1000]]);
  });

  test("retries a failed request", async () => {
    const core = makeCore();
    const fetch = makeFetch(httpError(500, "Internal Server Error"), ok());

    await getIDToken({ core, env: makeEnv(), fetch, sleep: mock() });

    expect(fetch).toHaveBeenCalledTimes(2);
    expect(core.setOutput).toHaveBeenCalledWith("github-jwt", TOKEN);
    expect(core.warning.mock.calls[0][0]).toContain(
      "the OIDC endpoint returned 500 Internal Server Error",
    );
  });

  test("retries a request that returns no token", async () => {
    const core = makeCore();
    const fetch = makeFetch(ok({}), ok({}), ok());

    await getIDToken({ core, env: makeEnv(), fetch, sleep: mock() });

    expect(fetch).toHaveBeenCalledTimes(3);
    expect(core.setOutput).toHaveBeenCalledWith("github-jwt", TOKEN);
    expect(core.warning.mock.calls[0][0]).toContain(
      "the OIDC endpoint returned no token",
    );
  });

  test("gives up after three attempts and backs off between them", async () => {
    const core = makeCore();
    const fetch = makeFetch(timeout(), timeout(), timeout());
    const sleep = mock();

    await getIDToken({ core, env: makeEnv(), fetch, sleep });

    expect(fetch).toHaveBeenCalledTimes(3);
    expect(sleep.mock.calls).toEqual([[1000], [2000]]);
    expect(core.warning).toHaveBeenCalledTimes(2);
    expect(core.setOutput).not.toHaveBeenCalled();
    expect(core.setFailed.mock.calls[0][0]).toBe(
      "❌ Failed to get OIDC token after 3 attempt(s): the OIDC endpoint did not respond within 30000ms",
    );
    expect(core.error.mock.calls[0][0]).toContain("id-token: write");
  });

  test("reports an unexpected request error as it is", async () => {
    const core = makeCore();
    const error = new TypeError("fetch failed");
    const fetch = makeFetch(error, error, error);

    await getIDToken({ core, env: makeEnv(), fetch, sleep: mock() });

    expect(core.setFailed.mock.calls[0][0]).toContain("fetch failed");
  });

  test("fails without retrying when the OIDC endpoint is not available", async () => {
    const core = makeCore();
    const fetch = makeFetch();

    await getIDToken({
      core,
      env: makeEnv({
        ACTIONS_ID_TOKEN_REQUEST_URL: undefined,
        ACTIONS_ID_TOKEN_REQUEST_TOKEN: undefined,
      }),
      fetch,
      sleep: mock(),
    });

    expect(fetch).not.toHaveBeenCalled();
    expect(core.warning).not.toHaveBeenCalled();
    expect(core.setOutput).not.toHaveBeenCalled();
    expect(core.setFailed.mock.calls[0][0]).toBe(
      "❌ Failed to get OIDC token after 1 attempt(s): Unable to get ACTIONS_ID_TOKEN_REQUEST_URL or ACTIONS_ID_TOKEN_REQUEST_TOKEN env variables",
    );
    expect(core.error.mock.calls[0][0]).toContain("id-token: write");
  });

  test.each(["https_proxy", "HTTPS_PROXY"])(
    "goes through core.getIDToken when %s is set",
    async (variable) => {
      const core = makeCore();
      const fetch = makeFetch();

      await getIDToken({
        core,
        env: makeEnv({ [variable]: "http://proxy.example:3128" }),
        fetch,
        sleep: mock(),
      });

      expect(fetch).not.toHaveBeenCalled();
      expect(core.getIDToken).toHaveBeenCalledWith(
        "vault-github-actions-grafana-ops",
      );
      expect(core.setSecret).toHaveBeenCalledWith(PROXY_TOKEN);
      expect(core.setOutput).toHaveBeenCalledWith("github-jwt", PROXY_TOKEN);
    },
  );

  test("retries core.getIDToken on proxied runners", async () => {
    const core = makeCore();
    core.getIDToken = mock(() => Promise.reject(new Error("Request timeout")));

    await getIDToken({
      core,
      env: makeEnv({ https_proxy: "http://proxy.example:3128" }),
      fetch: makeFetch(),
      sleep: mock(),
    });

    expect(core.getIDToken).toHaveBeenCalledTimes(3);
    expect(core.setFailed.mock.calls[0][0]).toContain("Request timeout");
  });
});
