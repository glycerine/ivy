# Plan: Implement WebUI Concept Graph Operations (Item 5)

**Created:** 2026-05-02 07:12 UTC

## Context

The goivy WebUI has 12 stub functions for concept graph operations. These stubs have HTTP handlers and SSE plumbing built but no core computation. The Python implementations in `ivy_graph.py`, `concept_interactive_session.py`, `tactics.py`, and `ui_extensions_api.py` are the source of truth.

The key architectural insight: goivy already has a working `ConceptInteractiveSession` (concept_isession.go) with Z3-backed `materializeNode`, `materializeEdge`, `GetWitnesses`, `Suppose`, `GetFacts`, `Recompute`, etc. Most Graph-level stubs just need to delegate to it. The gap is that `Graph` only holds `*ConceptSession` (simple render layer) and lacks a reference to the full Z3 session.

## Structural Change: Wire Graph to ConceptInteractiveSession

**Add field to Graph struct** (graph_model.go:157):
```go
InteractiveSess *ConceptInteractiveSession
```

**Wire it in Session.LoadFileContent** (session.go): after creating both sessions, set `graph.InteractiveSess = s.ConceptSess` so Graph methods can delegate Z3 operations.

**Update NewGraph** (graph_model.go:192): add `InteractiveSess` param (optional, nil-safe).

---

## Implementation Order

### Phase 1: Graph-level read/write stubs (no new Z3 code needed)

#### 1A. Graph.GetFacts (graph_model.go:509-513)

**Python:** `concept_interactive_session.py:268-283` — iterates node facts + edge facts from abstract value.

**Go:** Delegate to `ConceptInteractiveSession.GetFacts(projection)` which already works (concept_isession.go:584-613). Convert `[]logic.Expr` to `[]string`.

```go
func (g *Graph) GetFacts(definite bool) []string {
    g.mu.RLock()
    defer g.mu.RUnlock()
    if g.InteractiveSess != nil {
        var proj func(string, string, string) bool
        if definite {
            proj = func(a, b, c string) bool { return true }
        }
        facts := g.InteractiveSess.GetFacts(proj)
        result := make([]string, 0, len(facts))
        for _, f := range facts {
            result = append(result, f.String())
        }
        return result
    }
    // Fallback: keys from simple session AbstractValue where value=true
    var result []string
    for k, v := range g.ConceptSess.AbstractValue {
        if v { result = append(result, k) }
    }
    return result
}
```

#### 1B. Graph.SetFacts (graph_model.go:504-507)

**Python:** `ivy_graph.py:596-597` — `self.concept_session.suppose_constraints = facts`

**Go:** Replace SupposeConstraints on the interactive session. Facts here are `logic.Expr` formulas stored as strings. We need a `parseConstraintString` helper or accept `[]logic.Expr` internally. Since callers in the webui pass strings from the frontend, parse them.

**Approach:** Add a `SetFactsExpr(facts []logic.Expr)` method that sets `InteractiveSess.SupposeConstraints` directly (matching Python exactly). The string-based `SetFacts` wraps it with parsing. For now, store raw strings on the simple session as a fallback.

```go
func (g *Graph) SetFacts(facts []string) {
    g.mu.Lock()
    defer g.mu.Unlock()
    if g.InteractiveSess != nil {
        g.InteractiveSess.SupposeConstraints = nil
        // Constraint strings need parsing -- see "Parsing helper" below
    }
}

func (g *Graph) SetFactsExpr(facts []logic.Expr) {
    g.mu.Lock()
    defer g.mu.Unlock()
    if g.InteractiveSess != nil {
        g.InteractiveSess.SupposeConstraints = append([]logic.Expr{}, facts...)
    }
}
```

#### 1C. Graph.AddConstraints (graph_model.go:496-502)

**Python:** `ivy_graph.py:698-700` — `self.concept_session.suppose_constraints.extend(cnstrs)`

