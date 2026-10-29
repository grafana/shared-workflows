const assert = require("node:assert/strict");
const { test } = require("node:test");
const { validatePolicy } = require("./check-policy.cjs");

test("rejects top-level enabled: false", () => {
  assert.throws(() => validatePolicy({ enabled: false }), /repository-wide/);
});

test("rejects force.enabled: false", () => {
  assert.throws(
    () => validatePolicy({ force: { enabled: false } }),
    /repository-wide/,
  );
});

test("rejects the disable preset among other presets", () => {
  assert.throws(
    () =>
      validatePolicy({ extends: ["config:recommended", ":disableRenovate"] }),
    /:disableRenovate/,
  );
});

test("allows defaults and explicitly enabled configurations", () => {
  for (const config of [
    {},
    { enabled: true },
    { force: { enabled: true } },
    { extends: ["config:recommended"] },
  ]) {
    assert.doesNotThrow(() => validatePolicy(config));
  }
});

test("allows scoped disabling of major Go updates", () => {
  assert.doesNotThrow(() =>
    validatePolicy({
      packageRules: [
        {
          matchManagers: ["gomod"],
          matchUpdateTypes: ["major"],
          enabled: false,
        },
      ],
    }),
  );
});

test("allows major configuration disabling", () => {
  assert.doesNotThrow(() => validatePolicy({ major: { enabled: false } }));
});
