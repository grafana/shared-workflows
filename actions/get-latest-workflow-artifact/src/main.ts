import * as core from "@actions/core";
import * as github from "@actions/github";
import { DefaultArtifactClient } from "@actions/artifact";
import * as filepath from "node:path";

async function main() {
  const ghToken = core.getInput("github-token", { required: true });
  const workflowId = core.getInput("workflow-id", { required: true });
  const [repoOwner, repoName] = core
    .getInput("repository", { required: true })
    .split("/");
  const prNumber = parseInt(core.getInput("pr-number", { required: true }), 10);
  const artifactName = core.getInput("artifact-name", { required: true });
  let path = core.getInput("path", { required: false });

  if (!path) {
    const workspace = process.env.GITHUB_WORKSPACE || process.cwd();
    if (!workspace) {
      throw new Error("GITHUB_WORKSPACE is not set");
    }
    path = workspace;
  }

  path = filepath.join(path, artifactName);

  const client = github.getOctokit(ghToken);

  const { data } = await client.rest.pulls.get({
    owner: repoOwner,
    repo: repoName,
    pull_number: prNumber,
  });

  const headSha = data.head.sha;
  // Now let's get all workflows associated with this sha.
  const { data: runs } = await client.rest.actions.listWorkflowRuns({
    repo: repoName,
    owner: repoOwner,
    head_sha: headSha,
    workflow_id: workflowId,
    status: "completed",
  });

  if (runs.workflow_runs.length <= 0) {
    throw new Error(`No workflow runs found for sha ${headSha}`);
  }

  const latestRun = runs.workflow_runs[0];
  core.info(`Latest run: ${latestRun.id} <${latestRun.html_url}>`);

  // Now that we have the run we can get a list of all artifacts there:
  const { data: artifacts } =
    await client.rest.actions.listWorkflowRunArtifacts({
      owner: repoOwner,
      repo: repoName,
      run_id: latestRun.id,
    });

  const artifact = new DefaultArtifactClient();
  const candidates = artifacts.artifacts.filter(
    (artifact) => artifact.name == artifactName,
  );
  if (candidates.length <= 0) {
    throw new Error(`No artifacts found with name ${artifactName}`);
  }

  const response = await artifact.downloadArtifact(candidates[0].id, {
    path: path,
    findBy: {
      repositoryName: repoName,
      repositoryOwner: repoOwner,
      workflowRunId: latestRun.id,
      token: ghToken,
    },
  });
  core.setOutput("artifact-download-path", response.downloadPath);
  core.setOutput("artifact-id", candidates[0].id);
  core.setOutput("workflow-run-id", latestRun.id);
}

main();