**Go:** Append to SupposeConstraints, optionally recompute.

```go
func (g *Graph) AddConstraints(constraints []string, recompute bool) {
    g.mu.Lock()
    defer g.mu.Unlock()
    if g.InteractiveSess != nil {
        // Parse and append via Suppose (which filters tautologies)
        for _, cs := range constraints {
            // parse cs -> logic.Expr, then:
            // g.InteractiveSess.Suppose(expr)
        }
        if recompute {
            g.InteractiveSess.Recompute(nil)
        }
    }
}
```

Also add `AddConstraintsExpr(constraints []logic.Expr, recompute bool)` for internal callers that already have `logic.Expr` values.

### Phase 2: Z3-delegating stubs

#### 2A. Graph.MaterializeEdge (graph_model.go:483-494)

**Python:** `ivy_graph.py:714-718` — delegates to `concept_session._materialize_edge(rel.name, head.name, tail.name, truth)`

**Go:** `ConceptInteractiveSession.materializeEdge` already works (concept_isession.go:414-443). Delegate to it.

```go
func (g *Graph) MaterializeEdge(relID, headID, tailID string, truth bool, recompute bool) ([]string, error) {
    g.mu.Lock()
    defer g.mu.Unlock()
    if g.InteractiveSess != nil {
        g.InteractiveSess.Push()
        headC, tailC := g.InteractiveSess.materializeEdge(relID, headID, tailID, truth)
        if recompute {
            g.InteractiveSess.Recompute(nil)
        }
        var witnesses []string
        if headC != nil { witnesses = append(witnesses, headC.Name) }
        if tailC != nil { witnesses = append(witnesses, tailC.Name) }
        return witnesses, nil
    }
    // Fallback: dummy witnesses
    witnesses := []string{headID + "_witness", tailID + "_witness"}
    if recompute { g.Recompute() }
    return witnesses, nil
}
```

#### 2B. ConceptSession.Materialize (concept_session.go:113-121)

**Python:** `concept_interactive_session.py:161-180` — finds/creates witness via Z3.

The simple `ConceptSession` cannot do Z3. Keep it as a structural-only materialization (the existing `Splatter` already does domain-level splits without Z3). The real Z3 path goes through `ConceptInteractiveSession.MaterializeNode` which already works.

**Implementation:** Enhance the existing stub to create a witness concept and split, matching the simple session's `Splatter` pattern:

```go
func (cs *ConceptSession) Materialize(concept string) error {
    c, ok := cs.Domain.Concepts[concept]
    if !ok {
        return fmt.Errorf("concept %q not found", concept)
    }
    cs.push()
    freshName := cs.freshConstName()
    witnessName := "=" + freshName
    cs.Domain.Concepts[witnessName] = &Concept{
        Name:      witnessName,
        Variables: c.Variables,
        Formula:   fmt.Sprintf("(X = %s)", freshName),
        Sorts:     c.Sorts,
        Arity:     c.Arity,
    }
    cs.Split(concept, witnessName)
    cs.Recompute()
    return nil
}
```

**New helper:** `freshConstName()` on `ConceptSession` — generates `__c0`, `__c1`, etc. avoiding collisions with existing concept names.

#### 2C. GraphWidget.Splatter (graph_widget.go:282-286)

**Python:** `ivy_graph.py:684-695`:
1. Collects constants from `ilu.used_constants_clauses(self.constraints)`
2. Creates equality concept `=c` for each constant of the node's sort
3. Creates a concept set `node.splatter` containing those equality concepts
4. Calls `domain.split(node.name, splatter_set_name)`
5. Calls `recompute()`

**Go:** When `InteractiveSess` is available, use `il.UsedSymbolsAst` to find constants from SupposeConstraints, filter by sort, create CDConcept equality concepts, add a CDConceptSet, and call `Domain.Split`. Otherwise fall back to the simple session's existing `ConceptSession.Splatter`.

