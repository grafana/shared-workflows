import { describe, expect, it } from "bun:test";
import { tmpFileAsync } from "./tempfile";
import { readFile } from "node:fs/promises";
import { existsSync } from "node:fs";
import { basename, dirname } from "node:path";
import type { FileOptions } from "tmp";

describe("tmpFileAsync", () => {
  it.each<{ description: string; options: FileOptions }>([
    { description: "with default options", options: {} },
    { description: "with custom prefix", options: { prefix: "test-" } },
    { description: "with custom postfix", options: { postfix: ".txt" } },
    { description: "with custom directory", options: { tmpdir: "/tmp" } },
  ])(
    "should create and cleanup temp file $description",
    async ({ options }) => {
      let name: string | undefined;

      {
        await using tmp = await tmpFileAsync(options);

        name = tmp.name;
        const handle = tmp.handle;

        expect(existsSync(tmp.name)).toBeTrue();

        const fileName = basename(tmp.name);

        if (options.prefix) {
          expect(fileName).toStartWith(options.prefix);
        }

        if (options.postfix) {
          expect(fileName).toEndWith(options.postfix);
        }

        const dir = dirname(tmp.name);

        if (options.tmpdir) {
          expect(dir).toBe(options.tmpdir);
        }

        const testData = "test content";
        await handle.writeFile(testData);
        const content = await readFile(tmp.name, "utf8");

        expect(content).toBe(testData);
      }

      expect(name).toBeDefined();

      // Verify file is cleaned up
      expect(existsSync(name), `${name} still exists`).toBeFalse();
    },
  );

  // Error cases
  it.each([
    {
      description: "invalid directory",
      options: { tmpdir: "/nonexistent" },
      expectedError: "ENOENT",
    },
  ])("should handle errors for $description", () => {
    expect(
      tmpFileAsync({
        tmpdir: "/nonexistent",
      }),
    ).rejects.toThrow(expect.objectContaining({ code: "ENOENT" }));
  });

  it("creates files with the permissions specified", async () => {
    try {
      await using tmp = await tmpFileAsync({
        // read-only for everyone
        mode: 0o444,
      });

      await tmp.handle.writeFile("nonsense");
    } catch (e) {
      expect(e).toMatchObject({ code: "EACCES" });
      return;
    }

    expect().fail("Expected an error to be thrown");
  });

  // Testing using statement
  it("should cleanup file when used with 'using'", async () => {
    let tmpName: string;

    {
      await using tmp = await tmpFileAsync({});
      tmpName = tmp.name;

      // Verify file exists within block
      expect(existsSync(tmp.name)).toBe(true);
    }

    // Verify file is cleaned up after block
    expect(existsSync(tmpName)).toBe(false);
  });
});
