#!/usr/bin/env python3
"""Shared helpers for the SEVG Python developer scripts."""

from __future__ import annotations

import os
import shutil
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent


def _existing(path: str | Path | None) -> Path | None:
    if not path:
        return None
    p = Path(path).expanduser()
    return p if p.is_file() else None


def find_go() -> Path | None:
    """Locate the Go toolchain without relying on the current shell PATH."""
    for key in ("GO", "GOROOT_GO"):
        p = _existing(os.environ.get(key))
        if p:
            return p

    for name in ("go", "go.exe"):
        p = shutil.which(name)
        if p:
            return Path(p)

    candidates: list[Path] = []
    goroot = os.environ.get("GOROOT")
    if goroot:
        candidates.append(Path(goroot) / "bin" / ("go.exe" if os.name == "nt" else "go"))

    home = Path.home()
    candidates += [
        home / "go" / "bin" / ("go.exe" if os.name == "nt" else "go"),
        Path("/usr/local/go/bin/go"),
        Path("/usr/bin/go"),
    ]

    if os.name == "nt":
        local = os.environ.get("LOCALAPPDATA")
        if local:
            candidates.append(Path(local) / "Programs" / "Go" / "bin" / "go.exe")
        candidates += [
            Path(r"C:\Program Files\Go\bin\go.exe"),
            Path(r"C:\Go\bin\go.exe"),
            Path(r"D:\Go\bin\go.exe"),
            Path(r"E:\my_apps\Go\bin\go.exe"),
        ]

    for candidate in candidates:
        if candidate.is_file():
            return candidate
    return None


def require_go() -> Path:
    go = find_go()
    if go is None:
        print(
            "go was not found. Install Go 1.27+ and add it to PATH, "
            "or set GOROOT/GO.",
            file=sys.stderr,
        )
        raise SystemExit(1)
    return go


def find_glslang() -> Path | None:
    """Locate glslangValidator/glslang from PATH or the optional local bundle."""
    for key in ("GLSLANG", "GLSLANG_VALIDATOR"):
        p = _existing(os.environ.get(key))
        if p:
            return p

    names = ("glslang", "glslang.exe", "glslangValidator", "glslangValidator.exe")
    for name in names:
        p = shutil.which(name)
        if p:
            return Path(p)

    local_names = (
        "glslang.exe",
        "glslang",
        "glslangValidator.exe",
        "glslangValidator",
    )
    for base in (ROOT / ".tools" / "glslang" / "bin", ROOT / "bin"):
        for name in local_names:
            p = base / name
            if p.is_file():
                return p
    return None


def run(cmd: list[str | Path], cwd: Path = ROOT) -> int:
    rendered = " ".join(str(part) for part in cmd)
    print(f"+ {rendered}", flush=True)
    try:
        return subprocess.run([str(part) for part in cmd], cwd=cwd).returncode
    except FileNotFoundError as exc:
        print(f"failed to start command: {exc}", file=sys.stderr)
        return 1
    except KeyboardInterrupt:
        return 130
