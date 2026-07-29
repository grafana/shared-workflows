import json
import tempfile
import unittest
from pathlib import Path
from unittest import mock

import run_zizmor


class ZizmorCmdTests(unittest.TestCase):
    def test_extra_args_use_shlex(self) -> None:
        env = {
            "ZIZMOR_VERSION": "1.24.1",
            "MIN_SEVERITY": "low",
            "MIN_CONFIDENCE": "low",
            "ZIZMOR_CACHE_DIR": "/tmp/z",
            "ZIZMOR_EXTRA_ARGS": '--audit "foo bar"',
        }
        with mock.patch.dict("os.environ", env, clear=False):
            cmd = run_zizmor._zizmor_cmd("plain", [".github/workflows/ci.yml"])
        self.assertIn("--audit", cmd)
        idx = cmd.index("--audit")
        self.assertEqual(cmd[idx + 1], "foo bar")


class MergeSarifTests(unittest.TestCase):
    def test_merge_two_parts_single_run(self) -> None:
        with tempfile.TemporaryDirectory() as d:
            dpath = Path(d)
            p1 = dpath / "a.sarif"
            p2 = dpath / "b.sarif"
            out = dpath / "out.sarif"
            p1.write_text(
                json.dumps(
                    {
                        "version": "2.1.0",
                        "$schema": "https://json.schemastore.org/sarif-2.1.0.json",
                        "runs": [
                            {
                                "tool": {
                                    "driver": {
                                        "name": "zizmor",
                                        "rules": [{"id": "template-injection"}],
                                    }
                                },
                                "results": [
                                    {"ruleId": "template-injection", "ruleIndex": 0, "message": {"text": "a"}}
                                ],
                            }
                        ],
                    }
                ),
                encoding="utf-8",
            )
            p2.write_text(
                json.dumps(
                    {
                        "version": "2.1.0",
                        "runs": [
                            {
                                "tool": {
                                    "driver": {
                                        "name": "zizmor",
                                        "rules": [{"id": "artipacked"}],
                                    }
                                },
                                "results": [{"ruleId": "artipacked", "ruleIndex": 0, "message": {"text": "b"}}],
                            }
                        ],
                    }
                ),
                encoding="utf-8",
            )
            run_zizmor._merge_sarif_parts([p1, p2], out)
            doc = json.loads(out.read_text(encoding="utf-8"))
            self.assertEqual(len(doc["runs"]), 1)
            run = doc["runs"][0]
            rules = run["tool"]["driver"]["rules"]
            self.assertEqual([r["id"] for r in rules], ["template-injection", "artipacked"])
            results = run["results"]
            self.assertEqual(len(results), 2)
            by_id = {r["ruleId"]: r for r in results}
            self.assertEqual(by_id["template-injection"]["ruleIndex"], 0)
            self.assertEqual(by_id["artipacked"]["ruleIndex"], 1)


class SarifEmptyExplicitTests(unittest.TestCase):
    def test_writes_minimal_sarif_when_explicit_paths_empty(self) -> None:
        with tempfile.TemporaryDirectory() as d:
            out = Path(d) / "r.sarif"
            env = {
                "USE_EXPLICIT_PATHS": "true",
                "PATHS_LIST": str(Path(d) / "empty.txt"),
                "RUNNER_TEMP": d,
                "ZIZMOR_VERSION": "1.24.1",
                "MIN_SEVERITY": "low",
                "MIN_CONFIDENCE": "low",
                "ZIZMOR_CACHE_DIR": str(Path(d) / "cache"),
            }
            (Path(d) / "empty.txt").write_text("\n\n", encoding="utf-8")
            with mock.patch.dict("os.environ", env, clear=False):
                rc = run_zizmor._run_sarif(None, 400, out)
            self.assertEqual(rc, 0)
            doc = json.loads(out.read_text(encoding="utf-8"))
            self.assertEqual(doc.get("version"), "2.1.0")
            self.assertEqual(doc.get("runs"), [])


class PlainGithubOutputTests(unittest.TestCase):
    def test_plain_github_output_empty_explicit_paths(self) -> None:
        with tempfile.TemporaryDirectory() as d:
            gh = Path(d) / "gh.txt"
            empty = Path(d) / "empty.txt"
            empty.write_text("\n\n", encoding="utf-8")
            env = {
                "GITHUB_OUTPUT": str(gh),
                "USE_EXPLICIT_PATHS": "true",
                "PATHS_LIST": str(empty),
                "ZIZMOR_VERSION": "1.24.1",
                "MIN_SEVERITY": "low",
                "MIN_CONFIDENCE": "low",
            }
            with mock.patch.dict("os.environ", env, clear=False):
                rc = run_zizmor.main(["run_zizmor.py", "plain-github-output", "--batch-size", "400"])
            self.assertEqual(rc, 0)
            body = gh.read_text(encoding="utf-8")
            self.assertIn("zizmor-results<<EOF", body)
            self.assertIn("EOF\n", body)
            self.assertIn("zizmor-exit-code=0", body)

    def test_plain_closes_heredoc_on_crash(self) -> None:
        with tempfile.TemporaryDirectory() as d:
            gh = Path(d) / "gh.txt"
            paths = Path(d) / "paths.txt"
            paths.write_text("./.github/workflows/ci.yml\n", encoding="utf-8")
            env = {
                "GITHUB_OUTPUT": str(gh),
                "USE_EXPLICIT_PATHS": "true",
                "PATHS_LIST": str(paths),
                "ZIZMOR_VERSION": "1.24.1",
                "MIN_SEVERITY": "low",
                "MIN_CONFIDENCE": "low",
                "ZIZMOR_CACHE_DIR": str(Path(d) / "cache"),
            }

            class FakeProc:
                returncode = 1
                stdout = "boom\n"

            with mock.patch.dict("os.environ", env, clear=False):
                with mock.patch("run_zizmor.subprocess.run", return_value=FakeProc()):
                    rc = run_zizmor._run_plain(["./.github/workflows/ci.yml"], 400)
            self.assertEqual(rc, 1)
            body = gh.read_text(encoding="utf-8")
            self.assertIn("zizmor-results<<EOF", body)
            self.assertTrue(body.rstrip().endswith("zizmor-exit-code=1") or "zizmor-exit-code=1\n" in body)
            self.assertIn("\nEOF\n", body)
