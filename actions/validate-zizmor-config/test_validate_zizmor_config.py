"""Unit tests for validate_zizmor_config policy logic."""

import unittest

import yaml

from validate_zizmor_config import UniqueKeyFullLoader, collect_violations


class CollectViolationsTests(unittest.TestCase):
    def test_parsed_none_ok(self) -> None:
        self.assertEqual(collect_violations(None), [])

    def test_no_rules_key_ok(self) -> None:
        data = yaml.safe_load("other: true\n")
        self.assertEqual(collect_violations(data), [])

    def test_empty_rules_ok(self) -> None:
        data = yaml.safe_load("rules: {}\n")
        self.assertEqual(collect_violations(data), [])

    def test_allows_grafana_style_unpinned(self) -> None:
        text = """
rules:
  unpinned-uses:
    config:
      policies:
        actions/*: any
        grafana/*: any
"""
        data = yaml.safe_load(text)
        self.assertEqual(collect_violations(data), [])

    def test_rejects_insecure_commands(self) -> None:
        data = yaml.safe_load(
            "rules:\n  insecure-commands:\n    ignore: [x.yml]\n",
        )
        v = collect_violations(data)
        self.assertEqual(len(v), 1)
        self.assertIn("insecure-commands", v[0])

    def test_rejects_template_injection(self) -> None:
        data = yaml.safe_load("rules:\n  template-injection:\n    disable: true\n")
        v = collect_violations(data)
        self.assertEqual(len(v), 1)
        self.assertIn("template-injection", v[0])

    def test_rejects_impostor_commit(self) -> None:
        data = yaml.safe_load("rules:\n  impostor-commit: {}\n")
        self.assertTrue(any("impostor-commit" in m for m in collect_violations(data)))

    def test_rejects_known_vulnerable_actions(self) -> None:
        data = yaml.safe_load("rules:\n  known-vulnerable-actions:\n    ignore: []\n")
        v = collect_violations(data)
        self.assertEqual(len(v), 1)
        self.assertIn("known-vulnerable-actions", v[0])

    def test_rejects_ref_confusion(self) -> None:
        data = yaml.safe_load("rules:\n  ref-confusion:\n    disable: true\n")
        v = collect_violations(data)
        self.assertEqual(len(v), 1)
        self.assertIn("ref-confusion", v[0])

    def test_multiple_violations_in_one_config(self) -> None:
        text = """
rules:
  insecure-commands:
    ignore: [a.yml]
  template-injection:
    ignore: [b.yml]
  unpinned-uses:
    disable: true
"""
        data = yaml.safe_load(text)
        v = collect_violations(data)
        self.assertGreaterEqual(len(v), 3, msg=v)
        joined = " ".join(v)
        self.assertIn("insecure-commands", joined)
        self.assertIn("template-injection", joined)
        self.assertIn("unpinned-uses.disable", joined)

    def test_rejects_unpinned_disable(self) -> None:
        data = yaml.safe_load("rules:\n  unpinned-uses:\n    disable: true\n")
        v = collect_violations(data)
        self.assertTrue(any("disable" in m for m in v))

    def test_rejects_star_any_policy(self) -> None:
        data = yaml.safe_load(
            'rules:\n  unpinned-uses:\n    config:\n      policies:\n        "*": any\n',
        )
        v = collect_violations(data)
        self.assertTrue(any("*" in m or "any" in m for m in v))

    def test_duplicate_mapping_keys_rejected_by_loader(self) -> None:
        text = "rules:\n  insecure-commands:\n    x: 1\n  insecure-commands:\n    y: 2\n"
        with self.assertRaises(yaml.YAMLError):
            yaml.load(text, Loader=UniqueKeyFullLoader)


if __name__ == "__main__":
    unittest.main()