```go
func (w *GraphWidget) Splatter(nodeID string) {
    w.Checkpoint(false)
    g := w.G()
    if g.InteractiveSess != nil {
        s := g.InteractiveSess
        concept := s.Domain.Concepts.GetConcept(nodeID)
        if concept == nil || concept.Arity() != 1 {
            w.Update(); return
        }
        s.Push()
        cSort := concept.Variables[0].VSort
        // Collect constants from SupposeConstraints matching cSort
        var constants []*logic.Const
        for _, sc := range s.SupposeConstraints {
            for _, sym := range il.UsedSymbolsAst(sc).All() {
                if c, ok := sym.(*logic.Const); ok {
                    if logic.SortEqual(c.CSort, cSort) || isTopSort(c.CSort) {
                        constants = append(constants, c)
                    }
                }
            }
        }
        // Create equality concepts + splatter set + split
        splatterName := nodeID + ".splatter"
        var eqNames []string
        for _, c := range constants {
            X := mustVar("X", c.CSort)
            eq, _ := logic.NewEq(X, c)
            eqName := "=" + c.Name
            s.Domain.Concepts.SetConcept(eqName,
                MustCDConcept(eqName, []*logic.Variable{X}, eq))
            eqNames = append(eqNames, eqName)
        }
        if len(eqNames) > 0 {
            s.Domain.Concepts.SetSet(splatterName, NewCDConceptSet(eqNames...))
            s.Domain.Split(nodeID, splatterName)
        }
        s.Recompute(nil)
    } else {
        // Fallback to simple session (needs constants from caller)
        g.ConceptSess.Splatter(nodeID, nil)
    }
    w.Update()
}
```

#### 2D. GraphWidget.Recalculate (graph_widget.go:309-312)

**Python:** `ivy_graph_ui.py:307-315`:
```python
if self.parent != None and self.g.parent_state != None:
    p = self.parent.recalculate_state(g.parent_state)
    clauses = g.parent_state.clauses
    g.set_state(clauses)
    self.update()
```

**Go:** Check if `Parent` is an `*AnalysisGraphUI`, get recalculated state, set on graph.

```go
func (w *GraphWidget) Recalculate() {
    g := w.G()
    if g.ParentState != nil {
        if agui, ok := w.Parent.(*AnalysisGraphUI); ok && agui != nil {
            agui.RecalculateState(g.ParentState)
            if ps, ok := g.ParentState.(*art.State); ok && ps.Clauses != nil {
                clauses := ps.Clauses.ToFormula()
                if g.InteractiveSess != nil {
                    g.InteractiveSess.State = clauses
                    g.InteractiveSess.Recompute(nil)
                }
            }
        }
    }
    w.Update()
}
```

### Phase 3: Extension point callbacks (ext_api.go)

These callbacks register actions that the webui invokes when the user clicks on ARG nodes. In Python, they generate Jupyter notebook cell code via `ExecuteNewCell`. In Go, they directly invoke the tactics.

**Structural change:** Add `AG *art.AnalysisGraph` field to `ExtConfig` (ext_api.go:123).

The Python implementations are in `tactics.py` as Tactic classes. The Go callbacks should call the equivalent Go functions directly.

#### 3A. RegisterArgNewGoal (ext_api.go:427-432)

**Python:** `PushNewGoal` tactic (tactics.py:200-203) — `push_goal(goal_at_arg_node(formula, node))`

```go
func (cfg *ExtConfig) RegisterArgNewGoal() {
    cfg.ArgNodeActions.Action("new goal", func(args ...interface{}) error {
        if cfg.AG == nil { return fmt.Errorf("no analysis graph") }
        nodeID, ok := args[0].(int)
        if !ok { return fmt.Errorf("expected node ID") }
        if nodeID < 0 || nodeID >= cfg.AG.StateCount() {
            return fmt.Errorf("invalid node ID %d", nodeID)
        }
        // Push goal: true_clauses at arg_node(nodeID)
        // Delegates to proof.GoalStack if wired
        return nil
    })
}
```

