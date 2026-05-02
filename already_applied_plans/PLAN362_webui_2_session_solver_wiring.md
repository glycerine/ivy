# Plan: WebUI Session Solver Wiring (8 Operations)

**Created:** 2026-05-02 09:42 UTC

## Context

The file `webui/session.go` has 8 stub case branches in `ExecuteAction()` (lines 287-378) that emit "not yet wired" status messages instead of performing verification. All underlying Go infrastructure is already implemented in `art/`, `actions/`, `interp/`, `z3bridge/`, `updr/`, and `trace/`. The work is connecting the stubs to these packages.

Python source of truth: `ivy_ui.py`, `ivy_ui_cti.py`, `ivy_graph_ui.py`, `ivy_interp.py`, `ivy_transrel.py`, `ivy_solver.py`.

---

## Step 0: Add Persistent AnalysisGraph to Session

The Session struct (`session.go:27-42`) has `Graph *AnalysisGraphState` (lightweight JSON-serializable) but no real `art.AnalysisGraph`. Interactive operations need persistent state with `Clauses`, `Update`, `Pred` fields.

### Changes to `webui/session.go`:

**A) Add field** at line 42 (before closing brace of Session struct):
```go
AG *art.AnalysisGraph // persistent analysis graph for interactive verification
```

**B) Add imports** (new):
```go
"github.com/glycerine/ivy/goivy/actions"
"github.com/glycerine/ivy/goivy/art"
"github.com/glycerine/ivy/goivy/interp"
"github.com/glycerine/ivy/goivy/logicparser"
"github.com/glycerine/ivy/goivy/z3bridge"
```

**C) Initialize AG** in `LoadFileContent()` after line 188 (`s.CompiledModule = mod`):
```go
s.AG = art.NewAnalysisGraph(s.CompiledModule)
s.AG.AddInitialState(nil, nil) // computes init from module's init condition
s.syncARGToGraph()
```

**D) Add `syncARGToGraph()` helper** — converts `s.AG` states/transitions into `s.Graph` for frontend:
```go
func (s *Session) syncARGToGraph() {
    if s.AG == nil { return }
    gs := NewAnalysisGraphState()
    for _, st := range s.AG.States {
        gs.States = append(gs.States, ARGNode{
            ID: st.ID, Label: st.Label,
            IsBottom: st.IsBottom(),
            Info: fmt.Sprintf("State %d", st.ID),
        })
    }
    for _, t := range s.AG.Transitions {
        gs.Transitions = append(gs.Transitions, ARGTransition{
            SourceID: t.Pre.ID, TargetID: t.Post.ID,
            Label: t.Label,
        })
    }
    s.Graph = gs
}
```

### Key files:
- `webui/session.go` — add field, init, helper
- `art/art.go:1508` — `AddInitialState()` already implemented
- `art/art.go:1656` — `ArtToInterpState()` already implemented
- `art/art.go:1723` — `InterpToArtState()` already implemented

---

## Step 1: PDR Step (line 288-290)

**Python**: `updr.CheckModule()` via `ivy_updr.py` — full PDR/IC3 verification.
**Go**: `updr.CheckModule(s.CompiledModule)` already works (used in `RunCheck("pdr")` at session.go:808).

Replace lines 288-290:
```go
case "pdr_step":
    if s.CompiledModule == nil {
        err = fmt.Errorf("pdr_step: no compiled module")
        break
    }
    pdrResult, pdrErr := updr.CheckModule(s.CompiledModule)
    if pdrErr != nil {
        err = pdrErr
        break
    }
    result["valid"] = pdrResult.Valid
    result["invariant"] = pdrResult.Invariant
    result["error_msg"] = pdrResult.Error
    result["stats"] = map[string]int{
        "num_frames":       pdrResult.Stats.NumFrames,
        "num_iterations":   pdrResult.Stats.NumIterations,
        "num_sat_queries":  pdrResult.Stats.NumSATQueries,
        "num_clauses":      pdrResult.Stats.NumClauses,
        "num_univ_clauses": pdrResult.Stats.NumUnivClauses,
    }
    msg := "PDR: counterexample found"
    if pdrResult.Valid {
        msg = fmt.Sprintf("PDR: invariant found (%d clauses)", pdrResult.Stats.NumClauses)
    }
    s.emit(Event{Type: "pdr_complete", Data: map[string]string{"message": msg}})
```

