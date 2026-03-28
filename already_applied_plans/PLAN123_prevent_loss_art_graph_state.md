# Plan: Fix DecomposeState Chain Conversion + Rename art.Expr → art.Provenance

**Created:** 2026-03-28 19:30

## Context

`art.DecomposeState` calls `interp.DecomposeActionApp`, which builds and returns an `interp.State` chain where each non-root state has its `.Expr` field (type `ast.Node`) set to an action-application expression. When `InterpToArtState` converts these back to `art.State` objects, it **silently drops** the `.Expr` field because the two packages use incompatible types for it. The result: the decomposition subgraph has states but zero transitions.

Separately, the name `Expr` is overloaded everywhere (`lg.Expr`, `ast.Node`, `interp.State.Expr`). We rename the art package's derivation-tracking type to `Provenance` for clarity.

---

## Terminology (used consistently throughout this plan)

| Term | What it is | Type |
|------|-----------|------|
| `interp.State.Expr` | Field on `interp.State`. Records how the state was derived. | `ast.Node` |
| `art.Provenance` | **New name** for the interface currently called `art.Expr`. Marker interface with `provenanceMarker()`. | interface |
| `art.State.Prov` | **New name** for the field currently called `art.State.Expr`. Records how the state was derived. | `art.Provenance` |
| `*art.ActionApp` | Implements `art.Provenance`. Wraps `*art.State` pointers. | struct |
| `*art.StateJoin` | Implements `art.Provenance`. Wraps `*art.State` pointers. | struct |
| `stateNode` | Internal to interp. Wraps `*interp.State` as `ast.Node`. | struct |

---

## Root Cause

`InterpToArtState` copies Clauses, Label, InScope, Action, Update, Pred — but never touches `interp.State.Expr`. It cannot, because `interp.State.Expr` is `ast.Node` (containing `stateNode` wrappers around `*interp.State` pointers), while `art.State.Prov` is `art.Provenance` (containing `*art.State` pointers). There is no conversion between these two expression tree shapes.

**Data flow today:**
```
interp.DecomposeActionApp returns interp.State chain
    each non-root state has .Expr = ast.Atom{Rep: actionName, Terms: [stateNode{pred}]}
        ↓
InterpToArtState converts to art.State
    copies everything EXCEPT .Expr → art.State.Prov is nil
        ↓
ConstructTransitionsFromExpressions finds .Prov == nil on every state
    produces zero transitions
```

---

## Pre-Step: Rename in art/ package

Rename throughout the `art/` package only:

| Before | After |
|--------|-------|
| `type Expr interface { exprMarker() }` | `type Provenance interface { provenanceMarker() }` |
| `art.State.Expr Expr` field | `art.State.Prov Provenance` field |
| `func (*ActionApp) exprMarker()` | `func (*ActionApp) provenanceMarker()` |
| `func (*StateJoin) exprMarker()` | `func (*StateJoin) provenanceMarker()` |
| `IsActionApp(e Expr)` | `IsActionApp(p Provenance)` |
| `IsStateJoin(e Expr)` | `IsStateJoin(p Provenance)` |

**Not renamed:** `interp.State.Expr` stays as `Expr`. It is a different package with a different type (`ast.Node`).

**External callers** that reference `art.State.Expr` or the `art.Expr` type must be updated. Search: `\.Expr` in files that import `art/`, and `art\.Expr`.

---

## Step 1: Rewrite InterpToArtState with memoization and Prov conversion

Replace the current non-memoized, Prov-dropping implementation:

