const { readFileSync } = require("node:fs");

function validatePolicy(config) {
  if (config.enabled === false || config.force?.enabled === false) {
    throw new Error(
      "Disabling Renovate repository-wide is prohibited. Use scoped packageRules to disable specific updates.",
    );
  }

  if (config.extends?.includes(":disableRenovate")) {
    throw new Error("The :disableRenovate preset is prohibited.");
  }
}

if (require.main === module) {
  try {
    validatePolicy(JSON.parse(readFileSync(process.argv[2], "utf8")));
  } catch (error) {
    console.error(error.message);
    process.exitCode = 1;
  }
}

module.exports = { validatePolicy };