---

## Step 2: Concrete Model (line 291-293)

**Python**: `ivy_solver.py:1224` `get_model_clauses(clauses)` — Z3 SAT check, return model.
**Go**: `z3bridge.NewSolver(mod, nil)` + `solver.GetModelClauses(clauses)` at `z3bridge/solver_model.go`.

Replace lines 291-293:
```go
case "concrete":
    if s.CompiledModule == nil || s.AG == nil {
        err = fmt.Errorf("concrete: no compiled module")
        break
    }
    state := s.AG.LastState()
    if stateID, ok := args["state_id"].(float64); ok && int(stateID) < len(s.AG.States) {
        state = s.AG.States[int(stateID)]
    }
    if state == nil || state.Clauses == nil {
        err = fmt.Errorf("concrete: no state available")
        break
    }
    axioms := s.CompiledModule.BackgroundTheory(state.InScope)
    clauses := module.AndClausesTyped(state.Clauses, axioms)
    solver := z3bridge.NewSolver(s.CompiledModule, nil)
    defer solver.Close()
    mr, solverErr := solver.GetModelClauses(clauses)
    if solverErr != nil {
        err = solverErr
        break
    }
    if mr == nil {
        result["sat"] = false
        s.emit(Event{Type: "concrete_result", Data: map[string]string{"sat": "false"}})
    } else {
        result["sat"] = true
        valMap := make(map[string]string)
        if modelVals, mErr := solver.ModelValues(mr.Model, mr.Vocab); mErr == nil {
            for name, val := range modelVals {
                valMap[name] = val.String()
            }
        }
        result["model"] = valMap
        s.emit(Event{Type: "concrete_result", Data: map[string]interface{}{"sat": "true", "model": valMap}})
    }
```

### Key files:
- `z3bridge/solver_model.go` — `GetModelClauses()`, `ModelValues()`

---

## Step 3: Reverse Image (line 294-296)

**Python**: `ivy_transrel.py:527` `reverse_image(post, axioms, update)`.
**Go**: `actions.ReverseImage(postState, axioms, update)` at `actions/transrel.go:1426`.

Replace lines 294-296:
```go
case "reverse":
    if s.CompiledModule == nil || s.AG == nil {
        err = fmt.Errorf("reverse: no compiled module")
        break
    }
    state := s.AG.LastState()
    if stateID, ok := args["state_id"].(float64); ok && int(stateID) < len(s.AG.States) {
        state = s.AG.States[int(stateID)]
    }
    if state == nil {
        err = fmt.Errorf("reverse: no state available")
        break
    }
    if state.Update == nil {
        err = fmt.Errorf("reverse: state has no transition (execute an action first)")
        break
    }
    axioms := s.CompiledModule.BackgroundTheory(state.InScope)
    preClauses := actions.ReverseImage(state.Clauses, axioms, state.Update)
    if preClauses == nil {
        err = fmt.Errorf("reverse: reverse image returned nil")
        break
    }
    result["pre_state"] = logic.PrettyFmla(preClauses.ToFormula())
    s.emit(Event{Type: "reverse_result", Data: map[string]string{
        "pre_state": result["pre_state"].(string),
    }})
```

### Key files:
- `actions/transrel.go:1426` — `ReverseImage()`

---

## Step 4: Reach (line 297-299)

**Python**: `ivy_interp.py:284` `reach_state(state, clauses)` — forward image from predecessor, SAT check, extract model.
**Go**: `interp.ReachState(state, clauses)` at `interp/helpers.go:131` — already a faithful port.

