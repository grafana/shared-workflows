#!/usr/bin/env python3
# security-appsec#326: optional .github/zizmor-collection-ignore — directory prefixes to skip when collecting zizmor inputs.

import argparse
import os
import sys
from pathlib import Path


def normalize_prefix_line(line: str) -> str | None:
    s = line.strip()
    if not s or s.startswith("#"):
        return None
    s = s.removeprefix("/")
    while True:
        old = s
        s = s.removesuffix("**").removesuffix("/*").rstrip("/")
        if s == old:
            break
    if "*" in s:
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


def want_file(path: Path) -> bool:
    if path.name in ("action.yml", "action.yaml", "dependabot.yml", "dependabot.yaml"):
        return True
    if path.suffix not in (".yml", ".yaml"):
        return False
    return path.parent.name == "workflows" and path.parent.parent.name == ".github"


def collect_paths(repo_root: Path, prefixes: list[str], out: Path) -> int:
    repo_root = repo_root.resolve()
    out.parent.mkdir(parents=True, exist_ok=True)
    skip = [(repo_root / p).resolve() for p in prefixes]
    hits = []

    for dirpath, dirnames, filenames in os.walk(repo_root, topdown=True):
        here = Path(dirpath).resolve()
        pruned = [d for d in dirnames if not excluded((here / d).resolve(), skip)]
        dirnames[:] = pruned
        for fn in filenames:
            f = (here / fn).resolve()
            if excluded(f, skip) or not want_file(f):
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

    n = collect_paths(root, prefs, args.paths_out)
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
