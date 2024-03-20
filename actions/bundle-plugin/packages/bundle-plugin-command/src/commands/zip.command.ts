import {
  existsSync,
  mkdirSync,
  cpSync,
  rmdirSync,
  readFileSync,
  writeFileSync
} from 'node:fs'
import path from 'node:path'
import { sign } from '../utils/sign.js'
import {
  absoluteToRelativePaths,
  addSha1ForFiles,
  generateFolder,
  getJsonMetadata,
  listFiles
} from '../utils/utils.js'
import { compressFilesToZip } from '../utils/zip.js'
import { URL } from 'url'

// Typescript interface for the zip command
export interface ZipArgs {
  distDir?: string
  outDir?: string
  signatureType?: string
  rootUrls?: string
  noSign?: boolean
}

// Wrapper to parse arguments, make sure they are valid, and call the zip function
// Also handles cleanup of the temporary build directory in case of crashes
export const zip = async (argv: ZipArgs) => {
  const distDir = argv.distDir ?? 'dist'
  const outDir = argv.outDir ?? '__to-upload__'
  const signatureType: string | undefined = argv.signatureType ?? undefined
  const rootUrls: string[] = argv.rootUrls?.split(',') ?? []
  const noSign = argv.noSign ?? false

  const pluginDistDir = path.resolve(distDir)

  if (!existsSync(pluginDistDir)) {
    throw new Error(
      `Plugin \`${distDir}\` directory is missing. Did you build the plugin before attempting to to zip it?`
    )
  }

  // This check is redundant with the one in utils/sign
  // It's kept here also to fail faster if the user tries to sign a plugin without the required env variables
  const GRAFANA_API_KEY = process.env.GRAFANA_API_KEY
  const GRAFANA_ACCESS_POLICY_TOKEN = process.env.GRAFANA_ACCESS_POLICY_TOKEN

  if (!GRAFANA_ACCESS_POLICY_TOKEN && !GRAFANA_API_KEY && !noSign) {
    throw new Error(
      'You must create a GRAFANA_ACCESS_POLICY_TOKEN env variable to sign plugins. Please see: https://grafana.com/developers/plugin-tools/publish-a-plugin/sign-a-plugin#generate-an-access-policy-token for instructions.\n' +
        'You can also use the --noSign flag to skip signing the plugin.'
    )
  }
  if (GRAFANA_API_KEY && !noSign) {
    console.warn(
      `\x1b[33m%s\x1b[0m`,
      'The usage of GRAFANA_API_KEY is deprecated, please consider using GRAFANA_ACCESS_POLICY_TOKEN instead. For more info visit https://grafana.com/developers/plugin-tools/publish-a-plugin/sign-a-plugin'
    )
  }

  const buildDir = generateFolder('package-zip')
  try {
    await zipWorker(
      outDir,
      signatureType,
      rootUrls,
      pluginDistDir,
      buildDir,
      noSign
    )
  } catch (err) {
    console.error(err)
    throw new Error('Failed to zip the plugin')
  } finally {
    try {
      rmdirSync(buildDir, { recursive: true })
      console.log('Cleanup successful.')
    } catch (cleanupErr) {
      console.error('Failed to clean up:', cleanupErr)
    }
  }
}

