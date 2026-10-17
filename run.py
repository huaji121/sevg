#!/usr/bin/env python3
"""Run SEVG with arguments forwarded to the game (python run.py -seed 42)."""

from __future__ import annotations

import sys

from dev import require_go, run


def main() -> int:
    go = require_go()
    return run([go, "run", ".", *sys.argv[1:]])


if __name__ == "__main__":
    sys.exit(main())
