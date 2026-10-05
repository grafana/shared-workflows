const { describe, test, expect } = require("bun:test");

const {
  validateOperation,
  validateVaultInstance,
  parseSecrets,
  roleName,
} = require("./lib.js");

describe("lib.js", () => {
  describe("validateOperation", () => {
    test.each(["publish_winget", "publish-artifact", "sign2", "a_b-c"])(
      "accepts %s",
      (op) => {
        expect(validateOperation(op)).toBe(op);
      },
    );

    test.each([
      "",
      "Publish",
      "publish__winget",
      "_publish",
      "publish-",
      "../common",
      "publish/winget",
      "publish winget",
    ])("rejects '%s'", (op) => {
      expect(() => validateOperation(op)).toThrow(
        /Invalid value for operation/,
      );
    });
  });

  describe("validateVaultInstance", () => {
    test.each(["dev", "ops"])("accepts %s", (instance) => {
      expect(validateVaultInstance(instance)).toBe(instance);
    });

    test.each(["prod", "", "OPS"])("rejects '%s'", (instance) => {
      expect(() => validateVaultInstance(instance)).toThrow(
        /Must be 'dev' or 'ops'/,
      );
    });
  });

  describe("parseSecrets", () => {
    test("builds paths under the operation", () => {
      expect(
        parseSecrets(
          "publish_winget",
          "WINGET_TOKEN=winget:token\n\n  OTHER=nested/path:key.1  \n",
        ),
      ).toEqual([
        {
          envName: "WINGET_TOKEN",
          path: "ci/data/operations/publish_winget/winget",
          key: "token",
          line: 1,
        },
        {
          envName: "OTHER",
          path: "ci/data/operations/publish_winget/nested/path",
          key: "key.1",
          line: 3,
        },
      ]);
    });

    test("handles CRLF line endings", () => {
      expect(parseSecrets("op", "A=x:k\r\nB=y:k")).toHaveLength(2);
    });

    test("rejects empty input", () => {
      expect(() => parseSecrets("op", "  \n ")).toThrow(/must not be empty/);
      expect(() => parseSecrets("op", undefined)).toThrow(/must not be empty/);
    });

    test.each([
      ["no separator", "WINGET_TOKEN"],
      ["no key", "A=winget"],
      ["empty key", "A=winget:"],
      ["bad name", "1A=winget:token"],
      ["name with dash", "A-B=winget:token"],
      ["absolute subpath", "A=/winget:token"],
      ["parent segment", "A=../common/x:token"],
      ["nested parent segment", "A=a/../../x:token"],
      ["empty segment", "A=a//b:token"],
      ["dot segment", "A=./winget:token"],
      ["bad key", "A=winget:to ken"],
      ["key with colon", "A=winget:a:b"],
    ])("rejects %s", (_name, line) => {
      expect(() => parseSecrets("op", line)).toThrow(/Invalid/);
    });

    test("rejects duplicate names", () => {
      expect(() => parseSecrets("op", "A=x:k\nA=y:k")).toThrow(
        /line 2: the name is already used/,
      );
    });

    test("errors name the line number, not the line's text", () => {
      expect(() => parseSecrets("op", "A=x:k\n\nMY_TOKEN=../x:k")).toThrow(
        /^Invalid secrets line 3: /,
      );
      try {
        parseSecrets("op", "MY_TOKEN=../x:k");
      } catch (err) {
        expect(err.message).not.toContain("MY_TOKEN");
        expect(err.message).not.toContain("../x");
      }
    });
  });

  describe("roleName", () => {
    // Expected hashes were checked against Terraform's
    // substr(sha256("<org>/<repo>:<operation>"), 0, 8), the formula used by
    // the CI gates Vault binding in deployment_tools.
    test("matches the CI gates Vault binding", () => {
      expect(
        roleName({ repository: "grafana/k6", operation: "publish_winget" }),
      ).toBe("blessed-publish_winget-grafana-k6-5494e3e0");
    });

    test("includes the org", () => {
      expect(
        roleName({
          repository: "grafana-staff-apps/my-repo",
          operation: "publish_artifact",
        }),
      ).toBe("blessed-publish_artifact-grafana-staff-apps-my-repo-b7a15e1c");
    });

    test.each(["", "grafana", "/k6", "grafana/"])(
      "rejects malformed repository '%s'",
      (repository) => {
        expect(() => roleName({ repository, operation: "op" })).toThrow(
          /GITHUB_REPOSITORY/,
        );
      },
    );
  });
});
