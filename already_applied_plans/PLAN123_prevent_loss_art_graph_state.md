# Plan: Fix DecomposeState Expr Chain and Add Comprehensive Tests

**Created:** 2026-03-28 18:45

## Context

`art.DecomposeState` calls `interp.DecomposeActionApp`, which returns an `interp.State` chain with properly set `Provenance` fields (type `ast.Node`). But when converting back via `InterpToArtState`, the `Provenance` field is **silently dropped** because:

1. `interp.State.Expr` is `ast.Node` (containing wrapped `*interp.State` pointers via `stateNode`)
2. `art.State.Expr` is `art.Provenance` (marker interface implemented by `*ActionApp` and `*StateJoin`, containing `*art.State` pointers)
3. Neither `InterpToArtState` nor `ArtToInterpState` converts the expression — both silently lose it

**Result:** The decomposition subgraph has states but no transitions, because `ConstructTransitionsFromExpressions` finds `Provenance == nil` on every state.

---

## Root Cause: Type Split

Python has one `State` class with one `expr` field. Go has two:

| | `art.State` | `interp.State` |
|---|---|---|
| Expr type | `art.Provenance` (ActionApp/StateJoin wrapping `*art.State`) | `ast.Node` (ast.Atom wrapping `*interp.State` via stateNode) |
| Pred access | `s.Pred` (explicit field) | `s.Pred()` (lazy, unwraps from `s.Expr`) |

The fix must bridge this gap during conversion.

---

## Pre-Step: Rename art.Expr → art.Provenance

The name `Expr` is massively overloaded across the codebase (`lg.Expr`, `ast.Node`, `interp.State.Expr`). Rename `art.Expr` to `art.Provenance` to make its purpose clear: it records *how a state was derived* (its provenance), not a logical expression.

**Scope of rename in `art/` package:**
- `type Expr interface` → `type Provenance interface`
- `art.State.Expr` field → `art.State.Provenance`
- All references: `IsActionApp(e Expr)` → `IsActionApp(e Provenance)`, etc.
- `NewActionApp(...) *ActionApp` return used as `Provenance`
- Test files updated

This is a package-internal rename (the `Expr` type was not exported as a standalone name consumers depend on — callers use `*ActionApp` and `*StateJoin` concretely).

## Approach: Convert Provenance in InterpToArtState

**Do NOT make `interp.State` implement `lg.Expr`.** The `interp.State` struct holds mutable session state (InScope, CachedPred, Universe, etc.) — it is not a formula node. Forcing it into the `lg.Expr` interface would violate the type's semantics and create confusion about ownership and identity. The Python State class doesn't implement any expression interface either — it's stored *in* expressions, not *as* one.

Instead, convert the `ast.Node` expression tree into the `art.Provenance` tree during `InterpToArtState`, translating `stateNode` pointers from interp-space to art-space along the way.

### Step 1: Add Provenance conversion to InterpToArtState

**File:** `~/goivy/art/art.go` — function `InterpToArtState` (line ~1565)

The current code:
```go
func InterpToArtState(is *interp.State) *State {
    s := NewState(is.Domain, is.Clauses)
    s.Label = is.Label
    s.InScope = is.InScope
    s.Action = is.Action
    if is.Update() != nil { s.Update = is.Update() }
    if is.Pred() != nil { s.Pred = InterpToArtState(is.Pred()) }
    return s
}
```

The problem: recursive `InterpToArtState(is.Pred())` creates art.State objects for predecessors but never links them via Provenance. We need a memo-ized conversion that:

1. Converts the full chain bottom-up (root first)
2. Creates `art.ActionApp` expressions pointing at the already-converted predecessor art.States
3. Creates `art.StateJoin` expressions for join nodes

**New approach** — add a helper that takes a memo map:

```go
// interpToArtStateMemo converts an interp.State chain to art.State chain,
// preserving Expr fields by converting ast.Node expressions to art.Provenance.
// The memo map prevents duplicate conversions and ensures pointer identity.
func interpToArtStateMemo(is *interp.State, memo map[*interp.State]*State) *State {
    if is == nil {
        return nil
    }
    if existing, ok := memo[is]; ok {
        return existing
    }

    s := NewState(is.Domain, is.Clauses)
    memo[is] = s // memo BEFORE recursing to handle cycles
    s.Label = is.Label
    s.InScope = is.InScope
    s.Action = is.Action
    s.ActionName = is.ActionName
    if is.Update() != nil {
        s.Update = is.Update()
    }

    // Convert predecessor
    if is.Pred() != nil {
        s.Pred = interpToArtStateMemo(is.Pred(), memo)
    }

    // Convert JoinOf
    if is.JoinOf != nil {
        for _, jo := range is.JoinOf {
            s.JoinOf = append(s.JoinOf, interpToArtStateMemo(jo, memo))
        }
    }

    // Convert Expr: translate ast.Node → art.Provenance
    if is.Expr != nil {
        s.Provenance = interpExprToArtProvenance(is.Expr, memo)
    }

    return s
}
```

