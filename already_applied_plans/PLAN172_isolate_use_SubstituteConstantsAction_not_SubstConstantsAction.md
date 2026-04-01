# Fix LoopAction to Call the Correct SubstituteConstantsAction

**Created**: 2026-04-01 19:30

## Context

Golden test diverges at line 156616. Go emits `check.CreateIsolate after_fix_initializers` while Python emits `actions.substitute_constants_action ENTER type=Sequence nargs=3`. This means Python's `fix_initializers` → `loop_action` → `lu.substitute_constants_ast` produces xtracer traces that Go's `FixInitializers` → `LoopAction` → `SubstConstantsAction` does not.

Root cause: Go has **two** substitute-constants-action implementations:

1. **`substituteConstantsAST`** in `actions/helpers.go:232` — faithful port of Python's `ivy_logic_utils.substitute_constants_ast`. Uses `node.Args()` / `node.Clone(newArgs)`. **Has the xtracer trace** (line 242).

2. **`SubstConstantsAction`** in `actions/update.go:2172` — an elaborated version that uses `action.ActionArgs()` / `action.ActionClone(newArgs)` and explicitly handles FormalParams/FormalReturns substitution. **Does NOT have the xtracer trace.** This goes beyond Python's behavior (Python's version does NOT substitute in formal_params/formal_returns).

Go's `LoopAction` (create.go:698) calls the **wrong** one — `SubstConstantsAction` (update.go) instead of `SubstituteConstantsAction` (helpers.go). This causes:
- Missing xtracer traces (the `actions.substitute_constants_action ENTER` lines)
- Different substitution behavior (Go substitutes in FormalParams/FormalReturns; Python does not)

## Changes

### 1. Change `LoopAction` to call `SubstituteConstantsAction`

**File**: `~/ivy/goivy/isolate/create.go` (line 698)

Current:
```go
result := actions.SubstConstantsAction(action, subst)
```

Change to:
```go
result := actions.SubstituteConstantsAction(action, subst)
```

This calls the helpers.go version which:
- Is the direct port of Python's `lu.substitute_constants_ast`
- Uses `node.Args()` and `node.Clone(newArgs)` (matching Python's `ast.args` / `ast.clone(...)`)
- Has the xtracer trace at helpers.go:242
- Does NOT substitute in FormalParams/FormalReturns (matching Python)

### 2. Change `applyActuals` to call `SubstituteConstantsAction`

**File**: `~/ivy/goivy/actions/update.go` (line 1926)

Current:
```go
renamedCallee = SubstConstantsAction(callee, substMap)
```

Change to:
```go
renamedCallee = SubstituteConstantsAction(callee, substMap)
```

### 3. Delete `SubstConstantsAction` and `substConstantsNode`

**File**: `~/ivy/goivy/actions/update.go`

Delete the following functions (lines ~2155-2258):
- `SubstConstantsAction` (lines 2155-2242) — the elaborated version that diverges from Python
- `substConstantsNode` (lines 2244-2258) — its helper (only called by `SubstConstantsAction` and itself)

These are replaced by the faithful port in helpers.go (`SubstituteConstantsAction` / `substituteConstantsAST`).

### Files to modify
- `~/ivy/goivy/isolate/create.go` — line 698
- `~/ivy/goivy/actions/update.go` — line 1926, delete lines 2155-2258

## Verification

```bash
cd ~/ivy/goivy && go build ./... && make golden
```

Check that trace divergence advances past 156616.
