#!/usr/bin/env python3
"""Run the SEVG Go tests; extra arguments are forwarded to go test."""

from __future__ import annotations

import sys

from dev import require_go, run


def main() -> int:
    go = require_go()
    return run([go, "test", "./...", *sys.argv[1:]])


if __name__ == "__main__":
    sys.exit(main())
