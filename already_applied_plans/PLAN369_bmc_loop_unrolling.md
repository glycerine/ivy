# Plan: BMC Loop Unrolling (Todo Item #7)

**Created:** 2026-05-02 06:12 UTC

## Context

BMC (Bounded Model Checking) requires finite unrolling of `While` loops. Python's `ivy_bmc.py` calls `action.unroll_loops(lambda x: n_unroll)` on every action before running the BMC check. The Go `bmc/bmc.go:244-248` stub `UnrollAction()` returns the action unchanged, meaning BMC silently treats every while loop as zero iterations — producing wrong results.

**Key discovery:** The core unrolling algorithm is **already implemented** in two places:
- `actions/transforms.go:603-694` — `UnrollLoops(action, card)` + `unrollWhile()` (standalone recursive tree walker, panics on errors)
- `actions/update.go:1612-1667` — `WhileAction.Unroll(card, body)` method (returns errors, used by `IntUpdate` with `UnrollContext`)

Neither has any callers from `bmc/`. The fix is primarily **wiring**, not algorithmic.

## Changes

### 1. Replace `bmc.UnrollAction` stub

**File:** `/Users/jaten/ivy/goivy/bmc/bmc.go` lines 241-248

Replace with a function that:
- Creates a `CardFunc` that returns `n` for any sort (matching Python's `lambda x: n_unroll`)
- Calls `actions.UnrollLoops(act, cardFunc)`
- Recovers from panics in `unrollWhile` (which panics on cardinality > 100 or < 0) and returns the original action if unrolling fails

```go
func UnrollAction(act actions.Action, n int) actions.Action {
    if act == nil {
        return nil
    }
    card := actions.CardFunc(func(s lg.Sort) int {
        return n
    })
    var result actions.Action
    func() {
        defer func() {
            if r := recover(); r != nil {
                result = act
            }
        }()
        result = actions.UnrollLoops(act, card)
    }()
    return result
}
```

No new imports needed — `actions` and `lg` are already imported.

### 2. Wire `check/vmt.go` for FSMC method

**File:** `/Users/jaten/ivy/goivy/check/vmt.go` lines 43-53

Python's VMT code enters an `UnrollContext` with `im.module.sort_card` before computing the update for the FSMC method. Add the same:

```go
func actionToTR(m *module.Module, action actions.Action, method string) (...) {
    bgt := m.BackgroundTheory(nil)

    // For fsmc method, unroll loops before computing the update.
    if method == "fsmc" && m.Cfg != nil && m.Cfg.ActCfg != nil {
        uc := actions.NewUnrollContext(m.SortCard, m, m.Cfg.ActCfg)
        uc.Enter()
        defer uc.Exit()
    }

    upd := computeUpdate(m, action)
    // ... rest unchanged
```

`actions` is already imported. `m.SortCard` is defined at `module/module.go:635` and matches Python's `im.module.sort_card`.

### 3. Update and expand `bmc/bmc_test.go`

**File:** `/Users/jaten/ivy/goivy/bmc/bmc_test.go`

Add `il "github.com/glycerine/ivy/goivy/ivylogic"` to imports.

Add helper:
```go
func mkWhileAction(sortName string) *actions.WhileAction {
    sortT := &lg.UninterpretedSort{Name: sortName}
    ltSym := lg.NewConst("<", il.RelationSort([]lg.Sort{sortT, sortT}))
    xSym := lg.NewConst("x", sortT)
    boundSym := lg.NewConst("bound", sortT)
    cond, _ := lg.NewApply(ltSym, xSym, boundSym)
    body := actions.NewAssignAction(xSym, xSym)
    return actions.NewWhileAction(cond, body)
}
```

Update existing `TestUnrollAction` and add:

| Test | What it verifies |
|------|-----------------|
| `TestUnrollAction_NonWhile` | AssumeAction passes through (cloned, structurally same) |
| `TestUnrollAction_WhileUnrolled` | WhileAction with n=3 produces nested IfAction |
| `TestUnrollAction_NilAction` | Returns nil for nil input |
| `TestUnrollAction_ZeroUnroll` | n=0 produces single base-case IfAction (if cond then assume false) |
| `TestUnrollAction_LargeUnroll` | n=200 recovers from panic, returns original action |
| `TestUnrollAction_WhileInsideSequence` | Sequence containing a WhileAction: while is unrolled, sequence preserved |
| `TestUnrollAction_NestedWhile` | While containing while: both get unrolled |
| `TestUnrollAction_FormalsPreserved` | FormalParams/FormalReturns survive unrolling |
| `TestCheckIsolateWithUnroll_WhileAction` | End-to-end: CheckIsolate with NUnroll set, module contains a while action, actions restored after |

### 4. Add `actions/transforms_test.go`

**File:** `/Users/jaten/ivy/goivy/actions/transforms_test.go` (new file)

Tests for `UnrollLoops` directly (package-internal access):

| Test | What it verifies |
|------|-----------------|
| `TestUnrollLoops_Nil` | Returns nil for nil |
| `TestUnrollLoops_NoWhile` | AssumeAction cloned but structurally identical |
| `TestUnrollLoops_SimpleWhile` | WhileAction with card=3: top-level is IfAction, 3 levels of nesting |
| `TestUnrollLoops_WhileInSequence` | Sequence(Assume, While, Assume): while replaced, assumes preserved |
| `TestUnrollLoops_WhileInIf` | If(cond, While, Assume): while in then-branch unrolled |
| `TestUnrollLoops_NestedWhile` | While(body=While(...)): inner unrolled first, then outer |
| `TestUnrollLoops_NotEqCondition` | While condition is `!(x = bound)`: extracts sort from Eq |
| `TestUnrollLoops_AndCondition` | While condition is `And(x < bound, flag)`: peels And to find `<` |
| `TestUnrollLoops_UnknownCondition` | While condition is just a boolean var: CardFunc gets nil sort |
| `TestUnrollLoops_FormalsPreserved` | FormalParams survive through clone |
| `TestUnrollLoops_PanicsOnLargeCard` | card returns 200: panics (test with recover) |
| `TestUnrollLoops_PanicsOnNegativeCard` | card returns -1: panics |
| `TestUnrollLoops_VerifyNesting` | card=2: verify exact structure: `if(cond, seq(body, if(cond, seq(body, if(cond, assume(false))))))` |
| `TestUnrollLoops_ChoiceAction` | ChoiceAction with while in one branch: while unrolled |

### 5. Verify structure of unrolled output

For `card=N`, the expected structure is:
```
IfAction(cond,                          -- iteration N
  Sequence(body,
    IfAction(cond,                      -- iteration N-1
      Sequence(body,
        ...
          IfAction(cond,                -- iteration 1
            Sequence(body,
              IfAction(cond,            -- base case
                AssumeAction(Or{}))))   -- assume false
```

Tests should walk the tree and count nesting depth to verify correctness.

## Verification

1. `cd ~/ivy/goivy && make test` — must pass all existing and new tests
2. Verify `TestUnrollAction_WhileUnrolled` produces IfAction (not WhileAction) at top level
3. Verify `TestUnrollAction_LargeUnroll` does not panic (recovers gracefully)
4. Verify `TestCheckIsolateWithUnroll` restores original actions after BMC completes

## Critical files to modify
- `/Users/jaten/ivy/goivy/bmc/bmc.go` — replace UnrollAction stub (lines 241-248)
- `/Users/jaten/ivy/goivy/bmc/bmc_test.go` — update/add tests
- `/Users/jaten/ivy/goivy/check/vmt.go` — wire fsmc unrolling (lines 43-53)
- `/Users/jaten/ivy/goivy/actions/transforms_test.go` — new test file

## Files to reuse (no changes)
- `/Users/jaten/ivy/goivy/actions/transforms.go:603-694` — `UnrollLoops`, `unrollWhile`, `CardFunc`
- `/Users/jaten/ivy/goivy/actions/update.go:1612-1667` — `WhileAction.Unroll` method
- `/Users/jaten/ivy/goivy/actions/phase3.go:42-51` — `NewUnrollContext`
- `/Users/jaten/ivy/goivy/module/module.go:635` — `Module.SortCard`
