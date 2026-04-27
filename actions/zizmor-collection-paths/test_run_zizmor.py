"""Unit tests for run_zizmor."""

from __future__ import annotations

import json
import os
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch

import run_zizmor as rz


class TestMergeSarifParts(unittest.TestCase):
    def test_merge_two_parts(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            tdir = Path(tmp)
            p1 = tdir / "1.sarif"
            p2 = tdir / "2.sarif"
            out = tdir / "out.sarif"

            p1.write_text(
                json.dumps(
                    {
                        "$schema": "s",
                        "version": "2.1.0",
                        "runs": [{"tool": {"driver": {"name": "zizmor"}}}],
                    }
                ),
                encoding="utf-8",
            )
            p2.write_text(
                json.dumps(
                    {
                        "$schema": "s",
                        "version": "2.1.0",
                        "runs": [{"invocations": [{"executionSuccessful": True}]}],
                    }
                ),
                encoding="utf-8",
            )

            rz._merge_sarif_parts([p1, p2], out)
            merged = json.loads(out.read_text(encoding="utf-8"))
            self.assertEqual(merged.get("$schema"), "s")
            self.assertEqual(merged.get("version"), "2.1.0")
            self.assertEqual(len(merged.get("runs", [])), 2)


class TestPlainGithubOutput(unittest.TestCase):
    def test_batches_and_max_exit_code(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            tdir = Path(tmp)
            paths = tdir / "paths.txt"
            paths.write_text("\n".join([f"./p{i}.yml" for i in range(5)]) + "\n", encoding="utf-8")

            gh_out = tdir / "github_output"
            env = {
                "GITHUB_OUTPUT": str(gh_out),
                "ZIZMOR_VERSION": "0.0.0",
                "MIN_SEVERITY": "low",
                "MIN_CONFIDENCE": "low",
                "ZIZMOR_CACHE_DIR": str(tdir / "cache"),
                "RUNNER_TEMP": str(tdir),
                "USE_EXPLICIT_PATHS": "true",
                "PATHS_LIST": str(paths),
            }

            calls: list[list[str]] = []

            class FakePopen:
                def __init__(self, cmd, stdout=None, stderr=None, text=False):
                    self.cmd = cmd
                    calls.append(list(cmd))
                    self.stdout = iter(["line\n"])

                def __enter__(self):
                    return self

                def __exit__(self, exc_type, exc, tb):
                    return None

                def wait(self):
                    # First batch rc=12, second batch rc=10
                    if len(calls) == 1:
                        return 12
                    if len(calls) == 2:
                        return 10
                    raise AssertionError("unexpected batch")

            with patch.dict(os.environ, env, clear=True):
                with patch("run_zizmor.subprocess.Popen", FakePopen):
                    rc = rz.cmd_plain_github_output(["--batch-size", "3"])
                    self.assertEqual(rc, 0)

            text = gh_out.read_text(encoding="utf-8")
            self.assertIn("zizmor-results<<EOF\n", text)
            self.assertIn("EOF\n", text)
            self.assertIn("zizmor-exit-code=12\n", text)
            self.assertEqual(len(calls), 2)
            self.assertEqual(calls[0][-3:], ["./p0.yml", "./p1.yml", "./p2.yml"])
            self.assertEqual(calls[1][-2:], ["./p3.yml", "./p4.yml"])


if __name__ == "__main__":
    unittest.main()