```go
// InterpToArtState converts an interp.State (and its predecessor chain)
// to an art.State, including converting interp.State.Expr (ast.Node)
// into art.State.Prov (art.Provenance).
func InterpToArtState(is *interp.State) *State {
    return interpToArtMemo(is, make(map[*interp.State]*State))
}

func interpToArtMemo(is *interp.State, memo map[*interp.State]*State) *State {
    if is == nil {
        return nil
    }
    if s, ok := memo[is]; ok {
        return s
    }

    s := NewState(is.Domain, is.Clauses)
    memo[is] = s // insert before recursing to break cycles

    s.Label = is.Label
    s.InScope = is.InScope
    s.Action = is.Action
    s.ActionName = is.ActionName
    if is.Update() != nil {
        s.Update = is.Update()
    }
    if is.Pred() != nil {
        s.Pred = interpToArtMemo(is.Pred(), memo)
    }
    if is.JoinOf != nil {
        for _, jo := range is.JoinOf {
            s.JoinOf = append(s.JoinOf, interpToArtMemo(jo, memo))
        }
    }

    // Convert interp.State.Expr (ast.Node) → art.State.Prov (art.Provenance)
    s.Prov = interpExprToProvenance(is.Expr, memo)

    return s
}
```

## Step 2: Add interpExprToProvenance helper

This converts the ast.Node expression tree used by interp into the art.Provenance tree, translating embedded `stateNode` wrappers into the corresponding `*art.State` via the memo map.

```go
// interpExprToProvenance converts an interp-package ast.Node expression
// into an art.Provenance value. Returns nil for unrecognized expressions.
//
// Mapping:
//   interp ActionApp (ast.Atom, 1 stateNode arg) → *art.ActionApp
//   interp StateJoin (ast.Or, stateNode args)     → *art.StateJoin
func interpExprToProvenance(expr ast.Node, memo map[*interp.State]*State) Provenance {
    if expr == nil {
        return nil
    }
    if interp.IsActionApp(expr) {
        atom := expr.(*ast.Atom)
        var rep interface{} = atom.Rep
        var args []*State
        for _, term := range atom.Terms {
            if is := interp.UnwrapState(term); is != nil {
                args = append(args, interpToArtMemo(is, memo))
            }
        }
        return &ActionApp{Rep: rep, Args: args}
    }
    if interp.IsStateJoin(expr) {
        or := expr.(*ast.Or)
        var args []*State
        for _, term := range or.Terms {
            if is := interp.UnwrapState(term); is != nil {
                args = append(args, interpToArtMemo(is, memo))
            }
        }
        return &StateJoin{Args: args}
    }
    return nil
}
```

## Step 3: Symmetric — rewrite ArtToInterpState with memoization and Expr conversion

```go
func ArtToInterpState(s *State) *interp.State {
    return artToInterpMemo(s, make(map[*State]*interp.State))
}

func artToInterpMemo(s *State, memo map[*State]*interp.State) *interp.State {
    if s == nil {
        return nil
    }
    if is, ok := memo[s]; ok {
        return is
    }

    sv := interp.NewStateValue(nil, s.Clauses, clauseops.FalseClauses(nil))
    is := interp.NewState(s.Domain, sv, nil, s.Label)
    memo[s] = is

    is.InScope = s.InScope
    is.Action = s.Action
    is.ActionName = s.ActionName
    if s.Update != nil {
        is.SetUpdate(s.Update)
    }
    if s.Pred != nil {
        is.SetPred(artToInterpMemo(s.Pred, memo))
    }
    // JoinOf
    if s.JoinOf != nil {
        for _, jo := range s.JoinOf {
            is.JoinOf = append(is.JoinOf, artToInterpMemo(jo, memo))
        }
    }

    // Convert art.State.Prov (art.Provenance) → interp.State.Expr (ast.Node)
    is.Expr = provenanceToInterpExpr(s.Prov, s.Domain, memo)

    return is
}
```

