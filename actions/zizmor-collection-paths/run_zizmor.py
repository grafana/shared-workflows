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
    if os.environ.get("RUNNER_DEBUG", "").strip().lower() in ("1", "true", "yes", "y", "on"):
        cmd.append("--verbose")
    extra = (os.environ.get("ZIZMOR_EXTRA_ARGS") or "").strip()
    if extra:
        cmd += shlex.split(extra)
    cmd += paths
    return cmd


def _scan_paths() -> list[str] | None:
    if os.environ.get("USE_EXPLICIT_PATHS", "").strip().lower() != "true":
        return ["."]
    lines = [
        ln.strip()
        for ln in Path(os.environ["PATHS_LIST"]).read_text(encoding="utf-8").splitlines()
        if ln.strip()
    ]
    return lines if lines else None


def _zizmor_rc(rc: int) -> int:
    # 0 = clean, 10–14 = findings by severity; only 1 is a crash.
    return 1 if rc == 1 else 0


def _merge_sarif_parts(parts: list[Path], dst: Path) -> None:
    docs = [json.loads(p.read_text(encoding="utf-8")) for p in parts]
    runs = [run for doc in docs for run in (doc.get("runs") or [])]
    # Code scanning rejects multiple runs with the same tool and category:
    # fold every batch into the first run, remapping ruleIndex.
    if len(runs) > 1:
        base = runs[0]
        rules = base.setdefault("tool", {}).setdefault("driver", {}).setdefault("rules", [])
        index_by_id = {r.get("id"): i for i, r in enumerate(rules)}
        for run in runs[1:]:
            run_rules = run.get("tool", {}).get("driver", {}).get("rules") or []
            for result in run.get("results") or []:
                rule_id = result.get("ruleId")
                if rule_id is not None and rule_id not in index_by_id:
                    old = result.get("ruleIndex")
                    rule = run_rules[old] if old is not None and old < len(run_rules) else {"id": rule_id}
                    index_by_id[rule_id] = len(rules)
                    rules.append(rule)
                if "ruleIndex" in result and rule_id in index_by_id:
                    result["ruleIndex"] = index_by_id[rule_id]
                base.setdefault("results", []).append(result)
        runs = [base]
    merged = {
        "$schema": docs[0].get("$schema"),
        "version": docs[0].get("version"),
        "runs": runs,
    }
    dst.parent.mkdir(parents=True, exist_ok=True)
    dst.write_text(json.dumps(merged), encoding="utf-8")


def _run_sarif(paths: list[str] | None, batch: int, out: Path) -> int:
    out.parent.mkdir(parents=True, exist_ok=True)
    if paths is None:
        out.write_text(json.dumps(_EMPTY_SARIF), encoding="utf-8")
        return 0

    chunks = list(batched(paths, batch))
    if len(chunks) == 1:
        with out.open("wb") as fh:
            rc = subprocess.run(_zizmor_cmd("sarif", list(chunks[0])), stdout=fh, check=False).returncode
        return _zizmor_rc(rc)

    tmp = os.environ["RUNNER_TEMP"]
    with tempfile.TemporaryDirectory(prefix="zizmor-sarif-", dir=tmp) as name:
        parts: list[Path] = []
        for i, chunk in enumerate(chunks):
            part = Path(name) / f"part-{i}.sarif"
            with part.open("wb") as fh:
                rc = subprocess.run(_zizmor_cmd("sarif", list(chunk)), stdout=fh, check=False).returncode
            if rc == 1:
                return 1
            parts.append(part)
        _merge_sarif_parts(parts, out)
    return 0


def _run_plain(paths: list[str] | None, batch: int) -> int:
    gh = os.environ.get("GITHUB_OUTPUT")
    if not gh:
        print("GITHUB_OUTPUT is not set", file=sys.stderr)
        return 2

    with open(gh, "a", encoding="utf-8") as fh:
        fh.write("zizmor-results<<EOF\n")
        code = 0
        crashed = False
        try:
            if paths is not None:
                for chunk in batched(paths, batch):
                    proc = subprocess.run(
                        _zizmor_cmd("plain", list(chunk)),
                        stdout=subprocess.PIPE,
                        stderr=subprocess.STDOUT,
                        text=True,
                        check=False,
                    )
                    fh.write(proc.stdout or "")
                    if proc.returncode == 1:
                        print("zizmor crashed; see output above.", file=sys.stderr)
                        crashed = True
                        break
                    code = max(code, proc.returncode)
        finally:
            fh.write("EOF\n")
            fh.write(f"zizmor-exit-code={1 if crashed else code}\n")
    return 1 if crashed else 0


def main(argv: list[str]) -> int:
    p = argparse.ArgumentParser(prog=Path(argv[0]).name if argv else "run_zizmor.py")
    sub = p.add_subparsers(dest="command", required=True)
    ps = sub.add_parser("sarif")
    ps.add_argument("--batch-size", type=int, default=400)
    ps.add_argument("--out", type=Path, default=Path("results.sarif"))
    pp = sub.add_parser("plain-github-output")
    pp.add_argument("--batch-size", type=int, default=400)
    if len(argv) < 2:
        p.print_help(sys.stderr)
        return 2

    ns = p.parse_args(argv[1:])
    paths = _scan_paths()
    if ns.command == "sarif":
        return _run_sarif(paths, ns.batch_size, ns.out)
    return _run_plain(paths, ns.batch_size)


if __name__ == "__main__":
    sys.exit(main(sys.argv))
