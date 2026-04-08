# Plan: Route Eq Through atomToZ3 to Match Python's Trace

**Created:** 2026-04-08 ~04:00 UTC

## Context

Test `TestOrdLive` compares Go and Python execution traces line-by-line. At trace line 236123 they diverge:

```
236122  go/py: ivy_solver.py:638 formula_to_z3_int() ENTER type=Eq     ← MATCH
236123  go   : ivy_solver.py:638 formula_to_z3_int() ENTER type=Apply  ← Go recurses into T1
        py   : ivy_solver.py:517 atom_to_z3() ENTER rep== nargs=2     ← Python routes Eq to atom_to_z3
```

**Root cause:** In Python, `is_atom()` returns `True` for `Eq` instances (explicit `isinstance(term, lg.Eq)` check at ivy_logic.py:307), so `formula_to_z3_int()` dispatches Eq to `atom_to_z3()` at line 645-646. Python duck-types `.rep`, `.args`, `.relname` properties onto Eq so atom_to_z3 can treat it like any other atom.

In Go, `Translate()` only routes `*logic.Apply` (with Boolean sort and non-zero Terms) to `atomToZ3()`. Eq falls through to `translateCore()` which has a dedicated `case *logic.Eq:` that processes it differently — recursively calling `Translate(T1)` and `Translate(T2)` then using `Ctx.Eq()`.

## Fix

### Step 1: Add `eqToAtomZ3` helper on `*Translator`

**File:** `z3bridge/translate.go` — add after `atomToZ3` (after line 483)

This method mirrors Python's duck-typing of `.rep`, `.args`, `.relname` on Eq. It constructs a temporary `*logic.Apply` with `Func=Const("=", FunctionSort(s1, s2, Boolean))` and `Terms=[T1, T2]`, then calls `atomToZ3()`.

```go
// eqToAtomZ3 routes Eq through atomToZ3, matching Python where
// is_atom() returns True for Eq and atom_to_z3 accesses Eq via
// duck-typed properties:
//   atom.rep = Symbol('=', RelationSort([x.sort for x in self.args]))
//   atom.args = [t1, t2]
//   atom.relname = equals (ivy_logic.py:1145-1147)
func (t *Translator) eqToAtomZ3(eq *logic.Eq) (Expr, error) {
	s1, s2 := eq.T1.NodeSort(), eq.T2.NodeSort()
	repSort, err := logic.NewFunctionSort(s1, s2, logic.Boolean)
	if err != nil {
		return Expr{}, fmt.Errorf("eqToAtomZ3: %w", err)
	}
	eqConst := logic.NewConst("=", repSort)
	pseudoApp, err := logic.NewApply(eqConst, eq.T1, eq.T2)
	if err != nil {
		return Expr{}, fmt.Errorf("eqToAtomZ3: %w", err)
	}
	return t.atomToZ3(pseudoApp)
}
```

### Step 2: Route Eq in `Translate()` to `eqToAtomZ3`

**File:** `z3bridge/translate.go` lines 212-222

Change from:
```go
	// Python line 645: if ivy_logic.is_atom(fmla): return atom_to_z3(fmla)
	// is_atom: Apply (or Symbol) with Boolean sort, or Eq.
	// Eq is handled by translateCore's *logic.Eq case. Here we only
	// intercept Apply with Boolean sort and non-zero args.
	if app, ok := n.(*logic.Apply); ok && len(app.Terms) > 0 {
		if _, isBool := n.NodeSort().(*logic.BooleanSort); isBool {
			return t.atomToZ3(app)
		}
	}

	return t.translateCore(n)
```

To:
```go
	// Python line 645: if ivy_logic.is_atom(fmla): return atom_to_z3(fmla)
	// is_atom: (Apply or Symbol) with Boolean sort, or Eq.
	if app, ok := n.(*logic.Apply); ok && len(app.Terms) > 0 {
		if _, isBool := n.NodeSort().(*logic.BooleanSort); isBool {
			return t.atomToZ3(app)
		}
	}
	// Python: isinstance(term, lg.Eq) in is_atom → atom_to_z3(fmla)
	// Eq has duck-typed .rep/.args/.relname in Python; we construct a
	// pseudo-Apply to pass through the same atomToZ3 code path.
	if eq, ok := n.(*logic.Eq); ok {
		return t.eqToAtomZ3(eq)
	}

	return t.translateCore(n)
```

### Step 3: Update `case *logic.Eq:` in `translateCore`

**File:** `z3bridge/translate.go` lines 292-312

After the routing change, Eq is intercepted in `Translate()` before reaching `translateCore`. The `case *logic.Eq:` becomes unreachable because:
- `Translate()` catches Eq before calling `translateCore` (Step 2)
- `TermToZ3()` routes Boolean-sorted nodes (which Eq always is) to `Translate()` at line 508-511
- `atomToZ3()` only calls `translateCore` with Apply nodes (line 430)

Replace the case body with a panic to catch unexpected paths:

```go
case *logic.Eq:
	// Eq is now routed through atomToZ3 by Translate() (matching Python's
	// is_atom dispatch). This case should be unreachable.
	return Expr{}, fmt.Errorf("translateCore: unexpected Eq (should be routed through atomToZ3 by Translate)")
```

### Step 4 (if needed): Canonical predKey for "=" in `atomToZ3`

Python caches ALL equalities under one key (`equals` singleton with TopSort-based sort) via `z3_predicates[atom.relname]`. Go uses sort-specific keys (`"=:" + c.CSort.Sexp()`). If different Eq comparisons have different operand sorts, Go will cache-miss where Python cache-hits, producing extra `lookup_native()` / `functionsort()` traces.

**If this causes further divergence**, add a canonical key for "=":

In `atomToZ3()` around line 447, change:
```go
predKey := logic.NodeKey(c.Name + ":" + string(c.CSort.Sexp()))
```
To:
```go
var predKey logic.NodeKey
if c.Name == "=" {
    // Python: atom.relname for Eq is always equals = Symbol('=', RelationSort([TopSort(), TopSort()]))
    // regardless of concrete sorts. Use a canonical key to match Python's caching.
    predKey = eqCanonPredKey
} else {
    predKey = logic.NodeKey(c.Name + ":" + string(c.CSort.Sexp()))
}
```

With a package-level read-only var (exempt per CLAUDE.md rule C.9):
```go
var eqCanonPredKey = logic.NodeKey("=:eq")
```

**Defer this step** — only implement if a subsequent trace line diverges on caching.

## Files to Modify

- **`z3bridge/translate.go`** — primary: Translate(), translateCore(), new eqToAtomZ3()

## Files to Read (reference only)

- `logic/formula.go:33-60` — Eq struct definition
- `logic/term.go:81-95` — Const struct and NewConst
- `logic/term.go:132-189` — Apply struct and NewApply
- `logic/sort.go:54-69` — NewFunctionSort
- `logic/sort.go:183-186` — FirstOrderSort

## Verification

1. `cd /Users/jaten/go/src/github.com/glycerine/ivy/goivy && go build ./...` — compile check
2. `cd parser && go test -run TestOrdLive -v -count=1` — the failing test
3. If TestOrdLive passes or diverges at a LATER line (progress), the fix is correct
4. If it diverges at the same line or earlier, investigate
