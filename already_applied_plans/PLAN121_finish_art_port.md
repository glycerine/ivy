# Plan: Complete the Port of ivy_art.py to Go art/

**Created:** 2026-03-28 (current session)

## Context

Python's `ivy_art.py` was partially ported to `~/goivy/art/`, but the critical execution path is broken: Go's `art.PostState()` computes forward images directly via `transrel.ForwardImage` — it **never calls `interp.ConcretePost`** and therefore **never checks preconditions**. The Go `interp.ConcretePost` function already exists and is fully implemented (with precondition checking via `compose_state_action` logic). The fix is to rewire `art.PostState` to delegate to `interp.ConcretePost`, and to fix `JoinStates` to delegate to `interp.ConcreteJoin`. Additionally, two Python methods and one module-level parameter were never ported.

---

## Gap Analysis: Python ivy_art.py vs Go art/

### Critical Bugs (Semantic Divergence)

| # | Issue | Python | Go (current) |
|---|-------|--------|-------------|
| 1 | **PostState skips precondition checking** | Calls `concrete_post(op.update(...), pre_state)` which calls `compose_state_action(..., check=context.check)` — raises `ActionFailed` if precondition SAT | Calls `transrel.ForwardImage` directly, never checks preconditions, never calls `interp.ConcretePost` |
| 2 | **PostState skips moded-symbol handling** | `compose_state_action` unpacks `(su,sc,sp)` triples and renames symbols already modified in state before computing forward image | Go ignores `Modified` field entirely — uses raw formula |
| 3 | **JoinStates uses simplified disjunction** | Calls `concrete_join(state1, state2)` which uses `transrel.JoinState` with differential frame conditions | Go uses `clauseops.OrClausesTyped` — a simple OR without frame condition handling |
| 4 | **RecalculateState same ForwardImage issue** | Calls `concrete_post(state.update, state.pred)` | Reimplements ForwardImage inline, same bugs as PostState |
| 5 | **Execute/ExecuteAction take checkPrecond but don't use it** | `checkPrecond` flows through to `concrete_post` → `compose_state_action(check=...)` | `checkPrecond` parameter is accepted but never forwarded to anything |

### Missing Functions

| # | Python Function | Line | Status in Go |
|---|----------------|------|-------------|
| 6 | `AnalysisGraph.call(name, op, prestate)` | 278-284 | **Not ported** — calls `call_action`, adds result to graph, creates transition |
| 7 | `AnalysisGraph.construct_transitions_from_expressions()` | 385-390 | **Not ported** — iterates states, creates transitions from ActionApp expressions |
| 8 | `option_abs_init = BooleanParameter("abs_init", False)` | 64 | **Not ported** — module-level mutable parameter (must go on Config per CLAUDE.md rules) |

### Extra Go-only Functions (Not in Python — Review)

| # | Go Function | Status |
|---|------------|--------|
| 9 | `CheckConstraints()` | Stub — not in Python. Keep as placeholder or remove. |
| 10 | `StratifyGoals()` | Stub — not in Python. Keep as placeholder or remove. |

---

## Implementation Plan

### Step 1: Fix PostState to call interp.ConcretePost

**File:** `~/goivy/art/art.go` — method `PostState` (line 407-446)

**Current code** (broken):
```go
func (ag *AnalysisGraph) PostState(op actions.Action, preState *State, abstractor Abstractor) *State {
    var update *transrel.Update
    if preState.Domain != nil {
        update = actions.GetUpdateForArt(op, preState.Domain, preState.InScope)
    }
    // ... directly calls transrel.ForwardImage, no precondition check ...
}
```

