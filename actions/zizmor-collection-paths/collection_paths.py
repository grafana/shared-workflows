#!/usr/bin/env python3
"""Build explicit zizmor scan paths when `.github/zizmor-collection-ignore` lists prefixes."""

import argparse
import os
import sys
from pathlib import Path


def _noncomment_lines(content: str) -> list[str]:
    return [
        line.strip()
        for line in content.splitlines()
        if line.strip() and not line.lstrip().startswith("#")
    ]


def normalize_prefix_line(line: str) -> str | None:
    s = line.strip()
    if not s or s.startswith("#"):
        return None
    s = s.removeprefix("/")
    while True:
        before = s
        s = s.removesuffix("**").removesuffix("/*").rstrip("/")
        if s == before:
            break
    if "*" in s:
        return None
    return s or None


def parse_prefixes_from_ignore(content: str) -> list[str]:
    return [p for raw in _noncomment_lines(content) if (p := normalize_prefix_line(raw))]


def _under_exclusion(path: Path, excluded: list[Path]) -> bool:
    return any(path == r or r in path.parents for r in excluded)


def _is_scan_file(path: Path) -> bool:
    if path.name in {"action.yml", "action.yaml", "dependabot.yml", "dependabot.yaml"}:
        return True
    if path.suffix not in {".yml", ".yaml"}:
        return False
    return path.parent.name == "workflows" and path.parent.parent.name == ".github"


def collect_paths(repo_root: Path, prefixes: list[str], paths_out: Path) -> int:
    """Write sorted scan paths; return how many lines were written."""
    repo_root = repo_root.resolve()
    paths_out.parent.mkdir(parents=True, exist_ok=True)
    roots = [(repo_root / p).resolve() for p in prefixes]
    found: list[str] = []

    for root, dirs, files in os.walk(repo_root, topdown=True):
        cur = Path(root).resolve()
        dirs[:] = [d for d in dirs if not _under_exclusion((cur / d).resolve(), roots)]
        for name in files:
            fp = (cur / name).resolve()
            if _under_exclusion(fp, roots) or not _is_scan_file(fp):
                continue
            found.append(f"./{fp.relative_to(repo_root).as_posix()}")

    lines = sorted(set(found))
    text = "\n".join(lines) + ("\n" if lines else "")
    paths_out.write_text(text, encoding="utf-8")
    return len(lines)


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

    if collect_paths(repo_root, prefixes, paths_out) == 0:
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