```go
// provenanceToInterpExpr converts an art.Provenance back to an ast.Node
// for interp.State.Expr.
func provenanceToInterpExpr(prov Provenance, domain *module.Module, memo map[*State]*interp.State) ast.Node {
    if prov == nil {
        return nil
    }
    cfg := ast.NewAstConfig()
    if domain != nil && domain.Cfg != nil && domain.Cfg.AstCfg != nil {
        cfg = domain.Cfg.AstCfg
    }
    switch p := prov.(type) {
    case *ActionApp:
        if len(p.Args) > 0 {
            interpPred := artToInterpMemo(p.Args[0], memo)
            actionName := fmt.Sprintf("%v", p.Rep)
            return interp.ActionApp(cfg, actionName, interp.WrapState(interpPred))
        }
    case *StateJoin:
        var terms []lg.Expr
        for _, s := range p.Args {
            terms = append(terms, interp.WrapState(artToInterpMemo(s, memo)))
        }
        return &ast.Or{Terms: terms}
    }
    return nil
}
```

Note: `interp.ActionApp(cfg, name, arg)` takes a single `ast.Node` arg, matching Python's `action_app(action, state)` which always has exactly one predecessor.

## Step 4: Add ActionName field to art.State

```go
type State struct {
    // ... existing fields ...
    ActionName string // name of the action (matches interp.State.ActionName)
}
```

This field exists on `interp.State` but was missing from `art.State`. Needed for faithful round-tripping.

## Step 5: Simplify DecomposeState

After Step 1, `InterpToArtState` now recursively converts the entire chain with Prov intact. DecomposeState can be simplified:

```go
func (ag *AnalysisGraph) DecomposeState(state *State) *AnalysisGraph {
    // ... nil checks, cache check (same as today) ...

    // Call interp.DecomposeActionApp (same as today)
    resultState, err := interp.DecomposeActionApp(...)
    if err != nil || resultState == nil {
        return nil
    }

    // Convert the entire chain at once (memo handles sharing)
    artResult := InterpToArtState(resultState)

    // Collect chain root-first
    var chain []*State
    for cur := artResult; cur != nil; cur = cur.Pred {
        chain = append(chain, cur)
    }
    // Reverse
    for i, j := 0, len(chain)-1; i < j; i, j = i+1, j-1 {
        chain[i], chain[j] = chain[j], chain[i]
    }

    // Build subgraph
    otherArt := NewAnalysisGraph(ag.Domain)
    for _, st := range chain {
        otherArt.Add(st, st.Prov)  // Prov is now populated!
    }
    otherArt.ConstructTransitionsFromExpressions()

    // Cache
    aa.Subgraph = otherArt
    return otherArt
}
```

The key difference: `st.Prov` is no longer nil, so `ConstructTransitionsFromExpressions` finds `ActionApp` provenance on each non-root state and creates the transitions.

## Step 6: Update ConstructTransitionsFromExpressions for rename

This function currently checks `state.Expr` — update to `state.Prov`:

```go
func (ag *AnalysisGraph) ConstructTransitionsFromExpressions() {
    for _, state := range ag.States {
        if state.Prov == nil {
            continue
        }
        aa, ok := state.Prov.(*ActionApp)
        // ... rest unchanged ...
    }
}
```

And update all other internal references from `.Expr` to `.Prov` throughout the art/ package.

---

## Files Modified

| File | Changes |
|------|---------|
| `~/goivy/art/art.go` | Rename `Expr`→`Provenance` type, `.Expr`→`.Prov` field; add `ActionName` to State; rewrite `InterpToArtState`/`ArtToInterpState` with memo + conversion helpers; simplify `DecomposeState` |
| `~/goivy/art/art_test.go` | Update `.Expr` → `.Prov` references |
| `~/goivy/art/precond_test.go` | Same field rename |
| `~/goivy/art/port_completeness_test.go` | Same field rename |
| `~/goivy/art/phase7.go` | If it references `.Expr` |
| `~/goivy/art/cyrender.go` | If it references `.Expr` |
| `~/goivy/art/decompose_test.go` | **New** — comprehensive tests |
| External callers (`check/`, `bmc/`, `trace/`, etc.) | Update any references to `art.State.Expr` → `art.State.Prov` |

## Existing Facilities to Reuse

