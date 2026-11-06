"""Real generator/build checks; set OPENAPI_GENERATOR_JAR to enable locally."""

import json
import os
from pathlib import Path
import subprocess
import tempfile
import unittest


ACTION = Path(__file__).resolve().parent


@unittest.skipUnless(os.environ.get("OPENAPI_GENERATOR_JAR"), "generator jar required")
class JavaScriptPackageTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        temp = tempfile.TemporaryDirectory()
        cls.addClassCleanup(temp.cleanup)
        cls.root = Path(temp.name)
        (cls.root / "openapi-generator-cli.jar").symlink_to(
            Path(os.environ["OPENAPI_GENERATOR_JAR"]).resolve()
        )
        spec = {
            "openapi": "3.0.3",
            "info": {"title": "Sample", "version": "1.0.0"},
            "paths": {
                "/item": {
                    "get": {
                        "operationId": "getItem",
                        "responses": {
                            "200": {
                                "description": "An item",
                                "content": {
                                    "application/json": {
                                        "schema": {"$ref": "#/components/schemas/Item"}
                                    }
                                },
                            }
                        },
                    }
                }
            },
            "components": {
                "schemas": {
                    "Item": {
                        "type": "object",
                        "required": ["display_name"],
                        "properties": {"display_name": {"type": "string"}},
                    }
                }
            },
        }
        (cls.root / "spec.json").write_text(json.dumps(spec))
        result = subprocess.run(
            ["bash", str(ACTION / "generate.sh")],
            cwd=cls.root,
            env={
                **{key: value for key, value in os.environ.items() if key != "REPO_NAME"},
                "CLIENT_LANGUAGE": "javascript",
                "OUTPUT_DIR": str(cls.root),
                "PACKAGE_NAME": "sample",
                "GITHUB_REPOSITORY": "grafana/public-clients",
                "SPEC_PATH": str(cls.root / "spec.json"),
                "GITHUB_ACTION_PATH": str(ACTION),
            },
            capture_output=True,
            text=True,
        )
        if result.returncode:
            raise RuntimeError(result.stdout + result.stderr)
        cls.package = cls.root / "js/sample"
        cls.consumer = cls.root / "consumer"
        (cls.consumer / "node_modules").mkdir(parents=True)
        (cls.consumer / "package.json").write_text('{"type":"module"}')
        (cls.consumer / "node_modules/public-clients-sample").symlink_to(cls.package)

    def run_checked(self, *command):
        result = subprocess.run(
            command, cwd=self.consumer, capture_output=True, text=True
        )
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)

    def test_single_esm_build(self):
        package = json.loads((self.package / "package.json").read_text())
        self.assertEqual(package["type"], "module")
        self.assertIs(package["sideEffects"], False)
        self.assertFalse((self.package / "tsconfig.esm.json").exists())
        self.assertFalse((self.package / "dist/esm").exists())
        self.assertTrue((self.package / "package-lock.json").exists())

    def test_import_and_require_package_entrypoint(self):
        for mode, statement in [
            ("module", "import { Configuration, DefaultApi } from 'public-clients-sample';"),
            ("commonjs", "const { Configuration, DefaultApi } = require('public-clients-sample');"),
        ]:
            with self.subTest(mode=mode):
                self.run_checked(
                    "node", "--input-type=" + mode, "-e",
                    statement + """
const api = new DefaultApi(new Configuration({
  basePath: 'https://example.test',
  fetchApi: async (url) => {
    if (url !== 'https://example.test/item') throw new Error(url);
    return Response.json({ display_name: 'sample' });
  },
}));
api.getItem().then(item => {
  if (item.display_name !== 'sample') throw new Error(JSON.stringify(item));
}).catch(error => { console.error(error); process.exitCode = 1; });
""",
                )

    def test_consumer_type_resolution(self):
        (self.consumer / "index.ts").write_text(
            "import { Configuration, DefaultApi, type Item } from 'public-clients-sample';\n"
            "const api = new DefaultApi(new Configuration());\n"
            "const item: Promise<Item> = api.getItem();\n"
        )
        for resolution, module in [
            ("node10", "ESNext"), ("bundler", "ESNext"), ("NodeNext", "NodeNext")
        ]:
            with self.subTest(resolution=resolution):
                self.run_checked(
                    "node", str(self.package / "node_modules/typescript/bin/tsc"),
                    "--noEmit", "--strict", "--target", "ES2015",
                    "--module", module, "--moduleResolution", resolution, "index.ts",
                )


if __name__ == "__main__":
    unittest.main()