**New code** — delegate to `interp.ConcretePost`:
```go
func (ag *AnalysisGraph) PostState(checkPrecond bool, op actions.Action, preState *State, abstractor Abstractor) *State {
    // Step 1: Compute update (transition relation) for this action.
    // Python: op.update(pre_state.domain, pre_state.in_scope)
    interpPre := ArtToInterpState(preState)

    // Step 2: Use interp.ApplyAction which calls ConcretePost with precondition checking.
    // This matches Python: concrete_post(op.update(pre_state.domain, pre_state.in_scope), pre_state)
    // via the apply_action path.
    interpPost, err := interp.ApplyAction(checkPrecond, nil, op.Name(), op, interpPre)
    if err != nil {
        // If precondition failed (ActionFailed), log and return nil or propagate.
        // Python raises IvyActionFailedError which callers catch.
        log.Printf("art.PostState: %v", err)
        return nil
    }

    s := InterpToArtState(interpPost)
    s.Action = op
    if abstractor != nil {
        abstractor.Abstract(s)
    }
    return s
}
```

**Key change:** `PostState` now takes `checkPrecond bool` parameter and delegates to `interp.ApplyAction` → `interp.ConcretePost` → `compose_state_action` equivalent, which:
- Checks preconditions when `checkPrecond=true`
- Handles moded symbols properly (renaming already-modified symbols)
- Computes `transrel.ForwardImage` correctly
- Returns proper `(su, img, sp)` state-value triple

**Alternative considered:** Calling `interp.ConcretePost` directly with a hand-computed update. Rejected because `interp.ApplyAction` already does the update computation + ConcretePost + error wrapping, matching the Python `apply_action` → `concrete_post` path exactly.

### Step 2: Fix JoinStates to call interp.ConcreteJoin

**File:** `~/goivy/art/art.go` — method `JoinStates` (line 449-466)

**Current code** (simplified):
```go
func (ag *AnalysisGraph) JoinStates(state1, state2 *State, abstractor Abstractor) *State {
    joinedClauses = clauseops.OrClausesTyped(state1.Clauses, state2.Clauses)
    // ...
}
```

**New code** — delegate to `interp.ConcreteJoin`:
```go
func (ag *AnalysisGraph) JoinStates(state1, state2 *State, abstractor Abstractor) *State {
    interp1 := ArtToInterpState(state1)
    interp2 := ArtToInterpState(state2)

    interpJoined, err := interp.ConcreteJoin(interp1, interp2)
    if err != nil {
        log.Printf("art.JoinStates: %v", err)
        // Fallback: simple OR (defensive, matches old behavior)
        var joinedClauses *clauseops.Clauses
        if state1.Clauses != nil && state2.Clauses != nil {
            joinedClauses = clauseops.OrClausesTyped(state1.Clauses, state2.Clauses)
        } else if state1.Clauses != nil {
            joinedClauses = state1.Clauses
        } else {
            joinedClauses = state2.Clauses
        }
        joined := NewState(state1.Domain, joinedClauses)
        joined.JoinOf = []*State{state1, state2}
        joined.Label = state1.Label
        if abstractor != nil {
            abstractor.Abstract(joined)
        }
        return joined
    }

    s := InterpToArtState(interpJoined)
    s.JoinOf = []*State{state1, state2}
    s.Label = state1.Label
    if abstractor != nil {
        abstractor.Abstract(s)
    }
    return s
}
```

### Step 3: Update all callers of PostState to pass checkPrecond

**File:** `~/goivy/art/art.go`

Methods that call `PostState` and need updating:

1. **`Execute`** (line 371) — already has `checkPrecond bool` param, just forward it:
   ```go
   poststate := ag.PostState(checkPrecond, op, prestate, abstractor)
   ```

2. **`Recalculate`** (line 608) — needs `checkPrecond` parameter added:
   ```go
   func (ag *AnalysisGraph) Recalculate(checkPrecond bool, t Transition, abstractor Abstractor) *State {
       // ...
       ps = ag.PostState(checkPrecond, act, t.Pre, abstractor)
   }
   ```

