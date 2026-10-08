import assert from "node:assert/strict";
import { test } from "node:test";
import { retarget } from "./retarget.mjs";

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
} = {}) {
  const calls = [];
  const pr = {
    state,
    base: { ref: base, repo: { full_name: "grafana/example" } },
    head: { ref: head, repo: headRepo ? { full_name: headRepo } : null },
    labels: labels.map((name) => ({ name })),
  };
  const api =
    (method, data = {}) =>
    async (args) => {
      calls.push({ method, args });
      return { data };
    };
  const github = {
    rest: {
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
      pull_request: { number: 42, base: { ref: "main" }, labels: [] },
      issue: { number: 42, pull_request: {} },
      comment: { body, user: { login: "maintainer" } },
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
  await f.run();
  assert.ok(f.methods().includes("add"));
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
