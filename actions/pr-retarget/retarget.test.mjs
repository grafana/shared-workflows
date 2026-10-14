import assert from "node:assert/strict";
import { test } from "node:test";
import { retarget, dispatchAutomatic } from "./retarget.mjs";

function fixture({
  base = "main",
  head = "feature",
  headRepo = "grafana/example",
  labels = [],
  state = "open",
  eventName = "pull_request_target",
  action = eventName === "issue_comment" ? "created" : "opened",
  body = "/target main",
  permission = "write",
  sender = { login: "maintainer", type: "User" },
  addedLabel = null,
} = {}) {
  const calls = [];
  const pr = {
    state,
    base: { ref: base, repo: { full_name: "grafana/example" } },
    head: { ref: head, repo: headRepo ? { full_name: headRepo } : null },
    labels: (addedLabel && labels.length === 0 ? [addedLabel] : labels).map(
      (name) => ({ name }),
    ),
  };
  const api =
    (method, data = {}) =>
    async (args) => {
      calls.push({ method, args });
      return { data };
    };
  const github = {
    rest: {
      actions: { createWorkflowDispatch: api("dispatch") },
      repos: {
        getCollaboratorPermissionLevel: api("permission", { permission }),
        getBranch: api("branch"),
      },
      pulls: { get: api("get", pr), update: api("update") },
      issues: {
        getLabel: api("label"),
        addLabels: api("add"),
        removeLabel: api("remove"),
        createComment: api("comment"),
      },
    },
  };
  const context = {
    eventName,
    repo: { owner: "grafana", repo: "example" },
    payload: {
      action,
      pull_request: { number: 42, base: pr.base, head: pr.head, labels: [] },
      issue: { number: 42, pull_request: {} },
      comment: { body, user: { login: "maintainer" } },
      ...(sender && { sender }),
      ...(addedLabel && { label: { name: addedLabel } }),
    },
  };
  return {
    github,
    context,
    pr,
    calls,
    run: (config = {}) => retarget({ github, context, ...config }),
    methods: () => calls.map(({ method }) => method),
  };
}

for (const action of ["opened", "reopened", "edited", "labeled", "unlabeled"]) {
  test(`redirects main to dev on ${action}`, async () => {
    const f = fixture({ action });
    assert.deepEqual(await f.run(), { changed: true, baseBranch: "dev" });
    assert.deepEqual(f.methods(), ["get", "branch", "update", "comment"]);
    assert.deepEqual(f.calls[2].args, {
      owner: "grafana",
      repo: "example",
      pull_number: 42,
      base: "dev",
    });
    assert.match(f.calls[3].args.body, /\/target main/);
  });
}

for (const options of [
  { base: "dev" },
  { base: "release" },
  { labels: ["allow-main"] },
  { state: "closed" },
  { head: "dev" },
  { head: "dev", headRepo: "GRAFANA/EXAMPLE" },
]) {
  test(`does not mutate exempt PR: ${JSON.stringify(options)}`, async () => {
    const f = fixture(options);
    assert.deepEqual(await f.run(), {
      changed: false,
      baseBranch: options.base ?? "main",
    });
    assert.deepEqual(f.methods(), ["get"]);
  });
}

test("configured same-repository release branches are exempt", async () => {
  const f = fixture({ head: "release/2026-10-13" });
  assert.deepEqual(await f.run({ exemptHeadPrefix: "release/" }), {
    changed: false,
    baseBranch: "main",
  });
  assert.deepEqual(f.methods(), ["get"]);
});

test("release exemption does not apply to fork branches", async () => {
  const f = fixture({
    head: "release/2026-10-13",
    headRepo: "outsider/example",
  });
  assert.equal((await f.run({ exemptHeadPrefix: "release/" })).changed, true);
});

test("release exemption is disabled by default", async () => {
  const f = fixture({ head: "release/2026-10-13" });
  assert.equal((await f.run()).changed, true);
});

for (const headRepo of ["outsider/example", null]) {
  test(`does not exempt dev branch with head repository ${headRepo}`, async () => {
    const f = fixture({ head: "dev", headRepo });
    assert.equal((await f.run()).changed, true);
  });
}

