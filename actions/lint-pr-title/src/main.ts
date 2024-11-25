import type { LintOutcome, QualifiedConfig } from "@commitlint/types";
import type {
  MergeGroupChecksRequestedEvent,
  PullRequestEvent,
} from "@octokit/webhooks-types";
import { resolve } from "path";
import fsPromises from "node:fs/promises";
import { default as commitLint } from "@commitlint/lint";
import { context } from "@actions/github";
import { format } from "@commitlint/format";
import load from "@commitlint/load";
import { tmpFileAsync } from "./tempfile";

export type Context = typeof context;
export type WebHookPayload = Context["payload"];

interface ActionConfig {
  configPath: string;
  titleOnly: boolean;
}

interface Config {
  actionConfig: ActionConfig;
  commitLintConfig: QualifiedConfig;
}

/**
 * The Octokit API methods used by this action. We specify a subset so that the
 * testsuite can mock just the methods we need.
 *
 * @private
 */
export type compareCommitsWithBaseheadResponse = {
  commits: {
    sha: string;
    commit: {
      message: string;
    };
    parents: object[];
  }[];
};

type compareCommitsWithBasehead = ({
  owner,
  repo,
  basehead,
}: {
  owner: string;
  repo: string;
  basehead: string;
}) => Promise<{ data: compareCommitsWithBaseheadResponse }>;

export interface Octokit {
  rest: {
    repos: {
      compareCommitsWithBasehead: compareCommitsWithBasehead;
    };
  };
}

/**
 * Format a list of strings for display: with commas and an "or" before the last
 * element.
 *
 * @param strings - The list of strings to join.
 * @returns The formatted string.
 */
function joinStringWithOr(strings: string[]): string {
  switch (strings.length) {
    case 0:
      return "";
    case 1:
      return strings[0];
    case 2:
      return strings.join(" or ");
    default:
      return `${strings.slice(0, -1).join(", ")}, or ${
        strings[strings.length - 1]
      }`;
  }
}

/**
 * Error thrown when the commit message is invalid.
 */
export class LintError extends Error {
  constructor(public readonly report: LintResult) {
    super(format(report));
  }
}

/**
 * Error thrown when the event type is not one of the types this action can
 * handle.
 */
export class WrongEventTypeError extends Error {
  private static HANDLED_TARGETS = [
    "merge_group",
    "pull_request_target",
    "pull_request",
  ];

  constructor(public readonly event: string) {
    const message = `This action can only be run on ${joinStringWithOr(
      WrongEventTypeError.HANDLED_TARGETS.map((target) => `\`${target}\``),
    )} events. Got event \`${event}\`.`;

    super(message);
  }
}

export interface LintResult {
  valid: boolean;
  errorCount: number;
  warningCount: number;
  results: (LintOutcome & { sha?: string })[];
}

/**
 * Lint the messages against the provided rules. If the message is invalid, an
 * error is thrown.
 *
 * @param messages - The list of commit or PR messages to lint.
 * @param config - The config object containing the rules to lint against.
 * @throws {LintError} If the title is invalid.
 */
export async function lint(
  messages: { sha?: string; message: string }[],
  { parserPreset, plugins, ignores, defaultIgnores, rules }: QualifiedConfig,
): Promise<LintResult> {
  if (messages.length === 0) {
    return {
      valid: true,
      errorCount: 0,
      warningCount: 0,
      results: [],
    };
  }

  const opts = {
    parserOpts: parserPreset?.parserOpts ?? {},
    plugins,
    ignores: ignores ?? [],
    defaultIgnores: defaultIgnores ?? true,
  };

  const results = await Promise.all(
    messages.map(async ({ sha, message }) => {
      console.log(`Linting message: ${message}`);
      return { sha, ...(await commitLint(message, rules, opts)) };
    }),
  );

  const report = results.reduce(
    (info, result) => {
      info.valid = result.valid ? info.valid : false;
      info.errorCount += result.errors.length;
      info.warningCount += result.warnings.length;
      info.results.push(result);

      return info;
    },
    {
      valid: true,
      errorCount: 0,
      warningCount: 0,
      results: <LintOutcome[]>[],
    },
  );

  return report;
}

/**
 * Handle the pull request event. The title of the pull request is linted
 * against the provided rules. If the title is invalid, an error is thrown.
 *
 * @param prCtx - The pull request context.
 * @param config - The configuration object.
 * @throws If the title is invalid.
 */