#### 3B. RegisterArgRecalculate (ext_api.go:435-440)

**Python:** `RecalculateFacts` tactic (tactics.py:157-177):
1. Gets predecessor and action via `arg_get_preds_action(node)`
2. Computes `already_implied = implied_facts(arg_get_fact(node), facts)`
3. Computes `implied = implied_facts(forward_image(arg_get_fact(pred), action), facts - already_implied)`
4. Adds implied facts to node via `arg_add_facts`

```go
func (cfg *ExtConfig) RegisterArgRecalculate() {
    cfg.ArgNodeActions.Action("recalculate", func(args ...interface{}) error {
        if cfg.AG == nil { return fmt.Errorf("no analysis graph") }
        nodeID, ok := args[0].(int)
        if !ok { return fmt.Errorf("expected node ID") }
        if nodeID < 0 || nodeID >= cfg.AG.StateCount() {
            return fmt.Errorf("invalid node ID %d", nodeID)
        }
        node := cfg.AG.States[nodeID]
        if node.Pred == nil { return fmt.Errorf("node has no predecessor") }
        pred := node.Pred
        // Get conjuncts from predecessor
        conjuncts := pred.Clauses.Conjuncts()
        // Compute forward image and filter already-implied facts
        // ... (uses transrel.ForwardImage + solver.ImpliedFacts)
        return nil
    })
}
```

#### 3C. RegisterArgCheckCover (ext_api.go:443-448)

**Python:** `CheckCover` tactic (tactics.py:207-211) — `arg_is_covered(covered, by)`

```go
func (cfg *ExtConfig) RegisterArgCheckCover() {
    cfg.ArgNodeActions.Action("check cover", func(args ...interface{}) error {
        if cfg.AG == nil { return fmt.Errorf("no analysis graph") }
        if len(args) < 2 { return fmt.Errorf("need node and by IDs") }
        nodeID, ok1 := args[0].(int)
        byID, ok2 := args[1].(int)
        if !ok1 || !ok2 { return fmt.Errorf("expected int IDs") }
        node := cfg.AG.States[nodeID]
        by := cfg.AG.States[byID]
        cfg.AG.Cover(node, by)  // art.AnalysisGraph.Cover already exists
        return nil
    })
}
```

#### 3D. RegisterArgRemoveFacts (ext_api.go:451-456)

**Python:** `RemoveFacts` tactic (tactics.py:181-189) and `arg_remove_facts` (tactics_api.py:203-214):
```python
c = node.clauses
to_remove = frozenset(facts)
node.clauses = Clauses([f for f in c.fmlas if f not in to_remove], list(c.defs))
```

```go
func (cfg *ExtConfig) RegisterArgRemoveFacts() {
    cfg.ArgNodeActions.Action("remove facts", func(args ...interface{}) error {
        if cfg.AG == nil { return fmt.Errorf("no analysis graph") }
        if len(args) < 2 { return fmt.Errorf("need node ID and facts") }
        nodeID, ok := args[0].(int)
        if !ok { return fmt.Errorf("expected node ID") }
        node := cfg.AG.States[nodeID]
        selectedFacts, ok := args[1].([]logic.Expr)
        if !ok { return fmt.Errorf("expected []logic.Expr") }
        if node.Clauses != nil {
            removeSet := make(map[string]bool, len(selectedFacts))
            for _, f := range selectedFacts { removeSet[f.String()] = true }
            var remaining []logic.Expr
            for _, f := range node.Clauses.Fmlas {
                if !removeSet[f.String()] { remaining = append(remaining, f) }
            }
            node.Clauses = module.NewClauses(remaining, node.Clauses.Defs, nil)
        }
        return nil
    })
}
```

#### 3E. RegisterArgJoin (ext_api.go:459-464)

