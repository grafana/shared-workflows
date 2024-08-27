/**
 * From: https://github.com/vercel/ncc/issues/791#issuecomment-1731283695
 *
 * This ESM module polyfills "require".
 *
 * It is needed e.g. when bundling ESM scripts via "@vercel/ncc" because of
 * https://github.com/vercel/ncc/issues/791.
 */

import "./polyfill-dirname-filename";

import { createRequire } from "module";

globalThis.require = createRequire(__filename);