export const zipWorker = async (
  outDir: string,
  signatureType: string | undefined,
  rootUrls: string[],
  pluginDistDir: string | URL,
  buildDir: string,
  // eslint-disable-next-line @typescript-eslint/no-inferrable-types
  noSign: boolean
) => {
  const pluginJson = JSON.parse(
    readFileSync(path.join(`${pluginDistDir}`, `plugin.json`), 'utf-8')
  )
  const {
    id: pluginId,
    info: { version: pluginVersion }
  } = pluginJson

  const copiedPath = path.join(process.cwd(), buildDir, pluginId)

  cpSync(pluginDistDir, copiedPath, { recursive: true })

  const filesWithZipPaths = absoluteToRelativePaths(copiedPath)
  if (!noSign) {
    await sign(copiedPath, rootUrls, signatureType)
  }

  const anyPlatformZipPath = path.join(
    `${buildDir}`,
    `${pluginVersion}`,
    `${pluginId}-${pluginVersion}.zip`
  )

  const anyManifest = noSign
    ? {}
    : { [path.join(copiedPath, 'MANIFEST.txt')]: 'MANIFEST.txt' }

  // Binary distribution for any platform
  await compressFilesToZip(path.join(anyPlatformZipPath), pluginId, {
    ...filesWithZipPaths,
    ...anyManifest
  })

  const anyPlatformJson = getJsonMetadata(anyPlatformZipPath)
  const anyPlatformJsonPath = path.join(
    `${buildDir}`,
    `${pluginVersion}`,
    `info.json`
  )
  const anyPlatformJsonString = JSON.stringify(anyPlatformJson, null, 2)
  mkdirSync(path.dirname(anyPlatformJsonPath), { recursive: true })
  const anyPlatformJsonBuffer = Buffer.from(anyPlatformJsonString)
  writeFileSync(anyPlatformJsonPath, anyPlatformJsonBuffer)

  // Take filesWithZipPaths and split them into goBuildFiles and nonGoBuildFiles
  const goBuildFiles: { [key: string]: string } = {}
  const nonGoBuildFiles: { [key: string]: string } = {}
  Object.keys(filesWithZipPaths).forEach((filePath: string) => {
    const zipPath = filesWithZipPaths[filePath]
    const fileName = filePath.split(path.sep).pop()
    if (!fileName) {
      throw new Error('fileName is undefined or null')
    }
    if (fileName.startsWith('gpx')) {
      goBuildFiles[filePath] = zipPath
    } else {
      nonGoBuildFiles[filePath] = zipPath
    }
  })

  // Noop if there are no go build files
  // Otherwise, compress each go build file along with all non-go files into a separate zip
  // Creates os/arch specific distributions
  for (const [filePath, relativePath] of Object.entries(goBuildFiles)) {
    const fileName = filePath
      .split(path.sep)
      .pop()
      ?.replace(/\.exe$/, '')

    if (fileName === null || fileName === undefined) {
      throw new Error('fileName is undefined or null')
    }

    const [goos, goarch] = fileName?.split('_').slice(2) ?? []

    // If any of these are null, throw an error
    if (fileName === null || goos === null || goarch === null) {
      throw new Error('fileName, goos, or goarch is undefined or null')
    }

    const outputName = `${pluginId}-${pluginVersion}.${goos}_${goarch}.zip`
    const zipDestination = path.join(
      `${buildDir}`,
      `${pluginVersion}`,
      `${goos}`,
      `${outputName}`
    )

    mkdirSync(path.dirname(zipDestination), { recursive: true })

    const workingDir = path.join(path.dirname(zipDestination), 'working')

    mkdirSync(workingDir, { recursive: true })

    // Copy filePath to workingDir/relativePath
    cpSync(filePath, path.join(workingDir, relativePath))

    // Copy all nonGoBuildFiles into workingDir
    Object.entries(nonGoBuildFiles).forEach(([absPath, relPath]) => {
      cpSync(absPath, path.join(workingDir, relPath))
    })

    // Add the manifest
    if (!noSign) {
      await sign(workingDir, rootUrls, signatureType)
    }
    const toCompress = absoluteToRelativePaths(workingDir)
    await compressFilesToZip(zipDestination, pluginId, toCompress)
    // Add json info file
    const json = getJsonMetadata(zipDestination)
    const jsonPath = path.join(
      path.dirname(zipDestination),
      `info-${goos}_${goarch}.json`
    )
    const jsonString = JSON.stringify(json, null, 2)
    const jsonBuffer = Buffer.from(jsonString)
    writeFileSync(jsonPath, jsonBuffer)
    rmdirSync(workingDir, { recursive: true })
  }

  // Copy all of the files from buildDir/pluginVersion to buildDir/latest
  // Removes pluginVersion from their path and filename and replaces it with latest
  const latestPath = path.join(`${buildDir}`, `latest`)
  const currentVersionPath = `${buildDir}/${pluginVersion}`
  mkdirSync(latestPath, { recursive: true })
  const filesToCopy = listFiles(currentVersionPath)
  filesToCopy.forEach(filePath => {
    const fileNameArray = filePath.split(path.sep)
    const newFileName = fileNameArray
      .pop()
      ?.replace(`${pluginVersion}`, 'latest')
    // If newfilename is null, then throw an error
    if (newFileName === null) {
      throw new Error('Bad filename while trying to copy files to latest')
    }
    if (newFileName) {
      const newFileSubdirectory = filePath
        .replace(currentVersionPath, latestPath)
        .split(path.sep)
        .slice(0, -1)
        .join(path.sep)
      const newFilePath = path.join(`${newFileSubdirectory}`, `${newFileName}`)
      mkdirSync(path.dirname(newFilePath), { recursive: true })
      cpSync(filePath, newFilePath)
    }
  })

  // Validate all zip files with sha1
  const zipFiles = listFiles(currentVersionPath).filter(file =>
    file.endsWith('.zip')
  )
  addSha1ForFiles(zipFiles)
  const latestZipFiles = listFiles(latestPath).filter(file =>
    file.endsWith('.zip')
  )
  addSha1ForFiles(latestZipFiles)

  // Move buildDir/latest and buildDir/pluginVersion to rootDir/${outDir}
  const toUploadPath = path.join(process.cwd(), outDir)
  try {
    mkdirSync(toUploadPath, { recursive: true })
    cpSync(latestPath, path.join(toUploadPath, 'latest'), { recursive: true })
    cpSync(currentVersionPath, path.join(toUploadPath, pluginVersion), {
      recursive: true
    })
  } catch (err) {
    // Clean up the toUploadPath if there was an error
    rmdirSync(toUploadPath, { recursive: true })
    console.error(err)
    process.exit(1)
  }
}
