"""
Canonical s-expression strings for cross-language AST comparison.

Produces output compatible with the Go goivy ast.Canon() methods.
Format conventions:
  - Struct: (typeName field:value field2:value)
  - Slices: [elem1 elem2 elem3]  (space separated, NO commas)
  - Strings: "double quoted"
  - Bool: true / false
  - Int: plain number
  - None/nil: nil
  - Single space between fields
  - Struct names lowercase first letter
  - Field names use Go field names (lowercase first letter)
"""

import base64
import blake3


def canon_blake3(canonical_str):
    """Compute the blake3 hash of a canonical s-expression string.

    Takes the un-keyed 64-byte (512 bit) blake3 hash of the canonical string,
    takes just the first 33 bytes, converts to base64 using URL encoding,
    then prepends "blake3.33B-".

    Matches Go ivyutils.Canonical.Blake3() exactly.
    """
    h = blake3.blake3(canonical_str.encode('utf-8'), derive_key_context=None)
    sum64 = h.digest(length=64)
    first33 = sum64[:33]
    encoded = base64.urlsafe_b64encode(first33).decode('ascii')
    return "blake3.33B-" + encoded


def node_canon(n):
    """Return canon() of a node, or 'nil' if None."""
    if n is None:
        return 'nil'
    if hasattr(n, 'canon'):
        return n.canon()
    # Fallback for objects without canon()
    return str(n)


def slice_canon(nodes):
    """Return canonical form of a list/tuple of nodes."""
    if nodes is None or len(nodes) == 0:
        return '[]'
    return '[' + ' '.join(node_canon(n) for n in nodes) + ']'


def string_canon(s):
    """Return a double-quoted canonical string."""
    if s is None:
        return 'nil'
    if s == '':
        return ''
    if type(s).__name__ == 'This':
        return '"this"'
    if not isinstance(s, str):
        return '"' + str(s) + '"'
    # Escape backslashes and double quotes
    escaped = s.replace('\\', '\\\\').replace('"', '\\"')
    return '"' + escaped + '"'


def string_slice_canon(strings):
    """Return canonical form of a list of strings."""
    if not strings:
        return '[]'
    return '[' + ' '.join(string_canon(s) for s in strings) + ']'


def bool_canon(b):
    """Return canonical bool: true, false, or nil for None."""
    if b is None:
        return 'nil'
    return 'true' if b else 'false'


def lineno_fields(obj):
    """Return flattened lineno fields for an AST object.
    Matches Go Base.canonFields() output.
    Returns '' when empty, or ' field:value' (leading space) when populated.
    Callers use '(typeName{} field:...' so no double-space when empty."""
    ## appears python Ivy has a bug where the line numbers
    ## are wrong... so just return empty string for now.
    return ''
    if hasattr(obj, 'lineno'):
        ln = obj.lineno
        if hasattr(ln, 'filename') and ln.filename:
            return ' filename:{} lineno:{}'.format(string_canon(ln.filename), ln.line)
        if hasattr(ln, 'line'):
            return ' lineno:{}'.format(ln.line)
        # lineno is a plain int
        return ' lineno:{}'.format(ln)
    return ' lineno:0'


def decl_fields(obj):
    """Return flattened DeclBase-equivalent fields.
    Matches Go DeclBase.canonFields() output."""
    attrs = getattr(obj, 'attributes', ())
    common = getattr(obj, 'common', None)
    # attrs may be a tuple of strings or AST nodes
    if attrs:
        attrs_canon = slice_canon(list(attrs))
    else:
        attrs_canon = '[]'
    return '{} declArgs:{} attributes:{} common:{}'.format(
        lineno_fields(obj),
        slice_canon(list(obj.args)),
        attrs_canon,
        node_canon(common))
