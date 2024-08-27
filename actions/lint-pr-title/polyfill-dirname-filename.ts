/**
 * This ESM module polyfills `__dirname` and `__filename`.
 *
 * This is necessary because ESM does not support `__dirname` and `__filename`
 * and some libraries which we use depend on these variables.
 */

import { dirname } from "path";
import { fileURLToPath } from "url";

globalThis.__filename = fileURLToPath(import.meta.url);
globalThis.__dirname = dirname(__filename);
