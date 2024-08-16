import * as core from "@actions/core";
import * as fs from "fs/promises";
import * as github from "@actions/github";

import { join } from "path";
import { default as lint } from "@commitlint/lint";
import { validateConfig } from "@commitlint/config-validator";

/**
 * Format a list of strings for display: with commas and an "or" before the last
 * element.
 *
 * @param {string[]} strings - The list of strings to join.
 * @returns {string} The formatted string.
 */
function joinStringWithOr(strings) {
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

/**
 * Handle the pull request event. The title of the pull request is linted
 * against the provided rules. If the title is invalid, an error is thrown.
 *
 * @param {*} prContext - The pull request context.
 * @param {*} octokit - The authenticated Octokit instance.
 * @param {*} config - The configuration object.
 * @throws If the title is invalid.
 */
async function handlePullRequest(prContext, octokit, config) {
  console.log(`Handling ${prContext.action} event for PR #${prContext.number}`);

  const { data: pullRequest } = await octokit.rest.pulls.get({
    owner: prContext.base.user.login,
    repo: prContext.base.repo.name,
    pull_number: prContext.number,
  });

  const result = await lint(pullRequest.title, config.rules);

  if (!result.valid) {
    const errorMessages = result.errors.map((error) => error.message);
    throw new Error(errorMessages.join("; "));
  }
}

/**
 * Handle the merge group event. The titles of each commit from this pull
 * request in the merge group are linted against the provided rules. If any
 * title is invalid, an error is thrown.
 *
 * @param {*} context - The pull request context.
 * @param {*} octokit- The authenticated Octokit instance.
 * @param {*} config - The configuration object.
 * @throws If any title is invalid.
 */
async function handleMergeGroup(context, octokit, config) {
  console.log(
    `Handling ${context.event.action} event for merge group ${context.event.merge_group.id}`,
  );

  const opts =
    await github.rest.repos.compareCommitsWithBasehead.endpoint.merge({
      owner: context.repo.owner,
      repo: context.repo.repo,
      basehead: `${context.event.merge_group.base_ref}...${context.sha}`,
    });

  const {
    data: { commits },
  } = await octokit.paginate(opts);

  // Filter out merge commits (> 1 parent), loop over commits and save each
  // commit's title.
  const commitTitles = commits.reduce((titles, commit) => {
    if (commit.parents.length > 1) {
      return titles;
    }

    titles.push(commit.commit.message.split("\n")[0]);

    return titles;
  }, []);

  const errorMessages = [];
  for (const title of commitTitles) {
    const result = await lint(title, config.rules);

    if (!result.valid) {
      errorMessages.push(
        result.errors.map((error) => error.message).join("; "),
      );
    }
  }

  if (errorMessages.length > 0) {
    throw new Error(errorMessages.join("; "));
  }
}

async function run() {
  try {
    const octokit = github.getOctokit(process.env.GITHUB_TOKEN);

    const cwd = core.getInput("cwd") || process.cwd();
    const configPath =
      core.getInput("config-path") || join(cwd, "commitlint.config.js");

    if (!configPath) {
      throw new Error("Config path is required.");
    }

    try {
      await fs.access(configPath, fs.constants.R_OK);
    } catch (e) {
      throw new Error(`Config file not found or not readable: ${configPath}`);
    }

    const config = await import(configPath);

    // Throws an error if the config is invalid
    validateConfig(configPath, config.rules);

    switch (github.context.eventName) {
      case "pull_request_target":
      case "pull_request":
        const prContext = github.context.payload.pull_request;

        handlePullRequest(prContext, octokit, config.rules);
        break;
      case "merge_group":
        const mergeGroupContext = github.context;
        handleMergeGroup(mergeGroupContext, octokit, config.rules);
        break;
      default:
        throw new Error(
          `This action can only be run on ${joinStringWithOr(
            HANDLED_TARGETS.map((target) => `\`${target}\``),
          )} events.`,
        );
    }
  } catch (error) {
    core.setFailed(error.message);
  }
}

run();