for (const permission of ["write", "maintain", "admin"]) {
  test(`${permission} can restore main and persist override before base update`, async () => {
    const f = fixture({ eventName: "issue_comment", base: "dev", permission });
    assert.deepEqual(await f.run(), { changed: true, baseBranch: "main" });
    assert.deepEqual(f.methods(), [
      "permission",
      "get",
      "branch",
      "label",
      "add",
      "update",
      "comment",
    ]);
    assert.deepEqual(f.calls[4].args.labels, ["allow-main"]);
  });
}

for (const permission of ["read", "triage", "none"]) {
  test(`${permission} cannot override routing even as PR author`, async () => {
    const f = fixture({ eventName: "issue_comment", permission });
    f.context.payload.issue.user = { login: "maintainer" };
    f.context.payload.comment.author_association = "MEMBER";
    assert.deepEqual(await f.run(), { changed: false });
    assert.deepEqual(f.methods(), ["permission"]);
  });
}

test("/target dev removes override before retargeting", async () => {
  const f = fixture({
    eventName: "issue_comment",
    body: "/target dev",
    labels: ["allow-main"],
  });
  assert.deepEqual(await f.run(), { changed: true, baseBranch: "dev" });
  assert.deepEqual(f.methods(), [
    "permission",
    "get",
    "branch",
    "remove",
    "update",
    "comment",
  ]);
});

test("/target main on main persists override without changing base", async () => {
  const f = fixture({ eventName: "issue_comment" });
  assert.deepEqual(await f.run(), { changed: false, baseBranch: "main" });
  assert.deepEqual(f.methods(), ["permission", "get", "label", "add"]);
});

test("/target dev on dev removes label without changing base", async () => {
  const f = fixture({
    eventName: "issue_comment",
    base: "dev",
    body: "/target dev",
    labels: ["allow-main"],
  });
  assert.deepEqual(await f.run(), { changed: false, baseBranch: "dev" });
  assert.deepEqual(f.methods(), ["permission", "get", "remove"]);
});

test("repeating an override is idempotent", async () => {
  const f = fixture({ eventName: "issue_comment", labels: ["allow-main"] });
  await f.run();
  assert.deepEqual(f.methods(), ["permission", "get"]);
});

test("queued opened event uses current labels instead of stale snapshot", async () => {
  const f = fixture({ labels: ["allow-main"] });
  assert.deepEqual(f.context.payload.pull_request.labels, []);
  assert.equal((await f.run()).changed, false);
  assert.deepEqual(f.methods(), ["get"]);
});

for (const body of [
  "hello",
  "/target release",
  "Please /target main",
  "/target main\nmore text",
  "/target main; echo hacked",
  "/TARGET main",
]) {
  test(`ignores non-command comment ${JSON.stringify(body)}`, async () => {
    const f = fixture({ eventName: "issue_comment", body });
    assert.deepEqual(await f.run(), { changed: false });
    assert.deepEqual(f.methods(), []);
  });
}

test("accepts whitespace around an exact command", async () => {
  const f = fixture({
    eventName: "issue_comment",
    body: " \n/target main\r\n",
  });
  const run = await f.run();
  assert.ok(f.methods().includes("add"));
  return run;
});

// --- enforcement of override-label provenance ---
for (const permission of ["triage", "read", "none"]) {
  test(`${permission} cannot self-grant the override label`, async () => {
    const f = fixture({
      action: "labeled",
      addedLabel: "allow-main",
      sender: { login: "helper", type: "User" },
      permission,
    });
    assert.deepEqual(await f.run(), { changed: false, baseBranch: "main" });
    assert.deepEqual(f.methods(), ["get", "permission", "remove", "comment"]);
    assert.deepEqual(f.calls[2].args, {
      owner: "grafana",
      repo: "example",
      issue_number: 42,
      name: "allow-main",
    });
    assert.match(f.calls[3].args.body, /write access/);
  });
}

for (const permission of ["write", "maintain", "admin"]) {
  test(`${permission} can add the override label directly`, async () => {
    const f = fixture({
      action: "labeled",
      addedLabel: "allow-main",
      sender: { login: "maintainer", type: "User" },
      permission,
    });
    assert.deepEqual(await f.run(), { changed: false, baseBranch: "main" });
    assert.deepEqual(f.methods(), ["get", "permission"]);
  });
}

