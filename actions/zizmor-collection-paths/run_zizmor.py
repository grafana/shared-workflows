#!/usr/bin/env python3
"""Run `zizmor` from reusable-zizmor.

Scan `.` by default, or an explicit path list from `collection_paths.py`. Long
lists are split into batches and SARIF outputs are merged to avoid argv limits.
"""

import argparse
import json
import os
import subprocess
import sys
import tempfile
from dataclasses import dataclass
from pathlib import Path
from typing import Any, Iterator, TextIO


def _truthy_env(name: str) -> bool:
    v = os.environ.get(name)
    return bool(v) and v.strip().lower() in {"1", "true", "yes", "y", "on"}


def _read_nonempty_lines(paths_file: Path) -> list[str]:
    out: list[str] = []
    for raw in paths_file.read_text(encoding="utf-8").splitlines():
        line = raw.strip()
        if line:
            out.append(line)
    return out


def _parse_extra_args(raw: str | None) -> list[str]:
    if not raw:
        return []
    # Mirror `${ZIZMOR_EXTRA_ARGS:+${ZIZMOR_EXTRA_ARGS}}` in bash: split on whitespace.
    return raw.split()


@dataclass(frozen=True)
class ZizmorEnv:
    zizmor_version: str
    min_severity: str
    min_confidence: str
    cache_dir: Path
    config_path: Path | None
    extra_args: list[str]
    use_explicit_paths: bool
    paths_list: Path | None

    @classmethod
    def from_environ(cls) -> "ZizmorEnv":
        use_explicit = os.environ.get("USE_EXPLICIT_PATHS", "").strip().lower() == "true"
        cfg_raw = os.environ.get("ZIZMOR_CONFIG_PATH", "").strip()
        return cls(
            zizmor_version=os.environ["ZIZMOR_VERSION"],
            min_severity=os.environ["MIN_SEVERITY"],
            min_confidence=os.environ["MIN_CONFIDENCE"],
            cache_dir=Path(os.environ["ZIZMOR_CACHE_DIR"]),
            config_path=Path(cfg_raw) if cfg_raw else None,
            extra_args=_parse_extra_args(os.environ.get("ZIZMOR_EXTRA_ARGS")),
            use_explicit_paths=use_explicit,
            paths_list=Path(os.environ["PATHS_LIST"]) if use_explicit else None,
        )


def _uvx_zizmor_cmd(env: ZizmorEnv, fmt: str, targets: list[str]) -> list[str]:
    cmd: list[str] = [
        "uvx",
        f"zizmor@{env.zizmor_version}",
        "--format",
        fmt,
        "--min-severity",
        env.min_severity,
        "--min-confidence",
        env.min_confidence,
        "--cache-dir",
        str(env.cache_dir),
    ]
    if env.config_path is not None:
        cmd.extend(["--config", str(env.config_path)])
    if _truthy_env("RUNNER_DEBUG"):
        cmd.append("--verbose")
    cmd.extend(env.extra_args)
    cmd.extend(targets)
    return cmd


def _run(cmd: list[str], *, stdout: Path | None) -> int:
    if stdout is None:
        return int(subprocess.run(cmd, check=False).returncode)
    stdout.parent.mkdir(parents=True, exist_ok=True)
    with stdout.open("wb") as fh:
        return int(subprocess.run(cmd, stdout=fh, stderr=None, check=False).returncode)


def _merge_sarif_parts(parts: list[Path], out: Path) -> None:
    if not parts:
        raise ValueError("no SARIF parts")

    docs: list[dict[str, Any]] = [json.loads(p.read_text(encoding="utf-8")) for p in parts]

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


def _chunks(items: list[str], size: int) -> Iterator[list[str]]:
    for i in range(0, len(items), size):
        yield items[i : i + size]


