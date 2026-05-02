# Plan: WebUI CTI (Counter-To-Induction) Implementation

**Created:** 2026-05-02 15:42 UTC

## Context

The CTI workflow is the interactive invariant strengthening loop in Ivy's web UI. The user sees a counterexample to induction (CTI), selects facts from the concept graph that distinguish good from bad states, gathers them into a conjecture, minimizes it, checks sufficiency, and adds it to the invariant. All 10 functions in this workflow are stubbed in Go (`webui/ui_cti.go`). The Python source of truth is `~/ivy/pyivy/ivy/ivy/ivy_ui_cti.py`.

Items 1 (UPDR) and 2 (Session solver wiring) are already marked done, so the dependencies for CTI are satisfied.

## Prerequisite: Struct Field Type Changes

The current Go struct stores conjectures as `[]string` and `CurrentConjecture` as `string`. Python stores `Clauses` objects. Change to match Python.

**File:** `webui/ui_cti.go` lines 16-41

```go
// BEFORE:
Conjectures       []string
CurrentConjecture string

// AFTER:
Conjectures       []*module.Clauses
CurrentConjecture *module.Clauses
```

Also add fields needed for solver and module access:

```go
Mod    *module.Module      // compiled module (axioms, sig, actions)
Solver *z3bridge.Solver     // Z3 solver instance
AG     *art.AnalysisGraph   // stored analysis graph from last CheckInductiveness
```

Update `NewCTIAnalysisGraphUI` to accept `*module.Module` and create the solver. Update `StartCTI` signature: `StartCTI(conjectures []*module.Clauses)`.

Add new import paths to `ui_cti.go`:
- `github.com/glycerine/ivy/goivy/module`
- `github.com/glycerine/ivy/goivy/art`
- `github.com/glycerine/ivy/goivy/trace`
- `github.com/glycerine/ivy/goivy/actions`
- `github.com/glycerine/ivy/goivy/z3bridge`
- `github.com/glycerine/ivy/goivy/bmc`
- `github.com/glycerine/ivy/goivy/logic`
- `github.com/glycerine/ivy/goivy/logicutil`
- `github.com/glycerine/ivy/goivy/ivylogic`
- `github.com/glycerine/ivy/goivy/ivyutils`
- `github.com/glycerine/ivy/goivy/interp`

## Prerequisite: Port `FormulaToConceptl` helper

Python's `concept_from_formula(fmla)` (ivy_graph.py:23-28) creates a `Concept` from a formula. No Go equivalent exists. Port it.

**File:** `webui/graph_model.go` (add method on `*Graph`)

```go
func (g *Graph) FormulaToConceptl(fmla logic.Expr) *CDConcept {
    // Python: vs = sorted(used_variables_ast(fmla), key=str)
    // name = ','.join(str(v)+':'+str(v.sort) for v in vs) + '.' + str(fmla)
    // return Concept(name, vs, fmla)
    vs := logicutil.UsedVariablesAsts([]logic.Expr{fmla})
    sort.Slice(vs, func(i, j int) bool { return vs[i].String() < vs[j].String() })
    // Build name: "V1:sort1,V2:sort2.formula"
    var parts []string
    for _, v := range vs {
        parts = append(parts, fmt.Sprintf("%v:%v", v, v.VSort))
    }
    name := strings.Join(parts, ",") + "." + fmt.Sprintf("%v", fmla)
    c, _ := NewCDConcept(name, vs, fmla)
    return c
}
```

Also need `AddDomainConcept` if not present (Python `add_domain_concept`, ivy_graph.py:30-36).

## Implementation: 10 Stub Functions

### 1. AutodetectTransitive (ui_cti.go:91)

**Python:** `ivy_ui_cti.py:72-103`

**Go APIs:**
- `ui.Mod.Sig.AllSymbols()` → `[]*lg.Const` (`ivylogic/sig.go:201`)
- `ui.Mod.BackgroundTheory(nil)` → `*module.Clauses` (`module/theory.go:26`)
- `ui.Solver.ClausesImply(axioms, t)` → `(bool, error)` (`z3bridge/solver.go:721`)
- `g.FormulaToConceptl(fmla)` → `*CDConcept` (new, above)
- `w.ShowRelation(concept, "T", true, false)` (`graph_widget.go:391`)

