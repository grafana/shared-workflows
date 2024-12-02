import { readFile, readdir } from "fs/promises";

/**
 * A representation of a directory entry, which can be either a file or a
 * directory. This is a subset of the `fs.Dirent` interface containing just the
 * parts we need.
 */
export interface DirectoryEntry {
  name: string;
  isDirectory: () => boolean;
  isFile: () => boolean;
}

/**
 * Abstraction of the filesystem for reading directories and files. This is to
 * allow for easier testing by providing an in-memory filesystem implementation.
 */
export interface FileSystem {
  readDirectory: (path: string) => Promise<DirectoryEntry[]>;
  readFile: (path: string) => Promise<string>;
}

/**
 * Implementation of the filesystem using Node.js's built-in `fs` module, used
 * in production.
 */
export class NodeFileSystem implements FileSystem {
  async readDirectory(path: string): Promise<DirectoryEntry[]> {
    return readdir(path, { withFileTypes: true });
  }

  async readFile(path: string): Promise<string> {
    return readFile(path, "utf-8");
  }
}
