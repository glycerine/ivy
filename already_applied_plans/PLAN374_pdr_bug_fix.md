# Fix PDR Bug: Conformant Error Action Construction + forwardClausesIvy Fix

**Created**: 2026-05-03 ~UTC

## Context

PDR reports false counterexamples for inductive conjectures. Two root causes:

1. **`forwardClausesIvy` has an extra filter not in Python** — Go line 52 skips `IsNew` and `IsSkolem` symbols, but Python's `forward_clauses` (ivy_updr.py:43-45) only filters `x != '='` and `x not in inflex`.

2. **Go builds error clauses via `DualClauses` with `@`-prefix skolemizer** (updr.go:100-106) when no explicit "error" action exists. This is a missing port. Python's `ivy_updr.py` (lines 53-59) uses `ag.actions["error"]` — a user-defined action in the Ivy source that follows the pattern `x := *; assume ~property` — and passes it through `.update()` → `reverse_image()`. The error action is compiled from the Ivy source like any other action (ivy_compiler.py:1651).

## What Python Does (the source of truth)

Python Ivy programs define an explicit error action (e.g., client_server.ivy:25-30):
```ivy
individual x,y,z
action error = {
    x := *;              # HavocAction → Modified=[x], TR=True
    y := *;              # HavocAction → Modified=[y], TR=True
    z := *;              # HavocAction → Modified=[z], TR=True
    assume x ~= y & conn(x,z) & conn(y,z)   # AssumeAction
}
```

The Python UPDR flow (ivy_updr.py:53-59):
```python
actions = [ag.actions[lab] for lab in ag.actions if lab != "error"]
err_act = ag.actions["error"].update(ag.domain, state.in_scope)
error = tr.reverse_image([], axioms, err_act)
```

Step-by-step what happens:
1. **Sequence.int_update** composes children via `compose_updates`
2. **compose_updates** renames the AssumeAction's clauses to `new_` vocabulary: `map2 = {v: new(v) for v in updated1}` renames the assume's symbols
3. The composed Update = `Modified=[x,y,z], TR=~safe(new_x,new_y,new_z)`
4. **reverse_image** builds `post_updated=[new_x,new_y,new_z]`, calls `exist_quant`
5. **exist_quant** renames `new_x → __0_new_x` etc. (fresh `__`-prefixed names via UniqueRenamer)
6. Result: clauses with only signature symbols (conn, s) and `__`-prefixed existential witnesses

## Plan

### Step 1: Remove extra filter from `forwardClausesIvy`

**File**: `updr/updr.go`, line 52

Delete these 3 lines:
```go
if actions.IsNew(c.Name) || actions.IsSkolem(c.Name) {
    continue
}
```

Python's `forward_clauses` only filters `x != '='` and `x not in inflex`. The `inflex` set already excludes `IsNew`/`IsSkolem` (both sides do this at the inflex-building step), so those symbols are NOT in `inflex` and DO get renamed in Python.

### Step 2: Construct the error action from negated conjectures

**File**: `updr/updr.go`, replace lines 93-107 (the `DualClauses` fallback)

When no explicit "error" action exists, construct the equivalent of what a Python error action does. Use the Go action system (which faithfully ports the Python action system):

```go
// Build error action matching Python's pattern:
//   Sequence(HavocAction(v1), ..., HavocAction(vN), AssumeAction(~conj))
if len(mod.LabeledConjs) > 0 {
    var conjFmlas []lg.Expr
    for _, lc := range mod.LabeledConjs {
        if lc.Formula != nil {
            conjFmlas = append(conjFmlas, lc.Formula.(lg.Expr))
        }
    }
    if len(conjFmlas) > 0 {
        // Get universally quantified variables from conjecture formulas
        conjClauses := module.NewClauses(conjFmlas, nil, nil)
        vars := module.UsedVariablesOrdered(conjClauses)

        // Create witness constants for each variable (like "individual x:sort")
        // and substitute variables→constants in the conjecture
        varSubs := make(map[lg.NodeKey]lg.Expr)
        var witnessConsts []*lg.Const
        for _, v := range vars {
            c := lg.NewConst(v.Name, v.VSort)
            witnessConsts = append(witnessConsts, c)
            varSubs[lg.Key(v)] = c
        }

        // Build negated conjecture with constants substituted for variables
        var conjAll lg.Expr
        if len(conjFmlas) == 1 {
            conjAll = conjFmlas[0]
        } else {
            conjAll = &lg.And{Terms: conjFmlas}
        }
        negConj := &lg.Not{Body: module.SubstituteConstantsExpr(conjAll, varSubs)}

        // Build Sequence: HavocAction for each witness, then AssumeAction(~conj)
        var seqElems []lg.Expr
        for _, c := range witnessConsts {
            seqElems = append(seqElems, actions.NewHavocAction(c))
        }
        seqElems = append(seqElems, actions.NewAssumeAction(negConj))
        errAction := actions.NewSequence(seqElems...)

        // Get update using standard action composition (ComposeUpdates)
        errUpdate := actions.GetUpdate(errAction, updateCtx)

        // ReverseImage existentially quantifies the witness constants
        errorClauses = actions.ReverseImage(
            module.TrueClauses(nil), axioms, errUpdate)
    }
}
```