test("only the configured routing App bypasses the permission lookup", async () => {
  const f = fixture({
    action: "labeled",
    addedLabel: "allow-main",
    sender: { login: "routing-app[bot]", type: "Bot" },
  });
  assert.deepEqual(await f.run({ trustedBot: "routing-app[bot]" }), {
    changed: false,
    baseBranch: "main",
  });
  assert.deepEqual(f.methods(), ["get"]);
});

for (const trustedBot of ["", "routing-app[bot]"]) {
  test(`other bots cannot self-grant overrides (trustedBot=${trustedBot})`, async () => {
    const f = fixture({
      action: "labeled",
      addedLabel: "allow-main",
      sender: { login: "other-app[bot]", type: "Bot" },
      permission: "none",
    });
    await f.run({ trustedBot });
    assert.deepEqual(f.methods(), ["get", "permission", "remove", "comment"]);
  });
}

test("a matching login is not trusted without GitHub's Bot actor type", async () => {
  const f = fixture({
    action: "labeled",
    addedLabel: "allow-main",
    sender: { login: "routing-app[bot]", type: "User" },
    permission: "none",
  });
  await f.run({ trustedBot: "routing-app[bot]" });
  assert.deepEqual(f.methods(), ["get", "permission", "remove", "comment"]);
});

test("stale label event is harmless when the label is already gone", async () => {
  const f = fixture({
    action: "labeled",
    addedLabel: "allow-main",
    sender: { login: "helper", type: "User" },
    permission: "triage",
  });
  f.pr.labels = [];
  await f.run();
  assert.deepEqual(f.methods(), ["get"]);
});

test("adding an unrelated label is not enforcement", async () => {
  const f = fixture({
    action: "labeled",
    addedLabel: "needs-review",
    sender: { login: "helper", type: "User" },
    permission: "triage",
  });
  assert.deepEqual(await f.run(), { changed: true, baseBranch: "dev" });
  assert.deepEqual(f.methods(), ["get", "branch", "update", "comment"]);
});

test("override-label enforcement is skipped on closed PRs and foreign bases", async () => {
  for (const options of [{ state: "closed" }, { base: "release" }]) {
    const f = fixture({
      action: "labeled",
      addedLabel: "allow-main",
      sender: { login: "helper", type: "User" },
      permission: "triage",
      ...options,
    });
    assert.deepEqual(await f.run(), {
      changed: false,
      baseBranch: options.base ?? "main",
    });
    assert.deepEqual(f.methods(), ["get"]);
  }
});

for (const options of [
  { eventName: "push" },
  { eventName: "pull_request" },
  { action: "closed" },
  { eventName: "issue_comment", action: "edited" },
]) {
  test(`ignores unsupported event ${JSON.stringify(options)}`, async () => {
    const f = fixture(options);
    assert.deepEqual(await f.run(), { changed: false });
    assert.deepEqual(f.methods(), []);
  });
}

test("ignores issue comments that are not on a PR", async () => {
  const f = fixture({ eventName: "issue_comment" });
  delete f.context.payload.issue.pull_request;
  assert.deepEqual(await f.run(), { changed: false });
  assert.deepEqual(f.methods(), []);
});

for (const options of [{ state: "closed" }, { base: "release" }]) {
  test(`override cannot change unrelated or closed PR: ${JSON.stringify(options)}`, async () => {
    const f = fixture({ eventName: "issue_comment", ...options });
    assert.equal((await f.run()).changed, false);
    assert.deepEqual(f.methods(), ["permission", "get"]);
  });
}

test("supports configured branch names and labels", async () => {
  const f = fixture({
    eventName: "issue_comment",
    base: "develop",
    body: "/target trunk",
  });
  assert.deepEqual(
    await f.run({
      sourceBranch: "trunk",
      targetBranch: "develop",
      overrideLabel: "allow-trunk",
    }),
    { changed: true, baseBranch: "trunk" },
  );
  assert.deepEqual(f.calls[4].args.labels, ["allow-trunk"]);
});

