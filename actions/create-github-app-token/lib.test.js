const {
  describe,
  test,
  expect,
  spyOn,
  beforeEach,
  afterEach,
} = require("bun:test");

const {
  getState,
  retry,
  setSecret,
  normalizeWorkflowRefSha,
} = require("./lib.js");

describe("lib.js", () => {
  describe("getState", () => {
    afterEach(() => {
      delete process.env.STATE_MY_KEY;
    });

    test("reads from STATE_ prefix", () => {
      process.env.STATE_MY_KEY = "my_value";
      expect(getState("MY_KEY")).toBe("my_value");
    });

    test("returns empty string if unset", () => {
      expect(getState("MY_KEY")).toBe("");
    });
  });

  describe("retry", () => {
    test("succeeds on first try", async () => {
      let calls = 0;
      const result = await retry(
        { label: "test", attempts: 3, baseDelayMs: 1 },
        async () => {
          calls++;
          return "success";
        },
      );
      expect(result).toBe("success");
      expect(calls).toBe(1);
    });

    test("fails twice then succeeds", async () => {
      let calls = 0;
      const result = await retry(
        { label: "test", attempts: 3, baseDelayMs: 1 },
        async () => {
          calls++;
          if (calls < 3) throw new Error("fail");
          return "success";
        },
      );
      expect(result).toBe("success");
      expect(calls).toBe(3);
    });

    test("throws on all attempts failing", async () => {
      expect(
        retry({ label: "test", attempts: 3, baseDelayMs: 1 }, async () => {
          throw new Error("fatal failure");
        }),
      ).rejects.toThrow("fatal failure");
    });
  });

  describe("setSecret", () => {
    let logSpy;
    beforeEach(() => {
      logSpy = spyOn(console, "log").mockImplementation(() => {});
    });
    afterEach(() => {
      logSpy.mockRestore();
    });

    test("logs ::add-mask:: if value is truthy", () => {
      setSecret("top-secret");
      expect(logSpy).toHaveBeenCalledWith("::add-mask::top-secret");
    });

    test("does not log if value is falsy", () => {
      setSecret("");
      expect(logSpy).not.toHaveBeenCalled();
    });
  });

  describe("normalizeWorkflowRefSha", () => {
    test("strips owner/repo prefix and @ref suffix, hashes the workflow path", () => {
      expect(
        normalizeWorkflowRefSha(
          "octocat/hello-world/.github/workflows/ci.yml@refs/heads/main",
        ),
      ).toBe(
        "b803fcb7f17ed9235f1e5cb1fcd2f5d3b2838429d4368ae4c57ce4436577f03f",
      );
    });

    test("returns the same hash regardless of owner, repo, or ref", () => {
      const fromMain = normalizeWorkflowRefSha(
        "octocat/hello-world/.github/workflows/ci.yml@refs/heads/main",
      );
      const fromTag = normalizeWorkflowRefSha(
        "octocat/hello-world/.github/workflows/ci.yml@refs/tags/v1.0.0",
      );
      const fromSha = normalizeWorkflowRefSha(
        "octocat/hello-world/.github/workflows/ci.yml@0123456789abcdef0123456789abcdef01234567",
      );
      const differentOwner = normalizeWorkflowRefSha(
        "other-owner/other-repo/.github/workflows/ci.yml@refs/heads/main",
      );
      expect(fromTag).toBe(fromMain);
      expect(fromSha).toBe(fromMain);
      expect(differentOwner).toBe(fromMain);
    });

    test("produces different hashes for different workflow paths", () => {
      const ci = normalizeWorkflowRefSha(
        "octocat/hello-world/.github/workflows/ci.yml@refs/heads/main",
      );
      const release = normalizeWorkflowRefSha(
        "octocat/hello-world/.github/workflows/release.yml@refs/heads/main",
      );
      expect(ci).not.toBe(release);
      expect(release).toBe(
        "87db21a973eed4fef5f32b267aa60fcee5cbdf03c67fafdc2a9b553bb0b15f34",
      );
    });

    test("preserves nested path segments beyond owner/repo", () => {
      // Only the first two segments (owner/repo) are stripped; the remaining
      // path is hashed as-is.
      expect(
        normalizeWorkflowRefSha(
          "octocat/hello-world/path/to/nested/workflow.yml@refs/heads/main",
        ),
      ).toBe(
        "95330058c75b9737b0ed5f61a4c3a6ccc4b4810eb5160209f6126ffa02445682",
      );
    });

    test("handles refs that contain '@' (the first @ ends the path)", () => {
      // .replace(/@.*$/, "") is greedy from the first @, so any @ in the ref
      // (e.g. ref names containing '@') is stripped along with the rest.
      const withSimpleRef = normalizeWorkflowRefSha(
        "octocat/hello-world/.github/workflows/ci.yml@refs/heads/main",
      );
      const withAtInRef = normalizeWorkflowRefSha(
        "octocat/hello-world/.github/workflows/ci.yml@refs/heads/feature@v2",
      );
      expect(withAtInRef).toBe(withSimpleRef);
    });

    test("returns a 64-character lowercase hex string", () => {
      const result = normalizeWorkflowRefSha(
        "octocat/hello-world/.github/workflows/ci.yml@refs/heads/main",
      );
      expect(result).toMatch(/^[0-9a-f]{64}$/);
    });

    test("hashes the empty string when input is just 'owner/repo/@ref'", () => {
      // Edge case: after stripping owner/repo/ and @ref, the workflow path is
      // empty. The function should still return a deterministic sha256.
      expect(normalizeWorkflowRefSha("owner/repo/@refs/heads/main")).toBe(
        "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
      );
    });
  });
});