**Logic:**
1. Clear TransitiveRelations/TransitiveRelationConcepts
2. Get `axioms = ui.Mod.BackgroundTheory(nil)`
3. For each `c` in `ui.Mod.Sig.AllSymbols()`:
   - Skip if sort is not `*logic.FunctionSort` with arity==2, domain[0]==domain[1], range==Boolean
   - Create variables X, Y, Z of domain[0] sort
   - Build `transitive = ForAll([X,Y,Z], Or(Not(c(X,Y)), Not(c(Y,Z)), c(X,Z)))`
   - Build `defined_symmetry = ForAll([X,Y], Or(c(X,X), Not(c(Y,Y))))`
   - Wrap both in `module.NewClauses([transitive, defined_symmetry], nil, nil)`
   - If `solver.ClausesImply(axioms, t)`: store name, create concept, call ShowRelation
4. If any found, call `ui.CurrentConceptGraph.Update()`

### 2. CheckInductiveness (ui_cti.go:105)

**Python:** `ivy_ui_cti.py:106-180`

**Go APIs:**
- `trace.MakeCheckArt(ui.Mod, "", precondSlice)` → `(*art.AnalysisGraph, *art.State, *art.State, error)` (`trace/trace.go:478`)
- `module.DualClauses(conj, witness, nil)` → `*module.Clauses` (`module/ops.go:470`)
- `module.TrueClauses(nil)` (`module/clauses.go`)
- `trace.CheckFinalCond(ag, post, clauses, relsToMin, true)` → `*trace.TraceBase` (`trace/trace.go:552`)
- `actions.EmptyAnnotation{}` for annotations

**Logic:**
1. Auto-detect relations to minimize from sig (boolean function symbols not in TransitiveRelations)
2. `ag, succeed, fail, err := trace.MakeCheckArt(ui.Mod, "", ui.Conjectures)`
3. `toTest := append([]*module.Clauses{nil}, ui.Conjectures...)` — nil = safety check
4. For each conj in toTest:
   - Build witness function: `func(v *lg.Variable) lg.Expr { return lg.NewConst("@"+v.Name, v.VSort) }`
   - If conj == nil: `clauses = module.TrueClauses(nil)`, `post = fail`
   - Else: `clauses = module.DualClauses(conj, witness, nil)`, `post = succeed`
   - Set `clauses.Annot = actions.EmptyAnnotation{}`
   - `res := trace.CheckFinalCond(ag, post, clauses, relsToMin, true)`
   - If res != nil: store CTI (`ui.CurrentConjecture = conj`, `ui.AG = res.AnalysisGraph`, `ui.HaveCTI = true`), call ShowUsedRelations, return (false, msg)
5. If all pass: `ui.HaveCTI = false`, return (true, msg)

### 3. BoundedCheck (ui_cti.go:132)

**Python:** `ivy_ui_cti.py:280-356`

**Go APIs:**
- `art.NewAnalysisGraph(ui.Mod)` (`art/art.go`)
- `bmc.EnvAction(ui.Mod)` → `actions.Action` (`bmc/bmc.go:202`)
- `ag.Execute(true, envAction, post, nil, "")` (`art/art.go`)
- `module.AndClausesTyped(a, b)` for conjoining conjectures
- `trace.CheckFinalCond(ag, post, clauses, nil, true)`

**Logic:**
1. Build conjecture: if `conjecture == nil`, `conj = module.AndClausesTyped(ui.Conjectures...)`
2. Build witness, compute `clauses = module.DualClauses(conj, witness, nil)`
3. Create fresh `ag := art.NewAnalysisGraph(ui.Mod)`
4. Add initial state: `ag.Add(art.NewState(ui.Mod, ag.InitCond), nil)`, `post = ag.States[0]`
5. If "initialize" action exists in module, execute it
6. Loop n=0..bound:
   - `res := trace.CheckFinalCond(ag, post, clauses, nil, true)`
   - If res != nil: return (true, "BMC found counter-example at step N")
   - `post, _ = ag.Execute(true, bmc.EnvAction(ui.Mod), post, nil, "")`
