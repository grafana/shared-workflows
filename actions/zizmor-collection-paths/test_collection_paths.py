"""Unit tests for collection_paths."""

import tempfile
import unittest
from pathlib import Path

import collection_paths as cp


class TestNormalize(unittest.TestCase):
    def test_trim_and_strip_leading_slash(self) -> None:
        self.assertEqual(cp.normalize_prefix_line("  /foo/bar/  "), "foo/bar")

    def test_strip_suffixes(self) -> None:
        self.assertEqual(cp.normalize_prefix_line("vendor/**"), "vendor")
        self.assertEqual(cp.normalize_prefix_line("vendor/*"), "vendor")

    def test_skip_glob(self) -> None:
        self.assertIsNone(cp.normalize_prefix_line("foo*bar"))

    def test_skip_comment(self) -> None:
        self.assertIsNone(cp.normalize_prefix_line("# x"))
        self.assertIsNone(cp.normalize_prefix_line(""))

    def test_parse_file_body(self) -> None:
        body = """
# c
/ksonnet/vendor/

  /terraform/modules/github.com/github-aws-runners/

"""
        self.assertEqual(
            cp.parse_prefixes_from_ignore(body),
            ["ksonnet/vendor", "terraform/modules/github.com/github-aws-runners"],
        )


class TestFindIntegration(unittest.TestCase):
    def test_prune_excludes_nested_workflow(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            (root / ".github" / "workflows").mkdir(parents=True)
            (root / ".github" / "workflows" / "keep.yml").write_text("# root\n", encoding="utf-8")
            (root / ".github" / "workflows" / "keep-nested").mkdir(parents=True)
            (root / ".github" / "workflows" / "keep-nested" / "skip.yaml").write_text(
                "# nested workflow\n", encoding="utf-8"
            )
            (root / ".github" / "dependabot.yml").write_text("version: 2\n", encoding="utf-8")
            (root / "actions").mkdir(parents=True)
            (root / "actions" / "action.yml").write_text("name: x\n", encoding="utf-8")
            (root / "README.md").write_text("# not scanned\n", encoding="utf-8")
            vend = root / "ksonnet" / "vendor" / "upstream" / ".github" / "workflows"
            vend.mkdir(parents=True)
            (vend / "nested.yml").write_text("# nested\n", encoding="utf-8")
            (root / "ksonnet" / "vendor" / "upstream" / "action.yaml").write_text(
                "name: vendored\n", encoding="utf-8"
            )

            ignore = root / ".github" / "zizmor-collection-ignore"
            ignore.write_text("ksonnet/vendor\n", encoding="utf-8")

            prefixes = cp.parse_prefixes_from_ignore(ignore.read_text(encoding="utf-8"))
            out = root / "list.txt"
            cp.collect_paths(root, prefixes, out)
            lines = [ln.strip() for ln in out.read_text(encoding="utf-8").splitlines() if ln.strip()]
            rels = sorted(
                Path(ln[2:] if ln.startswith("./") else ln).as_posix() for ln in lines
            )
            self.assertIn(".github/workflows/keep.yml", rels)
            self.assertIn(".github/dependabot.yml", rels)
            self.assertIn("actions/action.yml", rels)
            self.assertNotIn(".github/workflows/keep-nested/skip.yaml", rels)
            self.assertNotIn(
                "ksonnet/vendor/upstream/.github/workflows/nested.yml",
                rels,
            )
            self.assertNotIn("ksonnet/vendor/upstream/action.yaml", rels)


if __name__ == "__main__":
    unittest.main()
