# Fix AlphaAvoidMap early return causing M vs M_a divergence

Created: 2025-05-01 ~UTC

## Context

XTRACE 1560701 in `log.syslive` shows a divergence in `proof.ApplyMatchAlt EXIT HASH`:
- Go: `(Variable name:M sort:(UninterpretedSort name:mem_type))`
- Python: `(Variable name:M_a sort:(UninterpretedSort name:mem_type))`

The match is **empty** (`nmatch=0`). The formula contains `M` both as a free variable (outside a ForAll) and as a bound variable (inside a ForAll). Python's `alpha_avoid` always renames bound variables that shadow free variables. Go's `AlphaAvoidMap` skips this entirely when the match RHS vars map is empty.

## Fix 1 (primary): Remove early return from AlphaAvoidMap

**File**: `/Users/jaten/ivy/goivy/ivylogic/util.go` lines 473-476

**Delete these 3 lines**:
```go
	if len(vs) == 0 {
		return fmla
	}
```

Python's `alpha_avoid` (`ivy_logic.py:1696-1706`) has no such early return. Even with empty `vs`, Python:
1. Collects free variables of the formula via `lu.free_variables(fmla)`
2. Reserves their names in the `VariableUniqifier`
3. Runs `vu.rec(fmla, vmap)` which renames bound variables that shadow free variables

The rest of Go's `AlphaAvoidMap` naturally handles empty `vs`:
- The loop over `vs` (line 479) iterates zero times — no-op
- `FreeVariablesList(fmla)` (line 488) still collects free vars and reserves names
- `vu.rec(fmla, vmap)` (line 500) renames bound vars that shadow free vars

Note: the sibling function `AlphaAvoid` (slice version, line 447) does NOT have this bug — it always processes free variables. The fix makes `AlphaAvoidMap` match.

## Fix 2 (secondary): Add sorts to FmlaVocab

**File**: `/Users/jaten/ivy/goivy/proof/phase5_goals.go` lines 228-239

Python's `fmla_vocab` (`ivy_proof.py:840-846`) collects sorts, symbols, AND variables:
```python
things = dict(lu.used_sorts_ast(fmla))
things.update(lu.used_symbols_ast(fmla))
things.update(lu.used_variables_ast(fmla))
```

Go's `FmlaVocab` only collects symbols and variables — **missing sorts**.

**Add** before the symbols loop:
```go
for k, s := range lu.SortsAst(fmla) {
    result[k] = s
}
```

`lu.SortsAst` exists at `logicutil/logic_utils.go:21`, returns `map[logic.NodeKey]logic.Sort`. `Sort` embeds `Expr` (`logic/sort.go:14-15`), so the assignment compiles. `lu` is already imported.

## Verification

Run `cd ~/ivy/goivy && make syslive` and check that XTRACE 1560701 no longer diverges (bound `M` should become `M_a` in Go, matching Python).

Also run `cd ~/ivy/goivy && make test` to check for regressions.