7. Return (false, "BMC with bound N did not find a counter-example")

### 4. Diagram (ui_cti.go:154)

**Python:** `ivy_ui_cti.py:207-227`

**Go APIs:**
- `module.DualClauses(conj, nil, nil)` or `module.TrueClauses(nil)`
- `ui.Mod.BackgroundTheory(nil)`
- `interp.UniverseConstraint(state)` (`interp/phase4.go:437`)
- `actions.ReverseImage(post, axioms, update)` (`actions/transrel.go:1426`)
- `module.AndClausesTyped(a, b)`
- `ui.Solver.GetModelClauses(clauses)` (`z3bridge/solver_model.go:37`)
- `ui.Solver.ClausesModelToDiagram(rev, trace.IsSkolem, axioms)` (`z3bridge/solver_clauses.go:143`)

**Logic:**
1. If !ui.HaveCTI: call checkInductivenessUnlocked(); if inductive, return
2. `conj = ui.CurrentConjecture`
3. `post = module.DualClauses(conj, nil, nil)` if conj != nil, else `module.TrueClauses(nil)`
4. `pre = ui.AG.States[0].Clauses`
5. `axioms = ui.Mod.BackgroundTheory(nil)`
6. `uc = interp.UniverseConstraint(ui.AG.States[0])` (needs interp.State conversion)
7. `axioms_uc = module.AndClausesTyped(axioms, uc)`
8. `rev = actions.ReverseImage(post, axioms, ui.AG.States[1].Update)`
9. `clauses = module.AndClausesTyped(module.AndClausesTyped(pre, rev), axioms_uc)`
10. `mod, _ = ui.Solver.GetModelClauses(clauses)` — assert mod != nil
11. `diag, _ = ui.Solver.ClausesModelToDiagram(rev, trace.IsSkolem, axioms)`
12. ViewState with diag, ShowUsedRelations(diag, true), GatherFacts()

### 5. ShowUsedRelations (ui_cti.go:218)

**Python:** `ivy_ui_cti.py:182-205`

**Go APIs:**
- `module.UsedSymbolsAST(clauses.ToFormula())` (`module/astutil.go:30`) — returns `*InsMap[NodeKey, Expr]`
- `ivylogic.NormalizeSymbol(sym, nil)` (`ivylogic/ivylogic.go:386`)
- `module.AppsClauses(clauses)` (`module/ops.go:1241`)
- `w.ClearEdges()`, `w.ShowRelation(rel, "+", true, false)`, `w.Update()` (`graph_widget.go`)
- `g.FormulaToConceptl(fmla)`, `g.NewRelation(concept)` (`graph_model.go`)

**Logic:**
1. Clear edges
2. Get relations from concept graph: `rels = ui.CurrentConceptGraph.G().Relations()`
3. Compute used constants from `clauses.ToFormula()` via `UsedSymbolsAST`
4. For each relation in rels:
   - Get formula constants, check if any is in used set and not '@'-prefixed
   - If match: `ShowRelation(rel, "+", true, false)`. If both: also `ShowRelation(rel, "-", true, false)`
5. Handle arity-3 apps: iterate `AppsClauses(clauses)`, build formula, create concept, add relation
6. Update

**Note:** Need to change signature to accept `*module.Clauses` instead of `string`:
```go
func (ui *CTIAnalysisGraphUI) ShowUsedRelations(clauses *module.Clauses, both bool)
```

### 6. GatherFacts (ui_cti.go:276)

**Python:** `ivy_ui_cti.py:675-722`

**Go APIs:**
- `g.ConceptSess.GetNodeFacts(node)` (`concept_isession.go:467`)
- `g.ConceptSess.GetEdgeFacts(edge, source, target, &polarity)` (`concept_isession.go:556`)
- `g.ConceptSess.AbstractValue` — `[]TagValue` with Tag fields
- `g.EdgeDisplayCheckboxes` — map[edge]map[class]*Option
- `g.SetFacts(facts)` (`graph_model.go:505`/`graph_widget.go:323`)
- `w.HighlightSelectedFacts()` (`graph_widget.go:447`)