### Step 2: Add interpExprToArtProvenance helper

```go
// interpExprToArtProvenance converts an interp-style ast.Node expression tree
// into an art.Provenance, translating wrapped interp.State pointers to art.State.
//
// interp uses:
//   - ast.Atom with 1 arg (stateNode) → ActionApp
//   - ast.Or with stateNode args → StateJoin
func interpExprToArtProvenance(expr ast.Node, memo map[*interp.State]*State) Provenance {
    if expr == nil {
        return nil
    }
    // Check for ActionApp: interp.IsActionApp checks isinstance(expr, Atom) && len(args)==1
    if interp.IsActionApp(expr) {
        atom := expr.(*ast.Atom)
        // Extract the action name/rep
        var rep interface{} = atom.Rep
        // Extract the predecessor state from the stateNode arg
        var args []*State
        for _, term := range atom.Terms {
            if is := interp.UnwrapState(term); is != nil {
                args = append(args, interpToArtStateMemo(is, memo))
            }
        }
        return &ActionApp{Rep: rep, Args: args}
    }
    // Check for StateJoin: interp.IsStateJoin checks isinstance(expr, Or)
    if interp.IsStateJoin(expr) {
        or := expr.(*ast.Or)
        var args []*State
        for _, term := range or.Terms {
            if is := interp.UnwrapState(term); is != nil {
                args = append(args, interpToArtStateMemo(is, memo))
            }
        }
        return &StateJoin{Args: args}
    }
    return nil
}
```

### Step 3: Update the public InterpToArtState to use memo

```go
func InterpToArtState(is *interp.State) *State {
    return interpToArtStateMemo(is, make(map[*interp.State]*State))
}
```

This preserves the public API while adding memoization and Provenance conversion.

### Step 4: Add ActionName field to art.State

Currently `art.State` has `Action actions.Action` but no `ActionName string`. The `interp.State` does have `ActionName`. We need it for the Provenance conversion (the `Rep` field on `ActionApp` often carries the action name as a string). Add:

```go
type State struct {
    // ... existing fields ...
    ActionName string // name of the action (matches interp.State.ActionName)
}
```

### Step 5: Symmetry — update ArtToInterpState to convert Expr

For completeness and future correctness, also convert `art.Provenance` → `ast.Node` in `ArtToInterpState`:

```go
func artProvenanceToInterpExpr(prov Provenance, cfg *ast.AstConfig, memo map[*State]*interp.State) ast.Node {
    if prov == nil {
        return nil
    }
    switch e := prov.(type) {
    case *ActionApp:
        // Convert to interp.ActionApp (ast.Atom with WrapState args)
        var args []ast.Node
        for _, s := range e.Args {
            args = append(args, interp.WrapState(artToInterpStateMemo(s, cfg, memo)))
        }
        return interp.ActionApp(cfg, fmt.Sprintf("%v", e.Rep), args...)
    case *StateJoin:
        var args []lg.Expr
        for _, s := range e.Args {
            args = append(args, interp.WrapState(artToInterpStateMemo(s, cfg, memo)))
        }
        return &ast.Or{Terms: args}
    }
    return nil
}
```

**Note:** Check if `interp.ActionApp` accepts variadic `ast.Node` or requires a single arg. Current signature at `interp/interp.go:346`:
```go
func ActionApp(cfg *ast.AstConfig, actionName string, arg ast.Node) *ast.Atom
```
It takes a single `arg ast.Node`. For the art→interp direction, we only need the first arg (the predecessor), which matches the Python `action_app(action, state)` pattern.

### Step 6: Fix DecomposeState to not double-convert

After the InterpToArtState fix, DecomposeState's current code (lines 1029-1041) already does the right thing — it walks the chain via `Pred()`, reverses, and adds with `artSt.Provenance`. Since `InterpToArtState` now preserves Expr, the `ConstructTransitionsFromExpressions()` call will find the transitions.

