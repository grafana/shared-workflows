#!/usr/bin/env python3
"""Fail if a repo-local zizmor.yml violates Grafana shared-workflows policy."""

import argparse
import sys
from collections.abc import Hashable
from pathlib import Path

import yaml
from yaml.constructor import ConstructorError
from yaml.loader import FullLoader
from yaml.nodes import MappingNode

# Stable prefix for CI log search (e.g. `zizmor-config-validator failed`).
_FAIL_LOG_PREFIX = "zizmor-config-validator failed:"

# Audits that must not appear under `rules` in repo-local zizmor.yml at all
# (no disable, ignore, or config — the whole block is forbidden).
_FORBIDDEN_RULE_AUDITS = (
    "insecure-commands",
    "template-injection",
    "impostor-commit",
    "known-vulnerable-actions",
    "ref-confusion",
)


class UniqueKeyFullLoader(FullLoader):
    """Like FullLoader but reject duplicate keys in any mapping (YAML 1.2 forbids them)."""

    def construct_mapping(self, node, deep=False):
        if isinstance(node, MappingNode):
            self.flatten_mapping(node)
        if not isinstance(node, MappingNode):
            raise ConstructorError(
                None,
                None,
                "expected a mapping node, but found %s" % node.id,
                node.start_mark,
            )
        mapping = {}
        for key_node, value_node in node.value:
            key = self.construct_object(key_node, deep=deep)
            if not isinstance(key, Hashable):
                raise ConstructorError(
                    "while constructing a mapping",
                    node.start_mark,
                    "found unhashable key",
                    key_node.start_mark,
                )
            if key in mapping:
                raise ConstructorError(
                    "while constructing a mapping",
                    node.start_mark,
                    "found duplicate key %r" % (key,),
                    key_node.start_mark,
                )
            value = self.construct_object(value_node, deep=deep)
            mapping[key] = value
        return mapping


def _github_error(message: str) -> None:
    escaped = message.replace("%", "%25").replace("\r", "%0D").replace("\n", "%0A")
    print(f"::error::{escaped}")


def collect_violations(data: object) -> list[str]:
    """Return policy violation messages (empty if the parsed config is allowed)."""
    violations: list[str] = []
    if data is None:
        return violations
    if not isinstance(data, dict):
        violations.append("top-level YAML must be a mapping")
        return violations

    rules = data.get("rules")
    if rules is None:
        return violations
    if not isinstance(rules, dict):
        violations.append("`rules` must be a mapping")
        return violations

    for audit_id in _FORBIDDEN_RULE_AUDITS:
        if audit_id in rules:
            violations.append(
                f"forbidden key `rules.{audit_id}` (remove this audit block from the config)",
            )

    unpinned = rules.get("unpinned-uses")
    if unpinned is None:
        return violations
    if not isinstance(unpinned, dict):
        violations.append("`rules.unpinned-uses` must be a mapping")
        return violations

    if "disable" in unpinned:
        violations.append("forbidden key `rules.unpinned-uses.disable`")

    cfg = unpinned.get("config")
    if cfg is None:
        return violations
    if not isinstance(cfg, dict):
        violations.append("`rules.unpinned-uses.config` must be a mapping")
        return violations

    policies = cfg.get("policies")
    if policies is None:
        return violations
    if not isinstance(policies, dict):
        violations.append("`rules.unpinned-uses.config.policies` must be a mapping")
        return violations

    for raw_key, raw_val in policies.items():
        key = _normalize_policy_pattern_key(raw_key)
        if key == "*" and raw_val == "any":
            violations.append(
                'forbidden `rules.unpinned-uses.config.policies` entry `"*": any` '
                "(universal unpinned policy); use scoped patterns instead (e.g. `actions/*: any`)",
            )

    return violations


def _normalize_policy_pattern_key(key: object) -> str | None:
    if key is None:
        return None
    if isinstance(key, bool):
        return None
    if isinstance(key, (int, float)):
        return str(key)
    if isinstance(key, str):
        return key.strip()
    return None


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("config_path", type=Path, help="Path to zizmor.yml or .github/zizmor.yml")
    args = parser.parse_args()
    path: Path = args.config_path

    if not path.is_file():
        full = f"{_FAIL_LOG_PREFIX} {path}: config file does not exist or is not a file"
        _github_error(full)
        print(full, file=sys.stderr)
        sys.exit(1)

    text = path.read_text(encoding="utf-8")

    try:
        data = yaml.load(text, Loader=UniqueKeyFullLoader)
    except yaml.YAMLError as exc:
        full = f"{_FAIL_LOG_PREFIX} {path}: invalid YAML: {exc}"
        _github_error(full)
        print(full, file=sys.stderr)
        sys.exit(1)

    violations = collect_violations(data)
    if violations:
        for msg in violations:
            full = f"{_FAIL_LOG_PREFIX} {path}: {msg}"
            _github_error(full)
            print(full, file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
