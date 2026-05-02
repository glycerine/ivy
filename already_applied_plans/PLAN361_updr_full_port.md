# Plan: Complete the UPDR/PDR Engine Port from Python to Go

**Created:** 2026-05-02 ~17:45 UTC

## Context

The Python UPDR (Unbounded Property-Directed Reachability) engine is the core
verification algorithm in Ivy. It uses Z3 to iteratively strengthen inductive
invariants until proving a property or finding a counterexample trace.

The Go port has substantial infrastructure already in place — the IC3/PDR core
algorithm, interactive UPDR, tactics, transrel, and Z3 bridge are all ported.
However, `CheckModule()` in `updr/updr.go` is a placeholder that returns
"always safe" without doing real verification, and the WebUI's PDR mode is
dead because this backend is hollow.

**Python source of truth:** `~/ivy/pyivy/ivy/ivy/ivy_updr.py` (155 lines)
**Python PDR algorithm:** `mini_pdr.py` (external, not in repo — derived from
Z3's `mini_ic3.py` at `~/go/src/github.com/Z3Prover/z3/examples/python/mini_ic3.py`)

## What's Already Done

| Component | File | Status |
|-----------|------|--------|
| IC3/PDR core algorithm | `updr/pdr.go` (567 lines, 34 tests) | Complete (from mini_ic3.py) |
| UPDR helpers | `updr/updr.go` (ForwardClauses, NumClauses, NumUnivClauses) | Complete |
| Interactive UPDR | `iupdr/iupdr.go` (660 lines, 4 tests) | Complete (from iupdr.py) |
| UPDR tactic | `tactics/tactics.go:421-529` | Complete (from tactics.py) |
| Transition relations | `actions/transrel.go` (ForwardImage, ReverseImage, FrameUpdate) | Complete |
| Z3 bridge | `z3bridge/` (ClausesToZ3, NativeSymbol, Solver) | Complete |
| WebUI scaffolding | `webui/` (buttons, modes, endpoints) | Scaffolding only |

## What's NOT Done

### 1. PDR constructor mismatch — `updr/pdr.go`

**Python mini_pdr.PDR constructor** (8 args):
```python
pdr = pd.PDR(init_z3, rho_z3, bad_z3, background_z3, gsyms, lsyms, relations_z3, True)
```

**Go NewPDR constructor** (7 args, missing background/symbols):
```go
func NewPDR(ctx *z3bridge.Z3Context, init, trans, bad z3bridge.Expr,
    x0, inputs, xn []z3bridge.Expr) *PDR
```

The Go PDR is missing:
- `background` (z3bridge.Expr) — background axioms asserted in every solver
- `gsyms` ([]z3bridge.Expr) — global/inflexible symbols
- `lsyms` ([][2]z3bridge.Expr) — local/flexible symbol pairs (current, next)
- `relations` ([]z3bridge.Expr) — relation symbols for generalization
- `useRelations` (bool) — flag controlling relation-based generalization

### 2. CheckModule() stub — `updr/updr.go:44-97`

Currently creates a trivial safe system and returns "always safe". Must be
replaced with the full Python ivy_updr.py `__main__` algorithm.

### 3. WebUI integration stubs

- `webui/session.go:798-804` — RunCheck("pdr") returns canned "pass"
- `webui/session.go:287-289` — "pdr_step" returns "not yet wired"

---

## Implementation Plan

### Step 1: Extend PDR constructor to match Python mini_pdr.PDR

**File:** `updr/pdr.go`

Add fields to the `PDR` struct:
```go
type PDR struct {
    // ... existing fields ...
    background   z3bridge.Expr      // background axioms
    gsyms        []z3bridge.Expr    // global/inflexible symbols
    lsyms        [][2]z3bridge.Expr // local symbol pairs: [0]=current, [1]=next
    relations    []z3bridge.Expr    // relation symbols
    useRelations bool               // flag for relation-based generalization
}
```

Update `NewPDR` signature:
```go
func NewPDR(ctx *z3bridge.Z3Context, init, trans, bad, background z3bridge.Expr,
    gsyms []z3bridge.Expr, lsyms [][2]z3bridge.Expr,
    relations []z3bridge.Expr, useRelations bool) *PDR
```

Wire background axioms into solver creation:
- In `NewPDR`: assert `background` in the F0 solver alongside init
- In `addFrame`: assert `background` in every new frame's solver
- In `unfold`, `isInductive`: background is already in the frame solvers

Wire lsyms into variable handling:
- Derive x0 and xn from lsyms: `x0[i] = lsyms[i][0]`, `xn[i] = lsyms[i][1]`
- Keep inputs separate (from gsyms that are not relations, if applicable)

Update all 34 existing tests to use new constructor signature (add nil/empty
for the new parameters so existing tests still pass unchanged).

### Step 2: Implement CheckModule() — faithful port of ivy_updr.py __main__

**File:** `updr/updr.go`

Replace the stub with a faithful port. Line-by-line mapping:

```
Python                                          Go
------                                          --
ag = ivy.ivy_init()                             mod *module.Module (parameter)
state = ag.states[0]                            state := mod.InitState() or equivalent
sig = ag.domain.sig                             sig := mod.Sig

# error states
err_act = ag.actions["error"].update(...)       errAct, _ := mod.Actions.Get2("error")
                                                errUpd := actions.GetUpdate(errAct, updateCtx)
error = tr.reverse_image([],axioms,err_act)     error := actions.ReverseImage(module.TrueClauses(), axioms, errUpd)

# actions (non-error)
actions = [ag.actions[lab]                      for name, act := range mod.Actions.All() {
    for lab in ag.actions if lab != "error"]         if name != "error" { ... }
                                                }

# updates
updates = [action.update(ag.domain,...)]        updates[i] = actions.GetUpdate(act, updateCtx)

# classify symbols
flex = set(sym for u in updates for sym in u[0])
inflex = set(sym not in flex, not new, not skolem)
skolems = set(sym if is_skolem)

# make frames explicit
updates = [tr.frame_update(u,flex,sig)]         updates[i] = actions.FrameUpdate(upd, flexConsts)

# build Z3
ns = sv.native_symbol                           ns := z3bridge.NativeSymbol
lsyms = [(ns(sym),ns(tr.new(sym))) for flex]
gsyms = [ns(sym) for sym in inflex]

init = forward_clauses(state.clauses, inflex)
error = forward_clauses(error, inflex)
axioms += forward_clauses(axioms, inflex)

init_z3 = sv.clauses_to_z3(init)               solver.ClausesToZ3(init)
rho_z3 = Or(*[sv.clauses_to_z3(simplify(a[1])) for a in updates])
bad_z3 = sv.clauses_to_z3(error)
background_z3 = sv.clauses_to_z3(axioms)
relations_z3 = [filter gsyms+lsyms by BoolSort]

pdr = pd.PDR(init_z3, rho_z3, bad_z3, background_z3, gsyms, lsyms, relations_z3, True)
outcome = pdr.run()
```

Key Go APIs to use:
- `actions.GetUpdate(action, ctx)` → `*actions.Update{Modified, TR, Pre}`
- `actions.ReverseImage(postState, axioms, update)` → `*module.Clauses`
- `actions.FrameUpdate(u, inScope)` → `*Update`
- `actions.New(name)`, `actions.IsNew(name)`, `actions.IsSkolem(name)`
- `z3bridge.NewSolver(mod, opts)` → solver with `ClausesToZ3()`
- `z3bridge.NativeSymbol(sig, sym)` → Z3 symbol
- `module.SimplifyClauses(c)` → simplified clauses

### Step 3: Wire WebUI session.RunCheck("pdr")

**File:** `webui/session.go:798-804`

Replace the canned response with:
```go
case "pdr":
    result, err := updr.CheckModule(s.Module)
    if err != nil {
        return &CheckResult{Result: "error", Message: err.Error()}
    }
    if result.Valid {
        return &CheckResult{
            Result:  "pass",
            Message: fmt.Sprintf("Invariant found (%d clauses, %d universal)",
                result.Stats.NumClauses, result.Stats.NumUnivClauses),
        }
    }
    return &CheckResult{
        Result:  "fail",
        Message: result.Error,
    }
```

### Step 4: Wire WebUI "pdr_step" action

**File:** `webui/session.go:287-289`

Wire the "pdr_step" action to a single PDR strengthening step. This should
use the UPDR tactic (tactics/tactics.go) or iupdr (iupdr/iupdr.go) depending
on the Python behavior. Need to check Python `ivy_graph_ui.py pdr_step()` for
the exact dispatch.

### Step 5: Comprehensive unit tests

**File:** `updr/updr_test.go` (extend existing)

A. Update all 34 existing tests for new NewPDR signature (add empty background/symbols).

B. Add CheckModule integration tests:
- `TestCheckModule_EmptyModule` — empty module returns valid (no conjectures to violate)
- `TestCheckModule_NilModule` — nil module returns error
- `TestCheckModule_NoActions` — module with init but no actions
- `TestCheckModule_NoError` — module without "error" action
- `TestCheckModule_TrivialSafe` — synthetic module with safe property
- `TestCheckModule_TrivialUnsafe` — synthetic module with violable property

C. Add PDR tests with background axioms (new parameter):
- `TestPDR_WithBackground_Safe` — background constrains state space, making property hold
- `TestPDR_WithBackground_Unsafe` — background doesn't prevent counterexample
- `TestPDR_EmptyBackground` — nil/true background behaves like no background

D. Add PDR tests with gsyms/lsyms/relations:
- `TestPDR_WithSymbolClassification` — verify symbol info is stored correctly
- `TestPDR_WithRelations` — verify relation-based generalization (if implemented)

---

## Files to Modify

| File | Change |
|------|--------|
| `updr/pdr.go` | Extend PDR struct + NewPDR signature with background, gsyms, lsyms, relations, useRelations. Wire background into solvers. |
| `updr/updr.go` | Replace CheckModule() stub with full implementation. |
| `updr/updr_test.go` | Update 34 existing tests for new signature. Add integration tests. |
| `webui/session.go` | Wire RunCheck("pdr") to updr.CheckModule(). Wire "pdr_step". |

## Files to Read (reference during implementation)

| File | Why |
|------|-----|
| `~/ivy/pyivy/ivy/ivy/ivy_updr.py` | Source of truth for CheckModule |
| `~/go/src/github.com/Z3Prover/z3/examples/python/mini_ic3.py` | Source of truth for PDR algorithm |
| `actions/transrel.go` | ForwardImage, ReverseImage, FrameUpdate APIs |
| `z3bridge/solver.go` | ClausesToZ3, NativeSymbol APIs |
| `z3bridge/quantifier.go` | Z3Context, Z3Solver, Expr types |
| `module/module.go` | Module struct, Actions, Sig |

## Risks and Mitigations

1. **mini_pdr.py not available** — The Python `import mini_pdr as pd` references
   a package not found on disk. We have `mini_ic3.py` (Z3 example) at
   `~/go/src/github.com/Z3Prover/z3/examples/python/mini_ic3.py` which the Go
   pdr.go is ported from. The additional parameters (background, gsyms, lsyms,
   relations) that mini_pdr.PDR takes are clear from the ivy_updr.py call site.
   We extend the Go PDR to accept them and assert background in frame solvers.

2. **UpdateContext construction** — CheckModule needs to create an
   `actions.UpdateContext` to call `GetUpdate()`. Verified: UpdateContext
   requires `Domain *module.Module`, `PVars map[string]bool` (in-scope vars),
   `ActCfg *ActionsConfig`, and `GetAction func(name string) Action`.
   Construct from the Module parameter:
   ```go
   ctx := &actions.UpdateContext{
       Domain: mod,
       PVars:  allSymbolNames(mod),
       GetAction: func(name string) actions.Action {
           if a, ok := mod.Actions.Get2(name); ok {
               return a.(actions.Action)
           }
           return nil
       },
   }
   ```

3. **ForwardClauses: two levels** — Python `forward_clauses` in ivy_updr.py
   renames at the Ivy Clauses level using `lu.rename_clauses` + `lu.used_symbols_clauses`.
   Go equivalents exist: `module.RenameClauses(clauses, subs)` and
   `module.UsedSymbolsClauses(clauses)`. The existing Go `ForwardClauses` in
   updr.go operates at Z3 level (`ctx.Substitute`); keep it for Z3-level use.
   In CheckModule, do the renaming at Ivy level (before ClausesToZ3), matching
   the Python flow exactly.

4. **Background axioms in PDR** — The background must be asserted in every
   frame solver so that all SAT checks respect axioms. In `NewPDR`: assert
   background in F0 solver. In `addFrame`: assert background + trans in each
   new solver. This matches how Python mini_pdr.PDR presumably uses it (and
   matches mini_ic3.py's pattern of asserting trans in each new State solver).

## Verification

1. `cd ~/ivy/goivy && make test` — all existing tests must pass
2. New updr tests exercise CheckModule with synthetic modules
3. WebUI manual test: load an Ivy file, click "PDR" check mode, verify it
   calls the real engine instead of returning canned results
