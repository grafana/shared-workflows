"""Exercise the actual trusted Python embedded in reusable-zizmor.yml."""

import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest

import yaml


ROOT = Path(__file__).resolve().parents[2]
WORKFLOW = yaml.safe_load((ROOT / ".github/workflows/reusable-zizmor.yml").read_text())
STEP = next(s for s in WORKFLOW["jobs"]["analysis"]["steps"] if s.get("id") == "renovate-policy")
SCRIPT = STEP["run"].split("<<'PY'\n", 1)[1].rsplit("\nPY", 1)[0]
CONFIG_FILES = (
    "renovate.json", "renovate.jsonc", "renovate.json5",
    ".github/renovate.json", ".github/renovate.jsonc", ".github/renovate.json5",
    ".gitlab/renovate.json", ".gitlab/renovate.jsonc", ".gitlab/renovate.json5",
    ".renovaterc", ".renovaterc.json", ".renovaterc.jsonc", ".renovaterc.json5",
    "package.json",
)


class RenovatePolicyTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name) / "repo"
        self.root.mkdir()
        self.summary = Path(self.temp.name) / "summary"

    def write(self, filename, contents):
        path = self.root / filename
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(contents, encoding="utf-8")
        return path

    def check(self, mode="enforce", event="pull_request"):
        return subprocess.run(
            [sys.executable, "-I", "-"],
            input=SCRIPT,
            text=True,
            capture_output=True,
            cwd=self.root,
            timeout=15,
            env={
                **os.environ,
                "GITHUB_WORKSPACE": str(self.root),
                "GITHUB_STEP_SUMMARY": str(self.summary),
                "GITHUB_EVENT_NAME": event,
                "RENOVATE_POLICY": mode,
            },
        )

    def assert_result(self, code, mode="enforce", event="pull_request", contains=""):
        result = self.check(mode, event)
        self.assertEqual(result.returncode, code, result.stdout + result.stderr)
        self.assertIn(contains, result.stdout)
        self.assertEqual(result.stderr, "")
        return result

    def test_direct_disablement_in_all_locations(self):
        for filename in CONFIG_FILES:
            with self.subTest(filename=filename):
                config = {"enabled": False}
                if filename == "package.json":
                    config = {"renovate": config}
                path = self.write(filename, json.dumps(config))
                self.assert_result(1, contains=f"{filename} disables Renovate")
                path.unlink()

    def test_all_direct_disable_forms(self):
        for config in (
            {"enabled": False},
            {"force": {"enabled": False}},
            {"extends": ["config:recommended", ":disableRenovate"]},
        ):
            with self.subTest(config=config):
                self.write("renovate.json", json.dumps(config))
                result = self.assert_result(1, contains="::error title=Renovate policy::")
                self.assertIn("/issues/791627", result.stdout)

    def test_json5_and_scoped_exclusions(self):
        self.write(".github/renovate.json5", """{
          // Scoped disablement remains allowed.
          enabled: true,
          npm: {enabled: false},
          packageRules: [{matchPackageNames: ['foo'], enabled: false}],
        }""")
        self.assert_result(0, contains="no explicit repository-wide")

    def test_inherited_presets_and_catch_all_rules_are_out_of_scope(self):
        self.write("renovate.json", json.dumps({
            "extends": ["github>example/presets"],
            "packageRules": [{"matchPackageNames": ["*"], "enabled": False}],
        }))
        self.assert_result(0, contains="Inherited presets and catch-all rules are not evaluated")

    def test_warn_is_nonblocking_and_has_summary(self):
        self.write("renovate.json", '{"enabled": false}')
        self.assert_result(0, mode="warn", contains="::warning title=Renovate policy::")
        summary = self.summary.read_text()
        self.assertIn("Mode: warn", summary)
        self.assertIn("/issues/791627", summary)

    def test_invalid_configuration_in_both_modes(self):
        for contents in (
            "not json", "null", "[]", '{"enabled": "false"}',
            '{"enabled": 0}', '{"force": null}', '{"force": {"enabled": 0}}',
            '{"extends": ":disableRenovate"}', '{"extends": [false]}',
        ):
            with self.subTest(contents=contents):
                self.write("renovate.json", contents)
                self.assert_result(2, contains="Unable to check")
                self.assert_result(0, mode="warn", contains="::warning")

    def test_off_does_not_inspect_configuration(self):
        self.write("renovate.json", "invalid")
        result = self.assert_result(0, mode="off")
        self.assertEqual(result.stdout, "")
        self.assertFalse(self.summary.exists())

    def test_invalid_mode_fails_without_echoing_input(self):
        result = self.assert_result(2, mode="invalid\n::notice::INJECTED", contains="must be off, warn, or enforce")
        self.assertNotIn("INJECTED", result.stdout)

    def test_first_config_takes_precedence(self):
        self.write("renovate.json", "{}")
        self.write(".github/renovate.json5", "{enabled: false}")
        self.write("package.json", '{"renovate": {"enabled": false}}')
        self.assert_result(0, contains="renovate.json: no explicit")
        self.write("renovate.json", "invalid")
        self.assert_result(2, contains="Cannot parse renovate.json")
        self.write("renovate.json", '{"enabled": false}')
        self.write(".github/renovate.json5", "{}")
        self.assert_result(1)

    def test_no_local_config_and_package_without_renovate_pass(self):
        self.assert_result(0, contains="global defaults apply")
        self.write("package.json", '{"name": "test"}')
        self.assert_result(0, contains="global defaults apply")

    def test_package_must_be_json_with_object_config(self):
        for contents in ("[]", '{"renovate": null}', "{renovate: {enabled: false}}"):
            with self.subTest(contents=contents):
                self.write("package.json", contents)
                self.assert_result(2, contains="Unable to check")

    def test_symlink_file_and_parent_are_rejected(self):
        target = Path(self.temp.name) / "outside"
        target.write_text('{"enabled": false}')
        link = self.root / "renovate.json"
        link.symlink_to(target)
        self.assert_result(2, contains="must not use symlinks")
        link.unlink()
        link.symlink_to(Path(self.temp.name) / "does-not-exist")
        self.assert_result(2, contains="must not use symlinks")
        link.unlink()
        (self.root / ".github").symlink_to(Path(self.temp.name), target_is_directory=True)
        self.assert_result(2, contains="must not use symlinks")

    def test_oversized_and_nonregular_configs_are_rejected(self):
        path = self.write("renovate.json", " " * (1024 * 1024 + 1))
        self.assert_result(2, contains="no larger than 1 MiB")
        path.unlink()
        path.mkdir()
        self.assert_result(2, contains="must be a regular file")

    def test_never_executes_or_prints_repository_code(self):
        self.write("renovate.json5", '{enabled: (() => { throw "SECRET_MARKER" })()}')
        self.write("json5.py", 'raise RuntimeError("SECRET_MARKER")')
        self.write("sitecustomize.py", 'raise RuntimeError("SECRET_MARKER")')
        result = self.assert_result(2, contains="Cannot parse")
        self.assertNotIn("SECRET_MARKER", result.stdout + self.summary.read_text())

    def test_pr_merge_queue_and_push_check_proposed_tree(self):
        self.write("renovate.json", '{"enabled": false}')
        for event in ("pull_request", "merge_group", "push"):
            with self.subTest(event=event):
                self.assert_result(1, event=event)

    def test_unsupported_base_branch_context_cannot_pass_enforcement(self):
        self.assert_result(2, event="pull_request_target", contains="requires a pull_request")
        self.assert_result(0, mode="warn", event="pull_request_target", contains="::warning")


if __name__ == "__main__":
    unittest.main()