**Logic:**
1. Get selectedNodes (or all if none selected)
2. For each node: collect `g.ConceptSess.GetNodeFacts(node)` → append to facts list
3. Collect edges from AbstractValue where tag[0]=="edge_info" and src/tgt in selectedNodes
4. For each (edge, source, target):
   - If `EdgeDisplayCheckboxes[edge]["all_to_all"].Value`:
     - Skip reflexive edges for transitive relations
     - Get positive edge facts via `GetEdgeFacts(edge, src, tgt, &true)`
   - If `EdgeDisplayCheckboxes[edge]["none_to_none"].Value` or edge=="=":
     - Get negative edge facts via `GetEdgeFacts(edge, src, tgt, &false)`
5. Filter: remove `Not(Eq(t1,t2))` where t1 >= t2 (string comparison)
6. Convert facts to string list, call `g.SetFacts(factStrings)`
7. `w.Update()`, `w.HighlightSelectedFacts()`

**Note:** GatherFacts works with `logic.Expr` internally but `SetFacts` takes `[]string`. Need to store the `[]logic.Expr` for use by GetSelectedConjecture. Add a field:
```go
type CTIConceptGraphWidget struct {
    *GraphWidget
    ParentCTI *CTIAnalysisGraphUI
    ActiveFactExprs []logic.Expr  // logic expressions for gathered facts
}
```

### 7. GetSelectedConjecture (ui_cti.go:314)

**Python:** `ivy_ui_cti.py:399-440`

**Go APIs:**
- `logicutil.FreeVariables(f)` (`logicutil/logicutil.go:61`)
- `logicutil.Substitute(f, subs)` (`logicutil/logicutil.go:201`)
- `module.Negate(f)` (`module/clauses.go:352`)
- `module.SimplifyClauses(clauses)` (`module/ops.go:1434`)
- `ivylogic.IsNumeral(c)` (`ivylogic/ivylogic.go:272`)
- `ivylogic.IsUninterpretedSort(sig, s)` (`ivylogic/globals.go:48`)
- `ivyutils.NewVariableGenerator()` (`ivyutils/renamer.go:45`)

**Logic:**
1. Get active facts from `w.ActiveFactExprs`
2. Assert no free variables: `logicutil.FreeVariables(facts...)` must be empty
3. Collect constants used in facts
4. For each numeral constant of uninterpreted sort: substitute with fresh variable via VariableGenerator
5. Negate each fact after substitution: `literals = [negate(substitute(f, subs)) for f in facts]`
6. Create disjunction: `lg.Or(literals...)`
7. Wrap in Clauses, simplify via `module.SimplifyClauses(result)`
8. Convert Or to Not(And(negate(lit) for lit in ...)): if result is Or, rewrite as `Not(And(negate(each_lit)))`
9. Return `*module.Clauses`

**Change return type** from `string` to `*module.Clauses`:
```go
func (w *CTIConceptGraphWidget) GetSelectedConjecture() *module.Clauses
```

### 8. MinimizeConjecture (ui_cti.go:326)

**Python:** `ivy_ui_cti.py:468-510`

**Go APIs:**
- `w.ParentCTI.BoundedCheck(bound, conj)` — first run BMC
- `art.NewAnalysisGraph(mod)`, `bmc.EnvAction(mod)`
- `module.AndClausesTyped(post.Clauses, axioms)`
- `ui.Solver.UnsatCore(clauses1, clauses2, nil, nil)` (`z3bridge/solver.go:786`)

**Logic:**
1. Run BMC first: if counter-example found, return early
2. Create fresh AG, execute env_action `n_steps` times (using `w.ParentCTI.CurrentBound`)
3. `postClauses = module.AndClausesTyped(post.Clauses, axioms)`
4. Get active facts from `w.ActiveFactExprs`
5. `core, _ = ui.Solver.UnsatCore(module.NewClauses(facts, nil, nil), postClauses, nil, nil)`
6. Filter facts to keep only those in core
7. Update: `g.SetFacts(...)`, `w.HighlightSelectedFacts()`
8. Return minimized conjecture string