| Function | Location | Purpose |
|----------|----------|---------|
| `interp.IsActionApp` | `interp/interp.go:334` | Check if ast.Node is an action application |
| `interp.IsStateJoin` | `interp/interp.go:340` | Check if ast.Node is a state join |
| `interp.UnwrapState` | `interp/interp.go:237` | Extract `*interp.State` from `stateNode` |
| `interp.WrapState` | `interp/interp.go:233` | Wrap `*interp.State` as `ast.Node` |
| `interp.ActionApp` | `interp/interp.go:346` | Create `ast.Atom` action application |

---

## Tests (`~/goivy/art/decompose_test.go`)

### Unit Tests

```go
// --- InterpToArtState conversion ---

// TestInterpToArtStatePreservesProv verifies that interp.State.Expr
// (type ast.Node) is converted to art.State.Prov (type art.Provenance).
func TestInterpToArtStatePreservesProv(t *testing.T) { ... }

// TestInterpToArtStateActionAppProv verifies an interp ActionApp
// expression becomes an *art.ActionApp with correct Rep and Args.
func TestInterpToArtStateActionAppProv(t *testing.T) { ... }

// TestInterpToArtStateJoinProv verifies an interp Or expression
// becomes an *art.StateJoin.
func TestInterpToArtStateJoinProv(t *testing.T) { ... }

// TestInterpToArtStateChainAllNonRootHaveProv verifies that converting
// a multi-step interp.State chain populates Prov on every non-root state.
func TestInterpToArtStateChainAllNonRootHaveProv(t *testing.T) { ... }

// TestInterpToArtStateMemoIdentity verifies the same interp.State pointer
// always maps to the same art.State pointer (no duplicates).
func TestInterpToArtStateMemoIdentity(t *testing.T) { ... }

// TestInterpToArtStateNilExprGivesNilProv verifies root states (Expr==nil)
// produce Prov==nil.
func TestInterpToArtStateNilExprGivesNilProv(t *testing.T) { ... }

// --- ArtToInterpState conversion ---

// TestArtToInterpStatePreservesExpr verifies art.State.Prov is converted
// to interp.State.Expr.
func TestArtToInterpStatePreservesExpr(t *testing.T) { ... }

// --- Round-trip ---

// TestRoundTripPreservesProvStructure verifies art→interp→art preserves
// the Provenance type (ActionApp stays ActionApp, rep and arg count match).
func TestRoundTripPreservesProvStructure(t *testing.T) { ... }

// --- DecomposeState ---

// TestDecomposeStateProducesTransitions verifies the subgraph has
// actual transitions (the whole point of this fix).
func TestDecomposeStateProducesTransitions(t *testing.T) { ... }

// TestDecomposeStateCachesSubgraph verifies second call returns cache.
func TestDecomposeStateCachesSubgraph(t *testing.T) { ... }

// TestDecomposeStateNilReturnsNil verifies nil/empty cases.
func TestDecomposeStateNilReturnsNil(t *testing.T) { ... }

// TestConstructTransitionsFromProvenance verifies that
// ConstructTransitionsFromExpressions uses Prov to build transitions.
func TestConstructTransitionsFromProvenance(t *testing.T) { ... }
```

### Fuzz Tests

```go
// FuzzInterpToArtChainProv builds interp.State chains of varying length
// and asserts every non-root converted art.State has non-nil Prov.
func FuzzInterpToArtChainProv(f *testing.F) { ... }

// FuzzRoundTripChain builds art.State chains, converts art→interp→art,
// and asserts Prov structure is preserved at every step.
func FuzzRoundTripChain(f *testing.F) { ... }
```

---

## Verification

1. `cd ~/goivy && go build ./...` — compilation (rename may break external callers)
2. `cd ~/goivy && make test` — full test suite
3. `go test ./art/ -run TestInterpToArtState -v`
4. `go test ./art/ -run TestDecomposeState -v`
5. `go test ./art/ -fuzz FuzzInterpToArtChainProv -fuzztime 30s`
