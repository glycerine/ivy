# PLAN205b: Fix elimDefinitions iteration order mismatch

**Created:** 2026-04-05 ~00:00 UTC

## Context

PLAN205 fixed `compose_updates` non-deterministic set operations but the divergence at xtrace line 166784 persisted. The real root cause is that `elimDefinitions` in Go iterates `clauses.Defs` order to convert dead defs to fmlas, while Python iterates the `dead` list order. When these differ, the converted fmlas end up in different positions.

## Root Cause: `elimDefinitions` Go vs Python iteration order

**Python** (`ivy_logic_utils.py:1310-1319`):
```python
def elim_definitions(clauses, dead):
    c2 = clauses.copy()
    fmlas = clauses.fmlas
    for sym in dead:                    # iterates DEAD-LIST order
        if sym in clauses.defidx:
            fmlas.append(c2.defidx[sym].to_constraint())
    deadset = set(dead)
    defs = [d for d in clauses.defs if d.defines() not in deadset]
    return Clauses(fmlas, defs, clauses.annot)
```

**Go** (`module/ops.go:826-843`):
```go
func elimDefinitions(clauses *Clauses, deadSet map[lg.NodeKey]bool) *Clauses {
    var fmlas []lg.Expr
    fmlas = append(fmlas, clauses.Fmlas...)
    for _, d := range clauses.Defs {    // iterates CLAUSE-DEFS order
        key := definesKey(d)
        if deadSet[key] {
            fmlas = append(fmlas, defToConstraint(d))
        }
    }
    ...
}
```

**Example**: If defs are [def_A, def_B, def_C] and dead=[C, A]:
- Python: fmlas += [constraint_C, constraint_A]  (dead order)
- Go: fmlas += [constraint_A, constraint_C]  (defs order)

This causes different formula ordering in `ite_clauses_int` output, which cascades through all subsequent `compose_updates` calls.

## Fix Plan

### Step 1: Change `elimDeadDefinitions` to pass ordered dead list

In `module/ops.go`, change the call from passing `deadSet` to passing the ordered `dead` slice:

```go
// Current:
result[i] = elimDefinitions(a, deadSet)

// Fixed:
result[i] = elimDefinitions(a, dead)
```

### Step 2: Rewrite `elimDefinitions` to iterate `dead` order

Change the function signature and iteration to match Python:

```go
func elimDefinitions(clauses *Clauses, dead []lg.NodeKey) *Clauses {
    var fmlas []lg.Expr
    fmlas = append(fmlas, clauses.Fmlas...)
    // Iterate dead in order, look up in clause's DefIdx (matches Python)
    for _, key := range dead {
        if idx, ok := clauses.DefIdx[key]; ok {
            fmlas = append(fmlas, defToConstraint(clauses.Defs[idx]))
        }
    }
    // Build deadSet for filtering remaining defs
    deadSet := make(map[lg.NodeKey]bool, len(dead))
    for _, key := range dead {
        deadSet[key] = true
    }
    var defs []*il.Definition
    for _, d := range clauses.Defs {
        key := definesKey(d)
        if !deadSet[key] {
            defs = append(defs, d)
        }
    }
    return NewClauses(fmlas, defs, clauses.Annot)
}
```

Key changes:
- Takes `dead []lg.NodeKey` instead of `deadSet map[lg.NodeKey]bool`
- Iterates `dead` list and looks up in `clauses.DefIdx[key]` to get the def
- Builds `deadSet` locally for filtering remaining defs
- `clauses.DefIdx` maps `lg.NodeKey` → index in `clauses.Defs`

### Step 3: Remove deadSet construction from `elimDeadDefinitions`

Remove lines 813-816 (the `deadSet` construction loop) since `elimDefinitions` now handles it internally.

### Step 4: Verify + test

```bash
go build ./...
go test ./module/... -count=1
go test ./actions/... -count=1
make golden
```

## Files Modified

- `module/ops.go` — rewrite `elimDefinitions` signature and iteration, update call site in `elimDeadDefinitions`
