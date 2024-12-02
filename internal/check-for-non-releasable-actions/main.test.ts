import { describe, it, expect } from "bun:test";
import { type DirectoryEntry, type FileSystem } from "./filesystem";

import { ReleaseConfigChecker } from "./main";

export type FileSystemStructure = {
  [key: string]: string | FileSystemStructure;
};

/**
 * An in-memory implementation of the filesystem for testing purposes.
 */
export class InMemoryFileSystem implements FileSystem {
  constructor(private readonly structure: FileSystemStructure = {}) {}

  private getEntry(path: string): string | FileSystemStructure | undefined {
    if (path === "" || path === ".") {
      return this.structure;
    }

    let current: FileSystemStructure | string = this.structure;
    for (const part of path.split("/")) {
      if (typeof current === "string" || !(part in current)) {
        return undefined;
      }

      current = current[part];
    }

    return current;
  }

  readDirectory(path: string): Promise<DirectoryEntry[]> {
    const node = this.getEntry(path);

    if (typeof node === "string" || node === undefined) {
      throw new Error(`Directory not found: ${path}`);
    }

    const res = Object.entries(node).map(([name, content]) => ({
      name,
      isDirectory: () => typeof content === "object",
      isFile: () => typeof content === "string",
    }));

    return Promise.resolve(res);
  }

  readFile(path: string): Promise<string> {
    const content = this.getEntry(path);

    if (typeof content !== "string") {
      throw new Error(`File not found: ${path}`);
    }

    return Promise.resolve(content);
  }
}

describe("processActionEntries", () => {
  const checker = new ReleaseConfigChecker(new InMemoryFileSystem());

  it.each<{
    name: string;
    entries: { name: string; isDirectory: () => boolean }[];
    expected: Set<string>;
  }>([
    {
      name: "filters and transforms directory entries",
      entries: [
        { name: "action1", isDirectory: () => true },
        { name: "not-dir", isDirectory: () => false },
        { name: "action2", isDirectory: () => true },
      ],
      expected: new Set(["actions/action1", "actions/action2"]),
    },
    {
      name: "handles empty entries",
      entries: [],
      expected: new Set(),
    },
  ])("$name", ({ entries, expected }) => {
    const result = checker.processActionEntries(
      entries.map((entry) => ({ ...entry, isFile: () => false })),
    );
    expect(result).toEqual(expected);
  });
});

describe("isWorkflowReusable", () => {
  const checker = new ReleaseConfigChecker(new InMemoryFileSystem());

  it.each([
    {
      name: "identifies reusable workflow",
      template: `
          on:
            workflow_call:
          jobs:
            test:
              runs-on: ubuntu-latest
              steps:
                - run: echo test
        `,
      expected: true,
    },
    {
      name: "identifies non-reusable workflow",
      template: `
          on:
            push:
          jobs:
            test:
              runs-on: ubuntu-latest
              steps:
                - run: echo test
        `,
      expected: false,
    },
  ])("$name", async ({ template, expected }) => {
    const workflowTemplate = await checker.parse(template);
    expect(checker.isWorkflowReusable(workflowTemplate)).toBe(expected);
  });
});

describe("ReleaseConfigChecker integration tests", () => {
  it("finds missing configurations in a complete repository", async () => {
    const fs = new InMemoryFileSystem({
      actions: {
        "test-action": {
          "action.yml": "name: Test Action",
        },
        "configured-action": {
          "action.yml": "name: Configured Action",
        },
      },
      ".github": {
        workflows: {
          "reusable.yml": `
              on:
                workflow_call:
              jobs:
                test:
                  runs-on: ubuntu-latest
                  steps:
                    - run: echo test
            `,
          "normal.yml": `
              on:
                push:
              jobs:
                test:
                  runs-on: ubuntu-latest
                  steps:
                    - run: echo test
            `,
        },
      },
    });

    const config = {
      packages: {
        "actions/configured-action": {},
        ".github/workflows/normal.yml": {},
      },
    };

    const checker = new ReleaseConfigChecker(fs, config);
    const missingConfigs = await checker.check();

    expect(missingConfigs).toContain("actions/test-action");
    expect(missingConfigs).toContain(".github/workflows/reusable.yml");
    expect(missingConfigs).not.toContain("actions/configured-action");
    expect(missingConfigs).not.toContain(".github/workflows/normal.yml");
  });
});
