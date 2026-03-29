# Fix ActionInterferenceCheck Divergence at xtrace line 146102

**Created:** 2026-03-28 ~21:45 PST

## Context

`make golden` diverges at xtrace line 146102 in `compiler.ActionInterferenceCheck`. Go and Python report completely different actions with different modify counts:

- **Python**: `index.impl.next[implement15] modifies=1`, `index.impl.prev[implement16] modifies=1`
- **Go**: `cfabric.init[after98] modifies=7`, `dramc_nb2.memc_arriving.raise[before214] modifies=3`

Root cause analysis reveals **three bugs** in Go's `CheckDefinitions` function at `compiler/ivy_compile.go:1475-1527`:

### Bug 1: Missing ActionsConfig — destructors not unwrapped (CORRECTNESS BUG)

Line 1480 calls `actions.Modifies(act)` **without** an `ActionsConfig`. This makes `isDestructor()` always return `false` (nil config → line 165-166 of `actions/transforms.go`). Go never unwraps destructor chains, so it finds the destructor symbol itself instead of the root object symbol. Python always unwraps via `ivy_module.module.destructor_sorts`.

This explains why Go reports different actions and higher counts — destructor assignments show up as separate modifications instead of being traced to the root symbol.

### Bug 2: No deduplication of modified symbols in xtrace count

Python uses `mod_syms = set()` and reports `len(mod_syms)` (unique count). Go uses `len(mods)` from the slice returned by `actions.Modifies()` (includes duplicates). If the same root symbol is assigned multiple times, Python counts 1, Go counts N.

### Bug 3: Non-deterministic map iteration order

Go iterates `mod.Actions` (a `map[string]Action`) in random order. Python iterates `mod.actions.items()` in dict insertion order (Python 3.7+). The xtrace output order must match for the golden test.

## Fix Plan

### Step 1: Fix destructor unwrapping in `modifiesRec` for HavocAction

**File:** `actions/transforms.go:146-151`

Go's HavocAction case doesn't unwrap destructors at all — it just returns `a.Target` directly. Python's `HavocAction.modifies()` has the same destructor unwrapping loop as `AssignAction.modifies()`. Add the destructor unwrapping loop to the HavocAction case, mirroring AssignAction.

### Step 2: Pass ActionsConfig to `actions.Modifies()` in CheckDefinitions

**File:** `compiler/ivy_compile.go:1478-1488`

Create an `ActionsConfig` with a context pointing to `mod`, so `isDestructor()` can check `mod.DestructorSorts`. The pattern already exists:

```go
acfg := &actions.ActionsConfig{
    Context: actions.NewActionContext(mod),
}
```

Then call `actions.Modifies(act, acfg)` instead of `actions.Modifies(act)`.

### Step 3: Match Python's xtrace structure — separate logic loop from xtrace loop, deduplicate

**File:** `compiler/ivy_compile.go:1478-1488`

Restructure to match Python's code structure:
1. First loop: build `modified` set (logic, no xtrace) — iterates `mod.Actions` in any order
2. Second loop: emit xtrace with deduplicated counts — iterates `mod.Actions` in sorted order

Use a `map[lg.NodeKey]bool` set for deduplication per action, and report `len(dedupSet)` for the xtrace count.

### Step 4: Add `ActionOrder` to Module for insertion-order iteration

**File:** `module/module.go`

Add `ActionOrder []string` field to `Module` struct. Add a `SetAction(name string, action Action)` method that:
1. Sets `m.Actions[name] = action`
2. Appends `name` to `m.ActionOrder` only if it's not already present

Update `Module.Clear()` to reset `ActionOrder`.
Update `Module.Copy()` to copy `ActionOrder`.

### Step 5: Update all non-test `mod.Actions[name] = ...` sites to use `SetAction`

**Key files with direct insertion (non-test only):**
- `compiler/ivy_compile.go`: lines 557, 794, 976, 1001
- `compiler/phase6.go`: lines 714, 1849, 2051, 2055
- `isolate/isolate.go`: lines 353, 1113
- `isolate/create.go`: lines 161, 250, 255, 331, 775
- `bmc/bmc.go`: lines 90, 561 (via mc/toaiger)
- `mc/toaiger.go`: line 561
- `vmt/vmt.go`: line 233

Test files can keep using direct map assignment since test xtrace output isn't compared in the golden test.

### Step 6: Use `ActionOrder` in CheckDefinitions xtrace loop

**File:** `compiler/ivy_compile.go`

In the xtrace loop (step 3), iterate `mod.ActionOrder` instead of sorting map keys. This faithfully matches Python's insertion-order dict iteration.

## Files to Modify

1. `actions/transforms.go` — Fix HavocAction destructor unwrapping in `modifiesRec`
2. `compiler/ivy_compile.go` — Pass ActionsConfig, restructure loops, use ActionOrder
3. `module/module.go` — Add `ActionOrder []string`, `SetAction()` method, update `Clear()`/`Copy()`
4. `compiler/phase6.go` — Use `SetAction`
5. `isolate/isolate.go` — Use `SetAction`
6. `isolate/create.go` — Use `SetAction`
7. `bmc/bmc.go` — Use `SetAction`
8. `mc/toaiger.go` — Use `SetAction`
9. `vmt/vmt.go` — Use `SetAction`

## Verification

1. `cd ~/go/src/github.com/glycerine/goivy && go build ./...` — must compile
2. `cd ~/goivy && make golden` — xtrace line 146102+ must now match
3. `cd ~/go/src/github.com/glycerine/goivy && go test ./actions/... ./compiler/... ./module/...` — existing tests pass
