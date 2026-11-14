#!/usr/bin/env python3
"""Check the library's source-floor and dependency boundaries without mutating it."""
import json
import os
from pathlib import Path
import subprocess

ROOT = Path(__file__).resolve().parents[1]
EXPECTED = 'module github.com/b1ug/colorkit\n\ngo 1.13\n'
if (ROOT / 'go.mod').read_text() != EXPECTED:
    raise SystemExit('go.mod must contain only the module path and go 1.13')
if (ROOT / 'go.sum').exists():
    raise SystemExit('Unexpected go.sum: review any new module dependency')

env = dict(os.environ, GOWORK='off', CGO_ENABLED='0')
raw = subprocess.check_output(
    ['go', 'list', '-mod=readonly', '-deps', '-json', './...'],
    cwd=str(ROOT), env=env, universal_newlines=True,
)
decoder, packages, offset = json.JSONDecoder(), [], 0
while offset < len(raw):
    if raw[offset].isspace():
        offset += 1
        continue
    package, offset = decoder.raw_decode(raw, offset)
    packages.append(package)
own = 'github.com/b1ug/colorkit'
for package in packages:
    path = package['ImportPath']
    if not package.get('Standard') and path != own and not path.startswith(own + '/'):
        raise SystemExit('External Go dependency: ' + path)
    if (path == own or path.startswith(own + '/')) and package.get('CgoFiles'):
        raise SystemExit('CGO source is outside the module contract: ' + path)
library = next(p for p in packages if p['ImportPath'] == own)
for path in library.get('Imports', []):
    if path in ['os', 'os/exec', 'syscall', 'plugin', 'unsafe'] or path == 'net' or path.startswith('net/'):
        raise SystemExit('Library import is outside the pure color boundary: ' + path)
print('Module policy passed: Go 1.13, standard library only, no CGO or direct I/O imports.')