export async function handlePullRequest(
  prCtx: WebHookPayload,
  { actionConfig: { titleOnly }, commitLintConfig }: Config,
): Promise<LintResult> {
  const prContext = prCtx as PullRequestEvent;

  console.log(
    `Handling ${prContext.action} event for PR #${prContext.number.toString()}`,
  );

  let message = prContext.pull_request.title;

  if (!titleOnly && prContext.pull_request.body) {
    message += `\n\n${prContext.pull_request.body}`;
  }

  return lint([{ message }], commitLintConfig);
}

/**
 * Handle the merge group event. The titles of each commit from this pull
 * request in the merge group are linted against the provided rules. If any
 * title is invalid, an error is thrown.
 *
 * @param context - The pull request context.
 * @param octokit - The authenticated Octokit instance.
 * @param config - The configuration object.
 * @throws If any title is invalid.
 */
async function handleMergeGroup(
  context: Context,
  octokit: Octokit,
  { actionConfig: { titleOnly }, commitLintConfig: config }: Config,
): Promise<LintResult> {
  console.log("linting merge group");
  const mergeGroup = context.payload as MergeGroupChecksRequestedEvent;
  console.log(`Handling ${mergeGroup.action} event for merge group`);

  const {
    data: { commits },
  } = await octokit.rest.repos.compareCommitsWithBasehead({
    owner: mergeGroup.repository.owner.login,
    repo: mergeGroup.repository.name,
    basehead: `${mergeGroup.merge_group.base_sha}...${mergeGroup.merge_group.head_sha}`,
  });

  // Filter out merge commits (> 1 parent), and save each commit's message.
  const commitMessages = commits.reduce<{ sha: string; message: string }[]>(
    (accumulatedTitles, commit) => {
      const {
        sha,
        commit: { message },
        parents,
      } = commit;
      if (parents.length > 1) {
        return accumulatedTitles;
      }

      const commitMessage = titleOnly ? message.split("\n")[0] : message;

      accumulatedTitles.push({ sha, message: commitMessage });

      return accumulatedTitles;
    },
    [],
  );

  if (commitMessages.length === 0) {
    console.log("No commits to lint");
    return {
      valid: true,
      errorCount: 0,
      warningCount: 0,
      results: [],
    };
  }

  console.log("Commit messages to lint:");
  commitMessages.forEach(({ message }) => {
    console.log(`  - ${message}`);
  });

  return lint(commitMessages, config);
}

/**
 * Load the config file and return the loaded config object. Throws an error if
 * the config file is invalid.
 *
 * @param configPath - The path to the config file.
 * @returns The loaded config object.
 * @throws If the config file is invalid.
 */
export async function loadConfig(configPath: string): Promise<QualifiedConfig> {
  const parentDirOfThisFile = resolve(__dirname, "..");

  const extension = configPath.split(".").pop();

  if (extension === undefined) {
    throw new Error(
      `Couldn't determine file type of "${configPath}" because it has no extension`,
    );
  }

  await using file = await tmpFileAsync({
    template: `commitlint.config.XXXXXX.${extension}`,
    tmpdir: parentDirOfThisFile,
  });

  const { name, handle } = file;

  // Read the config file and write it to the temporary file
  const configPathContents = await fsPromises.readFile(configPath, "utf-8");

  await handle.writeFile(configPathContents);

  const f = await load(
    {},
    {
      file: name,
      cwd: parentDirOfThisFile,
    },
  );

  return f;
}

/**
 * Validate the commits or PR titles according to the event and the provided
 * rules.
 *
 * @param context - The event context containing the event type and payload.
 * @param octokit - The authenticated Octokit instance to access the GitHub API.
 * @param config - The commitlint configuration object.
 * @throws If the commit/pr title is invalid.
 */
export async function handleEvent(
  context: Context,
  octokit: Octokit,
  config: Config,
): Promise<LintResult> {
  console.log(`Handling event: ${context.eventName}`);

  switch (context.eventName) {
    case "pull_request_target":
    case "pull_request":
      return handlePullRequest(context.payload, config);
    case "merge_group":
      return handleMergeGroup(context, octokit, config);
    default:
      throw new WrongEventTypeError(context.eventName);
  }
}