**Why this works**: 
- `HavocAction(c).ActionUpdate()` puts c in Modified, TR=True
- `AssumeAction(~conj(c1,...,cN)).ActionUpdate()` sets TR=skolemize(~conj(c1,...,cN))
- `Sequence.IntUpdate` → `ComposeUpdates` renames assume's clauses to new_ vocabulary
- `ReverseImage` → `ExistQuantClauses` quantifies away the new_ witnesses
- Result: clauses with `__`-prefixed witnesses and signature symbols only

**Key detail — variable→constant substitution**: The conjecture `flag(X)` uses variable X. The error action body needs constant X (like Python's `individual x`). We substitute Variable→Const using `module.SubstituteConstantsExpr` with `varSubs` map keyed by `lg.Key(variable)`. Since `substituteNodesRec` (ops.go:706-721) checks `lg.Key(n)` against the subs map for ALL node types (not just constants), this substitution works for variables too despite the function name.

### Step 3: Remove debug print statements

**File**: `updr/updr.go`, lines ~217-283

Remove the `fmt.Printf("PDR DEBUG ...")` statements.

### Step 4: Verify `updateCtx` is available

The `updateCtx` variable (from `makeUpdateContext(mod)`) is already created at line 114 of `CheckModule`. The error clause construction at lines 93-107 happens before it. Need to move `updateCtx` creation before the error clause block, or create a second one. The simplest fix: move the `updateCtx := makeUpdateContext(mod)` line to before the error clause construction.

Currently:
```
lines 80-111: build errorClauses (updateCtx NOT yet created)
line 114:     updateCtx := makeUpdateContext(mod)
```

Move to:
```
line ~80:     updateCtx := makeUpdateContext(mod)
lines 81-111: build errorClauses (can now use updateCtx)
```

The first `updateCtx` at line 114 can be removed since we'll reuse the earlier one.

## Files to Modify

- `updr/updr.go` — all changes:
  - Move `updateCtx` creation earlier (~line 80)
  - Line 52: remove `IsNew || IsSkolem` filter
  - Lines 93-107: replace `DualClauses` with Havoc+Assume+Sequence+GetUpdate+ReverseImage
  - Lines ~217-283: remove debug prints

## Existing Functions Reused (no new code outside updr.go)

| Function | File | Purpose |
|----------|------|---------|
| `actions.NewHavocAction(target)` | actions/action.go:330 | Create havoc action |
| `actions.NewAssumeAction(fmla)` | actions/action.go:158 | Create assume action |
| `actions.NewSequence(elems...)` | actions/action.go:125 | Compose into sequence |
| `actions.GetUpdate(action, ctx)` | actions/update.go:2190 | Compute update (calls IntUpdate+ComposeUpdates) |
| `actions.ReverseImage(post, ax, u)` | actions/transrel.go:1426 | Compute pre-image (calls ExistQuantClauses) |
| `module.UsedVariablesOrdered(cl)` | module/ops.go:998 | Extract free vars from clauses |
| `module.SubstituteConstantsExpr(e, s)` | module/astutil.go:135 | Substitute nodes (works for vars too via substituteNodesRec) |

## Verification

Run `cd ~/ivy/goivy && make test`:
- `TestCheckModule_InductiveConjectureValid` — propositional PDR, no universal vars
- `TestCheckModule_InductiveWithInitializer` — full pipeline with `conjecture flag(X)`, exercises the Havoc+Assume error construction with universal variable X
- All 40+ other PDR unit tests still pass
