const prActions = new Set([
  "opened",
  "reopened",
  "edited",
  "labeled",
  "unlabeled",
]);
const overridePermissions = new Set(["write", "maintain", "admin"]);

/** Metadata-only routing. Never execute PR code or interpolate comments into scripts. */
export async function retarget({
  github,
  context,
  sourceBranch = "main",
  targetBranch = "dev",
  overrideLabel = "allow-main",
}) {
  for (const value of [sourceBranch, targetBranch, overrideLabel]) {
    if (!value || value !== value.trim()) {
      throw new Error(
        "Branches and override-label must be nonempty and trimmed.",
      );
    }
  }
  if (sourceBranch === targetBranch) {
    throw new Error("source-branch and target-branch must be different.");
  }

  const ignored = { changed: false };
  const { payload, eventName, repo } = context;
  let number;
  let command;
  if (eventName === "issue_comment") {
    if (payload.action !== "created" || !payload.issue?.pull_request) {
      return ignored;
    }
    const body = payload.comment?.body?.trim();
    if (body === `/target ${sourceBranch}`) command = sourceBranch;
    else if (body === `/target ${targetBranch}`) command = targetBranch;
    else return ignored;
    number = payload.issue.number;

    // Check current repository permissions, not author_association or PR authorship.
    const { data } = await github.rest.repos.getCollaboratorPermissionLevel({
      ...repo,
      username: payload.comment.user.login,
    });
    if (!overridePermissions.has(data.permission)) return ignored;
  } else if (
    eventName === "pull_request_target" &&
    prActions.has(payload.action) &&
    payload.pull_request
  ) {
    number = payload.pull_request.number;
  } else {
    return ignored;
  }

  // Event snapshots can be stale (e.g. a comment restores main while an older
  // opened event is queued). Always use the current base, labels and state.
  const { data: pr } = await github.rest.pulls.get({
    ...repo,
    pull_number: number,
  });
  const unchanged = { changed: false, baseBranch: pr.base.ref };
  if (pr.state !== "open") return unchanged;
  if (![sourceBranch, targetBranch].includes(pr.base.ref)) return unchanged;

  const hasOverride = pr.labels.some((label) => label.name === overrideLabel);
  if (!command) {
    if (pr.base.ref !== sourceBranch || hasOverride) return unchanged;
    // Do not redirect dev → main rollups into a PR against their own branch.
    // A fork's branch named dev must NOT inherit this exemption.
    if (
      pr.head.ref === targetBranch &&
      pr.head.repo?.full_name?.toLowerCase() ===
        pr.base.repo.full_name.toLowerCase()
    ) {
      return unchanged;
    }
  }

  const destination = command ?? targetBranch;
  const changeBase = pr.base.ref !== destination;
  if (changeBase) {
    // Fail before altering labels if the destination does not exist.
    await github.rest.repos.getBranch({ ...repo, branch: destination });
  }

  if (command === sourceBranch && !hasOverride) {
    // Persist the exception BEFORE changing base, so the resulting edited event
    // cannot bounce the PR back to the development branch.
    // Require the configured label to exist; GitHub can silently ignore unknown
    // labels when adding them. Provision it once in each caller repository.
    await github.rest.issues.getLabel({ ...repo, name: overrideLabel });
    await github.rest.issues.addLabels({
      ...repo,
      issue_number: number,
      labels: [overrideLabel],
    });
  } else if (command === targetBranch && hasOverride) {
    await github.rest.issues.removeLabel({
      ...repo,
      issue_number: number,
      name: overrideLabel,
    });
  }

  if (!changeBase) return unchanged;
  await github.rest.pulls.update({
    ...repo,
    pull_number: number,
    base: destination,
  });
  await github.rest.issues.createComment({
    ...repo,
    issue_number: number,
    body: command
      ? `Changed the base branch to \`${destination}\` at a maintainer's request. The \`${overrideLabel}\` override ${command === sourceBranch ? "is enabled" : "is disabled"}.`
      : `This repository routes feature PRs to \`${targetBranch}\`, so I changed the base from \`${sourceBranch}\` to \`${targetBranch}\`. A maintainer with write access can comment \`/target ${sourceBranch}\` to restore \`${sourceBranch}\` and persist the \`${overrideLabel}\` override, or \`/target ${targetBranch}\` to remove it. Same-repository \`${targetBranch} → ${sourceBranch}\` rollups are exempt.`,
  });
  return { changed: true, baseBranch: destination };
}
