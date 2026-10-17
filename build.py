#!/usr/bin/env python3
"""Build the SEVG executable."""

from __future__ import annotations

import os
import sys

from dev import ROOT, require_go, run


def main() -> int:
    go = require_go()
    output = ROOT / ("sevg.exe" if os.name == "nt" else "sevg")
    rc = run([go, "build", "-trimpath", "-o", output, "."])
    if rc != 0:
        return rc
    print(f"built {output}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
