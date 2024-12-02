import { join } from "path";
import { z } from "zod";
import {
  convertWorkflowTemplate,
  parseWorkflow,
  type WorkflowTemplate,
} from "@actions/workflow-parser";
import { error, info, debug as verbose } from "@actions/core";
import type { FileSystem, DirectoryEntry } from "./filesystem";

/*
 * Define the schema for validating release-please configuration
 */
const ReleasePleaseConfig = z.object({
  packages: z.record(z.string(), z.unknown()),
});

/**
 * Check that all GitHub Actions and reusable workflows in a repository
 * are properly configured for release in release-please-config.json
 *
 * @param fs The filesystem to use for reading files and directories
 * @param config The parsed release-please configuration object
 */
export class ReleaseConfigChecker {
  constructor(
    private fs: FileSystem,
    private config: unknown = {},
  ) {}

  /**
   * Process directory entries to identify action packages
   *
   * @param entries Array of directory entries from the actions directory
   * @returns A Set of package paths relative to the repository root
   */
  processActionEntries(entries: DirectoryEntry[]): Set<string> {
    return new Set(
      entries
        .filter((entry) => entry.isDirectory())
        .map((entry) => join("actions", entry.name)),
    );
  }

  /**
   * Determine if a workflow template represents a reusable workflow
   *
   * @param workflowTemplate The parsed workflow template
   * @returns true if the workflow has a workflow_call event trigger
   */
  isWorkflowReusable(workflowTemplate: WorkflowTemplate): boolean {
    return workflowTemplate.events.workflow_call !== undefined;
  }

  /**
   * Compare configured packages against discovered packages to find missing ones
   *
   * @param configuredPackages Set of packages configured in release-please-config.json
   * @param discoveredPackages Set of packages found in the repository
   * @returns Array of package paths that are missing from the configuration
   */
  findMissingConfigurations(
    configuredPackages: Set<string>,
    discoveredPackages: Set<string>,
  ): string[] {
    return Array.from(discoveredPackages.difference(configuredPackages));
  }

  /**
   * Parse a workflow file content into a WorkflowTemplate
   *
   * @param content The raw YAML content of the workflow file
   * @returns Promise resolving to the parsed WorkflowTemplate
   * @throws Error if parsing fails
   */
  parse(content: string): Promise<WorkflowTemplate> {
    const { context, value } = parseWorkflow(
      { name: "inline", content },
      { error, info, verbose },
    );

    if (value === undefined) {
      throw new Error("Failed to parse workflow");
    }

    return convertWorkflowTemplate(context, value);
  }

  /**
   * Scan the actions directory to find all action packages
   *
   * @returns Promise resolving to a Set of action package paths
   */
  async getActionPackages(): Promise<Set<string>> {
    verbose("Scanning actions directory for packages...");

    const entries = await this.fs.readDirectory("actions");
    verbose(`Found ${entries.length.toString()} items in actions directory`);

    const packages = this.processActionEntries(entries);
    info(`Found ${packages.size.toString()} action packages`);

    return packages;
  }

  /**
   * Scan the workflows directory to find all reusable workflows
   *
   * @returns Promise resolving to a Set of reusable workflow paths
   */
  async getReusableWorkflows(): Promise<Set<string>> {
    verbose("Scanning .github/workflows directory for reusable workflows...");
    const packages = new Set<string>();

    const files = await this.fs.readDirectory(join(".github", "workflows"));
    verbose(`Found ${files.length.toString()} files in workflows directory`);

    for (const file of files) {
      if (!file.isFile() || !/\.ya?ml$/.test(file.name)) {
        verbose(`Skipping ${file.name} - not a YAML file`);
        continue;
      }

      const filePath = join(".github", "workflows", file.name);
      const content = await this.fs.readFile(filePath);

      try {
        const template = await this.parse(content);
        if (this.isWorkflowReusable(template)) {
          packages.add(filePath);
          verbose(`Added reusable workflow: ${filePath}`);
        }
      } catch (e) {
        error(
          `Failed to parse workflow ${filePath}: ${e instanceof Error ? e.message : String(e)}`,
        );
        throw e;
      }
    }

    info(`Found ${packages.size.toString()} reusable workflows`);
    return packages;
  }

  /**
   * Find all missing configurations
   *
   * @returns Promise resolving to an array of paths missing from the configuration
   * @throws Error if config is invalid or filesystem operations fail
   */
  async check(): Promise<string[]> {
    verbose("Checking for missing configurations...");

    const parsedConfig = ReleasePleaseConfig.parse(this.config);
    const configuredPackages = new Set(Object.keys(parsedConfig.packages));
    verbose(
      `Found ${configuredPackages.size.toString()} configured packages in release-please-config.json`,
    );

    const [actionPackages, reusableWorkflows] = await Promise.all([
      this.getActionPackages(),
      this.getReusableWorkflows(),
    ]);

    return this.findMissingConfigurations(
      configuredPackages,
      new Set([...actionPackages, ...reusableWorkflows]),
    );
  }
}

/**
 * Run the checker and handle the result
 *
 * @param fs The filesystem to use for reading files and directories
 * @param config The parsed release-please configuration object
 * @returns Promise resolving to 0 if all packages are configured, 1 otherwise
 */
export async function main(fs: FileSystem, config: unknown): Promise<number> {
  try {
    const checker = new ReleaseConfigChecker(fs, config);
    const missingConfigs = await checker.check();

    if (missingConfigs.length === 0) {
      info("All items are releasable!");
      return 0;
    }

    error(
      `Found ${missingConfigs.length.toString()} items missing from release-please-config.json:\n` +
        `${missingConfigs.join("\n")}\n` +
        `Please add them to release-please-config.json!`,
    );
    return 1;
  } catch (e) {
    error(`Fatal error: ${e instanceof Error ? e.message : String(e)}`);
    return 1;
  }
}