3. **`RecalculateState`** (line 1102) — rewrite to use `interp.ConcretePost` instead of inline ForwardImage:
   ```go
   func (ag *AnalysisGraph) RecalculateState(checkPrecond bool, state *State, abstractor Abstractor) {
       if state.Pred != nil && state.Update != nil {
           interpPre := ArtToInterpState(state.Pred)
           interpPost, err := interp.ConcretePost(checkPrecond, state.Update, interpPre, nil)
           if err != nil {
               log.Printf("art.RecalculateState: %v", err)
               return
           }
           ps := InterpToArtState(interpPost)
           if abstractor != nil {
               abstractor.Abstract(ps)
           }
           ag.ReplaceState(state, ps)
       } else if state.JoinOf != nil && len(state.JoinOf) >= 2 {
           ps := ag.JoinStates(state.JoinOf[0], state.JoinOf[1], abstractor)
           ag.ReplaceState(state, ps)
       }
   }
   ```

### Step 4: Handle PostState error propagation in Execute

Python's `execute` doesn't catch `ActionFailed` from `post_state` — it lets it propagate. But Python's `check_safety` catches `IvyActionFailedError`. In Go, since `PostState` returns `*State` (not error), we need to decide:

**Option A (recommended):** Change `PostState` to return `(*State, error)` so callers can handle precondition failures properly.

```go
func (ag *AnalysisGraph) PostState(checkPrecond bool, op actions.Action, preState *State, abstractor Abstractor) (*State, error) {
    // ... as above, but return the error ...
}

func (ag *AnalysisGraph) Execute(checkPrecond bool, op actions.Action, prestate *State, abstractor Abstractor, label string) (*State, error) {
    // ...
    poststate, err := ag.PostState(checkPrecond, op, prestate, abstractor)
    if err != nil {
        return nil, err
    }
    // ...
}
```

This requires updating all callers of `Execute`, `ExecuteAction`, `PostState`, `Recalculate`, and `RecalculateState` to handle errors. Affected callers exist in:
- `art/art.go` (internal calls)
- `art/art_test.go` (tests)
- Any external callers in webui, check, etc. — search with `Grep` for `\.Execute\(|\.PostState\(|\.Recalculate\(`

### Step 5: Port missing method — Call

**File:** `~/goivy/art/art.go`

Python (line 278-284):
```python
def call(self, name, op, prestate=None):
    if prestate == None:
        prestate = self.states[len(self.states)-1]
    op, poststate = self.call_action(name, prestate)
    self.add(poststate)
    self.transitions.append((prestate, op, name, poststate))
    return poststate
```

Go port:
```go
func (ag *AnalysisGraph) Call(name string, op func(*AnalysisGraph), prestate *State) *State {
    if prestate == nil {
        prestate = ag.LastState()
    }
    if prestate == nil {
        return nil
    }
    poststate := ag.CallAction(name, op, prestate)
    if poststate == nil {
        return nil
    }
    ag.Add(poststate, nil)
    ag.Transitions = append(ag.Transitions, Transition{
        Pre:   prestate,
        Op:    nil, // op is a func, not an Action
        Label: name,
        Post:  poststate,
    })
    return poststate
}
```

### Step 6: Port missing method — ConstructTransitionsFromExpressions

**File:** `~/goivy/art/art.go`

Python (line 385-390):
```python
def construct_transitions_from_expressions(self):
    for state in self.states:
        if hasattr(state,'expr') and is_action_app(state.expr):
            expr = state.expr
            prestate,op,label,poststate = expr.args[0],expr.rep,label_from_action(expr.rep),state
            self.transitions.append((prestate,op,label,poststate))
```