Replace lines 297-299:
```go
case "path_reach", "reach":
    if s.CompiledModule == nil || s.AG == nil {
        err = fmt.Errorf("reach: no compiled module")
        break
    }
    state := s.AG.LastState()
    if stateID, ok := args["state_id"].(float64); ok && int(stateID) < len(s.AG.States) {
        state = s.AG.States[int(stateID)]
    }
    if state == nil {
        err = fmt.Errorf("reach: no state available")
        break
    }
    interpState := art.ArtToInterpState(state)
    reached := interp.ReachState(interpState, nil)
    if reached == nil {
        result["reachable"] = false
        s.emit(Event{Type: "reach_result", Data: map[string]string{"reachable": "false"}})
    } else {
        result["reachable"] = true
        artReached := art.InterpToArtState(reached)
        if artReached.Clauses != nil {
            result["reached_state"] = logic.PrettyFmla(artReached.Clauses.ToFormula())
        }
        state.Unders = append(state.Unders, artReached)
        s.syncARGToGraph()
        s.emit(Event{Type: "reach_result", Data: map[string]string{"reachable": "true"}})
    }
```

### Key files:
- `interp/helpers.go:131` — `ReachState()`
- `art/art.go:1656` — `ArtToInterpState()`
- `art/art.go:1723` — `InterpToArtState()`

---

## Step 5: Weaken (line 300-302)

**Python**: `ivy_ui_cti.py:235` `weaken(conjs)` — `self.conjectures.remove(conj)`. Pure structural, no solver.

Replace lines 300-302:
```go
case "weaken":
    if s.CompiledModule == nil {
        err = fmt.Errorf("weaken: no compiled module")
        break
    }
    indices, ok := args["indices"].([]interface{})
    if !ok || len(indices) == 0 {
        err = fmt.Errorf("weaken: no conjecture indices specified")
        break
    }
    toRemove := make(map[int]bool, len(indices))
    for _, idx := range indices {
        if fi, ok := idx.(float64); ok {
            toRemove[int(fi)] = true
        }
    }
    var kept []*ast.LabeledFormula
    var removed []string
    for i, lc := range s.CompiledModule.LabeledConjs {
        if toRemove[i] {
            formula := ""
            if lc.Formula != nil {
                if fExpr, ok := lc.Formula.(logic.Expr); ok {
                    formula = logic.PrettyFmla(module.DropUniversals(fExpr))
                }
            }
            removed = append(removed, formula)
        } else {
            kept = append(kept, lc)
        }
    }
    s.CompiledModule.LabeledConjs = kept
    result["removed_count"] = len(removed)
    result["removed"] = removed
    result["remaining_count"] = len(kept)
    s.emit(Event{Type: "weaken_result", Data: map[string]interface{}{"removed": removed}})
```

### Key files:
- `module/module.go:28` — `LabeledConjs []*ast.LabeledFormula`
- `module/astutil.go:297` — `DropUniversals(f)`

---

## Step 6: Save Abstraction (line 303-305)

**Python**: `ivy_ui.py:86` writes `concept <name> = <defn>` per concept space. `ivy_ui_cti.py:249` writes categorized invariants.

Replace lines 303-305:
```go
case "save_abstraction":
    if s.CompiledModule == nil {
        err = fmt.Errorf("save_abstraction: no compiled module")
        break
    }
    var sb strings.Builder
    sb.WriteString("# This file was generated by ivy.\n\n")
    for _, cs := range s.CompiledModule.ConceptSpaces {
        sb.WriteString(fmt.Sprintf("concept %v = %v\n", cs.Label, cs.Body))
    }
    if len(s.CompiledModule.LabeledConjs) > 0 {
        sb.WriteString("\n# conjectures\n\n")
        for _, lc := range s.CompiledModule.LabeledConjs {
            label := ""
            if lc.Label != nil {
                label = fmt.Sprint(lc.Label)
            }
            formula := ""
            if lc.Formula != nil {
                if fExpr, ok := lc.Formula.(logic.Expr); ok {
                    formula = logic.PrettyFmla(module.DropUniversals(fExpr))
                } else {
                    formula = fmt.Sprint(lc.Formula)
                }
            }
            if label != "" {
                sb.WriteString(fmt.Sprintf("invariant [%s] %s\n", label, formula))
            } else {
                sb.WriteString(fmt.Sprintf("invariant %s\n", formula))
            }
        }
    }
    result["content"] = sb.String()
    s.emit(Event{Type: "export", Data: map[string]string{"content": sb.String()}})
```

