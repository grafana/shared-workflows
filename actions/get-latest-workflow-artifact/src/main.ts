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
  const considerInProgress = core.getInput("consider-inprogress") === "true";
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

  // Now let's get all workflows associated with this sha:
  // We start with in-progress ones if those are to be considered.
  let found = null;

  if (considerInProgress) {
    found = await getLatestArtifact(
      client,
      repoOwner,
      repoName,
      workflowId,
      headSha,
      "in_progress",
      artifactName,
    );
    core.setOutput("workflow-run-status", "in_progress");
  }

  if (!found) {
    found = await getLatestArtifact(
      client,
      repoOwner,
      repoName,
      workflowId,
      headSha,
      "completed",
      artifactName,
    );
    core.setOutput("workflow-run-status", "completed");
  }

  if (!found) {
    throw new Error(`No artifacts found with name ${artifactName}`);
  }
  const { artifact: foundArtifact, run } = found;
  const artifact = new DefaultArtifactClient();
  const response = await artifact.downloadArtifact(foundArtifact.id, {
    path: path,
    findBy: {
      repositoryName: repoName,
      repositoryOwner: repoOwner,
      workflowRunId: run.id,
      token: ghToken,
    },
  });
  core.setOutput("artifact-download-path", response.downloadPath);
  core.setOutput("artifact-id", foundArtifact.id);
  core.setOutput("workflow-run-id", run.id);
}

async function getLatestArtifact(
  client,
  repoOwner,
  repoName,
  workflowId,
  headSha: string,
  status: string,
  artifactName: string,
): Promise<{ artifact: any; run: any } | null> {
  const { data } = await client.rest.actions.listWorkflowRuns({
    repo: repoName,
    owner: repoOwner,
    head_sha: headSha,
    workflow_id: workflowId,
    status: status,
  });
  if (data.workflow_runs.length === 0) {
    console.log(`No ${status} runs found`);
    return null;
  }
  const run = data.workflow_runs[0];
  // Since these are pending workflows, the artifact might not be there yet. For this scenario, let's try this a couple of times:
  for (let attempt = 0; attempt < 5; attempt++) {
    const { data: artifacts } =
      await client.rest.actions.listWorkflowRunArtifacts({
        owner: repoOwner,
        repo: repoName,
        run_id: run.id,
      });
    const artifact = artifacts.artifacts.find(
      (art) => art.name == artifactName,
    );
    if (artifact) {
      console.log(`Found ${status} artifact`);
      return { artifact, run };
    }
    console.log("No artifact found, retrying in 10s");
    await new Promise((resolve) => {
      setTimeout(() => resolve(), 10000);
    });
  }
}
main();
