/**
 * From: https://github.com/vercel/ncc/issues/791#issuecomment-1731283695
 *
 * This ESM module polyfills "require".
 *
 * It is needed e.g. when bundling ESM scripts via "@vercel/ncc" because of
 * https://github.com/vercel/ncc/issues/791.
 */
import { createRequire } from "module";
import { resolve } from "path";
import url from "url";

function createRequireWithFallback(): NodeRequire {
  const require = createRequire(url.fileURLToPath(import.meta.url));
  const customPaths = ["node_modules"];

  function enhancedRequire(id: string) {
    try {
      // eslint-disable-next-line @typescript-eslint/no-unsafe-return
      return require(id);
    } catch (error) {
      // Is it ERR_MODULE_NOT_FOUND?
      if (
        error instanceof Error &&
        "code" in error &&
        error.code === "MODULE_NOT_FOUND"
      ) {
        for (const customPath of customPaths) {
          try {
            // eslint-disable-next-line @typescript-eslint/no-unsafe-return
            return require(resolve(customPath, id));
          } catch (innerError) {
            if (
              innerError instanceof Error &&
              "code" in innerError &&
              innerError.code !== "MODULE_NOT_FOUND"
            )
              throw innerError;
          }
        }
      }
      throw error;
    }
  }

  Object.setPrototypeOf(enhancedRequire, require);

  return enhancedRequire as NodeRequire;
}

globalThis.require = createRequireWithFallback();
