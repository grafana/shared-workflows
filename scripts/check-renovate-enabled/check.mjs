import { appendFileSync, lstatSync, readFileSync } from "node:fs";
import { isAbsolute, join, resolve } from "node:path";
import { pathToFileURL } from "node:url";
import JSON5 from "json5";

// Renovate's default search order. Check the first match, not every preset in
// the repository. Keep this in sync when upgrading the self-hosted runner.
// https://docs.renovatebot.com/configuration-options/#locations-for-configuration-filenames
export const CONFIG_FILES = [
  "renovate.json",
  "renovate.jsonc",
  "renovate.json5",
  ".github/renovate.json",
  ".github/renovate.jsonc",
  ".github/renovate.json5",
  ".gitlab/renovate.json",
  ".gitlab/renovate.jsonc",
  ".gitlab/renovate.json5",
  ".renovaterc",
  ".renovaterc.json",
  ".renovaterc.jsonc",
  ".renovaterc.json5",
  "package.json",
];

const ISSUE = "https://github.com/grafana/deployment_tools/issues/791627";
const MAX_CONFIG_BYTES = 1024 * 1024;

function isObject(value) {
  return value !== null && typeof value === "object" && !Array.isArray(value);
}

export function violations(config) {
  if (!isObject(config)) {
    throw new Error("Renovate configuration must be an object.");
  }
  if (Object.hasOwn(config, "enabled") && typeof config.enabled !== "boolean") {
    throw new Error("Renovate enabled must be a boolean.");
  }
  if (Object.hasOwn(config, "force") && !isObject(config.force)) {
    throw new Error("Renovate force must be an object.");
  }
  if (
    Object.hasOwn(config.force ?? {}, "enabled") &&
    typeof config.force.enabled !== "boolean"
  ) {
    throw new Error("Renovate force.enabled must be a boolean.");
  }
  if (
    Object.hasOwn(config, "extends") &&
    (!Array.isArray(config.extends) ||
      config.extends.some((preset) => typeof preset !== "string"))
  ) {
    throw new Error("Renovate extends must be an array of preset names.");
  }

  const reasons = [];
  if (config.enabled === false) reasons.push("enabled: false");
  if (config.force?.enabled === false) reasons.push("force.enabled: false");
  if (config.extends?.includes(":disableRenovate")) {
    reasons.push("extends includes :disableRenovate");
  }
  return reasons;
}

function readConfig(root, filename) {
  // Config is untrusted data. Never follow a symlink out of the checkout,
  // including a symlink used for .github or .gitlab itself.
  let current = root;
  const parts = filename.split("/");
  for (const [index, part] of parts.entries()) {
    current = join(current, part);
    let stat;
    try {
      stat = lstatSync(current);
    } catch (error) {
      if (error.code === "ENOENT") return undefined;
      throw new Error(`Cannot read ${filename}.`, { cause: error });
    }
    if (stat.isSymbolicLink()) {
      throw new Error(`Configuration path ${filename} must not use symlinks.`);
    }
    if (index < parts.length - 1) {
      if (!stat.isDirectory()) throw new Error(`Cannot read ${filename}.`);
    } else if (!stat.isFile() || stat.size > MAX_CONFIG_BYTES) {
      throw new Error(
        `${filename} must be a regular file no larger than 1 MiB.`,
      );
    }
  }
  try {
    const contents = readFileSync(current, "utf8");
    // package.json is JSON; standalone Renovate files also support JSON5.
    return filename === "package.json"
      ? JSON.parse(contents)
      : JSON5.parse(contents);
  } catch {
    // Do not echo file contents or parser errors from an untrusted config.
    throw new Error(`Cannot parse ${filename} as JSON or JSON5.`);
  }
}

export function inspectRepository(root) {
  if (!isAbsolute(root) || !lstatSync(root).isDirectory()) {
    throw new Error("An existing absolute repository directory is required.");
  }
  for (const filename of CONFIG_FILES) {
    let config = readConfig(root, filename);
    if (config === undefined) continue;
    if (filename === "package.json") {
      if (!isObject(config)) throw new Error("package.json must be an object.");
      if (!Object.hasOwn(config, "renovate")) continue;
      config = config.renovate;
    }
    return { configFile: filename, reasons: violations(config) };
  }
  return { configFile: null, reasons: [] };
}

function report(kind, message) {
  const escaped = message
    .replaceAll("%", "%25")
    .replaceAll("\r", "%0D")
    .replaceAll("\n", "%0A");
  console.log(`::${kind} title=Renovate policy::${escaped}`);
  if (process.env.GITHUB_STEP_SUMMARY) {
    appendFileSync(process.env.GITHUB_STEP_SUMMARY, `${message}\n\n`);
  }
}

function main() {
  if (process.argv.length !== 3) {
    report("error", "Usage: node check.mjs /absolute/path/to/repository");
    process.exitCode = 2;
    return;
  }
  try {
    const { configFile, reasons } = inspectRepository(resolve(process.argv[2]));
    if (reasons.length > 0) {
      report(
        "error",
        `${configFile} disables Renovate repository-wide (${reasons.join("; ")}). Remove the repository-wide disablement; scoped packageRules remain allowed. See ${ISSUE}.`,
      );
      process.exitCode = 1;
    } else {
      report(
        "notice",
        configFile
          ? `${configFile}: no explicit repository-wide Renovate disablement found. Inherited presets and catch-all rules are not evaluated.`
          : "No repository Renovate configuration found; the self-hosted global defaults apply.",
      );
    }
  } catch (error) {
    report("error", `Unable to check Renovate configuration: ${error.message}`);
    process.exitCode = 2;
  }
}

if (
  process.argv[1] &&
  import.meta.url === pathToFileURL(resolve(process.argv[1])).href
) {
  main();
}
