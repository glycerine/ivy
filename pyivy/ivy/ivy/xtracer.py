"""Execution tracer for parallel conformance testing.

Go counterpart: goivy/xtracer/ package with build tag "xtracer".

Both emit the same format:
    XTRACE: <category>.<function> <ENTER|EXIT> [<detail>]

Guard with `if __debug__:` so that `python3 -O` eliminates all traces
at bytecode compile time (zero overhead in production).

Usage:
    from . import xtracer
    if __debug__: xtracer.trace("compiler.IvyCompile ENTER decls=%d" % n)
"""

import os
import sys

from .canon import canon_blake3

_hash_verbose = True

enabled = os.environ.get('XTRACE_OFF') != '1'

def trace(msg, *args):
    """Print an XTRACE line to stdout, flushed immediately."""
    if enabled:
    #if args: # turn off everything except vocab calls for a moment
            print("XTRACE: " + msg, file=sys.stdout, flush=True)


def normalize_filename(f):
    """Replace include/examples directory paths with canonical placeholders.

    Matches Go's normalizeLine in golden_test.go so that golden test
    comparisons don't diverge on absolute paths.
    """
    if f is None:
        return f
    import os.path
    # Build list of include directory prefixes to normalize.
    # The include path may be under the source tree or under a venv.
    prefixes = []
    from . import ivy_utils as iu
    std_dir = iu.get_std_include_dir()
    if std_dir:
        base_dir = os.path.dirname(std_dir)
        if base_dir and not base_dir.endswith(os.sep):
            base_dir += os.sep
        prefixes.append(base_dir)
    # Also check the directory where this module is installed
    # (handles venv case where include/ is under site-packages/ivy/).
    _mod_dir = os.path.dirname(os.path.abspath(__file__))
    _mod_include = os.path.join(_mod_dir, 'include') + os.sep
    if _mod_include not in prefixes and os.path.isdir(_mod_include):
        prefixes.append(_mod_include)
    for prefix in prefixes:
        if f.startswith(prefix):
            return '<IVY_INCLUDE>/' + f[len(prefix):]
    examples_dir = os.environ.get('IVY_EXAMPLES_DIR', '')
    if examples_dir:
        if not examples_dir.endswith(os.sep):
            examples_dir += os.sep
        if f.startswith(examples_dir):
            return '<IVY_EXAMPLES>/' + f[len(examples_dir):]
    return f


import re

_re_lineno = re.compile(r'\s*lineno:\d+')
_re_filename = re.compile(r'\s*filename:"[^"]*"')

def _strip_locations(s):
    """Remove lineno:N and filename:"..." from a canonical s-expression."""
    s = _re_lineno.sub('', s)
    s = _re_filename.sub('', s)
    while '  ' in s:
        s = s.replace('  ', ' ')
    s = s.replace('( ', '(')
    return s


class MerkleState:
    """Rolling Merkle root for incremental state verification."""
    def __init__(self):
        self.prev_root = ''

    def add_leaf(self, canonical_str):
        """Add a leaf to the Merkle tree and return (leaf_hash, root_hash).

        (Update: we just omit line numbers atm).
        Line numbers and filenames are stripped before hashing so that
        Go and Python can have different line numbers without causing
        hash mismatches. The full canon string is still available for display.
        """
        ## we just omit
        ##stripped = _strip_locations(canonical_str)
        ##leaf = canon_blake3(stripped)
        leaf = canon_blake3(canonical_str)
        self.prev_root = canon_blake3(self.prev_root + leaf)
        return leaf, self.prev_root
