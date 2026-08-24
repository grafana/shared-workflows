import { LintError, LintResult, handleEvent, loadConfig } from "./main";
import { access, constants } from "fs/promises";
import { context, getOctokit } from "@actions/github";
import { getInput, setFailed } from "@actions/core";

import { resolve } from "path";

/**
 * All code to interact with the environment is in this file.
 */

/**
 * Run the action. The action runs in the context of a GitHub event and will
 * load a configuration file and then lint the commit or PR title against the
 * provided rules.
 *
 * @throws If the configuration file is not found or not readable, or if the
 * commit or PR title is invalid.
 */
export async function run(): Promise<LintResult> {
  const configPath = resolve(getInput("config_path") || "commitlint.config.js");
  const titleOnly = getInput("title_only") === "true";

  const gitHubToken = process.env.GITHUB_TOKEN;
  if (gitHubToken === undefined) {
    throw new Error("GITHUB_TOKEN is required.");
  }

  const actionConfig = {
    configPath,
    titleOnly,
  };

  // Throws an error if the config is invalid
  try {
    await access(configPath, constants.R_OK);
  } catch {
    throw new Error(`Config file ${configPath} not found or not readable`);
  }
  console.log(`Loading config from ${configPath}`);
  const commitLintConfig = await loadConfig(configPath);

  console.log(`Will lint ${titleOnly ? "titles only" : "titles and bodies"}`);

  return handleEvent(context, getOctokit(gitHubToken), {
    actionConfig,
    commitLintConfig,
  });
}

export async function main() {
  try {
    const result = await run();
    if (!result.valid) {
      throw new LintError(result);
    }
  } catch (error) {
    if (error instanceof Error) {
      setFailed(error.message);
      return;
    }
    throw error;
  }
}

if (import.meta.main) {
  await main();
}
