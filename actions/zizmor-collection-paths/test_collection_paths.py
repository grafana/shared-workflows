import os
import shutil
import sys
import tempfile
import unittest
from pathlib import Path
from unittest import mock

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

    def test_rejects_double_slash_segment(self) -> None:
        self.assertIsNone(collection_paths.normalize_prefix_line("a//b"))

    def test_normalizes_backslashes(self) -> None:
        self.assertEqual(collection_paths.normalize_prefix_line("ksonnet\\vendor"), "ksonnet/vendor")


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

    def test_skips_helper_checkout_tree(self) -> None:
        self._write(".github/workflows/ci.yml")
        self._write(
            "_shared-workflows-zizmor/actions/example/action.yaml",
            "name: x\nruns:\n  using: composite\n  steps: []\n",
        )
        out = self.tmp / "out.txt"
        n = collection_paths.collect_paths(self.tmp, [], out)
        body = out.read_text(encoding="utf-8")
        self.assertIn("./.github/workflows/ci.yml", body)
        self.assertNotIn("_shared-workflows-zizmor", body)
        self.assertEqual(n, 1)

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

    @unittest.skipUnless(os.name == "posix", "requires POSIX path joining with absolute second operand")
    def test_collect_paths_rejects_absolute_resolved_prefix(self) -> None:
        self._write(".github/workflows/ci.yml")
        out = self.tmp / "out.txt"
        with self.assertRaises(ValueError) as ctx:
            collection_paths.collect_paths(self.tmp, ["/etc"], out)
        self.assertIn("outside repo root", str(ctx.exception).lower())


class UnsafePrefixTests(unittest.TestCase):
    def test_rejects_parent_segments(self) -> None:
        self.assertIsNone(collection_paths.normalize_prefix_line("../foo"))
        self.assertIsNone(collection_paths.normalize_prefix_line("foo/../bar"))

    def test_rejects_absolute(self) -> None:
        self.assertIsNone(collection_paths.normalize_prefix_line("/etc/passwd"))


class CliTests(unittest.TestCase):

    def setUp(self) -> None:
        self.tmp = Path(tempfile.mkdtemp())

    def tearDown(self) -> None:
        shutil.rmtree(self.tmp, ignore_errors=True)

    def test_main_missing_ignore_file(self) -> None:
        gh = self.tmp / "gh.txt"
        root = self.tmp / "repo"
        (root / ".github").mkdir(parents=True)
        ignore = root / ".github" / "zizmor-collection-ignore"
        paths_out = self.tmp / "paths.txt"
        argv = [
            "collection_paths.py",
            "--repo-root",
            str(root),
            "--ignore-file",
            str(ignore),
            "--paths-out",
            str(paths_out),
        ]
        with mock.patch.dict(os.environ, {"GITHUB_OUTPUT": str(gh)}, clear=False):
            with mock.patch.object(sys, "argv", argv):
                rc = collection_paths.main()
        self.assertEqual(rc, 0)
        self.assertIn("use_explicit_paths=false", gh.read_text(encoding="utf-8"))

    def test_main_ignore_only_comments(self) -> None:
        gh = self.tmp / "gh2.txt"
        root = self.tmp / "repo2"
        (root / ".github").mkdir(parents=True)
        ignore = root / ".github" / "zizmor-collection-ignore"
        ignore.write_text("# nothing\n\n", encoding="utf-8")
        paths_out = self.tmp / "paths2.txt"
        argv = [
            "collection_paths.py",
            "--repo-root",
            str(root),
            "--ignore-file",
            str(ignore),
            "--paths-out",
            str(paths_out),
        ]
        with mock.patch.dict(os.environ, {"GITHUB_OUTPUT": str(gh)}, clear=False):
            with mock.patch.object(sys, "argv", argv):
                rc = collection_paths.main()
        self.assertEqual(rc, 0)
        self.assertIn("use_explicit_paths=false", gh.read_text(encoding="utf-8"))

    def test_main_explicit_paths_when_prefix_skips_vendor(self) -> None:
        gh = self.tmp / "gh3.txt"
        root = self.tmp / "repo3"
        (root / ".github" / "workflows").mkdir(parents=True)
        (root / ".github" / "workflows" / "ci.yml").write_text(
            "on: push\njobs: x: {runs-on: ubuntu-latest, steps: [{run: echo}]}\n",
            encoding="utf-8",
        )
        (root / "vendor" / ".github" / "workflows").mkdir(parents=True)
        (root / "vendor" / ".github" / "workflows" / "nested.yml").write_text(
            "on: push\njobs: x: {runs-on: ubuntu-latest, steps: [{run: echo}]}\n",
            encoding="utf-8",
        )
        ignore = root / ".github" / "zizmor-collection-ignore"
        ignore.write_text("vendor\n", encoding="utf-8")
        paths_out = self.tmp / "paths3.txt"
        argv = [
            "collection_paths.py",
            "--repo-root",
            str(root),
            "--ignore-file",
            str(ignore),
            "--paths-out",
            str(paths_out),
        ]
        with mock.patch.dict(os.environ, {"GITHUB_OUTPUT": str(gh)}, clear=False):
            with mock.patch.object(sys, "argv", argv):
                rc = collection_paths.main()
        self.assertEqual(rc, 0)
        out_txt = gh.read_text(encoding="utf-8")
        self.assertIn("use_explicit_paths=true", out_txt)
        self.assertIn("paths_list=", out_txt)
        listed = paths_out.read_text(encoding="utf-8")
        self.assertIn("ci.yml", listed)
        self.assertNotIn("nested", listed)
