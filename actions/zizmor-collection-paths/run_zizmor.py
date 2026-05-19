#!/usr/bin/env python3

import argparse
import json
import os
import shlex
import subprocess
import sys
import tempfile
from pathlib import Path

try:
    from itertools import batched
except ImportError:

    def batched(iterable, n):  # Python <3.12
        seq = list(iterable)
        for i in range(0, len(seq), n):
            yield seq[i : i + n]


_EMPTY_SARIF = {
    "$schema": "https://json.schemastore.org/sarif-2.1.0.json",
    "version": "2.1.0",
    "runs": [],
}


def _merge_sarif_parts(parts, dst: Path):
    if not parts:
        raise ValueError("no SARIF parts")
    docs = [json.loads(p.read_text(encoding="utf-8")) for p in parts]
    if len(docs) == 1:
        merged = docs[0]
    else:
        runs = []
        for doc in docs:
            r = doc.get("runs")
            if isinstance(r, list):
                runs.extend(r)
        merged = {"$schema": docs[0].get("$schema"), "version": docs[0].get("version"), "runs": runs}
    dst.write_text(json.dumps(merged), encoding="utf-8")


def _zizmor_cmd(fmt: str, paths: list[str]) -> list[str]:
    cmd = [
        "uvx",
        f"zizmor@{os.environ['ZIZMOR_VERSION']}",
        "--format",
        fmt,
        "--min-severity",
        os.environ["MIN_SEVERITY"],
        "--min-confidence",
        os.environ["MIN_CONFIDENCE"],
        "--cache-dir",
        os.environ["ZIZMOR_CACHE_DIR"],
    ]
    cfg = os.environ.get("ZIZMOR_CONFIG_PATH", "").strip()
    if cfg:
        cmd += ["--config", cfg]
    dbg = os.environ.get("RUNNER_DEBUG", "").strip().lower()
    if dbg in ("1", "true", "yes", "y", "on"):
        cmd.append("--verbose")
    extra = (os.environ.get("ZIZMOR_EXTRA_ARGS") or "").strip()
    if extra:
        cmd += shlex.split(extra)
    cmd += paths
    return cmd


def _run(cmd: list[str], out: Path | None) -> int:
    if out is None:
        return subprocess.run(cmd, check=False).returncode
    out.parent.mkdir(parents=True, exist_ok=True)
    with out.open("wb") as fh:
        return subprocess.run(cmd, stdout=fh, check=False).returncode


def _scan_paths():
    if os.environ.get("USE_EXPLICIT_PATHS", "").strip().lower() != "true":
        return ["."]
    text = Path(os.environ["PATHS_LIST"]).read_text(encoding="utf-8")
    lines = [ln.strip() for ln in text.splitlines() if ln.strip()]
    return lines if lines else None


def _pipe_plain(cmd: list[str], fh) -> int:
    proc = subprocess.Popen(cmd, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True)
    if proc.stdout is None:
        return 1
    with proc.stdout:
        for line in proc.stdout:
            fh.write(line)
    return proc.wait()


def _sarif(batch: int, out: Path) -> int:
    paths = _scan_paths()
    if paths is None:
        out.write_text(json.dumps(_EMPTY_SARIF), encoding="utf-8")
        return 0

    chunks = [list(c) for c in batched(paths, batch)]
    if len(chunks) == 1:
        rc = _run(_zizmor_cmd("sarif", chunks[0]), out)
        return 1 if rc == 1 else 0

    with tempfile.TemporaryDirectory(prefix="zizmor-sarif-", dir=os.environ["RUNNER_TEMP"]) as name:
        t = Path(name)
        parts = []
        for i, c in enumerate(chunks):
            part = t / f"part-{i}.sarif"
            if _run(_zizmor_cmd("sarif", c), part) == 1:
                return 1
            parts.append(part)
        _merge_sarif_parts(parts, out)
    return 0


def _plain(batch: int) -> int:
    gh = os.environ.get("GITHUB_OUTPUT")
    if not gh:
        print("GITHUB_OUTPUT is not set", file=sys.stderr)
        return 2

    paths = _scan_paths()
    with open(gh, "a", encoding="utf-8") as fh:
        fh.write("zizmor-results<<EOF\n")
        code = 0
        if paths is not None:
            for chunk in batched(paths, batch):
                rc = _pipe_plain(_zizmor_cmd("plain", list(chunk)), fh)
                if rc == 1:
                    print("zizmor crashed; see output above.", file=sys.stderr)
                    return 1
                code = max(code, rc)
        fh.write("EOF\n")
        fh.write(f"zizmor-exit-code={code}\n")
    return 0


def main(argv: list[str]) -> int:
    prog = Path(argv[0]).name if argv else "run_zizmor.py"
    p = argparse.ArgumentParser(prog=prog)
    sub = p.add_subparsers(dest="command", required=True)
    ps = sub.add_parser("sarif", help="run zizmor with SARIF output")
    ps.add_argument("--batch-size", type=int, default=400)
    ps.add_argument("--out", type=Path, default=Path("results.sarif"))
    pp = sub.add_parser("plain-github-output", help="append zizmor plain output to GITHUB_OUTPUT")
    pp.add_argument("--batch-size", type=int, default=400)
    if len(argv) < 2:
        p.print_help(sys.stderr)
        return 2
    ns = p.parse_args(argv[1:])
    if ns.command == "sarif":
        return _sarif(ns.batch_size, ns.out)
    return _plain(ns.batch_size)


if __name__ == "__main__":
    sys.exit(main(sys.argv))
