# PLAN211: Fix fmlaPair/arc lineno divergence in fragment canon output

NOTE: not applied. We manually set lineno = 0 with // TODO revert when 
// we can confidently fix python's broken lineno mechanism. The Go is
// currently correct, the python is not. But we don't want to mess with
// the python and risk changing its logic, for now; until the Go port
// is confident of its replication.

**Created:** 2026-04-07 ~03:30 UTC

## Context

The golden test (`make golden`) diverges at `log.red:81960`. Go's `fmlaPair` canon outputs `lineno:0` while Python's outputs `lineno:<IVY_EXAMPLES>/doc/examples/apple/ord_live.ivy: line 312:`.

**Root cause**: Python stores `lineno` as a `LocationTuple` object (a tuple of `(filename, line_number)`) on AST nodes like `LabeledFormula`. When Python's `fmla_pair_canon()` in `canon_fragment.py:156` formats `lineno:{}`, Python calls `str()` on the `LocationTuple`, which outputs the full path like `<IVY_EXAMPLES>/doc/examples/apple/ord_live.ivy: line 312:`. Go stores `fmlaPair.lineno` as `int` and formats it with `lineno:%d`.

**Existing precedent**: Both Go's `Base.canonFields()` (`ast/ast.go:208-218`) and Python's `lineno_fields()` (`canon.py:83-99`) are **disabled** — they return empty strings, with the comment "python's line numbers are off, omit for now." The `fmlaPair.Sexp()` and `arc.Sexp()` bypass this mechanism and emit lineno directly.

## Approach: Normalize lineno to 0 on both sides in fragment canon

Extend the "lineno disabled for matching" precedent from the AST canon system to the fragment canon system. Both Go and Python will always emit `lineno:0` in `fmlaPair` and `arc` canon output.

## File changes

### 1. Go: `fragment/canon.go` — always emit `lineno:0`

**`fmlaPair.Sexp()`** (line 132-139): Change `f.lineno` to hardcoded `0`:

```go
func (f *fmlaPair) Sexp() lg.NodeKey {
    sourceStr := "nil"
    if f.source != nil {
        sourceStr = string(f.source.Canon())
    }
    // Normalize lineno to 0 — matches disabled lineno_fields()/canonFields()
    return lg.NodeKey(fmt.Sprintf("(fmlaPair fmla:%s source:%s lineno:0)",
        exprSexp(f.fmla), sourceStr))
}
```

**`arc.Sexp()`** (line 80-84): Change `a.lineno` to hardcoded `0`:

```go
func (a *arc) Sexp() lg.NodeKey {
    // Normalize lineno to 0 — matches disabled lineno_fields()/canonFields()
    return lg.NodeKey(fmt.Sprintf("(arc from:%s to:%s fmla:%s lineno:0 argIdx:%d hasIdx:%v)",
        ufNodeSexp(a.from), ufNodeSexp(a.to), exprSexp(a.fmla),
        a.argIdx, a.hasIdx))
}
```

### 2. Python: `~/ivy/pyivy/ivy/ivy/canon_fragment.py` — always emit `lineno:0`

**`fmla_pair_canon()`** (line 148-157):

```python
def fmla_pair_canon(fmla, source, lineno):
    """Serialize a formula/source pair. Matches Go fmlaPair.Sexp()."""
    source_str = 'nil'
    if source is not None:
        if hasattr(source, 'canon'):
            source_str = source.canon()
        else:
            source_str = str(source)
    # Normalize lineno to 0 — matches disabled lineno_fields()/canonFields()
    return '(fmlaPair fmla:{} source:{} lineno:0)'.format(
        node_canon(fmla), source_str)
```

**`arc_canon()`** (lines 87-100):

```python
def arc_canon(arc_tuple):
    """Serialize an arc tuple. Matches Go arc.Sexp().
    Python arcs are 4-tuples (from, to, fmla, lineno)
    or 5-tuples (from, to, fmla, lineno, argIdx)."""
    if len(arc_tuple) == 5:
        v, anode, fmla, lineno, idx = arc_tuple
        # Normalize lineno to 0 — matches disabled lineno_fields()/canonFields()
        return '(arc from:{} to:{} fmla:{} lineno:0 argIdx:{} hasIdx:true)'.format(
            uf_node_canon(v), uf_node_canon(anode),
            node_canon(fmla), idx)
    else:
        v, anode, fmla, lineno = arc_tuple
        return '(arc from:{} to:{} fmla:{} lineno:0 argIdx:-1 hasIdx:false)'.format(
            uf_node_canon(v), uf_node_canon(anode),
            node_canon(fmla))
```

### 3. Go: `fragment/canon_test.go` — update cross-language test expectations

The cross-language test creates arcs/fmlaPairs with specific lineno values (15, 42, etc.) and compares against Python output. After both sides normalize to 0, the existing test data will still match (both sides emit `lineno:0` regardless of input). No test code changes needed — the test constructs Go Sexp() and Python Sexp() independently, and both will now produce `lineno:0`.

### 4. Python: `~/ivy/goivy/pytesthelper/emit_fragment_sexp.py` — no change needed

The `fmla_pair_canon(eq, X, 15)` call will still work; it just ignores the `15` and outputs `lineno:0`. The Go side's `fmlaPair{lineno: 15}.Sexp()` will also output `lineno:0`. Both match.

## Verification

```bash
# Go build and fragment tests
cd ~/ivy/goivy && go build ./... && go test ./fragment/...

# Golden test (should advance past log.red line 81960)
cd ~/ivy/goivy && make golden
```
