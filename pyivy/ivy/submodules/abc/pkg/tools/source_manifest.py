#!/usr/bin/env python3
"""Hash the local ABC source without invoking make, compilers, or version control.

Build membership is deliberately textual, not a claim about active configuration.
Run from any directory; stdout is stable JSON. Progress belongs in a separate ledger.
"""
import hashlib
import json
from pathlib import Path
import re

ROOT = Path(__file__).resolve().parents[2]
BUILD_ARTIFACT_EXTENSIONS = {'.o', '.d'}


def digest(data):
    return hashlib.sha256(data).hexdigest()


def manifest():
    paths = set()
    for path in (ROOT / 'src').rglob('*'):
        if path.is_file() and path.suffix not in BUILD_ARTIFACT_EXTENSIONS:
            paths.add(path)
    for name in ('Makefile', 'Makefile.emscripten', 'CMakeLists.txt', 'arch_flags.c',
                 'copyright.txt', 'README.md', 'readmeaig', 'abc.rc'):
        paths.add(ROOT / name)
    memberships = {}
    for path in sorted(paths):
        if path.name == 'module.make':
            text = path.read_text(errors='replace')
            text = '\n'.join(line.split('#', 1)[0] for line in text.splitlines())
            for member in re.findall(r'\bsrc/[\w./+-]+\.(?:c|cc|cpp|y)\b', text):
                memberships.setdefault(member, []).append(path.relative_to(ROOT).as_posix())
    files = []
    for path in sorted(paths):
        data = path.read_bytes()
        name = path.relative_to(ROOT).as_posix()
        files.append({'path': name, 'sha256': digest(data), 'bytes': len(data),
                      'physical_lines': len(data.splitlines()),
                      'textual_module_memberships': memberships.get(name, [])})
    aggregate = ''.join(f"{item['sha256']}  {item['path']}\n" for item in files).encode()
    return {'schema_version': 2,
            'scope': 'all src files except .o/.d build artifacts, plus selected build/lifecycle documentation',
            'membership_caveat': 'Textual module.make references only; not preprocessed build selection.',
            'source_identity_sha256': digest(aggregate), 'files': files}


if __name__ == '__main__':
    print(json.dumps(manifest(), indent=2, sort_keys=True))
