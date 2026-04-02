# Fix: `distinctObjRenaming` skips identity mappings — missing `substitute_constants_action` trace

**Created**: 2026-04-02 00:15
**Previous fixes**: Module.Copy() ActCfg nil fix (manually applied), SubstituteConstantsAST consolidation (moved divergence 157358→157557)

## Context

Golden test diverges at line 157557:
```
157557  go : XTRACE: actions.Sequence.int_update ENTER
        py : XTRACE: actions.substitute_constants_action ENTER type=Sequence nargs=2
```

After `CallAction.int_update ENTER` (line 157556), Python calls `substitute_constants_ast(v, subst)` on the callee action (a Sequence with 2 args), emitting a trace. Go skips the substitution entirely because its `distinctObjRenaming` returned an empty map.

## Root cause

**Python** `distinct_obj_renaming` (`ivy_utils.py:186-188`):
```python
def distinct_obj_renaming(names1, names2):
    rn = UniqueRenamer('', names2)
    return dict((s, s.rename(rn)) for s in names1)
```
Returns entries for **ALL** formals — including identity mappings (same name → same name). UniqueRenamer tracks used names to prevent inter-formal collisions.

**Go** `distinctObjRenaming` (`actions/update.go:2021-2041`):
```go
for _, sym := range formals {
    if _, used := usedNames[name]; !used {
        usedNames[name] = true
        continue  // ← BUG: no entry added to result for non-conflicting formals
    }
    // Only conflicting formals get an entry
    result[sym] = lg.NewConst(newName, sym.CSort)
}
```
Skips non-conflicting formals entirely. When no formals conflict with vocab, `renaming` is empty.

**Consequence in `applyActuals`** (`update.go:1930`):
```go
if len(renaming) > 0 {  // false when no conflicts exist
    renamedCallee = SubstituteConstantsAction(callee, substMap)  // SKIPPED
}
```
Go never calls `SubstituteConstantsAction`, so no trace is emitted. Python unconditionally calls it (line 1332 of `ivy_actions.py`).

## Changes

### Step 1: Fix `distinctObjRenaming` — include ALL formals in result

**File: `actions/update.go`**, function `distinctObjRenaming` (lines 2021-2041):

Change the non-conflicting branch to add an identity mapping instead of skipping:

```go
func distinctObjRenaming(formals []*lg.Const, vocabNames map[string]bool) map[*lg.Const]*lg.Const {
    result := make(map[*lg.Const]*lg.Const)
    usedNames := make(map[string]bool)
    for k := range vocabNames {
        usedNames[k] = true
    }
    for _, sym := range formals {
        name := sym.Name
        if _, used := usedNames[name]; !used {
            usedNames[name] = true
            // No conflict — identity mapping (matches Python which always adds all formals)
            result[sym] = lg.NewConst(name, sym.CSort)
            continue
        }
        // Need a fresh name
        newName := unusedNameWithBase(name, usedNames)
        usedNames[newName] = true
        result[sym] = lg.NewConst(newName, sym.CSort)
    }
    return result
}
```

### Step 2: Remove the `if len(renaming) > 0` guard in `applyActuals`

**File: `actions/update.go`**, function `applyActuals` (lines 1928-1942):

Python unconditionally calls `substitute_constants_ast(v, subst)`. Go should too. Remove the guard:

```go
// Before (lines 1929-1942):
renamedCallee := callee
if len(renaming) > 0 {
    substMap := make(map[lg.NodeKey]lg.Expr)
    for oldSym, newSym := range renaming {
        substMap[lg.Key(oldSym)] = newSym
        oldOfOld := lg.NewConst("old("+oldSym.Name+")", oldSym.CSort)
        substMap[lg.Key(oldOfOld)] = lg.NewConst("old("+newSym.Name+")", newSym.CSort)
    }
    renamedCallee = SubstituteConstantsAction(callee, substMap)
}

// After:
substMap := make(map[lg.NodeKey]lg.Expr)
for oldSym, newSym := range renaming {
    substMap[lg.Key(oldSym)] = newSym
    oldOfOld := lg.NewConst("old("+oldSym.Name+")", oldSym.CSort)
    substMap[lg.Key(oldOfOld)] = lg.NewConst("old("+newSym.Name+")", newSym.CSort)
}
renamedCallee := SubstituteConstantsAction(callee, substMap)
```

Note: With Step 1 fix, `renaming` is always non-empty when there are formals, so the guard would rarely trigger anyway. Removing it matches Python's unconditional call pattern.

## Key files

- `actions/update.go:2021-2041` — `distinctObjRenaming` (missing identity mappings)
- `actions/update.go:1895-2016` — `applyActuals` (guarded SubstituteConstantsAction call)
- Python: `ivy_utils.py:186-188` — `distinct_obj_renaming` (returns ALL formals)
- Python: `ivy_actions.py:1318-1358` — `apply_actuals` (unconditional substitute_constants_ast)

## Verification

```bash
cd ~/ivy/goivy && go build ./... && go test ./... && make golden
```

Check that:
1. `go build` and `go test` pass
2. The divergence in `log.red` advances past line 157557
