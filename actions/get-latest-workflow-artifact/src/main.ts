import { Temporal, toTemporalInstant } from "@js-temporal/polyfill";
Date.prototype.toTemporalInstant = toTemporalInstant;

import * as core from "@actions/core";
import { getOctokit } from "@actions/github";
import { DefaultArtifactClient } from "@actions/artifact";
import * as filepath from "node:path";

import { components } from "@octokit/openapi-types";

type WorkflowRunStatus = "in_progress" | "completed";
type Artifact = components["schemas"]["artifact"];
type GitHubClient = ReturnType<typeof getOctokit>;
type WorkflowRun = components["schemas"]["workflow-run"];
type IssueComment = components["schemas"]["issue-comment"];
type PullRequest = components["schemas"]["pull-request-simple"];
type GetLatestArtifactResponse = {
  artifact: Artifact;
  run: WorkflowRun;
};

async function main() {
  const ghToken = core.getInput("github-token", { required: true });
  const workflowId = core.getInput("workflow-id", { required: true });
  const [repoOwner, repoName] = core
    .getInput("repository", { required: true })
    .split("/");
  const prNumber = parseInt(core.getInput("pr-number", { required: true }), 10);
  const artifactName = core.getInput("artifact-name", { required: true });
  const considerInProgress = core.getInput("consider-inprogress") === "true";
  const considerComments = core.getInput("consider-comments") === "true";
  let path = core.getInput("path", { required: false });

  if (!path) {
    const workspace = process.env.GITHUB_WORKSPACE || process.cwd();
    if (!workspace) {
      throw new Error("GITHUB_WORKSPACE is not set");
    }
    path = workspace;
  }

  path = filepath.join(path, artifactName);

  const client = getOctokit(ghToken);

  const { data } = await client.rest.pulls.get({
    owner: repoOwner,
    repo: repoName,
    pull_number: prNumber,
  });

  const headSha = data.head.sha;

  // Now let's get all workflows associated with this sha:
  // We start with in-progress ones if those are to be considered.
  let found: GetLatestArtifactResponse | null = null;
  let foundForComment: GetLatestArtifactResponse | null = null;

  if (considerComments) {
    // If we need to consider workflow runs triggered by comments, then we
    // cannot filter by head_sha but need to really go through all the runs of
    // a particular workflow and hope it's still available through paging.
    // Otherwise, we can only fallback and potentially get an outdated
    // artifact:
    //
    // First we need to get a list of all the comments inside a PR but only
    // those that happened before the PR was merged:
    const mergedAt = data.merged_at;
    const comment = await getRelevantComment(
      client,
      repoOwner,
      repoName,
      prNumber,
      mergedAt,
    );
    if (comment) {
      // Now that we have a comment, we need to find workflows triggered by it.
      // The problem is, that there is no clear association between a workflow
      // run and a comment triggering it available through the API. For this we
      // need to look at workflows runs with the title of the PR (ideally at the
      // time of the comment) and filter for all workflow runs that were trigger
      // between [comment_time, comment_time+delta]. This is based on the
      // assumption, that the title of a PR is stable around the time the fetch
      // is made and the comment is created
      const workflowRun = await getWorkflowRunForComment(
        client,
        repoOwner,
        repoName,
        workflowId,
        data,
        comment,
        considerInProgress,
      );
      if (workflowRun) {
        foundForComment = await getWorkflowRunArtifact(
          client,
          repoOwner,
          repoName,
          workflowRun,
          artifactName,
        );
      }
    }
  }

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
  }

  if (
    found &&
    foundForComment &&
    found.run.created_at < foundForComment.run.created_at
  ) {
    found = foundForComment;
  }

  if (!found) {
    throw new Error(`No artifacts found with name ${artifactName}`);
  }

  core.setOutput("workflow-run-status", found.run.status);

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
  client: GitHubClient,
  repoOwner: string,
  repoName: string,
  workflowId: string,
  headSha?: string,
  status: WorkflowRunStatus,
  artifactName: string,
): Promise<GetLatestArtifactResponse | null> {
  const {
    data: { workflow_runs: workflowRuns },
  }: { data: { workflow_runs: WorkflowRun[] } } =
    await client.rest.actions.listWorkflowRuns({
      repo: repoName,
      owner: repoOwner,
      head_sha: headSha,
      workflow_id: workflowId,
      status: status,
    });

  if (workflowRuns.length === 0) {
    console.log(`No ${status} runs found`);
    return null;
  }
  const run = workflowRuns[0];
  return getWorkflowRunArtifact(client, repoOwner, repoName, run, artifactName);
}

async function getRelevantComment(
  client: GitHubClient,
  owner: string,
  repo: string,
  prNumber: number,
  mergedAt: string | null,
): Promise<IssueComment | null> {
  const comments: IssueComment[] = await client.paginate(
    client.rest.issues.listComments,
    {
      owner,
      repo,
      issue_number: prNumber,
      per_page: 100,
    },
  );
  comments.reverse();
  // Now drop all that are made *after* the PR was merged
  for (const comment of comments) {
    if (mergedAt && comment.created_at >= mergedAt) {
      continue;
    }
    return comment;
  }
  return null;
}

async function getWorkflowRunForComment(
  client: GitHubClient,
  owner: string,
  repo: string,
  workflowId: string,
  pr: PullRequest,
  comment: IssueComment,
  considerInProgress: boolean,
): Promise<WorkflowRun | null> {
  console.log(
    "Searching for workflows connected to comments. This can take a while…",
  );
  const commentCreatedAt = Temporal.Instant.from(comment.created_at);
  const runsIterator = client.paginate.iterator(
    client.rest.actions.listWorkflowRuns,
    {
      repo,
      owner,
      workflow_id: workflowId,
      per_page: 100,
      created: `${commentCreatedAt.toString()}..${commentCreatedAt.add({ minutes: 1 }).toString()}`,
    },
  );

  let abort = false;
  let page = 0;

  const allowedStatus = ["completed", "success"];
  if (considerInProgress) {
    allowedStatus.push("in_progress");
  }

  for await (const runs of runsIterator) {
    for (const run of runs.data) {
      if (run.event !== "issue_comment") {
        continue;
      }
      if (!allowedStatus.includes(run.status)) {
        continue;
      }
      if (run.display_title !== pr.title) {
        continue;
      }
      return run;
    }
    if (++page > 100) {
      console.log("Aborting search after 100 pages");
      abort = true;
    }
    if (abort) {
      break;
    }
  }
}

async function getWorkflowRunArtifact(
  client: GitHubClient,
  owner: string,
  repo: string,
  run: WorkflowRun,
  artifactName: string,
): Promise<GetLatestArtifactResponse | null> {
  const maxAttempts = run.status === "in_progress" ? 5 : 1;
  for (let attempt = 1; attempt <= maxAttempts; attempt++) {
    const {
      data: { artifacts },
    }: { data: { artifacts: Artifact[] } } =
      await client.rest.actions.listWorkflowRunArtifacts({
        owner,
        repo,
        run_id: run.id,
      });
    const artifact = artifacts.find((art) => art.name == artifactName);
    if (artifact) {
      console.log(`Found ${run.status} artifact`);
      return { artifact, run };
    }
    if (attempt < maxAttempts) {
      console.log("No artifact found, retrying in 10s");
      await new Promise((resolve) => setTimeout(resolve, 10_000));
    }
  }
  console.log("No artifact found");
  return null;
}

await main();
