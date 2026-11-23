import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import {
  mkdirSync,
  mkdtempSync,
  rmSync,
  symlinkSync,
  writeFileSync,
} from "node:fs";
import { tmpdir } from "node:os";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import { test } from "node:test";
import { CONFIG_FILES, inspectRepository, violations } from "./check.mjs";

function repository(t, files = {}) {
  const root = mkdtempSync(join(tmpdir(), "renovate-policy-"));
  t.after(() => rmSync(root, { recursive: true, force: true }));
  for (const [name, contents] of Object.entries(files)) {
    mkdirSync(dirname(join(root, name)), { recursive: true });
    writeFileSync(join(root, name), contents);
  }
  return root;
}

test("detects each explicit repository-wide disablement", () => {
  for (const config of [
    { enabled: false },
    { force: { enabled: false } },
    { extends: ["config:recommended", ":disableRenovate"] },
  ]) {
    assert.equal(violations(config).length, 1);
  }
});

test("allows defaults, scoped exclusions and unresolved presets", () => {
  for (const config of [
    {},
    { enabled: true },
    { extends: ["local>org/preset"] },
    { major: { enabled: false } },
    { npm: { enabled: false } },
    { packageRules: [{ matchManagers: ["gomod"], enabled: false }] },
    {
      packageRules: [{ matchPackageNames: ["pkg"], force: { enabled: false } }],
    },
    // Deliberately outside the first policy's scope: later rules can re-enable packages.
    { packageRules: [{ matchPackageNames: ["*"], enabled: false }] },
  ]) {
    assert.deepEqual(violations(config), []);
  }
});

test("invalid policy fields cannot silently pass", () => {
  for (const config of [
    null,
    [],
    false,
    { enabled: "false" },
    { force: null },
    { force: { enabled: "false" } },
    { extends: ":disableRenovate" },
    { extends: [null] },
  ]) {
    assert.throws(() => violations(config));
  }
});

test("all supported config locations detect disabling", (t) => {
  for (const filename of CONFIG_FILES) {
    const contents =
      filename === "package.json"
        ? '{"renovate":{"enabled":false}}'
        : '{"enabled":false}';
    const result = inspectRepository(repository(t, { [filename]: contents }));
    assert.equal(result.configFile, filename);
    assert.deepEqual(result.reasons, ["enabled: false"]);
  }
});

test("JSON5 comments, unquoted keys, single quotes and trailing commas", (t) => {
  const root = repository(t, {
    "renovate.json5":
      "{ /* policy */ extends: ['config:recommended', ':disableRenovate',], }",
  });
  assert.deepEqual(inspectRepository(root).reasons, [
    "extends includes :disableRenovate",
  ]);
});

test("uses first config by Renovate precedence, not other presets", (t) => {
  const root = repository(t, {
    "renovate.json": "{}",
    ".github/renovate.json5": "{enabled:false}",
    "presets/disabled.json": '{"enabled":false}',
  });
  assert.deepEqual(inspectRepository(root), {
    configFile: "renovate.json",
    reasons: [],
  });
  rmSync(join(root, "renovate.json"));
  assert.deepEqual(inspectRepository(root), {
    configFile: ".github/renovate.json5",
    reasons: ["enabled: false"],
  });
});

test("missing config and ordinary package.json use global defaults", (t) => {
  for (const files of [
    {},
    { "package.json": '{"name":"app","enabled":false}' },
  ]) {
    assert.deepEqual(inspectRepository(repository(t, files)), {
      configFile: null,
      reasons: [],
    });
  }
});

test("invalid config fails instead of falling through to another file", (t) => {
  for (const invalid of ["{broken", "null", "[]"]) {
    const root = repository(t, {
      "renovate.json": invalid,
      ".renovaterc": "{}",
    });
    assert.throws(() => inspectRepository(root));
  }
});

test("package.json renovate value must be an object", (t) => {
  const root = repository(t, { "package.json": '{"renovate":false}' });
  assert.throws(() => inspectRepository(root), /must be an object/);
});

test("config symlinks, including dangling links, are rejected", (t) => {
  const root = repository(t, { "target.json": "{}" });
  for (const target of ["target.json", "missing.json"]) {
    symlinkSync(target, join(root, "renovate.json"));
    assert.throws(() => inspectRepository(root), /symlinks/);
    rmSync(join(root, "renovate.json"));
  }
});

test("parent directory symlinks cannot escape the repository", (t) => {
  const outside = repository(t, { "renovate.json": "{}" });
  const root = repository(t);
  symlinkSync(outside, join(root, ".github"));
  assert.throws(() => inspectRepository(root), /symlinks/);
});

test("rejects oversized files and directories used as config", (t) => {
  const root = repository(t, { "renovate.json": " ".repeat(1024 * 1024 + 1) });
  assert.throws(() => inspectRepository(root), /1 MiB/);
  rmSync(join(root, "renovate.json"));
  mkdirSync(join(root, "renovate.json"));
  assert.throws(() => inspectRepository(root), /regular file/);
});

test("does not execute repository JavaScript", (t) => {
  const root = repository(t, {
    "renovate.config.js": "throw new Error('must not execute');",
  });
  assert.deepEqual(inspectRepository(root), { configFile: null, reasons: [] });
});

test("CLI distinguishes passes, policy failures and invalid configuration", (t) => {
  const script = fileURLToPath(new URL("check.mjs", import.meta.url));
  for (const [contents, status] of [
    ["{}", 0],
    ["{enabled:false}", 1],
    ["{ secret: DO_NOT_PRINT_ME", 2],
  ]) {
    const root = repository(t, { "renovate.json5": contents });
    const result = spawnSync(process.execPath, [script, root], {
      encoding: "utf8",
      env: { ...process.env, GITHUB_STEP_SUMMARY: "" },
    });
    assert.equal(result.status, status, result.stderr);
    assert.ok(!result.stdout.includes("DO_NOT_PRINT_ME"));
    if (status === 1) assert.match(result.stdout, /issues\/791627/);
  }
});