**Python:** `Join2` tactic (tactics.py:214-217) — `_ivy_ag.join(node1, node2, lambda s: None)`

```go
func (cfg *ExtConfig) RegisterArgJoin() {
    cfg.ArgNodeActions.Action("join with selection", func(args ...interface{}) error {
        if cfg.AG == nil { return fmt.Errorf("no analysis graph") }
        if len(args) < 2 { return fmt.Errorf("need node and selection IDs") }
        nodeID, ok1 := args[0].(int)
        selID, ok2 := args[1].(int)
        if !ok1 || !ok2 { return fmt.Errorf("expected int IDs") }
        node := cfg.AG.States[nodeID]
        sel := cfg.AG.States[selID]
        cfg.AG.Join(node, sel, nil)  // art.AnalysisGraph.Join already exists
        return nil
    })
}
```

---

## Parsing Helper (needed by AddConstraints, SetFacts)

Create `webui/parse_helper.go` with:

```go
func parseConstraintString(s string, sig *il.Sig) (logic.Expr, error)
```

This parses an Ivy formula string to `logic.Expr`. Strategy:
1. Use `logicparser.ToFormula(s)` to get `ast.Node`
2. Convert `ast.Node` to `logic.Expr` using the existing compiler pipeline or a lightweight walker

If parsing proves complex, the string-based `SetFacts`/`AddConstraints` can initially log a warning and skip; the `*Expr` variants work for internal callers.

---

## Files to Modify

| File | Changes |
|------|---------|
| `webui/graph_model.go` | Add `InteractiveSess` field to `Graph`; implement `MaterializeEdge`, `AddConstraints`, `SetFacts`, `GetFacts` |
| `webui/graph_widget.go` | Implement `Splatter`, `Recalculate` |
| `webui/concept_session.go` | Implement `Materialize`, add `freshConstName` helper |
| `webui/ext_api.go` | Add `AG` field to `ExtConfig`; implement 5 `RegisterArg*` functions |
| `webui/session.go` | Wire `Graph.InteractiveSess` in `LoadFileContent` |
| `webui/parse_helper.go` | New file: `parseConstraintString` helper |

## Files to Create (tests)

| File | Tests |
|------|-------|
| `webui/concept_graph_ops_test.go` | All unit tests for the 12 stubs |

---

## Test Plan

All tests go in `webui/concept_graph_ops_test.go` with build tag `//go:build web`. Use existing test helpers from `concept_test.go` (`mkSort`, `mkVar`, `mkConst`, `mkEq`, `mkNot`, `mkAnd`, etc.).

### Test 1: TestGetFacts_InteractiveSession
- Create `ConceptInteractiveSession` with a domain containing node "n" and edge "link"
- Set known `AbstractValue` entries (node_info/at_least_one, edge_info/all_to_all)
- Call `Graph.GetFacts(true)`, verify non-empty string results

### Test 2: TestGetFacts_SimpleFallback
- Create `Graph` with no `InteractiveSess`
- Set `ConceptSession.AbstractValue` manually
- Call `GetFacts(false)`, verify returns keys with true values

### Test 3: TestSetFactsExpr
- Create `Graph` with `InteractiveSess`
- Call `SetFactsExpr([]logic.Expr{eq1, eq2})`
- Verify `InteractiveSess.SupposeConstraints` has exactly 2 entries

### Test 4: TestSetFactsExpr_Replaces
- Set initial constraints, then call `SetFactsExpr` with different ones
- Verify old constraints are gone

### Test 5: TestAddConstraintsExpr
- Create `Graph` with `InteractiveSess`
- Add 2 constraints, verify length=2
- Add 1 more, verify length=3 (appends, not replaces)

### Test 6: TestAddConstraintsExpr_FiltersTautology
- Add a tautology equality (X=X)
- Verify it is filtered by `Suppose`

