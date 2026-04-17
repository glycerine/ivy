# PLAN: Fix missing `always` parameter on rewrite_atom() methods

Created: 2026-04-17, 07:00

## Context

After fixing the `optwith` parser build issue, `ivy_check tilelink1.ivy` now parses successfully but crashes during checking with:
```
TypeError: AstRewriteSubstConstantsParams.rewrite_atom() got an unexpected keyword argument 'always'
```

The crash occurs at `ivy_ast.py:1741` where `ast_rewrite` calls `rewrite.rewrite_atom(atom, always=always)`. Five rewrite classes exist; three accept `always`, two don't.

## Root cause

`ivy_ast.py` line 1741:
```python
always = not(hasattr(rewrite,'local') and rewrite.local)
arg0 = rewrite.rewrite_atom(atom,always=always)
```

**Has `always=False`**: `AstRewriteSubstPrefix` (line 1663), `AstRewritePostfix` (line 1683), `AstRewriteAddParams` (line 1691)

**Missing `always`**: `AstRewriteSubstConstants` (line 1637), `AstRewriteSubstConstantsParams` (line 1647)

These two classes do simple substitution lookups — the `always` flag doesn't affect their behavior, but they need to accept it to satisfy the call signature.

## Fix

Add `always=False` to `rewrite_atom()` in both classes:

In `/Users/jaten/go/src/github.com/glycerine/ivy/pyivy/ivy/ivy/ivy_ast.py`:

```python
# Line ~1639: AstRewriteSubstConstants.rewrite_atom
def rewrite_atom(self, atom, always=False):  # was: def rewrite_atom(self, atom):

# Line ~1649: AstRewriteSubstConstantsParams.rewrite_atom
def rewrite_atom(self, atom, always=False):  # was: def rewrite_atom(self, atom):
```

## Go side — no change needed

Go's `Rewriter` interface (`ast/rewrite.go:336`) already requires `RewriteAtom(atom *Atom, always bool) Node`, and both `AstRewriteSubstConstants` (line 355) and `AstRewriteSubstConstantsParams` (line 382) implement it with the `always` parameter.

## Critical file

- `/Users/jaten/go/src/github.com/glycerine/ivy/pyivy/ivy/ivy/ivy_ast.py`

## Verification

```bash
cd ~/ivy/ivy-lang-examples/examples/tilelink
python3 -O $(which ivy_check) tilelink1.ivy
```

Should get past the `ast_rewrite` crash. May hit further issues in the file, but this specific TypeError will be resolved.
