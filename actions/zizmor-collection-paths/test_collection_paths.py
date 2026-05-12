"""Unit tests for collection_paths (security-appsec#326). Run: python3 -m unittest discover -v"""

import shutil
import tempfile
import unittest
from pathlib import Path

import collection_paths


class NormalizePrefixTests(unittest.TestCase):
    def test_strips_glob_suffixes(self) -> None:
        self.assertEqual(collection_paths.normalize_prefix_line("ksonnet/vendor/**/*"), "ksonnet/vendor")
        self.assertEqual(collection_paths.normalize_prefix_line("terraform/modules/foo/*"), "terraform/modules/foo")

    def test_comments_and_blank(self) -> None:
        self.assertIsNone(collection_paths.normalize_prefix_line(""))
        self.assertIsNone(collection_paths.normalize_prefix_line("  # ignore"))
        self.assertIsNone(collection_paths.normalize_prefix_line("# full line"))

    def test_rejects_unescaped_glob(self) -> None:
        self.assertIsNone(collection_paths.normalize_prefix_line("foo/*/bar"))


class ParsePrefixesTests(unittest.TestCase):
    def test_parse_multiline(self) -> None:
        text = """
# skip vendor
ksonnet/vendor

terraform/modules/github.com/github-aws-runners/**/*
"""
        got = collection_paths.parse_prefixes_from_ignore(text)
        self.assertEqual(got, ["ksonnet/vendor", "terraform/modules/github.com/github-aws-runners"])


class CollectPathsTests(unittest.TestCase):
    def setUp(self) -> None:
        self.tmp = Path(tempfile.mkdtemp())

    def tearDown(self) -> None:
        shutil.rmtree(self.tmp, ignore_errors=True)

    def _write(self, rel: str, content: str = "on: push\njobs: x: {runs-on: ubuntu-latest, steps: [{run: echo}]}\n") -> None:
        p = self.tmp / rel
        p.parent.mkdir(parents=True, exist_ok=True)
        p.write_text(content, encoding="utf-8")

    def test_includes_first_party_workflow(self) -> None:
        self._write(".github/workflows/ci.yml")
        out = self.tmp / "out.txt"
        n = collection_paths.collect_paths(self.tmp, ["ksonnet/vendor"], out)
        self.assertEqual(n, 1)
        body = out.read_text(encoding="utf-8")
        self.assertIn("./.github/workflows/ci.yml", body)

    def test_skips_nested_workflow_under_prefix(self) -> None:
        self._write(".github/workflows/ci.yml")
        self._write("ksonnet/vendor/pkg/.github/workflows/nested.yml")
        out = self.tmp / "out.txt"
        n = collection_paths.collect_paths(self.tmp, ["ksonnet/vendor"], out)
        self.assertEqual(n, 1)
        body = out.read_text(encoding="utf-8")
        self.assertNotIn("nested.yml", body)
        self.assertEqual(body.count("./"), 1)

    def test_dependabot_at_root(self) -> None:
        self._write(".github/dependabot.yml", "version: 2\nupdates: []\n")
        out = self.tmp / "out.txt"
        n = collection_paths.collect_paths(self.tmp, [], out)
        self.assertGreaterEqual(n, 1)
        body = out.read_text(encoding="utf-8")
        self.assertIn("dependabot.yml", body)

    def test_dependabot_not_outside_dot_github(self) -> None:
        self._write("vendor/dep/.github/dependabot.yml", "version: 2\nupdates: []\n")
        self._write(".github/workflows/ci.yml")
        out = self.tmp / "out.txt"
        n = collection_paths.collect_paths(self.tmp, [], out)
        body = out.read_text(encoding="utf-8")
        self.assertNotIn("dependabot", body)
        self.assertIn("./.github/workflows/ci.yml", body)
        self.assertEqual(n, 1)


class UnsafePrefixTests(unittest.TestCase):
    def test_rejects_parent_segments(self) -> None:
        self.assertIsNone(collection_paths.normalize_prefix_line("../foo"))
        self.assertIsNone(collection_paths.normalize_prefix_line("foo/../bar"))

    def test_rejects_absolute(self) -> None:
        self.assertIsNone(collection_paths.normalize_prefix_line("/etc/passwd"))