Go port:
```go
func (ag *AnalysisGraph) ConstructTransitionsFromExpressions() {
    for _, state := range ag.States {
        if state.Expr == nil {
            continue
        }
        aa, ok := state.Expr.(*ActionApp)
        if !ok {
            continue
        }
        if len(aa.Args) == 0 {
            continue
        }
        prestate := aa.Args[0]
        var action actions.Action
        var label string
        switch rep := aa.Rep.(type) {
        case actions.Action:
            action = rep
            label = LabelFromAction(rep)
        case string:
            label = rep
            if a, ok := ag.Actions[rep]; ok {
                if act, ok2 := a.(actions.Action); ok2 {
                    action = act
                }
            }
        }
        ag.Transitions = append(ag.Transitions, Transition{
            Pre:   prestate,
            Op:    action,
            Label: label,
            Post:  state,
        })
    }
}
```

### Step 7: Port option_abs_init as Config field

**File:** `~/goivy/art/art.go` (new or existing config)

Per CLAUDE.md rules, Python globals must become Config fields. Since `option_abs_init` is defined but **never used** in `ivy_art.py` (only defined at line 64, never referenced), we add it as a field on an `ArtConfig` struct but don't need to thread it yet:

```go
// ArtConfig holds per-session mutable state for the art package.
// Python module-level globals live here (per CLAUDE.md rule C).
type ArtConfig struct {
    // AbsInit controls whether to abstract the initial state.
    // Python: option_abs_init = iu.BooleanParameter("abs_init", False)
    AbsInit bool
}

func NewArtConfig() *ArtConfig {
    return &ArtConfig{AbsInit: false}
}
```

Add `ArtCfg *ArtConfig` field to `module.Config` if it doesn't already exist.

### Step 8: Remove Go-only stubs not in Python

Remove `CheckConstraints()` and `StratifyGoals()` from `art.go` — they are not in Python's `ivy_art.py` and violate rule B.7 ("Do not add abstractions, interfaces, or helper types that don't exist in Python").

### Step 9: Fix DecomposeState to match Python more faithfully

Current Go `DecomposeState` (line 952-1019) builds sub-states with intermediate clauses, but Python's `decompose_state` (line 392-401) delegates to `interp.decompose_action_app` which does the actual decomposition with BMC satisfaction checking. The Go should call `interp.DecomposeActionApp` which already exists at `interp/phase4.go:172`.

```go
func (ag *AnalysisGraph) DecomposeState(state *State) *AnalysisGraph {
    if state == nil || state.Expr == nil {
        return nil
    }
    // Check cached subgraph
    if aa, ok := state.Expr.(*ActionApp); ok && aa.Subgraph != nil {
        return aa.Subgraph
    }

    // Python: other_art = AnalysisGraph(self.domain)
    //         with AC(other_art):
    //             res = decompose_action_app(state, state.expr)
    interpState := ArtToInterpState(state)
    // Build the ast.Node expression from the ActionApp
    var exprNode ast.Node
    if aa, ok := state.Expr.(*ActionApp); ok {
        var actionName string
        switch rep := aa.Rep.(type) {
        case string:
            actionName = rep
        case actions.Action:
            actionName = rep.Name()
        }
        if actionName != "" && len(aa.Args) > 0 {
            interpPre := ArtToInterpState(aa.Args[0])
            exprNode = interp.ActionApp(ag.Domain.Cfg.AstCfg, actionName, interp.WrapState(interpPre))
        }
    }
    if exprNode == nil {
        return nil
    }

    resultState, err := interp.DecomposeActionApp(true, ag.Domain.Cfg.IuCfg, interpState, exprNode)
    if err != nil || resultState == nil {
        return nil
    }

    // Build sub-graph from decomposed result
    otherArt := NewAnalysisGraph(ag.Domain)
    // ... populate from resultState ...
    otherArt.ConstructTransitionsFromExpressions()

    // Cache
    if aa, ok := state.Expr.(*ActionApp); ok {
        aa.Subgraph = otherArt
    }
    return otherArt
}
```

---

## Step 10: Update All External Callers

Search for all callers of the changed signatures across the codebase:

```
Grep: \.PostState\(   → update to pass checkPrecond, handle error return
Grep: \.Execute\(     → update to handle error return
Grep: \.ExecuteAction\( → update to handle error return
Grep: \.Recalculate\(  → update to pass checkPrecond, handle error return
Grep: \.RecalculateState\( → update to pass checkPrecond
```