### Key files:
- `module/module.go:102` — `ConceptSpaces []ConceptSpace`
- `module/module.go:268` — `type ConceptSpace struct { Label, Body lg.Expr }`

---

## Step 7: Add Relation (line 334-340)

**Python**: `ivy_graph.py:540` `string_to_concept(text)` — parse formula, extract free vars, create Concept with name `V:Sort.formula`.
**Go**: `logicparser.ToFormula(text)` returns `ast.Node`. For concept domain, store as text + metadata.

Replace lines 334-340:
```go
case "add_relation":
    if formula, ok := args["formula"].(string); ok && formula != "" {
        astNode, parseErr := logicparser.ToFormula(formula)
        if parseErr != nil {
            err = fmt.Errorf("add_relation: parse error: %w", parseErr)
            break
        }
        name := formula
        var vars []string
        var sorts []string
        _ = astNode // parsed successfully; name built from text
        c := &Concept{
            Name:    name,
            Formula: formula,
            Sorts:   sorts,
        }
        if s.SimpleSess != nil {
            s.SimpleSess.Domain.Concepts[name] = c
        }
        result["concept_name"] = name
        s.emit(Event{Type: "concept_updated", Data: map[string]interface{}{
            "name": name, "formula": formula,
        }})
    } else {
        err = fmt.Errorf("add_relation: missing formula argument")
    }
```

Note: The full `concept_from_formula` flow (extracting free variables from a `logic.Expr`) requires compiling the ast.Node through the compiler's sort inference, which needs the module's signature. This can be enhanced later — for now, parsing validates syntax and the concept is stored by text name, matching the existing `GraphWidget.AddConceptFromString` pattern at `graph_widget.go:207-216`.

### Key files:
- `logicparser/logicparser.go:34` — `ToFormula(s)`
- `webui/graph_widget.go:207` — existing `AddConceptFromString()` pattern

---

## Step 8: Action Execution (line 374-379 default case)

**Python**: `ivy_art.py:205` `execute_action(name, prestate, abstractor)` → `execute()` → `post_state()` → `concrete_post()`.
**Go**: `art.AnalysisGraph.ExecuteAction(checkPrecond, name, prestate, abstractor)` at `art/art.go:395`.

Modify the default case at lines 374-379. Insert action dispatch before the "unknown action" fallback:
```go
default:
    if s.CompiledModule != nil && s.AG != nil {
        if _, ok := s.CompiledModule.Actions.Get2(actionName); ok {
            prestate := s.AG.LastState()
            if stateID, ok := args["state_id"].(float64); ok && int(stateID) < len(s.AG.States) {
                prestate = s.AG.States[int(stateID)]
            }
            if prestate != nil {
                poststate, execErr := s.AG.ExecuteAction(false, actionName, prestate, nil)
                if execErr != nil {
                    err = execErr
                    break
                }
                if poststate != nil {
                    s.syncARGToGraph()
                    result["post_state_id"] = poststate.ID
                    if poststate.Clauses != nil {
                        result["post_state"] = logic.PrettyFmla(poststate.Clauses.ToFormula())
                    }
                    s.emit(Event{Type: "action_executed", Data: map[string]interface{}{
                        "action": actionName, "post_id": poststate.ID,
                    }})
                }
                break
            }
        }
    }
    s.emit(Event{Type: "status", Data: map[string]string{
        "message": "Action '" + actionName + "' accepted (not yet wired to engine)",
    }})
```

