# Fix l2s fairCycle Implies(And(),...) wrapping divergence

**Created:** 2026-04-29 ~UTC

## Context

XTRACE 28252224 shows Go and Python diverge in the body of a `ForAll` inside the "idle" ActionTerm during `l2s.modPass`. The diff:

- **Go:** `body:(Iff t1:(Apply func:(NamedBinder name:l2s_s ...`
- **Python:** `body:(Implies t1:(And terms:[]) t2:(Iff t1:(Apply func:(NamedBinder name:l2s_s ...`

Python always wraps with `Implies(And(*filtered), body)` when there are quantified variables, even when the filter produces zero items (all vars are finite-sorted), yielding `Implies(And(), Iff(...))`. Go skips the `Implies` wrapper when the filter produces zero conjuncts.

## Root cause

`/Users/jaten/ivy/goivy/check/l2s.go` — fair cycle construction has two locations with the same bug:

**Relations (lines 678–682):**
```go
if len(aConjs) > 0 {
    fairCycle = append(fairCycle,
        forall(vb.Vars, &lg.Implies{T1: makeAnd(aConjs...), T2: iff}))
} else {
    fairCycle = append(fairCycle, forall(vb.Vars, iff))  // ← skips Implies wrapper
}
```

**Functions/constants (lines 718–722):**
```go
if len(aConjs) > 0 {
    fairCycle = append(fairCycle,
        forall(vb.Vars, &lg.Implies{T1: makeAnd(aConjs...), T2: eq}))
} else {
    fairCycle = append(fairCycle, forall(vb.Vars, eq))  // ← skips Implies wrapper
}
```

Python (`ivy_l2s.py` lines 980–1006) has no such conditional — it always wraps:
```python
forall(vs, lg.Implies(
    lg.And(*(l2s_a(v.sort)(v) for v in vs if v.sort.name not in finite_sorts)),
    lg.Iff(l2s_s(vs, t)(*vs), t)
))
```

Also: `makeAnd()` with zero args returns `lg.True` (not `&lg.And{Terms: nil}`), which would produce the wrong canon even if we removed the conditional.

## Fix

### Change 1 — Relations (lines 678–682)

Replace the 5-line if/else with:
```go
fairCycle = append(fairCycle,
    forall(vb.Vars, &lg.Implies{T1: &lg.And{Terms: aConjs}, T2: iff}))
```

### Change 2 — Functions/constants (lines 718–722)

Replace the 5-line if/else with:
```go
fairCycle = append(fairCycle,
    forall(vb.Vars, &lg.Implies{T1: &lg.And{Terms: aConjs}, T2: eq}))
```

Both use `&lg.And{Terms: aConjs}` directly (not `makeAnd`) so that when `aConjs` is nil, we get `(And terms:[])` in the canon output, matching Python's `lg.And()`.

## Files modified

- `/Users/jaten/ivy/goivy/check/l2s.go` — lines 678–682 and 718–722

## Verification

```
cd ~/ivy/goivy && make test
```

Confirm XTRACE 28252224 no longer diverges and no new divergences appear.