test("missing destination fails before label or base mutations", async () => {
  const f = fixture({ eventName: "issue_comment", base: "dev" });
  f.github.rest.repos.getBranch = async () => {
    throw new Error("Branch not found");
  };
  await assert.rejects(f.run(), /Branch not found/);
  assert.deepEqual(f.methods(), ["permission", "get"]);
});

test("missing override label fails before restoring main", async () => {
  const f = fixture({ eventName: "issue_comment", base: "dev" });
  f.github.rest.issues.getLabel = async () => {
    throw new Error("Label not found");
  };
  await assert.rejects(f.run(), /Label not found/);
  assert.deepEqual(f.methods(), ["permission", "get", "branch"]);
});

test("permission lookup errors fail closed", async () => {
  const f = fixture({ eventName: "issue_comment" });
  f.github.rest.repos.getCollaboratorPermissionLevel = async () => {
    throw new Error("Forbidden");
  };
  await assert.rejects(f.run(), /Forbidden/);
  assert.deepEqual(f.methods(), []);
});

test("failed base update propagates and does not post success", async () => {
  const f = fixture();
  f.github.rest.pulls.update = async () => {
    throw new Error("Cannot change base");
  };
  await assert.rejects(f.run(), /Cannot change base/);
  assert.deepEqual(f.methods(), ["get", "branch"]);
});

for (const config of [
  { sourceBranch: "dev" },
  { sourceBranch: "" },
  { targetBranch: " dev" },
  { overrideLabel: "" },
]) {
  test(`rejects invalid configuration ${JSON.stringify(config)}`, async () => {
    const f = fixture();
    await assert.rejects(f.run(config));
    assert.deepEqual(f.methods(), []);
  });
}

function dispatchContext(f) {
  const { action, sender, label } = f.context.payload;
  f.context.eventName = "workflow_dispatch";
  f.context.payload = {
    inputs: {
      "pr-number": "42",
      "routing-event": JSON.stringify({
        action,
        sender,
        ...(label && { label }),
      }),
    },
  };
}

for (const action of ["opened", "reopened", "edited", "labeled", "unlabeled"]) {
  test(`tokenless ${action} dispatcher queues metadata on main`, async () => {
    const f = fixture({
      action,
      addedLabel: ["labeled", "unlabeled"].includes(action)
        ? "needs-review"
        : null,
    });
    await dispatchAutomatic({ github: f.github, context: f.context });
    assert.deepEqual(f.methods(), ["dispatch"]);
    const args = f.calls[0].args;
    assert.equal(args.workflow_id, "pr-retarget.yml");
    assert.equal(args.ref, "main");
    assert.equal(args.inputs["pr-number"], "42");
    const metadata = JSON.parse(args.inputs["routing-event"]);
    assert.equal(metadata.action, action);
    assert.equal(metadata.sender.login, "maintainer");
    assert.equal(metadata.comment, undefined);
    assert.equal(metadata.command, undefined);
    assert.equal(metadata.pull_request, undefined);
  });

  test(`dispatched ${action} routes using fresh PR state`, async () => {
    const f = fixture({
      action,
      addedLabel: ["labeled", "unlabeled"].includes(action)
        ? "needs-review"
        : null,
    });
    dispatchContext(f);
    assert.deepEqual(await f.run(), { changed: true, baseBranch: "dev" });
    assert.deepEqual(f.methods(), ["get", "branch", "update", "comment"]);
  });
}

for (const options of [
  { base: "dev", head: "main" },
  { base: "dev", head: "backmerge/main-after-release" },
  { base: "release" },
  { head: "dev" },
  { head: "release/2026-10-14" },
  { eventName: "issue_comment" },
  { action: "closed" },
]) {
  test(`dispatcher skips irrelevant/exempt event ${JSON.stringify(options)}`, async () => {
    const f = fixture(options);
    await dispatchAutomatic({
      github: f.github,
      context: f.context,
      exemptHeadPrefix: "release/",
    });
    assert.deepEqual(f.methods(), []);
  });
}

test("dispatcher does not exempt fork release branches", async () => {
  const f = fixture({
    head: "release/2026-10-14",
    headRepo: "outsider/example",
  });
  await dispatchAutomatic({
    github: f.github,
    context: f.context,
    exemptHeadPrefix: "release/",
  });
  assert.deepEqual(f.methods(), ["dispatch"]);
});

