#!/usr/bin/env python3
"""Build explicit zizmor scan paths when ignore prefixes are configured."""

from __future__ import annotations

import argparse
import os
import sys
from pathlib import Path


def _lines_after_sed_filter(content: str) -> list[str]:
    """Drop blank lines and full-line comments."""
    out: list[str] = []
    for line in content.splitlines():
        if not line.strip():
            continue
        if line.lstrip().startswith("#"):
            continue
        out.append(line)
    return out


def normalize_prefix_line(line: str) -> str | None:
    """Normalize one ignore entry to a repo-relative prefix."""
    s = line.strip()
    if not s or s.startswith("#"):
        return None
    if s.startswith("/"):
        s = s[1:]
    if s.endswith("**"):
        s = s[:-2]
    if s.endswith("/*"):
        s = s[:-2]
    while s.endswith("/"):
        s = s[:-1]
    if "*" in s:
        return None
    return s or None


def parse_prefixes_from_ignore(content: str) -> list[str]:
    prefixes: list[str] = []
    for raw in _lines_after_sed_filter(content):
        p = normalize_prefix_line(raw)
        if p:
            prefixes.append(p)
    return prefixes


def _is_in_excluded_tree(path: Path, excluded_roots: list[Path]) -> bool:
    return any(path == root or root in path.parents for root in excluded_roots)


def _is_target_file(path: Path) -> bool:
    name = path.name
    if name in {"action.yml", "action.yaml", "dependabot.yml", "dependabot.yaml"}:
        return True

    if path.suffix not in {".yml", ".yaml"}:
        return False
    parent = path.parent
    return parent.name == "workflows" and parent.parent.name == ".github"


def collect_paths(repo_root: Path, prefixes: list[str], paths_out: Path) -> None:
    repo_root = repo_root.resolve()
    paths_out.parent.mkdir(parents=True, exist_ok=True)
    excluded_roots = [(repo_root / p).resolve() for p in prefixes]

    collected: list[str] = []
    for current_root, dirs, files in os.walk(repo_root, topdown=True):
        current_path = Path(current_root).resolve()
        dirs[:] = [
            d for d in dirs if not _is_in_excluded_tree((current_path / d).resolve(), excluded_roots)
        ]

        for filename in files:
            file_path = (current_path / filename).resolve()
            if _is_in_excluded_tree(file_path, excluded_roots):
                continue
            if not _is_target_file(file_path):
                continue
            rel = file_path.relative_to(repo_root).as_posix()
            collected.append(f"./{rel}")

    lines = sorted(set(collected))
    output = "\n".join(lines)
    if output:
        output += "\n"
    paths_out.write_text(output, encoding="utf-8")


def append_github_output(github_output: Path, use_explicit: bool, paths_list: str) -> None:
    with github_output.open("a", encoding="utf-8") as fh:
        fh.write(f"use_explicit_paths={'true' if use_explicit else 'false'}\n")
        fh.write(f"paths_list={paths_list}\n")


def main() -> int:
    p = argparse.ArgumentParser()
    p.add_argument("--repo-root", type=Path, required=True)
    p.add_argument("--ignore-file", type=Path, required=True)
    p.add_argument("--paths-out", type=Path, required=True)
    args = p.parse_args()

    gh_out = os.environ.get("GITHUB_OUTPUT")
    if not gh_out:
        print("GITHUB_OUTPUT is not set", file=sys.stderr)
        return 2

    repo_root = args.repo_root.resolve()
    ignore_file = args.ignore_file
    paths_out = args.paths_out

    if not ignore_file.is_file():
        append_github_output(Path(gh_out), False, "")
        return 0

    prefixes = parse_prefixes_from_ignore(ignore_file.read_text(encoding="utf-8"))
    if not prefixes:
        append_github_output(Path(gh_out), False, "")
        return 0

    collect_paths(repo_root, prefixes, paths_out)
    text = paths_out.read_text(encoding="utf-8")
    nonempty_lines = [ln for ln in text.splitlines() if ln.strip()]
    if not nonempty_lines:
        print(
            "::error::.github/zizmor-collection-ignore excluded every zizmor input. "
            "Remove or relax a prefix.",
            file=sys.stderr,
        )
        return 1

    append_github_output(Path(gh_out), True, str(paths_out.resolve()))
    return 0


if __name__ == "__main__":
    sys.exit(main())