### 9. IsSufficient (ui_cti.go:333)

**Python:** `ivy_ui_cti.py:516-562`

**Go APIs:**
- `w.GetSelectedConjecture()` → `*module.Clauses`
- `w.ParentCTI.CurrentConjecture` → target
- `module.AndClausesTyped(conj, andOfConjectures)`
- `art.NewAnalysisGraph(mod)`, `ag.Execute(...)`
- `module.DualClauses(target, witness, nil)`
- `trace.CheckFinalCond(ag, post, clauses, nil, true)`

**Logic:**
1. `conj := w.GetSelectedConjecture()`
2. `targetConj := w.ParentCTI.CurrentConjecture`
3. Create fresh AG
4. Build pre-state: `pre.Clauses = module.AndClausesTyped(conj, and_of_all_conjectures)`
5. Set annotation: `pre.Clauses.Annot = actions.EmptyAnnotation{}`
6. Execute env_action: `post, _ = ag.Execute(true, envAction, pre, nil, "")`
7. `post.Clauses = module.TrueClauses(nil)` with annotation
8. Build witness function, `clauses = module.DualClauses(targetConj, witness, nil)`
9. `res = trace.CheckFinalCond(ag, post, clauses, nil, true)`
10. If res != nil: return (false, "does not imply"), else return (true, "implies")

### 10. IsInductive (ui_cti.go:341)

**Python:** `ivy_ui_cti.py:565-609`

Identical to IsSufficient except `targetConj = conj` (self-induction check). Extract shared helper:

```go
func (w *CTIConceptGraphWidget) checkInductionHelper(conj, targetConj *module.Clauses) (bool, string)
```

Both IsSufficient and IsInductive call this with different targetConj.

## Callers to Update

After changing `Conjectures` from `[]string` to `[]*module.Clauses` and `CurrentConjecture` from `string` to `*module.Clauses`, update all callers:

- `StartCTI` signature
- `Weaken` — operates on `[]*module.Clauses` instead of `[]string`
- `SaveConjectures` — use `module.DropUniversals(conj.ToFormula())` for formatting
- `Strengthen` — receives `*module.Clauses` from GetSelectedConjecture
- `widget_analysis_session.go` — any callers that pass string conjectures

## Files Modified

| File | Changes |
|------|---------|
| `webui/ui_cti.go` | Main implementation: all 10 stubs, struct changes, new imports |
| `webui/graph_model.go` | Add `FormulaToConceptl` method |
| `webui/ui_cti_test.go` | New file: comprehensive unit tests |

## Testing Plan

**File:** `webui/ui_cti_test.go` (new)

Follow existing patterns from `webui/concept_test.go` (mk* helpers, testDomainSetup).

### Test Categories:

**A. AutodetectTransitive tests:**
- `TestAutodetectTransitiveFindsLE` — module with `le(X,Y)` axiom that's transitive: verify detection
- `TestAutodetectTransitiveSkipsNonTransitive` — binary relation without transitivity axiom: verify not detected
- `TestAutodetectTransitiveSkipsUnary` — unary predicate: verify skipped
- `TestAutodetectTransitiveEmpty` — module with no symbols: no panic, empty results

**B. CheckInductiveness tests:**
- `TestCheckInductivenessAllInductive` — module where all conjectures are inductive: returns (true, msg)
- `TestCheckInductivenessFindsCtI` — module with non-inductive conjecture: returns (false, msg), HaveCTI==true, CurrentConjecture set
- `TestCheckInductivenessSafetyFirst` — safety check (nil conjecture) fails: returns (false, msg)
- `TestCheckInductivenessRelationsToMinimize` — auto-detection of relations to minimize

**C. BoundedCheck tests:**
- `TestBoundedCheckNoCounterExample` — trivially safe with bound=0: returns (false, msg)
- `TestBoundedCheckFindsCounterExample` — conjecture violated at step 1: returns (true, msg)
- `TestBoundedCheckWithInitialize` — module with 'initialize' action