test("override-label enforcement is dispatched even on dev", async () => {
  const f = fixture({
    base: "dev",
    action: "labeled",
    addedLabel: "allow-main",
    permission: "triage",
    sender: { login: "helper", type: "User" },
  });
  await dispatchAutomatic({ github: f.github, context: f.context });
  assert.deepEqual(f.methods(), ["dispatch"]);
  dispatchContext(f);
  await f.run();
  assert.deepEqual(f.methods(), [
    "dispatch",
    "get",
    "permission",
    "remove",
    "comment",
  ]);
  assert.equal(f.calls[2].args.username, "helper");
});

test("dispatched routing never overrides a freshly restored main exception", async () => {
  const f = fixture({ labels: ["allow-main"] });
  dispatchContext(f);
  await f.run();
  assert.deepEqual(f.methods(), ["get"]);
});

test("dispatched trusted App label preserves the explicitly configured bot exemption", async () => {
  const f = fixture({
    action: "labeled",
    addedLabel: "allow-main",
    sender: { login: "routing-app[bot]", type: "Bot" },
  });
  dispatchContext(f);
  await f.run({ trustedBot: "routing-app[bot]" });
  assert.deepEqual(f.methods(), ["get"]);
});

test("empty manual dispatch is a no-op for the action", async () => {
  const f = fixture({ eventName: "workflow_dispatch" });
  f.context.payload.inputs = { "pr-number": "", "routing-event": "" };
  assert.deepEqual(await f.run(), { changed: false });
  assert.deepEqual(f.methods(), []);
});

const validAutomaticMetadata = JSON.stringify({
  action: "opened",
  sender: { login: "maintainer", type: "User" },
});
for (const inputs of [
  { "pr-number": 42, "routing-event": validAutomaticMetadata },
  { "pr-number": "0", "routing-event": validAutomaticMetadata },
  { "pr-number": "-1", "routing-event": validAutomaticMetadata },
  { "pr-number": "42;echo hacked", "routing-event": validAutomaticMetadata },
  { "pr-number": "9007199254740993", "routing-event": validAutomaticMetadata },
  { "pr-number": "42" },
  { "routing-event": validAutomaticMetadata },
  { "pr-number": "42", "routing-event": "not-json" },
  { "pr-number": "42", "routing-event": "null" },
  { "pr-number": "42", "routing-event": "[]" },
  { "pr-number": "42", "routing-event": JSON.stringify({ action: "opened" }) },
  {
    "pr-number": "42",
    "routing-event": JSON.stringify({
      action: "issue_comment",
      sender: { login: "maintainer", type: "User" },
    }),
  },
  {
    "pr-number": "42",
    "routing-event": JSON.stringify({
      action: "opened",
      sender: { login: "maintainer", type: "User" },
      command: "/target main",
    }),
  },
  {
    "pr-number": "42",
    "routing-event": JSON.stringify({
      action: "opened",
      sender: { login: "maintainer", type: "User" },
      comment: { body: "/target main" },
    }),
  },
  {
    "pr-number": "42",
    "routing-event": JSON.stringify({
      action: "labeled",
      sender: { login: "maintainer", type: "User" },
    }),
  },
  {
    "pr-number": "42",
    "routing-event": JSON.stringify({
      action: "opened",
      sender: { login: "maintainer", type: "Organization" },
    }),
  },
]) {
  test(`invalid dispatch fails before GitHub mutations: ${JSON.stringify(inputs)}`, async () => {
    const f = fixture({ eventName: "workflow_dispatch" });
    f.context.payload.inputs = inputs;
    await assert.rejects(f.run(), /Dispatch/);
    assert.deepEqual(f.methods(), []);
  });
}

test("dispatch errors propagate without any PR writes", async () => {
  const f = fixture();
  f.github.rest.actions.createWorkflowDispatch = async () => {
    throw new Error("Dispatch forbidden");
  };
  await assert.rejects(
    dispatchAutomatic({ github: f.github, context: f.context }),
    /Dispatch forbidden/,
  );
  assert.deepEqual(f.methods(), []);
});