Known callers from current art_test.go that need updating:
- `TestAnalysisGraphPostState` — add checkPrecond arg, handle error
- `TestAnalysisGraphPostStateWithAbstractor` — same
- `TestAnalysisGraphExecuteAction` — handle error return
- `TestAnalysisGraphRecalculate` — add checkPrecond arg, handle error
- `TestAnalysisGraphCheckSafety` — handle error return from inner calls

---

## Step 11: Comprehensive Unit Tests

### 11a: Port-Completeness Tests (verify every Python function has a Go equivalent)

**File:** `~/goivy/art/port_completeness_test.go`

```go
// TestPortCompleteness_AllPythonMethodsExist verifies that every method
// from Python ivy_art.py has a corresponding Go method on AnalysisGraph.
func TestPortCompleteness_AllPythonMethodsExist(t *testing.T) {
    // Use reflect to check all expected methods exist on *AnalysisGraph
    expectedMethods := []string{
        "Context", "AddInitialState", "Initialize",
        "StateActions", "DoStateAction", "Recalculate",
        "PostState", "ReplaceState", "RecalculateState",
        "Execute", "ExecuteAction", "JoinStates", "Join",
        "Cover", "IsCovered", "Unreachable", "ShowCore",
        "Add", "CallAction", "Call",
        "Delete", "RemoveMarkedStates",
        "ConceptGraph", "TransitionTo", "GetHistory",
        "BMC", "CheckBoundedSafety", "CopyPath",
        "CheckSafety", "ConstructTransitionsFromExpressions",
        "DecomposeState", "MakeConcreteTrace", "DecomposeEdge",
        "UncoveredStates", "FixedpointCandidate",
        "StateExtensions", "AsCyElements",
    }
    // ... reflect check ...
}
```

Also check standalone functions: `LabelFromAction`, `NewAC`, `NewAnalysisSubgraph`, etc.

### 11b: Precondition Checking Tests

**File:** `~/goivy/art/precond_test.go`

```go
// TestPostStateChecksPrecondition verifies that PostState with checkPrecond=true
// returns an error when the action's precondition is violated.
func TestPostStateChecksPrecondition(t *testing.T) { ... }

// TestPostStateSkipsPreconditionWhenFalse verifies checkPrecond=false skips checking.
func TestPostStateSkipsPreconditionWhenFalse(t *testing.T) { ... }

// TestExecuteActionPropagatesPrecondError verifies ExecuteAction returns error
// when precondition fails.
func TestExecuteActionPropagatesPrecondError(t *testing.T) { ... }

// TestCheckSafetyCatchesPrecondFailure verifies CheckSafety catches
// IvyActionFailedError and returns a Counterexample.
func TestCheckSafetyCatchesPrecondFailure(t *testing.T) { ... }

// TestAddInitialStateCheckFalse verifies AddInitialState uses check=false
// (matching Python's EvalContext(check=False)).
func TestAddInitialStateCheckFalse(t *testing.T) { ... }
```

### 11c: JoinStates Tests

```go
// TestJoinStatesUsesConcreteJoin verifies JoinStates delegates to interp.ConcreteJoin
// with proper frame conditions (not just simple OR).
func TestJoinStatesUsesConcreteJoin(t *testing.T) { ... }

// TestJoinStatesPreservesLabel verifies the joined state gets state1's label.
func TestJoinStatesPreservesLabel(t *testing.T) { ... }
```

### 11d: New Method Tests

```go
// TestCall verifies the Call method adds state and transition to graph.
func TestCall(t *testing.T) { ... }

// TestConstructTransitionsFromExpressions verifies transitions are built
// from existing state expressions.
func TestConstructTransitionsFromExpressions(t *testing.T) { ... }
```

### 11e: RecalculateState Tests