**D. Diagram tests:**
- `TestDiagramAfterCTI` — after CheckInductiveness finds CTI, Diagram produces non-empty result
- `TestDiagramTriggersCheckFirst` — when HaveCTI==false, calls CheckInductiveness first

**E. ShowUsedRelations tests:**
- `TestShowUsedRelationsEnablesMatchingEdges` — clauses using relation R: R's checkbox enabled
- `TestShowUsedRelationsSkipsSkolem` — '@'-prefixed constants skipped
- `TestShowUsedRelationsArity3` — ternary application creates new concept

**F. GatherFacts tests:**
- `TestGatherFactsAllNodes` — no selection: uses all nodes
- `TestGatherFactsSelectedNodes` — specific nodes selected
- `TestGatherFactsFiltersDuplicateEqualities` — Not(Eq(a,b)) where a>=b filtered
- `TestGatherFactsEdgePolarity` — all_to_all vs none_to_none polarity

**G. GetSelectedConjecture tests:**
- `TestGetSelectedConjectureBasic` — known facts produce expected negated universal conjecture
- `TestGetSelectedConjectureSubstitution` — numeral constants substituted with variables
- `TestGetSelectedConjectureNoFreeVars` — asserts no free variables in result
- `TestGetSelectedConjectureEmptyFacts` — returns nil when no facts

**H. MinimizeConjecture tests:**
- `TestMinimizeConjectureReducesFacts` — unsat core is smaller than original fact set
- `TestMinimizeConjectureEarlyExitOnBMC` — if BMC finds counter-example, returns early

**I. IsSufficient/IsInductive tests:**
- `TestIsSufficientTrue` — conjecture implies target: returns (true, msg)
- `TestIsSufficientFalse` — conjecture does not imply: returns (false, msg)
- `TestIsInductiveSelfCheck` — self-induction check with target_conj==conj
- `TestIsInductiveFalse` — non-inductive conjecture

**J. Integration tests:**
- `TestCTIWorkflowEndToEnd` — full loop: CheckInductiveness → Diagram → GatherFacts → GetSelectedConjecture → Strengthen → CheckInductiveness again

### Test Infrastructure:

```go
// testCTIModule creates a module with:
// - sort T with elements 0,1
// - binary relation le(X,Y) with transitivity axiom
// - init: le(0,0) & le(0,1) & le(1,1)
// - action: nondeterministic step
// - conjecture: forall X. le(X,X)  (inductive)
// - conjecture: forall X,Y. le(X,Y) | le(Y,X) (not inductive)
func testCTIModule() *module.Module

// testCTISetup creates a full CTIAnalysisGraphUI with solver wired up
func testCTISetup() (*CTIAnalysisGraphUI, *CTIConceptGraphWidget)
```

## Verification

1. `cd ~/ivy/goivy && make test` — all existing tests pass
2. `XTRACE_OFF=1 go test ./webui/ -run TestCTI -v -count=1` — new CTI tests pass
3. Manual: start webui, load an .ivy file with conjectures, click "Check induction" → should produce CTI or "Inductive invariant found"

## Package Dependencies (no new external deps)

All required Go packages already exist in the codebase:
- `trace/trace.go` — MakeCheckArt, CheckFinalCond
- `art/art.go` — AnalysisGraph, State
- `module/ops.go` — DualClauses, SimplifyClauses, AndClausesTyped
- `actions/transrel.go` — ReverseImage
- `z3bridge/solver.go` — ClausesImply, UnsatCore
- `z3bridge/solver_model.go` — GetModelClauses
- `z3bridge/solver_clauses.go` — ClausesModelToDiagram
- `bmc/bmc.go` — EnvAction
- `interp/phase4.go` — UniverseConstraint
- `logicutil/logicutil.go` — FreeVariables, Substitute
- `ivylogic/` — IsNumeral, IsUninterpretedSort, NormalizeSymbol, DropUniversals
- `ivyutils/renamer.go` — NewVariableGenerator
- `conceptspace/` — used internally by concept graph (not called directly by CTI)