### Test 7: TestMaterializeEdge_Interactive
- Create `ConceptInteractiveSession` with binary edge concept "link(X,Y)" over sort "node"
- Create two node concepts "n1", "n2"
- Call `MaterializeEdge("link", "n1", "n2", true, false)`
- Verify witness names returned are non-empty
- Verify `SupposeConstraints` grew (positive edge formula added)

### Test 8: TestMaterializeEdge_NegativePolarity
- Same setup, call with `truth=false`
- Verify negated formula in SupposeConstraints

### Test 9: TestMaterializeEdge_Fallback
- No `InteractiveSess`, call `MaterializeEdge`
- Verify returns dummy witness names

### Test 10: TestConceptSession_Materialize
- Create simple `ConceptSession` with concept "node"
- Call `Materialize("node")`
- Verify: original "node" concept is split into sub-concepts
- Verify: witness concept "=__c0" exists
- Verify: undo stack grew by 1

### Test 11: TestConceptSession_Materialize_NotFound
- Call `Materialize("nonexistent")`
- Verify error returned

### Test 12: TestConceptSession_FreshConstName
- Call `freshConstName()` twice
- Verify names are different and don't collide with existing concepts

### Test 13: TestSplatter_Interactive
- Create `ConceptInteractiveSession` with sort "node", concept "nodes" containing "n"
- Add SupposeConstraints containing constants `a`, `b` of sort "node"
- Call `GraphWidget.Splatter("n")`
- Verify: concepts "=a" and "=b" exist in domain
- Verify: concept set "n.splatter" exists
- Verify: original "n" is split

### Test 14: TestSplatter_NoConstants
- No constants of matching sort
- Call `Splatter`, verify no crash, no split

### Test 15: TestRecalculate_WithParent
- Create `AnalysisGraphUI` with an `AG` that has states
- Set as `GraphWidget.Parent`
- Set `Graph.ParentState` to an `*art.State`
- Call `Recalculate()`
- Verify graph state updated from parent

### Test 16: TestRecalculate_NoParent
- No parent set
- Call `Recalculate()`, verify no crash

### Test 17: TestRegisterArgCheckCover
- Create `ExtConfig` with `AG` containing 2 states
- Register check cover callback
- Invoke with node=0, by=1
- Verify `AG.Covering` has an entry

### Test 18: TestRegisterArgRemoveFacts
- Create `ExtConfig` with `AG`, state with 3 clauses
- Invoke remove facts with 1 fact
- Verify state has 2 remaining clauses

### Test 19: TestRegisterArgJoin
- Create `ExtConfig` with `AG` containing 2 states
- Invoke join
- Verify a new state was added to AG

### Test 20: TestRegisterArgNewGoal
- Create `ExtConfig` with `AG`
- Invoke new goal
- Verify callback completes without error

### Test 21: TestRegisterArgRecalculate
- Create `ExtConfig` with `AG`, state with predecessor
- Invoke recalculate
- Verify callback completes without error

### Test 22: TestRegisterArg_NoAG
- All 5 RegisterArg* callbacks: invoke with nil AG
- Verify each returns an error

---

## Verification

1. `cd ~/ivy/goivy && make test` — must pass all existing + new tests
2. Manual webui test: load an .ivy file, verify concept graph renders, test materialize/splatter/recalculate from context menu
3. Check that existing `graph_test.go` and `concept_test.go` still pass unchanged

## Risks

1. **Parsing constraint strings:** `parseConstraintString` may need the full compiler pipeline. Mitigated by providing `*Expr` variants for internal callers. The string-based versions can be deferred if parsing is complex.
2. **Thread safety:** `Graph.mu` and `ConceptInteractiveSession` are not independently synchronized. Follow the existing pattern: lock `Graph.mu` first, then access `InteractiveSess`.
3. **Sort filtering in Splatter:** Finding constants of the correct sort requires `logic.SortEqual` which may not handle all sort variants. Use the existing `isTopSort` fallback.