However, the current code re-creates the predecessor chain twice (once via InterpToArtState's recursive Pred conversion, once by walking the chain manually). We should simplify: just convert the deepest state via `InterpToArtState` (which recursively converts the whole chain with memo), then collect the chain from the art side.

---

## Files Modified

| File | Changes |
|------|---------|
| `~/goivy/art/art.go` | Add `ActionName` to State; rewrite `InterpToArtState`/`ArtToInterpState` with memo+Provenance conversion; simplify `DecomposeState` |
| `~/goivy/art/art_test.go` | Update existing tests if needed |
| `~/goivy/art/decompose_test.go` | **New** — comprehensive tests for DecomposeState Provenance chain |

## Existing Facilities to Reuse

| Function | Location | Purpose |
|----------|----------|---------|
| `interp.IsActionApp` | `interp/interp.go:334` | Check if ast.Node is an action application |
| `interp.IsStateJoin` | `interp/interp.go:340` | Check if ast.Node is a state join |
| `interp.UnwrapState` | `interp/interp.go:237` | Extract *interp.State from stateNode |
| `interp.WrapState` | `interp/interp.go:233` | Wrap *interp.State as ast.Node |
| `interp.ActionApp` | `interp/interp.go:346` | Create ast.Atom action application |
| `interp.DecomposeActionApp` | `interp/phase4.go:172` | The decomposition function |

---

## Tests

### Unit Tests (`~/goivy/art/decompose_test.go`)

```go
// TestInterpToArtStatePreservesExpr verifies that InterpToArtState
// converts interp.State.Expr (ast.Node) into art.State.Expr (art.Provenance).
func TestInterpToArtStatePreservesExpr(t *testing.T) { ... }

// TestInterpToArtStateActionApp verifies that an interp ActionApp expression
// is converted to an art.ActionApp with correct Rep and Args.
func TestInterpToArtStateActionApp(t *testing.T) { ... }

// TestInterpToArtStateJoinExpr verifies that an interp Or expression
// is converted to an art.StateJoin.
func TestInterpToArtStateJoinExpr(t *testing.T) { ... }

// TestInterpToArtStateChainPreservesAllExprs verifies that converting
// a multi-step interp.State chain preserves Expr on every non-root state.
func TestInterpToArtStateChainPreservesAllExprs(t *testing.T) { ... }

// TestInterpToArtStateMemoPreservesIdentity verifies that the same
// interp.State pointer converts to the same art.State pointer.
func TestInterpToArtStateMemoPreservesIdentity(t *testing.T) { ... }

// TestDecomposeStateProducesTransitions verifies that DecomposeState
// returns a subgraph with actual transitions (not just states).
func TestDecomposeStateProducesTransitions(t *testing.T) { ... }

// TestDecomposeStateCachesSubgraph verifies that the second call
// returns the cached subgraph.
func TestDecomposeStateCachesSubgraph(t *testing.T) { ... }

// TestDecomposeStateNilExprReturnsNil verifies nil/empty cases.
func TestDecomposeStateNilExprReturnsNil(t *testing.T) { ... }

// TestConstructTransitionsAfterDecompose verifies that
// ConstructTransitionsFromExpressions finds the transitions
// that DecomposeState's Provenance chain provides.
func TestConstructTransitionsAfterDecompose(t *testing.T) { ... }

// TestArtToInterpStatePreservesExpr verifies the reverse direction.
func TestArtToInterpStatePreservesExpr(t *testing.T) { ... }

// TestRoundTripArtInterpArtPreservesExprSemantics verifies that
// converting art→interp→art preserves the Expr structure (ActionApp with
// correct rep and arg count).
func TestRoundTripArtInterpArtPreservesExprSemantics(t *testing.T) { ... }
```

### Fuzz Tests

```go
// FuzzInterpToArtStateChain fuzz-tests conversion of interp.State chains
// of varying lengths, asserting every non-root state has a non-nil Expr.
func FuzzInterpToArtStateChain(f *testing.F) { ... }

// FuzzDecomposeRoundTrip fuzz-tests that decompose + construct_transitions
// produces the same number of transitions as there are non-root states.
func FuzzDecomposeRoundTrip(f *testing.F) { ... }
```

## Verification

1. `cd ~/goivy && go build ./art/...` — compilation
2. `cd ~/goivy && make test` — full test suite
3. Specifically: `go test ./art/ -run TestInterpToArtState -v` — new tests
4. Specifically: `go test ./art/ -run TestDecomposeState -v` — decompose tests
5. `go test ./art/ -fuzz FuzzInterpToArtStateChain -fuzztime 30s`
