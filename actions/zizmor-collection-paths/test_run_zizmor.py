"""Unit tests for run_zizmor (security-appsec#326). Run: python3 -m unittest discover -v"""

import json
import tempfile
import unittest
from pathlib import Path
from unittest import mock

import run_zizmor


class MergeSarifTests(unittest.TestCase):
    def test_merge_two_parts(self) -> None:
        with tempfile.TemporaryDirectory() as d:
            dpath = Path(d)
            p1 = dpath / "a.sarif"
            p2 = dpath / "b.sarif"
            out = dpath / "out.sarif"
            p1.write_text(
                json.dumps({"version": "2.1.0", "runs": [{"tool": {"driver": {"name": "a"}}}]}),
                encoding="utf-8",
            )
            p2.write_text(
                json.dumps({"version": "2.1.0", "runs": [{"tool": {"driver": {"name": "b"}}}]}),
                encoding="utf-8",
            )
            run_zizmor._merge_sarif_parts([p1, p2], out)
            doc = json.loads(out.read_text(encoding="utf-8"))
            self.assertEqual(len(doc["runs"]), 2)


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
                rc = run_zizmor._sarif(400, out)
            self.assertEqual(rc, 0)
            doc = json.loads(out.read_text(encoding="utf-8"))
            self.assertEqual(doc.get("version"), "2.1.0")
            self.assertEqual(doc.get("runs"), [])
