import "./polyfill-require-in-esm";
import "./polyfill-dirname-filename";

import type {
  MergeGroupChecksRequestedEvent,
  PullRequestEvent,
} from "@octokit/webhooks-types";
import { QualifiedConfig, QualifiedRules } from "@commitlint/types";
import { basename, join, resolve } from "path";
import { constants, promises } from "fs";
import { context, getOctokit } from "@actions/github";
import { getInput, setFailed } from "@actions/core";

import { default as lint } from "@commitlint/lint";
import load from "@commitlint/load";

type Context = typeof context;
type Octokit = ReturnType<typeof getOctokit>;
type WebHookPayload = Context["payload"];

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

const HANDLED_TARGETS = ["merge_group", "pull_request_target", "pull_request"];

class LintError extends Error {}

/**
 * Lint the commit or PR title against the provided rules. If the title is
 * invalid, an error is thrown.
 *
 * @param title - The commit or PR title to lint.
 * @param rules - The rules to lint against.
 * @throws {LintError} If the title is invalid.
 */
async function lintTitle(title: string, rules: QualifiedRules) {
  const result = await lint(title, rules);

  if (!result.valid) {
    const errorMessages = result.errors.map((error) => error.message);

    throw new LintError(`title: ${title}. errors: ${errorMessages.join("; ")}`);
  }
}

/**
 * Handle the pull request event. The title of the pull request is linted
 * against the provided rules. If the title is invalid, an error is thrown.
 *
 * @param prCtx - The pull request context.
 * @param octokit - The authenticated Octokit instance.
 * @param config - The configuration object.
 * @throws If the title is invalid.
 */
async function handlePullRequest(
  prCtx: WebHookPayload,
  octokit: Octokit,
  config: QualifiedConfig,
) {
  const prContext = prCtx as PullRequestEvent;

  console.log(
    `Handling ${prContext.action} event for PR #${prContext.number.toString()}`,
  );

  const { data: pullRequest } = await octokit.rest.pulls.get({
    owner: prContext.pull_request.base.user.login,
    repo: prContext.pull_request.base.repo.name,
    pull_number: prContext.number,
  });

  await lintTitle(pullRequest.title, config.rules);
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
  config: QualifiedConfig,
) {
  const mergeGroup = context.payload as MergeGroupChecksRequestedEvent;
  console.log(`Handling ${mergeGroup.action} event for merge group`);

  const comparedCommits = await octokit.rest.repos.compareCommitsWithBasehead({
    owner: mergeGroup.repository.owner.login,
    repo: mergeGroup.repository.name,
    basehead: `${mergeGroup.merge_group.base_sha}...${mergeGroup.merge_group.head_sha}`,
  });

  // Filter out merge commits (> 1 parent), and save each commit's title.
  const commitTitles = comparedCommits.data.commits.reduce<string[]>(
    (titles, commit) => {
      if (commit.parents.length > 1) {
        return titles;
      }

      titles.push(commit.commit.message.split("\n")[0]);

      return titles;
    },
    [],
  );

  if (commitTitles.length === 0) {
    console.log("No commits to lint");
    return;
  }

  console.log("Commit titles to lint:");
  commitTitles.forEach((title) => {
    console.log(`  - ${title}`);
  });

  const errorMessages = [];
  for (const title of commitTitles) {
    try {
      await lintTitle(title, config.rules);
    } catch (error) {
      if (error instanceof LintError) {
        errorMessages.push(error.message);

        continue;
      }

      throw error;
    }
  }

  if (errorMessages.length > 0) {
    throw new LintError(errorMessages.join("; "));
  }
}

async function run() {
  try {
    if (process.env.GITHUB_TOKEN === undefined) {
      throw new Error("GITHUB_TOKEN is required.");
    }

    const octokit = getOctokit(process.env.GITHUB_TOKEN);

    const configPath = resolve(
      getInput("config-path") || "commitlint.config.js",
    );

    const destConfig = join(import.meta.dirname, basename(configPath));

    await promises.copyFile(configPath, destConfig);

    try {
      await promises.access(configPath, constants.R_OK);
    } catch {
      throw new Error(`Config file ${configPath} not found or not readable`);
    }

    console.log(`Loading config from ${configPath}`);

    // Throws an error if the config is invalid
    const config = await load(
      {},
      {
        file: destConfig,
      },
    );

    switch (context.eventName) {
      case "pull_request_target":
      case "pull_request":
        await handlePullRequest(context.payload, octokit, config);
        break;
      case "merge_group":
        await handleMergeGroup(context, octokit, config);
        break;
      default:
        throw new Error(
          `This action can only be run on ${joinStringWithOr(
            HANDLED_TARGETS.map((target) => `\`${target}\``),
          )} events.`,
        );
    }
  } catch (error) {
    if (error instanceof Error) {
      setFailed(error.message);
    }
  }
}

await run();
