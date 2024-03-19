import { existsSync, mkdirSync, cpSync, rmdirSync, readFileSync, writeFileSync } from 'node:fs';
import path from 'node:path';
import minimist from 'minimist';
import { sign } from '../utils/sign.js';
import { absoluteToRelativePaths, addSha1ForFiles, generateFolder, listFiles } from '../utils/utils.js';
import { compressFilesToZip } from '../utils/zip.js';
import { createHash } from 'node:crypto';

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

// Typescript interface for the zip command
export interface ZipArgs {
  distDir?: string;
  outDir?: string;
  signatureType?: string;
  rootUrls?: string;
}

export const zip = async (argv: ZipArgs) => {
  const distDir = argv.distDir ?? 'dist';
  const outDir = argv.outDir ?? '__to-upload__';
  const signatureType: string | undefined = argv.signatureType ?? undefined;
  const rootUrls: string[] = argv.rootUrls?.split(',') ?? [];
  const pluginDistDir = path.resolve(distDir);

  if (!existsSync(pluginDistDir)) {
    throw new Error(
      `Plugin \`${distDir}\` directory is missing. Did you build the plugin before attempting to to zip it?`
    );
  }

  const buildDir = generateFolder('package-zip');
  const pluginJson = JSON.parse(readFileSync(path.join(`${pluginDistDir}`, `plugin.json`), 'utf-8'));
  const {
    id: pluginId,
    info: { version: pluginVersion },
  } = pluginJson;

  const copiedPath = path.join(process.cwd(), buildDir, pluginId);

  cpSync(pluginDistDir, copiedPath, { recursive: true });

  const filesWithZipPaths = absoluteToRelativePaths(copiedPath);
  await sign(copiedPath, rootUrls, signatureType);

  const anyPlatformZipPath = path.join(`${buildDir}`, `${pluginVersion}`, `${pluginId}-${pluginVersion}.zip`);
  
  // Binary distribution for any platform
  await compressFilesToZip(
    path.join(anyPlatformZipPath),
    pluginId,
    { ...filesWithZipPaths, [path.join(copiedPath, 'MANIFEST.txt')]: 'MANIFEST.txt' }
  );

  const anyPlatformJson = getJsonMetadata(anyPlatformZipPath);
  const anyPlatformJsonPath = path.join(`${buildDir}`, `${pluginVersion}`, `info.json`);
  const anyPlatformJsonString = JSON.stringify(anyPlatformJson, null, 2);
  mkdirSync(path.dirname(anyPlatformJsonPath), { recursive: true });
  const anyPlatformJsonBuffer = Buffer.from(anyPlatformJsonString);
  writeFileSync(anyPlatformJsonPath, anyPlatformJsonBuffer);

  // Take filesWithZipPaths and split them into goBuildFiles and nonGoBuildFiles
  const goBuildFiles: { [key: string]: string } = {};
  const nonGoBuildFiles: { [key: string]: string } = {};
  Object.keys(filesWithZipPaths).forEach((filePath: string) => {
    const zipPath = filesWithZipPaths[filePath];
    const fileName = filePath.split(path.sep).pop();
    if (!fileName) {
      throw new Error('fileName is undefined or null');
    }
    if (fileName.startsWith('gpx')) {
      goBuildFiles[filePath] = zipPath;
    } else {
      nonGoBuildFiles[filePath] = zipPath;
    }
  });

  // Noop if there are no go build files
  // Otherwise, compress each go build file along with all non-go files into a separate zip
  // Creates os/arch specific distributions
  for (let [filePath, relativePath] of Object.entries(goBuildFiles)) {
    const fileName = filePath
      .split(path.sep)
      .pop()
      ?.replace(/\.exe$/, '');

    if (fileName === null || fileName === undefined) {
      throw new Error('fileName is undefined or null');
    }

    const [goos, goarch] = fileName?.split('_').slice(2) ?? [];

    // If any of these are null, throw an error
    if (fileName === null || goos === null || goarch === null) {
      throw new Error('fileName, goos, or goarch is undefined or null');
    }

    const outputName = `${pluginId}-${pluginVersion}.${goos}_${goarch}.zip`;
    const zipDestination = path.join(`${buildDir}`, `${pluginVersion}`, `${goos}`, `${outputName}`);

    mkdirSync(path.dirname(zipDestination), { recursive: true });

    const workingDir = path.join(path.dirname(zipDestination), 'working');

    mkdirSync(workingDir, { recursive: true });

    // Copy filePath to workingDir/relativePath
    cpSync(filePath, path.join(workingDir, relativePath));

    // Copy all nonGoBuildFiles into workingDir
    Object.entries(nonGoBuildFiles).forEach(([absPath, relPath]) => {
      cpSync(absPath, path.join(workingDir, relPath));
    });

    // Add the manifest
    await sign(workingDir, rootUrls, signatureType);
    const toCompress = absoluteToRelativePaths(workingDir);
    await compressFilesToZip(zipDestination, pluginId, toCompress);
    // Add json info file
    const json = getJsonMetadata(zipDestination);
    const jsonPath = path.join(path.dirname(zipDestination), `info-${goos}_${goarch}.json`);
    const jsonString = JSON.stringify(json, null, 2);
    const jsonBuffer = Buffer.from(jsonString);
    writeFileSync(jsonPath, jsonBuffer);
    rmdirSync(workingDir, { recursive: true });
  }

  // Copy all of the files from buildDir/pluginVersion to buildDir/latest
  // Removes pluginVersion from their path and filename and replaces it with latest
  const latestPath = path.join(`${buildDir}`, `latest`);
  const currentVersionPath = `${buildDir}/${pluginVersion}`;
  mkdirSync(latestPath, { recursive: true });
  const filesToCopy = listFiles(currentVersionPath);
  filesToCopy.forEach((filePath) => {
    const fileNameArray = filePath.split(path.sep);
    const newFileName = fileNameArray.pop()?.replace(`${pluginVersion}`, 'latest');
    // If newfilename is null, then throw an error
    if (newFileName === null) {
      throw new Error('Bad filename while trying to copy files to latest');
    }
    if (newFileName) {
      const newFileSubdirectory = filePath
        .replace(currentVersionPath, latestPath)
        .split(path.sep)
        .slice(0, -1)
        .join(path.sep);
      const newFilePath = path.join(`${newFileSubdirectory}`, `${newFileName}`);
      mkdirSync(path.dirname(newFilePath), { recursive: true });
      cpSync(filePath, newFilePath);
    }
  });

  // Sign all zip files with sha1
  const zipFiles = listFiles(currentVersionPath).filter((file) => file.endsWith('.zip'));
  addSha1ForFiles(zipFiles);
  const latestZipFiles = listFiles(latestPath).filter((file) => file.endsWith('.zip'));
  addSha1ForFiles(latestZipFiles);

  // Move buildDir/latest and buildDir/pluginVersion to rootDir/${outDir}
  const toUploadPath = path.join(process.cwd(), outDir);
  mkdirSync(toUploadPath, { recursive: true });
  cpSync(latestPath, path.join(toUploadPath, 'latest'), { recursive: true });
  cpSync(currentVersionPath, path.join(toUploadPath, pluginVersion), { recursive: true });

  // Clean up after yourself
  rmdirSync(buildDir, { recursive: true });
};