def _resolve_scan_targets(env: ZizmorEnv) -> list[str] | None:
    """Return targets to scan, or None when explicit mode has no inputs."""
    if not env.use_explicit_paths:
        return ["."]
    assert env.paths_list is not None
    targets = _read_nonempty_lines(env.paths_list)
    if not targets:
        return None
    return targets


def _cmd_sarif(env: ZizmorEnv, args: argparse.Namespace) -> int:
    targets = _resolve_scan_targets(env)
    if targets is None:
        args.out.write_text("", encoding="utf-8")
        return 0

    def run_batch(batch: list[str], out: Path) -> int:
        return _run(_uvx_zizmor_cmd(env, "sarif", batch), stdout=out)

    batches = list(_chunks(targets, args.batch_size))
    if len(batches) == 1:
        rc = run_batch(batches[0], args.out)
        return 1 if rc == 1 else 0

    runner_temp = Path(os.environ["RUNNER_TEMP"])
    with tempfile.TemporaryDirectory(prefix="zizmor-sarif-", dir=runner_temp) as tmpdir:
        tmp = Path(tmpdir)
        parts: list[Path] = []
        for batch in batches:
            part = tmp / f"part-{len(parts):05d}.sarif"
            rc = run_batch(batch, part)
            if rc == 1:
                return 1
            parts.append(part)
        _merge_sarif_parts(parts, args.out)
    return 0


def _stream_plain_to_github(env: ZizmorEnv, targets: list[str], gh_fh: TextIO) -> int:
    cmd = _uvx_zizmor_cmd(env, "plain", targets)
    proc = subprocess.Popen(cmd, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True)
    if proc.stdout is None:
        return 1
    with proc.stdout:
        for line in proc.stdout:
            gh_fh.write(line)
    return int(proc.wait())


def _cmd_plain_github_output(env: ZizmorEnv, args: argparse.Namespace) -> int:
    gh_out = os.environ.get("GITHUB_OUTPUT")
    if not gh_out:
        print("GITHUB_OUTPUT is not set", file=sys.stderr)
        return 2

    targets = _resolve_scan_targets(env)
    out_path = Path(gh_out)

    with out_path.open("a", encoding="utf-8") as gh_fh:
        gh_fh.write("zizmor-results<<EOF\n")

        if targets is None:
            zizmor_exit_code = 0
        elif len(targets) <= args.batch_size:
            zizmor_exit_code = _stream_plain_to_github(env, targets, gh_fh)
        else:
            zizmor_exit_code = 0
            for batch in _chunks(targets, args.batch_size):
                rc = _stream_plain_to_github(env, batch, gh_fh)
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


def _sarif_args_parser() -> argparse.ArgumentParser:
    p = argparse.ArgumentParser()
    p.add_argument("--batch-size", type=int, default=400)
    p.add_argument("--out", type=Path, default=Path("results.sarif"))
    return p


def _plain_args_parser() -> argparse.ArgumentParser:
    p = argparse.ArgumentParser()
    p.add_argument("--batch-size", type=int, default=400)
    return p


def cmd_sarif(argv: list[str]) -> int:
    return _cmd_sarif(ZizmorEnv.from_environ(), _sarif_args_parser().parse_args(argv))


def cmd_plain_github_output(argv: list[str]) -> int:
    return _cmd_plain_github_output(ZizmorEnv.from_environ(), _plain_args_parser().parse_args(argv))


def main(argv: list[str]) -> int:
    if len(argv) < 2:
        print("usage: run_zizmor.py {sarif|plain-github-output} ...", file=sys.stderr)
        return 2

    sub = argv[1]
    rest = argv[2:]
    env = ZizmorEnv.from_environ()
    if sub == "sarif":
        return _cmd_sarif(env, _sarif_args_parser().parse_args(rest))
    if sub == "plain-github-output":
        return _cmd_plain_github_output(env, _plain_args_parser().parse_args(rest))
    print(f"unknown command: {sub}", file=sys.stderr)
    return 2


if __name__ == "__main__":
    sys.exit(main(sys.argv))
