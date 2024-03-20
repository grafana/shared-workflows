import { existsSync, mkdirSync, readdirSync, statSync, readFileSync, writeFileSync } from 'node:fs';
import path from 'node:path';
import crypto, { createHash } from 'node:crypto';

export function generateFolder(prefix: string): string {
  const randomHash = crypto.createHash('md5').update(new Date().getTime().toString()).digest('hex');
  const folderName = `${prefix}-${randomHash}`;

  if (!existsSync(folderName)) {
    mkdirSync(folderName);
  } else {
    throw new Error(`Folder ${folderName} already exists`);
  }
  return folderName;
}

// Takes a directory, gives absolute paths for all files in it
// and its subdirectories
export function listFiles(dir: string): string[] {
  const out: string[] = [];
  readdirSync(dir).forEach((file) => {
    if (statSync(path.join(dir, file)).isDirectory()) {
      out.push(...listFiles(path.join(dir, file)));
    } else {
      out.push(path.join(dir, file));
    }
  });
  return out;
}

export function addSha1ForFiles(files: string[]) {
  files.forEach((file) => {
    const fileContent = readFileSync(file);
    const sha1 = crypto.createHash('sha1').update(fileContent).digest('hex');
    writeFileSync(`${file}.sha1`, sha1);
  });
}
export const absoluteToRelativePaths = (dir: string) => {
  const out: { [key: string]: string } = {};
  listFiles(dir).forEach((file) => {
    out[file] = file.replace(dir, '');
  });
  return out;
};

export const getJsonMetadata = (zipPath: string): {
  plugin: {
    md5: string,
    name: string,
    sha1: string,
    size: number
  }
} => {
  const name = zipPath.split(path.sep).pop();
  if (name === null || name === undefined) {
    throw new Error('name is undefined or null');
  }
  const md5 = createHash('md5').update(readFileSync(zipPath)).digest('hex');
  const sha1 = createHash('sha1').update(readFileSync(zipPath)).digest('hex');
  const size = readFileSync(zipPath).byteLength;
  return {
    "plugin": {
      "md5": md5,
      "name": name,
      "sha1": sha1,
      "size": size
    }
  }
}
