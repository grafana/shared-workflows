const core = require("@actions/core");
const github = require("@actions/github");
const lint = require("@commitlint/lint").default;

async function run() {
  try {
    const octokit = github.getOctokit(process.env.GITHUB_TOKEN);
    const eventName = github.context.eventName;
    let pullRequest;

    if (eventName === "pull_request" || eventName === "pull_request_target") {
      const contextPullRequest = github.context.payload.pull_request;
      if (!contextPullRequest) {
        throw new Error(
          "This action can only be invoked in `pull_request_target` or `pull_request` events. Otherwise, the pull request can't be inferred.",
        );
      }
      const { data } = await octokit.rest.pulls.get({
        owner: contextPullRequest.base.user.login,
        repo: contextPullRequest.base.repo.name,
        pull_number: contextPullRequest.number,
      });
      pullRequest = data;
    } else if (eventName === "merge_group") {
      const mergeGroupContext = github.context.payload.merge_group;
      if (!mergeGroupContext) {
        throw new Error(
          "This action can only be invoked in `merge_group` events. Otherwise, the merge group can't be inferred.",
        );
      }
      print(mergeGroupContext)
      const { data } = await octokit.rest.pulls.get({
        owner: mergeGroupContext.base.user.login,
        repo: mergeGroupContext.base.repo.name,
        pull_number: mergeGroupContext.head_sha,
      });
      pullRequest = data;
    } else {
      throw new Error(
        "This action can only be invoked in `pull_request_target`, `pull_request`, or `merge_group` events.",
      );
    }

    const configPath = core.getInput("config-path");
    const config = configPath
      ? require(configPath)
      : require("./commitlint.config.js");
    const result = await lint(pullRequest.title, config.rules);
    if (!result.valid) {
      const errorMessages = result.errors.map((error) => error.message);
      throw new Error(errorMessages.join("; "));
    }
  } catch (error) {
    core.setFailed(error.message);
  }
}

run();
