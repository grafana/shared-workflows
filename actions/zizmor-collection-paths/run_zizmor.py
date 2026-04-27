#!/usr/bin/env python3
"""Run `zizmor` from reusable-zizmor with the same behavior as the old bash.

Either scan `.`, or scan an explicit list of YAML inputs collected by
`collection_paths.py`. In the explicit case, split argv into batches and merge
SARIF so we do not hit OS argv limits.
"""

from __future__ import annotations

import argparse
import json
import os
import subprocess
import sys
import tempfile
from pathlib import Path
from typing import Any, TextIO


def _truthy_env(name: str) -> bool:
    v = os.environ.get(name)
    return bool(v) and v.strip().lower() in {"1", "true", "yes", "y", "on"}


def _read_nonempty_lines(paths_file: Path) -> list[str]:
    out: list[str] = []
    for raw in paths_file.read_text(encoding="utf-8").splitlines():
        line = raw.strip()
        if not line:
            continue
        out.append(line)
    return out


def _zizmor_cmd(
    *,
    zizmor_version: str,
    fmt: str,
    min_severity: str,
    min_confidence: str,
    cache_dir: Path,
    config_path: Path | None,
    extra_args: list[str],
    targets: list[str],
) -> list[str]:
    cmd: list[str] = [
        "uvx",
        f"zizmor@{zizmor_version}",
        "--format",
        fmt,
        "--min-severity",
        min_severity,
        "--min-confidence",
        min_confidence,
        "--cache-dir",
        str(cache_dir),
    ]

    if config_path is not None:
        cmd.extend(["--config", str(config_path)])

    if _truthy_env("RUNNER_DEBUG"):
        cmd.append("--verbose")

    cmd.extend(extra_args)
    cmd.extend(targets)
    return cmd


def _run(cmd: list[str], *, stdout: Path | None) -> int:
    if stdout is None:
        p = subprocess.run(cmd, check=False)
        return int(p.returncode)

    stdout.parent.mkdir(parents=True, exist_ok=True)
    with stdout.open("wb") as fh:
        p = subprocess.run(cmd, stdout=fh, stderr=None, check=False)
    return int(p.returncode)


def _merge_sarif_parts(parts: list[Path], out: Path) -> None:
    if not parts:
        raise ValueError("no SARIF parts")

    docs: list[dict[str, Any]] = []
    for part in parts:
        docs.append(json.loads(part.read_text(encoding="utf-8")))

    if len(docs) == 1:
        merged = docs[0]
    else:
        schema = docs[0].get("$schema")
        version = docs[0].get("version")
        runs: list[Any] = []
        for doc in docs:
            r = doc.get("runs")
            if isinstance(r, list):
                runs.extend(r)
        merged = {"$schema": schema, "version": version, "runs": runs}

    out.write_text(json.dumps(merged), encoding="utf-8")


def _parse_extra_args(raw: str | None) -> list[str]:
    if not raw:
        return []
    # Mirror `${ZIZMOR_EXTRA_ARGS:+${ZIZMOR_EXTRA_ARGS}}` in bash: split on whitespace.
    return raw.split()


def cmd_sarif(argv: list[str]) -> int:
    p = argparse.ArgumentParser()
    p.add_argument("--batch-size", type=int, default=400)
    p.add_argument("--out", type=Path, default=Path("results.sarif"))
    args = p.parse_args(argv)

    zizmor_version = os.environ["ZIZMOR_VERSION"]
    min_severity = os.environ["MIN_SEVERITY"]
    min_confidence = os.environ["MIN_CONFIDENCE"]
    cache_dir = Path(os.environ["ZIZMOR_CACHE_DIR"])

    cfg_raw = os.environ.get("ZIZMOR_CONFIG_PATH", "").strip()
    config_path = Path(cfg_raw) if cfg_raw else None

    extra_args = _parse_extra_args(os.environ.get("ZIZMOR_EXTRA_ARGS"))

    use_explicit = os.environ.get("USE_EXPLICIT_PATHS", "").strip().lower() == "true"
    if not use_explicit:
        cmd = _zizmor_cmd(
            zizmor_version=zizmor_version,
            fmt="sarif",
            min_severity=min_severity,
            min_confidence=min_confidence,
            cache_dir=cache_dir,
            config_path=config_path,
            extra_args=extra_args,
            targets=["."],
        )
        rc = _run(cmd, stdout=args.out)
        if rc == 1:
            return 1
        return 0

    paths_list = Path(os.environ["PATHS_LIST"])
    targets = _read_nonempty_lines(paths_list)
    if not targets:
        args.out.write_text("", encoding="utf-8")
        return 0

    if len(targets) <= args.batch_size:
        cmd = _zizmor_cmd(
            zizmor_version=zizmor_version,
            fmt="sarif",
            min_severity=min_severity,
            min_confidence=min_confidence,
            cache_dir=cache_dir,
            config_path=config_path,
            extra_args=extra_args,
            targets=targets,
        )
        rc = _run(cmd, stdout=args.out)
        if rc == 1:
            return 1
        return 0

    runner_temp = Path(os.environ["RUNNER_TEMP"])
    tmp = Path(tempfile.mkdtemp(prefix="zizmor-sarif-", dir=runner_temp))
    try:
        parts: list[Path] = []
        for idx in range(0, len(targets), args.batch_size):
            batch = targets[idx : idx + args.batch_size]
            part = tmp / f"part-{len(parts):05d}.sarif"
            cmd = _zizmor_cmd(
                zizmor_version=zizmor_version,
                fmt="sarif",
                min_severity=min_severity,
                min_confidence=min_confidence,
                cache_dir=cache_dir,
                config_path=config_path,
                extra_args=extra_args,
                targets=batch,
            )
            rc = _run(cmd, stdout=part)
            if rc == 1:
                return 1
            parts.append(part)

        _merge_sarif_parts(parts, args.out)
        return 0
    finally:
        # Best-effort cleanup; keep the SARIF output even if cleanup fails.
        for child in tmp.iterdir():
            try:
                child.unlink()
            except OSError:
                pass
        try:
            tmp.rmdir()
        except OSError:
            pass


