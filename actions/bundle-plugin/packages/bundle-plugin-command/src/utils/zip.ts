import archiver from 'archiver';
import { createReadStream, createWriteStream, existsSync, mkdirSync } from 'fs';
import path from 'path';

export function compressFilesToZip(zipFilePath: string, pluginId: string, fileMapping: { [key: string]: string }) {
  return new Promise<void>((resolve, reject) => {
    // Create the folder for output if it does not exist
    const outputDir = path.dirname(zipFilePath);
    if (!existsSync(outputDir)) {
      mkdirSync(outputDir, { recursive: true });
    }
    // create a write stream for the output zip file
    console.log('Creating zip write stream for ' + zipFilePath);
    const output = createWriteStream(zipFilePath);
    const archive = archiver('zip', {
      zlib: { level: 9 }, // Sets the compression level.
    });

    // listen for all archive data to be written
    output.on('close', function () {
      console.log(archive.pointer() + ' total bytes');
      console.log(`archiver for ${zipFilePath} has been finalized and the output file descriptor has closed.`);
      resolve();
    });

    // handle errors
    output.on('error', reject);

    // pipe archive data to the file
    archive.pipe(output);

    // append files to the archive
    Object.keys(fileMapping).forEach((filePath) => {
      const fileName = path.join(pluginId, fileMapping[filePath]); // get the file name
      archive.append(createReadStream(filePath), { name: fileName, mode: 0o755 });
    });

    // finalize the archive
    archive.finalize();
  });
}