```go
// TestRecalculateStateUsesConcreatePost verifies RecalculateState
// calls interp.ConcretePost (not inline ForwardImage).
func TestRecalculateStateUsesConcretePost(t *testing.T) { ... }

// TestRecalculateStateWithPrecondCheck verifies precondition checking
// is threaded through RecalculateState.
func TestRecalculateStateWithPrecondCheck(t *testing.T) { ... }
```

### 11f: ArtConfig Tests

```go
// TestArtConfigDefaults verifies default config values.
func TestArtConfigDefaults(t *testing.T) {
    cfg := NewArtConfig()
    if cfg.AbsInit != false {
        t.Error("AbsInit should default to false")
    }
}
```

### 11g: Fuzz Tests

```go
// FuzzPostStateCheckPrecond fuzz-tests PostState with random action/state combinations.
func FuzzPostStateCheckPrecond(f *testing.F) { ... }

// FuzzJoinStatesSymmetry verifies join is order-independent for clauses content.
func FuzzJoinStatesSymmetry(f *testing.F) { ... }

// FuzzConstructTransitions verifies round-trip: build expressions → construct transitions.
func FuzzConstructTransitions(f *testing.F) { ... }

// FuzzExecuteAndCheckSafety fuzz-tests the execute→check_safety pipeline.
func FuzzExecuteAndCheckSafety(f *testing.F) { ... }
```

---

## Files Modified

| File | Changes |
|------|---------|
| `~/goivy/art/art.go` | Fix `PostState`, `JoinStates`, `RecalculateState`, `Execute`, `Recalculate`; add `Call`, `ConstructTransitionsFromExpressions`, `ArtConfig`; fix `DecomposeState`; remove stubs; add `checkPrecond` params |
| `~/goivy/art/art_test.go` | Update all tests for new signatures (checkPrecond, error returns) |
| `~/goivy/art/precond_test.go` | **New** — precondition checking tests |
| `~/goivy/art/port_completeness_test.go` | **New** — verify every Python function is ported |
| `~/goivy/module/config.go` | Add `ArtCfg *ArtConfig` field to `module.Config` |

## Existing Facilities to Reuse (Do NOT Duplicate)

| Function | Location | Purpose |
|----------|----------|---------|
| `interp.ConcretePost` | `interp/eval.go:27` | Full concrete post with precondition check |
| `interp.ConcreteJoin` | `interp/eval.go:79` | Full concrete join with frame conditions |
| `interp.ApplyAction` | `interp/eval.go:147` | Action update + ConcretePost + error wrapping |
| `interp.DecomposeActionApp` | `interp/phase4.go:172` | Full action decomposition with BMC |
| `ArtToInterpState` | `art/art.go:1486` | Adapter: art.State → interp.State |
| `InterpToArtState` | `art/art.go:1504` | Adapter: interp.State → art.State |
| `actions.GetUpdateForArt` | `actions/update.go:2235` | Compute transition relation from action |
| `transrel.ForwardImage` | `transrel/` | Forward image computation (called by ConcretePost) |
| `interp.FailAction` | `interp/helpers.go:59` | Fail action wrapper (for fail_expr) |
| `interp.FailExpr` | `interp/helpers.go:661` | Build fail expression |

## Verification

1. **`go build ./art/...`** — confirms compilation
2. **`go test ./art/ -v`** — runs all unit tests including new ones
3. **`go test ./art/ -fuzz FuzzPostStateCheckPrecond -fuzztime 30s`** — fuzz testing
4. **`go test ./art/ -fuzz FuzzJoinStatesSymmetry -fuzztime 30s`** — fuzz testing
5. **`go test ./interp/ -v`** — ensure interp tests still pass (no regressions)
6. **`go test ./... 2>&1 | tail -20`** — full project test suite
7. **Manual:** Create an AnalysisGraph with a precondition-bearing action, call Execute with checkPrecond=true, verify error is returned
