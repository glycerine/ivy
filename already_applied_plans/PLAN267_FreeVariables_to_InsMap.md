# Fix ForAll variable nesting order: FreeVariables → InsMap

Created: 2026-04-12 ~04:00 UTC

## Context

Go and Python diverge at xtrace line 255185 in TestOrdLive. The l2s tactic creates nested single-variable ForAlls in different nesting orders:

- **Go**: `ForAll vars:[A] → ForAll vars:[M] → ForAll vars:[P]` (random/alphabetical — map iteration)
- **Python**: `ForAll vars:[P] → ForAll vars:[M] → ForAll vars:[A]` (DFS traversal order)

**Root cause**: `FreeVariablesList()` in Go iterates a `map[logic.NodeKey]logic.Expr` returned by `FreeVariables()`. Go maps have non-deterministic iteration order. Python's equivalent (`variables_ast → unique → tuple`) preserves first-occurrence DFS traversal order.

**Fix**: Change `FreeVariables()` to return `*iu.InsMap[logic.NodeKey, logic.Expr]` (insertion-ordered map from `ivyutils/insmap.go`). Since `freeVariablesRec` traverses depth-first and inserts variables on first encounter, InsMap preserves DFS first-occurrence order — matching Python.

## Files to modify

### 1. `logicutil/logicutil.go` — Core change

- **`FreeVariables()`** (line 58): Change return type from `map[logic.NodeKey]logic.Expr` to `*iu.InsMap[logic.NodeKey, logic.Expr]`
- **`freeVariablesRec()`** (line 82): Change `result` parameter from `map[logic.NodeKey]logic.Expr` to `*iu.InsMap[logic.NodeKey, logic.Expr]`. Use `result.Set(key, n)` instead of `result[key] = n`. Use `result.Get2(key)` for membership checks.
- **`FreeVariablesList()`** (line 614): Iterate with `for _, node := range fv.All()` instead of `for _, node := range fv`.
- **`copyVarSet()`** (line 627): Keep as-is — only used for `bound` parameter which stays as plain map.
- Internal callers at lines 379, 411: update map access to InsMap API.

### 2. Callers of `FreeVariables()` — Mechanical map→InsMap translation

Each caller needs: `map[key]` → `.Get(key)`, `_, ok := map[key]` → `.Get2(key)`, `for k,v := range map` → `for k,v := range m.All()`, `len(map)` → `.Len()`, `map[key] = val` → `.Set(key, val)`.

Production files:
- `fragment/fragment.go` (lines 479, 538)
- `ivylogic/constructors.go` (lines 198, 251)
- `ivylogic/classify_ext.go` (lines 100, 120)
- `ivylogic/classify.go` (line 48)
- `ivylogic/util.go` (lines 256, 265, 452) — via `FreeVariablesList`, no direct change needed
- `z3bridge/solver_herbrand.go` (line 241)
- `proof/goal.go` (line 281)
- `proof/match.go` (line 540)
- `logicutil/logic_utils.go` (lines 212, 1653, 1676)
- `module/astutil.go` (line 278) — wraps `FreeVariables`, return type changes
- `module/ops.go` (lines 1068, 1077)

### 3. Callers of `FreeVariablesList()` — No change needed

`FreeVariablesList` still returns `[]*logic.Variable`. Its internals change (iterate InsMap instead of map), but its signature is unchanged. All 16 callers are unaffected.

### 4. Test files — Same mechanical translation

- `logicutil/logicutil_test.go` — multiple call sites
- `fragment/divergence5_test.go` — 3 call sites
- `ivylogic/varuniq_test.go`, `ivylogic/new_funcs_test.go`
- `module/clauseops_test.go`
- `proof/proof_test.go`

## What stays unchanged

- **`deduplicateAndSortVars`**: Only called from `NewForAll`/`NewExists`, not in the l2s path. Both Go `varsSexp` and Python `_vars_sexp` sort alphabetically for multi-var nodes. Not the cause of this divergence.
- **`forall()` helper in `check/l2s.go`**: Creates raw ForAll, no sorting. Receives variables from `FreeVariablesList` which will now be in correct order.
- **`bound` parameter in `freeVariablesRec`**: Stays as plain `map` — only used for membership lookups, not iteration order.
- **Python side**: `_vars_sexp` already sorts alphabetically. No changes needed.

## Verification

1. `cd ~/ivy/goivy && go build ./...` — confirm compilation
2. `cd ~/ivy/goivy/parser && go test -run TestOrdLive -v -count=1` — the failing test should now pass (line 255185 divergence resolved)
3. `cd ~/ivy/goivy && go test ./logicutil/... -v -count=1` — existing unit tests pass
