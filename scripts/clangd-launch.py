#!/usr/bin/env python3
import json
import os
import sys
from pathlib import Path

CLANGD = "/usr/bin/clangd"

DRIVERS = {
    ".c": "clang",
    ".cc": "clang++",
    ".cpp": "clang++",
    ".cxx": "clang++",
}


def refresh_compile_commands(root):
    sources = sorted(
        os.path.relpath(os.path.join(dirpath, name), root)
        for dirpath, _, filenames in os.walk(root)
        for name in filenames
        if os.path.splitext(name)[1] in DRIVERS
    )
    payload = (
        json.dumps(
            [
                {
                    "directory": root,
                    "arguments": [DRIVERS[os.path.splitext(src)[1]], "-c", src],
                    "file": src,
                }
                for src in sources
            ],
            indent=4,
        )
        + "\n"
    )
    output = Path(root, "compile_commands.json")
    if not output.is_file() or output.read_text(encoding="utf-8") != payload:
        output.write_text(payload, encoding="utf-8")


try:
    refresh_compile_commands(os.getcwd())
except Exception as exc:  # noqa: BLE001
    print(f"refresh_compile_commands failed: {exc!r}", file=sys.stderr)

os.execv(CLANGD, [CLANGD] + sys.argv[1:])
