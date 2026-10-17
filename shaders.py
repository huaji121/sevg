#!/usr/bin/env python3
"""Compile the GLSL sources in internal/render/shaders to SPIR-V."""

from __future__ import annotations

import sys

from dev import ROOT, find_glslang, run


def main() -> int:
    glslang = find_glslang()
    if glslang is None:
        print(
            "glslang was not found. Install glslang and add it to PATH, set GLSLANG, "
            "or put glslang(.exe) in .tools/glslang/bin.",
            file=sys.stderr,
        )
        return 1

    shader_dir = ROOT / "internal" / "render" / "shaders"
    sources = sorted(
        path
        for path in shader_dir.iterdir()
        if path.is_file() and path.suffix in {".vert", ".frag"}
    )
    if not sources:
        print(f"no shader sources found in {shader_dir}", file=sys.stderr)
        return 1

    for source in sources:
        output = source.with_name(source.name + ".spv")
        rc = run(
            [
                glslang,
                "-V",
                "--target-env",
                "vulkan1.0",
                "-o",
                output,
                source,
            ]
        )
        if rc != 0:
            return rc
    print(f"compiled {len(sources)} shaders in {shader_dir}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