def cmd_plain_github_output(argv: list[str]) -> int:
    p = argparse.ArgumentParser()
    p.add_argument("--batch-size", type=int, default=400)
    args = p.parse_args(argv)

    gh_out = os.environ.get("GITHUB_OUTPUT")
    if not gh_out:
        print("GITHUB_OUTPUT is not set", file=sys.stderr)
        return 2

    zizmor_version = os.environ["ZIZMOR_VERSION"]
    min_severity = os.environ["MIN_SEVERITY"]
    min_confidence = os.environ["MIN_CONFIDENCE"]
    cache_dir = Path(os.environ["ZIZMOR_CACHE_DIR"])

    cfg_raw = os.environ.get("ZIZMOR_CONFIG_PATH", "").strip()
    config_path = Path(cfg_raw) if cfg_raw else None

    extra_args = _parse_extra_args(os.environ.get("ZIZMOR_EXTRA_ARGS"))

    use_explicit = os.environ.get("USE_EXPLICIT_PATHS", "").strip().lower() == "true"

    out_path = Path(gh_out)

    def run_plain(targets: list[str], *, gh_fh: TextIO) -> int:
        cmd = _zizmor_cmd(
            zizmor_version=zizmor_version,
            fmt="plain",
            min_severity=min_severity,
            min_confidence=min_confidence,
            cache_dir=cache_dir,
            config_path=config_path,
            extra_args=extra_args,
            targets=targets,
        )
        p = subprocess.Popen(cmd, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True)
        assert p.stdout is not None
        try:
            for line in p.stdout:
                gh_fh.write(line)
        finally:
            stdout_io = p.stdout
            close = getattr(stdout_io, "close", None)
            if callable(close):
                try:
                    close()
                except OSError:
                    pass
        return int(p.wait())

    with out_path.open("a", encoding="utf-8") as gh_fh:
        gh_fh.write("zizmor-results<<EOF\n")

        zizmor_exit_code = 0
        if not use_explicit:
            zizmor_exit_code = run_plain(["."], gh_fh=gh_fh)
        else:
            paths_list = Path(os.environ["PATHS_LIST"])
            targets = _read_nonempty_lines(paths_list)
            if not targets:
                zizmor_exit_code = 0
            elif len(targets) <= args.batch_size:
                zizmor_exit_code = run_plain(targets, gh_fh=gh_fh)
            else:
                for idx in range(0, len(targets), args.batch_size):
                    batch = targets[idx : idx + args.batch_size]
                    rc = run_plain(batch, gh_fh=gh_fh)
                    if rc == 1:
                        print(
                            "zizmor itself failed - check the above output. failing the workflow.",
                            file=sys.stderr,
                        )
                        return 1
                    if rc > zizmor_exit_code:
                        zizmor_exit_code = rc

        gh_fh.write("EOF\n")
        gh_fh.write(f"zizmor-exit-code={zizmor_exit_code}\n")

    return 0


def main(argv: list[str]) -> int:
    if len(argv) < 2:
        print("usage: run_zizmor.py {sarif|plain-github-output} ...", file=sys.stderr)
        return 2

    sub = argv[1]
    rest = argv[2:]
    if sub == "sarif":
        return cmd_sarif(rest)
    if sub == "plain-github-output":
        return cmd_plain_github_output(rest)

    print(f"unknown command: {sub}", file=sys.stderr)
    return 2


if __name__ == "__main__":
    sys.exit(main(sys.argv))
