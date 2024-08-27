import { dirname } from "path";
import { fileURLToPath } from "url";

globalThis.__filename = fileURLToPath(import.meta.url);
globalThis.__dirname = dirname(__filename);
