import os
from pathlib import Path
import subprocess
import tempfile
import unittest


ACTION = Path(__file__).resolve().parent


class GenerateClientTest(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)
        self.bin = self.root / "bin"
        self.bin.mkdir()
        self.env = {
            **os.environ,
            "PATH": f"{self.bin}:{os.environ['PATH']}",
            "OUTPUT_DIR": str(self.root / "clients with spaces"),
            "PACKAGE_NAME": "sample",
            "SPEC_PATH": str(self.root / "spec with spaces.yaml"),
            "REPO_NAME": "public-clients",
            "GITHUB_ACTION_PATH": str(ACTION),
            "TEST_LOG": str(self.root / "commands"),
        }
        self.stub("java", '''
if [ "${FAIL_GENERATION:-}" = 1 ]; then exit 42; fi
while [ "$#" -gt 0 ]; do
  if [ "$1" = -o ]; then mkdir -p "$2"; touch "$2/generated"; fi
  shift
done
''')
        self.stub("npm", '''
printf 'npm %s in %s\n' "$*" "$PWD" >> "$TEST_LOG"
if [ "${FAIL_BUILD:-}" = 1 ] && [ "$1" = run ]; then exit 43; fi
''')
        self.stub("go", 'printf "go %s\\n" "$*" >> "$TEST_LOG"')
        self.stub("goimports", ':')

    def stub(self, name, body):
        path = self.bin / name
        path.write_text("#!/usr/bin/env bash\nset -eu\n" + body + "\n")
        path.chmod(0o755)

    def generate(self, **env):
        return subprocess.run(
            ["bash", str(ACTION / "generate.sh")],
            cwd=self.root,
            env={**self.env, **env},
            capture_output=True,
            text=True,
        )

    def test_default_still_generates_go_without_npm(self):
        result = self.generate()
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertTrue((Path(self.env["OUTPUT_DIR"]) / "go/sample/generated").exists())
        self.assertEqual((self.root / "commands").read_text(), "go mod tidy\n")

    def test_javascript_replaces_only_its_package_and_builds(self):
        output = Path(self.env["OUTPUT_DIR"])
        for name in ["js/sample/stale", "js/other/keep", "go/sample/keep"]:
            path = output / name
            path.parent.mkdir(parents=True, exist_ok=True)
            path.touch()
        result = self.generate(CLIENT_LANGUAGE="javascript")
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertFalse((output / "js/sample/stale").exists())
        self.assertTrue((output / "js/sample/generated").exists())
        self.assertTrue((output / "js/other/keep").exists())
        self.assertTrue((output / "go/sample/keep").exists())
        commands = (self.root / "commands").read_text()
        self.assertIn("npm install --ignore-scripts --no-audit --no-fund", commands)
        self.assertIn(f"npm run build in {output / 'js/sample'}", commands)
        self.assertNotIn("go mod", commands)

    def test_generation_failure_does_not_install_or_build(self):
        result = self.generate(CLIENT_LANGUAGE="javascript", FAIL_GENERATION="1")
        self.assertEqual(result.returncode, 42)
        self.assertFalse((self.root / "commands").exists())

    def test_build_failure_is_propagated(self):
        result = self.generate(CLIENT_LANGUAGE="javascript", FAIL_BUILD="1")
        self.assertEqual(result.returncode, 43)

    def test_unknown_language_fails_before_generation(self):
        result = self.generate(CLIENT_LANGUAGE="typo")
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("Unsupported client language", result.stderr)
        self.assertFalse(Path(self.env["OUTPUT_DIR"]).exists())


if __name__ == "__main__":
    unittest.main()
