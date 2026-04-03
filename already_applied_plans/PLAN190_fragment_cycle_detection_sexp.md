# Fix infinite loop in checker_canon()

NOTE: NOT IMPLEMENTED.

**Created:** 2026-04-02 (current session)

## Context

The `checker_canon()` function in `~/ivy/pyivy/ivy/ivy/canon_fragment.py:160` hangs when called from the xtracer trace at `ivy_fragment.py:473`. The function serializes fragment checker state (strat_map, arcs, universally_quantified_variables, etc.) into canonical s-expressions by calling `node_canon()` which dispatches to `.canon()`/`.sexp()` methods on various AST and logic objects.

The root cause is that `node_canon()` in `canon.py:37` has **no cycle detection**. It recursively calls `.canon()` on objects, which in turn call `node_canon()` on their children. If any object graph contains a cycle (e.g., an AST node whose `.sort` or `.args` field eventually references itself), this recurses infinitely.

AST nodes (from `ivy_ast.py`) are mutable Python objects — unlike logic types (from `logic.py`) which use immutable `recstruct`. This means AST object graphs CAN contain cycles if a field gets set to reference an ancestor node during compilation or transformation.

## Fix: Add cycle detection to `node_canon()`

### File: `~/ivy/pyivy/ivy/ivy/canon.py`

Add a thread-local (or module-level) visited set to `node_canon()` that tracks `id()` of objects currently being serialized. If the same object is encountered again during serialization, emit a sentinel like `"<cycle>"` instead of recursing.

```python
_canon_in_progress = set()  # tracks id() of objects being serialized

def node_canon(n):
    """Return canon() of a node, or 'nil' if None."""
    if n is None:
        return 'nil'
    obj_id = id(n)
    if obj_id in _canon_in_progress:
        return '<cycle:{}:{}>'.format(type(n).__name__, obj_id)
    if hasattr(n, 'canon'):
        _canon_in_progress.add(obj_id)
        try:
            return n.canon()
        finally:
            _canon_in_progress.discard(obj_id)
    # Fallback for objects without canon()
    return str(n)
```

This is safe because:
- `node_canon` is the central dispatch for ALL canonicalization (both `canon_ast.py` and `canon_fragment.py` route through it)
- The visited set uses `id()` which is the object's memory address — unique per live object
- The `try/finally` ensures cleanup even if `.canon()` raises
- Objects without `.canon()` use `str()` which doesn't go through `node_canon`, so no tracking needed
- The `<cycle:...>` sentinel makes cycles visible in debug output rather than silently truncating

### Files involved

- **`~/ivy/pyivy/ivy/ivy/canon.py:37-44`** — modify `node_canon()` to add cycle detection (the only change needed)

## Verification

1. Run the Ivy test that triggers the `checker_canon()` call at `ivy_fragment.py:473`
2. Confirm it no longer hangs
3. If `<cycle:...>` appears in the output, that identifies the specific object type causing the cycle, which can be investigated further as a separate issue
