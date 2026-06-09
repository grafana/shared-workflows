import eslint from "@eslint/js";
import eslintPluginJest from "eslint-plugin-jest";
import eslintPluginPrettierRecommended from "eslint-plugin-prettier/recommended";
import js from "@eslint/js";
import tseslint from "typescript-eslint";
import { defineConfig } from "eslint/config";

export default defineConfig([
  js.configs.recommended,
  eslint.configs.recommended,
  tseslint.configs.strict,
  eslintPluginPrettierRecommended,
  {
    // Allow unused vars if they start with an underscore
    rules: {
      "@typescript-eslint/no-unused-vars": [
        "error",
        {
          varsIgnorePattern: "^_",
          argsIgnorePattern: "^_",
        },
      ],

      "@typescript-eslint/restrict-template-expressions": [
        "error",
        {
          allowNumber: true,
        },
      ],
    },
    languageOptions: {
      parserOptions: {
        projectService: true,
      },
    },
  },
  {
    files: ["**/*.js", "**/*.mjs"],
    ...tseslint.configs.disableTypeChecked,
  },
  {
    // The create-github-app-token action ships as a plain node24 action that
    // runs straight from a git checkout with no bundler, so its sources must
    // use CommonJS (require/module.exports) and rely on Node.js runtime
    // globals. Declare those here instead of excluding the files from linting.
    files: ["actions/create-github-app-token/**/*.js"],
    languageOptions: {
      sourceType: "commonjs",
      globals: {
        require: "readonly",
        module: "writable",
        exports: "writable",
        __dirname: "readonly",
        __filename: "readonly",
        process: "readonly",
        console: "readonly",
        Buffer: "readonly",
        URL: "readonly",
        URLSearchParams: "readonly",
        fetch: "readonly",
        setTimeout: "readonly",
        clearTimeout: "readonly",
        setInterval: "readonly",
        clearInterval: "readonly",
      },
    },
    rules: {
      "@typescript-eslint/no-require-imports": "off",
    },
  },
  {
    files: ["test/**/*.ts"],
    ...eslintPluginJest.configs["flat/recommended"],
  },
  {
    ignores: ["coverage/", "dist/", "node_modules/"],
  },
]);
