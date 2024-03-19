import { existsSync, mkdirSync, readdirSync, statSync, readFileSync, writeFileSync } from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
export function generateFolder(prefix) {
    const randomHash = crypto.createHash('md5').update(new Date().getTime().toString()).digest('hex');
    const folderName = `${prefix}-${randomHash}`;
    if (!existsSync(folderName)) {
        mkdirSync(folderName);
    }
    else {
        throw new Error(`Folder ${folderName} already exists`);
    }
    return folderName;
}
export function listFiles(dir) {
    const out = [];
    readdirSync(dir).forEach((file) => {
        if (statSync(path.join(dir, file)).isDirectory()) {
            out.push(...listFiles(path.join(dir, file)));
        }
        else {
            out.push(path.join(dir, file));
        }
    });
    return out;
}
export function addSha1ForFiles(files) {
    files.forEach((file) => {
        const fileContent = readFileSync(file);
        const sha1 = crypto.createHash('sha1').update(fileContent).digest('hex');
        writeFileSync(`${file}.sha1`, sha1);
    });
}
export const absoluteToRelativePaths = (dir) => {
    const out = {};
    listFiles(dir).forEach((file) => {
        out[file] = file.replace(dir, '');
    });
    return out;
};
