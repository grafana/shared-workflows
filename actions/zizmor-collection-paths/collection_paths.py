#!/usr/bin/env python3
"""Collect zizmor input paths, honoring optional `.github/zizmor-collection-ignore`."""

import argparse
import os
import sys
from pathlib import Path


def _unsafe_prefix_reason(rel: str) -> str | None:
    """Reject prefixes that escape the repository root."""
    if not rel:
        return "empty"
    if rel.startswith(("/", "\\")):
        return "absolute path"
    norm = rel.replace("\\", "/")
    parts = norm.split("/")
    if ".." in parts:
        return "parent segment"
    if any(p == "" for p in parts):
        return "empty path segment"
    return None


def normalize_prefix_line(line: str) -> str | None:
    raw = line.strip()
    if not raw or raw.startswith("#"):
        return None
    if raw.startswith(("/", "\\")):
        return None
    s = raw
    while True:
        old = s
        s = s.removesuffix("**").removesuffix("/*").rstrip("/")
        if s == old:
            break
    if "*" in s:
        return None
    s = s.replace("\\", "/")
    if _unsafe_prefix_reason(s):
        return None
    return s or None


def parse_prefixes_from_ignore(content: str) -> list[str]:
    prefixes = []
    for line in content.splitlines():
        line = line.strip()
        if not line or line.lstrip().startswith("#"):
            continue
        p = normalize_prefix_line(line)
        if p:
            prefixes.append(p)
    return prefixes


def excluded(path: Path, roots: list[Path]) -> bool:
    return any(path == r or r in path.parents for r in roots)


def want_file(path: Path, repo_root: Path) -> bool:
    """True for repo-root workflows, root dependabot config, and composite action manifests."""
    try:
        rel = path.resolve().relative_to(repo_root)
    except ValueError:
        return False
    parts = rel.parts
    if path.name in ("action.yml", "action.yaml"):
        return True
    if path.name in ("dependabot.yml", "dependabot.yaml"):
        return len(parts) == 2 and parts[0] == ".github"
    if path.suffix not in (".yml", ".yaml"):
        return False
    return len(parts) >= 3 and parts[0] == ".github" and parts[1] == "workflows"


def collect_paths(repo_root: Path, prefixes: list[str], out: Path) -> int:
    repo_root = repo_root.resolve()
    out.parent.mkdir(parents=True, exist_ok=True)
    skip = []
    for p in prefixes:
        cand = (repo_root / p).resolve()
        try:
            cand.relative_to(repo_root)
        except ValueError as e:
            raise ValueError(f"ignore prefix {p!r} resolves outside repo root") from e
        skip.append(cand)
    hits = []

    for dirpath, dirnames, filenames in os.walk(repo_root, topdown=True):
        here = Path(dirpath).resolve()
        pruned = [d for d in dirnames if not excluded((here / d).resolve(), skip)]
        dirnames[:] = pruned
        for fn in filenames:
            f = (here / fn).resolve()
            if excluded(f, skip) or not want_file(f, repo_root):
                continue
            hits.append("./" + str(f.relative_to(repo_root)).replace("\\", "/"))

    lines = sorted(set(hits))
    out.write_text("\n".join(lines) + ("\n" if lines else ""), encoding="utf-8")
    return len(lines)


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--repo-root", type=Path, required=True)
    ap.add_argument("--ignore-file", type=Path, required=True)
    ap.add_argument("--paths-out", type=Path, required=True)
    args = ap.parse_args()

    gh = os.environ.get("GITHUB_OUTPUT")
    if not gh:
        print("GITHUB_OUTPUT is not set", file=sys.stderr)
        return 2

    root = args.repo_root.resolve()
    if not args.ignore_file.is_file():
        with open(gh, "a", encoding="utf-8") as f:
            f.write("use_explicit_paths=false\npaths_list=\n")
        return 0

    prefs = parse_prefixes_from_ignore(args.ignore_file.read_text(encoding="utf-8"))
    if not prefs:
        with open(gh, "a", encoding="utf-8") as f:
            f.write("use_explicit_paths=false\npaths_list=\n")
        return 0

    try:
        n = collect_paths(root, prefs, args.paths_out)
    except ValueError as e:
        print(f"::error::{e}", file=sys.stderr)
        return 1
    if n == 0:
        print(
            "::error::.github/zizmor-collection-ignore excluded every zizmor input; remove or relax a prefix.",
            file=sys.stderr,
        )
        return 1

    with open(gh, "a", encoding="utf-8") as f:
        f.write("use_explicit_paths=true\n")
        f.write(f"paths_list={args.paths_out.resolve()}\n")
    return 0


if __name__ == "__main__":
    sys.exit(main())
