#!/usr/bin/env python3
"""Run zizmor from reusable-zizmor: scan `.` or explicit paths; batch + merge SARIF when needed."""

import argparse
import json
import os
import subprocess
import sys
import tempfile
from pathlib import Path
from types import SimpleNamespace
from typing import Any, TextIO


def _env() -> SimpleNamespace:
    ex = os.environ.get("USE_EXPLICIT_PATHS", "").strip().lower() == "true"
    cfg = os.environ.get("ZIZMOR_CONFIG_PATH", "").strip()
    return SimpleNamespace(
        ver=os.environ["ZIZMOR_VERSION"],
        sev=os.environ["MIN_SEVERITY"],
        conf=os.environ["MIN_CONFIDENCE"],
        cache=Path(os.environ["ZIZMOR_CACHE_DIR"]),
        cfg=Path(cfg) if cfg else None,
        extra=(os.environ.get("ZIZMOR_EXTRA_ARGS") or "").split(),
        explicit=ex,
        plist=Path(os.environ["PATHS_LIST"]) if ex else None,
    )


def _targets(e: SimpleNamespace) -> list[str] | None:
    if not e.explicit:
        return ["."]
    assert e.plist is not None
    lines = [ln.strip() for ln in e.plist.read_text(encoding="utf-8").splitlines() if ln.strip()]
    return lines or None


def _uvx(e: SimpleNamespace, fmt: str, tgts: list[str]) -> list[str]:
    cmd = [
        "uvx",
        f"zizmor@{e.ver}",
        "--format",
        fmt,
        "--min-severity",
        e.sev,
        "--min-confidence",
        e.conf,
        "--cache-dir",
        str(e.cache),
    ]
    if e.cfg:
        cmd += ["--config", str(e.cfg)]
    dbg = os.environ.get("RUNNER_DEBUG", "")
    if dbg and dbg.strip().lower() in {"1", "true", "yes", "y", "on"}:
        cmd.append("--verbose")
    return cmd + e.extra + tgts


def _run(cmd: list[str], *, out: Path | None) -> int:
    if out is None:
        return int(subprocess.run(cmd, check=False).returncode)
    out.parent.mkdir(parents=True, exist_ok=True)
    with out.open("wb") as fh:
        return int(subprocess.run(cmd, stdout=fh, stderr=None, check=False).returncode)


def _merge_sarif_parts(parts: list[Path], dst: Path) -> None:
    if not parts:
        raise ValueError("no SARIF parts")
    docs = [json.loads(p.read_text(encoding="utf-8")) for p in parts]
    if len(docs) == 1:
        merged = docs[0]
    else:
        runs: list[Any] = []
        for d in docs:
            r = d.get("runs")
            if isinstance(r, list):
                runs.extend(r)
        merged = {"$schema": docs[0].get("$schema"), "version": docs[0].get("version"), "runs": runs}
    dst.write_text(json.dumps(merged), encoding="utf-8")


def _plain_stream(e: SimpleNamespace, tgts: list[str], fh: TextIO) -> int:
    proc = subprocess.Popen(_uvx(e, "plain", tgts), stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True)
    if proc.stdout is None:
        return 1
    with proc.stdout:
        for line in proc.stdout:
            fh.write(line)
    return int(proc.wait())


def _sarif(e: SimpleNamespace, batch: int, out: Path) -> int:
    tg = _targets(e)
    if tg is None:
        out.write_text("", encoding="utf-8")
        return 0
    n = len(tg)
    if n <= batch:
        rc = _run(_uvx(e, "sarif", tg), out=out)
        return 1 if rc == 1 else 0
    parts: list[Path] = []
    with tempfile.TemporaryDirectory(prefix="zizmor-sarif-", dir=os.environ["RUNNER_TEMP"]) as td:
        tdir = Path(td)
        for i in range(0, n, batch):
            p = tdir / f"p{len(parts):05d}.sarif"
            rc = _run(_uvx(e, "sarif", tg[i : i + batch]), out=p)
            if rc == 1:
                return 1
            parts.append(p)
        _merge_sarif_parts(parts, out)
    return 0


def _plain(e: SimpleNamespace, batch: int) -> int:
    gh = os.environ.get("GITHUB_OUTPUT")
    if not gh:
        print("GITHUB_OUTPUT is not set", file=sys.stderr)
        return 2
    tg = _targets(e)
    with Path(gh).open("a", encoding="utf-8") as fh:
        fh.write("zizmor-results<<EOF\n")
        code = 0
        if tg is None:
            pass
        elif len(tg) <= batch:
            code = _plain_stream(e, tg, fh)
        else:
            for i in range(0, len(tg), batch):
                rc = _plain_stream(e, tg[i : i + batch], fh)
                if rc == 1:
                    print(
                        "zizmor itself failed - check the above output. failing the workflow.",
                        file=sys.stderr,
                    )
                    return 1
                code = max(code, rc)
        fh.write("EOF\n")
        fh.write(f"zizmor-exit-code={code}\n")
    return 0


def cmd_sarif(argv: list[str]) -> int:
    p = argparse.ArgumentParser()
    p.add_argument("--batch-size", type=int, default=400)
    p.add_argument("--out", type=Path, default=Path("results.sarif"))
    a = p.parse_args(argv)
    return _sarif(_env(), a.batch_size, a.out)


def cmd_plain_github_output(argv: list[str]) -> int:
    p = argparse.ArgumentParser()
    p.add_argument("--batch-size", type=int, default=400)
    return _plain(_env(), p.parse_args(argv).batch_size)


def main(argv: list[str]) -> int:
    if len(argv) < 2:
        print("usage: run_zizmor.py {sarif|plain-github-output} ...", file=sys.stderr)
        return 2
    sub, rest = argv[1], argv[2:]
    if sub == "sarif":
        return cmd_sarif(rest)
    if sub == "plain-github-output":
        return cmd_plain_github_output(rest)
    print(f"unknown command: {sub}", file=sys.stderr)
    return 2


if __name__ == "__main__":
    sys.exit(main(sys.argv))
