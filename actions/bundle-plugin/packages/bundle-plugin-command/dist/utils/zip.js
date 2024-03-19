import archiver from 'archiver';
import { createReadStream, createWriteStream, existsSync, mkdirSync } from 'fs';
import path from 'path';
export function compressFilesToZip(zipFilePath, pluginId, fileMapping) {
    return new Promise((resolve, reject) => {
        const outputDir = path.dirname(zipFilePath);
        if (!existsSync(outputDir)) {
            mkdirSync(outputDir, { recursive: true });
        }
        console.log('Creating zip write stream for ' + zipFilePath);
        const output = createWriteStream(zipFilePath);
        const archive = archiver('zip', {
            zlib: { level: 9 },
        });
        output.on('close', function () {
            console.log(archive.pointer() + ' total bytes');
            console.log(`archiver for ${zipFilePath} has been finalized and the output file descriptor has closed.`);
            resolve();
        });
        output.on('error', reject);
        archive.pipe(output);
        Object.keys(fileMapping).forEach((filePath) => {
            const fileName = path.join(pluginId, fileMapping[filePath]);
            archive.append(createReadStream(filePath), { name: fileName, mode: 0o755 });
        });
        archive.finalize();
    });
}
