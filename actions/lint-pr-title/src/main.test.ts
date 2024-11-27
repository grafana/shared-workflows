import load from "@commitlint/load";
import { QualifiedConfig, RuleConfigSeverity } from "@commitlint/types";
import { beforeAll, describe, expect, it } from "bun:test";

import {
  Octokit,
  compareCommitsWithBaseheadResponse,
  handleEvent,
  lint,
  loadConfig,
} from "./main";
import { compareCommitsWithBaseheadResponses } from "./testUtils/compareCommitsWithBaseheadresponses";
import { mergeQueueContext } from "./testUtils/mergeGroupContext";
import { expect_toBeDefined, newContextFromPullRequest } from "./testUtils";
import { tmpFileAsync } from "./tempfile";

const config = await loadConfig("commitlint.config.js");

function mockOctokit(
  data: compareCommitsWithBaseheadResponse = { commits: [] },
): Octokit {
  const compareCommitsWithBasehead = () => {
    return Promise.resolve({ data });
  };

  return {
    rest: {
      repos: {
        compareCommitsWithBasehead,
      },
    },
  };
}

describe("lint", () => {
  it("should accept a valid title", () => {
    expect(
      lint([{ message: "fix(ci): fix CI" }], config),
    ).resolves.toMatchObject({
      valid: true,
    });
  });

  it("should reject a commit title that ends with a period", () => {
    expect(
      lint([{ message: "fix(ci): fix CI." }], config),
    ).resolves.toMatchObject({
      valid: false,
    });
  });

  it("can handle custom rules", async () => {
    const mockConfig = await load(
      {
        extends: ["@commitlint/config-conventional"],
        rules: {
          // This overrides the default rule from
          // @commitlint/config-conventional to only allow "fix" types.
          "type-enum": [RuleConfigSeverity.Error, "always", ["fix"]],
        },
      },
      {
        // Prevent commitlint from discovering and loading our default config file.
        cwd: "/",
      },
    );

    const typeEnumRule = mockConfig.rules["type-enum"];
    expect_toBeDefined(typeEnumRule);

    expect(typeEnumRule).not.toBe([RuleConfigSeverity.Disabled]);

    const [, , typeEnumRuleConfig] = typeEnumRule;

    expect(typeEnumRuleConfig).toEqual(["fix"]);

    expect(
      lint([{ message: "chore(no): this is not allowed" }], mockConfig),
    ).resolves.toMatchObject({
      valid: false,
    });
  });
});

describe("pull_request", () => {
  let mockConfig: QualifiedConfig;

  beforeAll(async () => {
    mockConfig = await load(
      {
        extends: ["@commitlint/config-conventional"],
      },
      {
        cwd: "/",
      },
    );
  });

  it.each<{
    name: string;
    title: string;
    body?: string;
    validTitleOnly: boolean;
    validTitleAndMessage: boolean;
  }>([
    {
      name: "valid title",
      title: "fix(CI): fix CI",
      body: "a good body",
      validTitleOnly: true,
      validTitleAndMessage: true,
    },
    {
      name: "trailing period",
      title: "fix(CI): fix CI.",
      validTitleOnly: false,
      validTitleAndMessage: false,
    },
    {
      name: "really long body line",
      title: "fix(CI): fix CI",
      body: "a".repeat(512),
      validTitleOnly: true,
      validTitleAndMessage: false,
    },
  ])(
    "pull request event",
    ({ title, body, validTitleOnly, validTitleAndMessage }) => {
      const ctx = newContextFromPullRequest(title, body);

      for (const titleOnly of [true, false]) {
        const lintResult = handleEvent(ctx, mockOctokit(), {
          actionConfig: {
            configPath: "commitlint.config.js",
            titleOnly: titleOnly,
          },
          commitLintConfig: mockConfig,
        });

        expect(lintResult).resolves.toMatchObject({
          valid: titleOnly ? validTitleOnly : validTitleAndMessage,
        });
      }
    },
  );
});

describe("merge_group", () => {
  let mockConfig: QualifiedConfig;

  beforeAll(async () => {
    mockConfig = await load(
      {
        extends: ["@commitlint/config-conventional"],
      },
      {
        cwd: "/",
      },
    );
  });

  it.each(compareCommitsWithBaseheadResponses)(
    "should lint properly",
    async ({ expectedCheckedCommits, commits, valid }) => {
      const lintResult = await handleEvent(
        mergeQueueContext,
        mockOctokit({ commits }),
        {
          actionConfig: {
            configPath: "commitlint.config.js",
            titleOnly: true,
          },
          commitLintConfig: mockConfig,
        },
      );

      expect(lintResult.valid).toBe(valid);

      // get the checked SHAs
      const checkedSHAs = lintResult.results.map((r) => r.sha);

      expect(checkedSHAs).toEqual(expectedCheckedCommits);
    },
  );
});

describe("loadConfig", () => {
  it("should load the default config", async () => {
    const config = await loadConfig("commitlint.config.js");

    expect(config).toMatchObject({
      extends: ["@commitlint/config-conventional"],
    });
  });

  it("should load a custom config", async () => {
    const cfg = {
      extends: ["@commitlint/config-conventional"],
      rules: {
        "body-max-line-length": [1, "always", 100],
      },
    };

    await using tempFile = await tmpFileAsync({
      template: "commitlint-XXXXXX.config.js",
    });

    const { name, handle } = tempFile;

    await handle.writeFile(`module.exports = ${JSON.stringify(cfg)}`);

    const config = await loadConfig(name);

    // `body-max-line-length` here differes from the default so we're making
    // sure we didn't just load that.
    expect(config).toMatchObject({
      extends: ["@commitlint/config-conventional"],
      rules: {
        "body-max-line-length": [1, "always", 100],
      },
    });
  });
});