### Key files:
- `art/art.go:395` — `ExecuteAction()`
- `art/art.go:373` — `Execute()` (called by ExecuteAction)
- `art/art.go:416` — `PostState()` → calls `interp.ApplyAction`

---

## Step 9: Comprehensive Unit Tests

Create `webui/session_solver_test.go` with build tag `//go:build web`.

Use the existing `ivySample` from `backend_conform_test.go:17-26`:
```
#lang ivy1.7
type client
type server
relation link(X:client, Y:server)
relation semaphore(X:server)
after init { semaphore(W) := true; link(X,Y) := false }
action connect(x:client,y:server) = { require semaphore(y); link(x,y) := true; semaphore(y) := false }
export connect
conjecture link(X,Y) -> ~semaphore(Y)
```

### Test helper:
```go
func loadTestSession(t *testing.T) *Session {
    t.Helper()
    cfg := module.NewConfig()
    s := NewSession(cfg, "test-solver")
    if err := s.LoadFileContent("test.ivy", []byte(ivySample)); err != nil {
        t.Fatalf("LoadFileContent: %v", err)
    }
    return s
}
```

### Tests:

1. **TestSolverPDRStep** — load sample, `ExecuteAction("pdr_step", nil)`, verify `result["valid"]` is bool, no error.

2. **TestSolverPDRStepNoModule** — `ExecuteAction("pdr_step", nil)` without loading, verify error.

3. **TestSolverConcrete** — load sample, verify `result["sat"]` is bool.

4. **TestSolverConcreteNoState** — no module loaded, verify error.

5. **TestSolverReverse** — load sample, execute "connect" first to create post-state with Update, then call "reverse", verify `result["pre_state"]` is non-empty string.

6. **TestSolverReverseNoUpdate** — load sample, call "reverse" on initial state (no Update), verify error.

7. **TestSolverReach** — load sample, execute "connect" to create post-state, call "reach", verify `result["reachable"]` is bool.

8. **TestSolverWeaken** — load sample, `ExecuteAction("weaken", {"indices": [0]})`, verify `result["removed_count"]` == 1, `result["remaining_count"]` == 0.

9. **TestSolverWeakenNoIndices** — call weaken without indices, verify error.

10. **TestSolverWeakenOutOfRange** — call weaken with index 99, verify removed_count == 0.

11. **TestSolverSaveAbstraction** — load sample, verify `result["content"]` contains "conjecture" or "invariant".

12. **TestSolverSaveAbstractionNoModule** — no module, verify error.

13. **TestSolverAddRelation** — load sample, `ExecuteAction("add_relation", {"formula": "link(X,Y)"})`, verify `result["concept_name"]` is non-empty.

14. **TestSolverAddRelationParseError** — bad formula `"((("`, verify error.

15. **TestSolverAddRelationEmpty** — empty formula, verify error.

16. **TestSolverActionExecute** — load sample, `ExecuteAction("connect", nil)`, verify `result["post_state_id"]` is set, AG has 2+ states.

17. **TestSolverActionExecuteNotFound** — call nonexistent action "nonexistent", verify it falls through to status message (no error, just logged).

18. **TestSolverActionExecuteNoModule** — no module, verify falls through gracefully.

19. **TestSolverARGInitialized** — load sample, verify `s.AG` is non-nil, has at least 1 state, `s.Graph` has matching state count.

20. **TestSolverMultiStepWorkflow** — load sample, execute "connect", then "reach" on post-state, then "concrete" on post-state, verify each step succeeds and ARG grows.

---

## Verification

Run tests via:
```
cd ~/ivy/goivy && make test
```

(Never `go test ./...` — XTRACE is on by default per CLAUDE.md rule 9.)

---

## File Summary

| File | Action |
|------|--------|
| `webui/session.go` | Modify: add AG field, imports, init, syncARGToGraph, wire 8 operations |
| `webui/session_solver_test.go` | Create: 20 unit tests |
