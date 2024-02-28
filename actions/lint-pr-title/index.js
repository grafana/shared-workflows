const core = require('@actions/core');
const github = require('@actions/github');
const lint = require('@commitlint/lint');
const config = require('./commitlint.config.js');

async function run() {
  try {
    const octokit = github.getOctokit(process.env.GITHUB_TOKEN);

    const contextPullRequest = github.context.payload.pull_request;
    if (!contextPullRequest) {
      throw new Error("This action can only be invoked in `pull_request_target` or `pull_request` events. Otherwise the pull request can't be inferred.");
    }
    const { data: pullRequest } = await octokit.rest.pulls.get({
      owner: contextPullRequest.base.user.login,
      repo: contextPullRequest.base.repo.name,
      pull_number: contextPullRequest.number
    });

    core.setOutput('Pr title is ' + pullRequest.title +' and the rules to check are: ' + JOSN.stringify(config.rules))
    // const result = await lint(pullRequest.title, config.rules);
    // if (!result.valid) {
    //   const errorMessages = result.errors.map((error) => error.message);
    //   throw new Error(errorMessages.join('; '));
    // }
  } catch (error) {
    core.setFailed(error.message);
  }
};

run();
