# Fix traceSymSet to Display All Symbol Types Using PrettyFmla

**Created:** 2026-03-31

## Context

`traceSymSet` in `isolate/helpers.go:24` only traces `*lg.Const` entries using `ConstSymDisplay`. Non-Const symbols (e.g., `*lg.Apply` reps from `SymbolsIluAst`) are silently omitted. Python's `_trace_sym_set` calls `str(x)` on every symbol.

### Key finding: Python's `str()` is NOT the logic.py `__str__`

Python's `ivy_logic.py:1440-1442` **monkey-patches** `__str__` on all logic types:

```python
for cls in [lg.Eq, lg.Not, lg.And, lg.Or, lg.Implies, lg.Iff, lg.Ite, lg.ForAll, lg.Exists,
            lg.Apply, lg.Var, lg.Const, lg.Lambda, lg.NamedBinder]:
    cls.__str__ = pretty_fmla
```

Where `pretty_fmla` calls `self.drop_annotations(False, set())` then `.ugly(0)`. This is the "pretty" infix format, not the Go `String()` methods which use a different structural format (e.g., `And(...)` vs `&`, `ForAll` vs `forall`).

Go already has `lg.PrettyFmla(expr)` in `logic/pretty.go` which exactly mirrors this Python `pretty_fmla` → `ugly(0)` path. This is the correct function to use.

Note: `ConstSymDisplay` is close but not identical to `PrettyFmla` for Const — it uses `c.CSort.String()` which gives `"Boolean"` while `constUgly`/`sortName` gives `"bool"`. `PrettyFmla` is the faithful match.

## Plan

### File to modify

`/Users/jaten/go/src/github.com/glycerine/ivy/goivy/isolate/helpers.go` — `traceSymSet` (lines 24-36)

### Change

Replace the Const-only loop with `lg.PrettyFmla(v)` for all entries:

```go
func traceSymSet(label string, syms map[lg.NodeKey]lg.Expr) {
	displayNames := make([]string, 0, len(syms))
	for _, v := range syms {
		displayNames = append(displayNames, lg.PrettyFmla(v))
	}
	sort.Strings(displayNames)
	for _, s := range displayNames {
		xtracer.Trace("%s.sym %s", label, s)
	}
	xtracer.Trace("%s n=%d", label, len(syms))
}
```

This removes the `actions` import dependency from `traceSymSet` (though other functions in helpers.go may still use it). The `iu` and `il` imports may also become unused if only used by the old code — will check at build time.

## Verification

1. `go build ./...`
2. `go test ./isolate/...`
