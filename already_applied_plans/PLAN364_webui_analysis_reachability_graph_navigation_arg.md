# Plan: WebUI ARG Navigation — Wire 15 Stub Functions

**Created:** 2026-05-02 07:12 UTC

## Context

The Go webui package has 15 stub functions in `AnalysisGraphUI` (file `webui/ui_main.go`) that implement the interactive ARG (Analysis Reachability Graph) exploration workflow. These are todo item #4 from the audit. The Python source of truth is `ivy_ui.py` and `ivy_art.py`. All underlying `art.AnalysisGraph` methods are **already implemented** in Go — the stubs just need to call through to them and sync the frontend rendering state.

A key architectural gap: `AnalysisGraphUI` currently has only a lightweight `G *AnalysisGraphState` rendering view but no reference to the real `art.AnalysisGraph` or `module.Module`. The completed CTI stubs (todo item #3 in `ui_cti.go`) solved this by adding `AG`, `Mod`, `Solver` fields to `CTIAnalysisGraphUI`. We follow the same pattern.

**Circular import constraint:** `alpha` imports `webui`, so `webui` cannot import `alpha`. The `GetAlpha()` abstractor factory must live on `Session` (which already has all needed imports) rather than on `AnalysisGraphUI`.

## Phase 1: Structural Changes

### 1A. Add fields to `AnalysisGraphUI` (`webui/ui_main.go:54-77`)

Add `AG *art.AnalysisGraph`, `Mod *module.Module`, `SyncCallback func()` fields. Add imports for `art`, `module`, `ast`, `logic`, `logicparser`.

### 1B. Add helper methods (`webui/ui_main.go`, new code)

- `stateByID(nodeID int) (*art.State, error)` — bounds-checked lookup into `ui.AG.States`
- `transitionByEndpoints(srcID, tgtID int) (*art.Transition, error)` — linear scan of `ui.AG.Transitions`
- `artToGraphState(ag *art.AnalysisGraph) *AnalysisGraphState` — shared AG→lightweight conversion (also used by `session.syncARGToGraph`)

### 1C. Add `GetAlpha()` on Session (`webui/session.go`)

Returns `art.Abstractor` based on mode. Session already imports `alpha`. For `ModeConcrete`: nil. For `ModeBounded`: top_alpha (sets clauses to true). For `ModeAbstract`/`ModeInduction`: `alpha.Alpha`. For `ModePDR`: `alpha.PredicateAlpha`.

### 1D. Add `AGUI *AnalysisGraphUI` field to Session (`webui/session.go`)

In `LoadFileContent`, after `s.AG = art.NewAnalysisGraph(...)`, create and wire the AGUI:
```
s.AGUI = NewAnalysisGraphUI()
s.AGUI.AG = s.AG
s.AGUI.Mod = s.CompiledModule
s.AGUI.SyncCallback = func() { s.syncARGToGraph() }
```

## Phase 2: Implement 15 Stub Functions

All in `webui/ui_main.go`, replacing existing stub bodies. Each follows the pattern: validate → delegate to `ui.AG.xxx()` → sync callback → return.

### Stub 1: `NodeExecuteCommands(nodeID int) []ActionEntry` (line 193)
**Python:** `ivy_ui.py:159` — `self.g.state_actions(n)`, sorted by label.
- Call `ui.AG.StateActions(state)`, sort by `stateEquationLabel(eq)`, return `[]ActionEntry`.
- Need helper `stateEquationLabel(eq *ast.Definition) string` extracting action name from Definition Lhs/Rhs.

### Stub 2: `CheckLocalSafety(nodeID int) (bool, string)` (line 256)
**Python:** `ivy_ui.py:357` — `self.g.check_safety(node)`.
- Call `ui.AG.CheckSafety(true, state)`. Return `result.Safe` and message from `result.Cex.Msg`.

### Stub 3: `CheckBoundedSafety(nodeID int) (bool, string)` (line 262)
**Python:** `ivy_ui.py:341` — `self.g.check_bounded_safety(node)`.
- Call `ui.AG.CheckBoundedSafety(state, nil)`. Return safety status. `result.Art` holds error trace AG.

### Stub 4: `FindExtension(nodeID int) (string, error)` (line 268)
**Python:** `ivy_ui.py:386` — `next(self.g.state_extensions(node))`, then `self.do_state_action(a)`.
- Call `ui.AG.StateExtensions(state, nil)`. If empty, return "state N is closed". Otherwise call `ui.AG.DoStateAction(false, extensions[0], abstractor)` and sync.

### Stub 5: `ExecuteAction(nodeID int, actionName string) error` (line 274)
**Python:** `ivy_ui.py:240` — `self.g.execute_action(a, n, self.get_alpha())`.
- Call `ui.AG.ExecuteAction(false, actionName, state, abstractor)` and sync. Abstractor obtained via session callback (see Phase 1C).

### Stub 6: `RecalculateAll()` (line 280)
**Python:** `ivy_ui.py:254` — iterates transitions, recalculates each unique target once.
- Iterate `ui.AG.Transitions`, track `done` set of post-state IDs, call `ui.AG.Recalculate(false, t, abstractor)` per unique target. Sync once at end.

### Stub 7: `RecalculateEdge(srcID, tgtID int)` (line 285)
**Python:** `ivy_ui.py:264` — `self.g.recalculate(transition, self.get_alpha())`.
- Find transition via `transitionByEndpoints`, call `ui.AG.Recalculate(false, *t, abstractor)`, sync.

### Stub 8: `DecomposeEdge(srcID, tgtID int) (*AnalysisGraphState, error)` (line 290)
**Python:** `ivy_ui.py:271` — `self.g.decompose_edge(transition)`.
- Find transition, call `ui.AG.DecomposeEdge(*t)`. Convert result sub-AG to `*AnalysisGraphState` via `artToGraphState`.

### Stub 9: `ViewSourceEdge(srcID, tgtID int) (string, int, error)` (line 296)
**Python:** `ivy_ui.py:289` — `act.lineno.filename, act.lineno.line`.
- Find transition, check `t.Op != nil && t.Op.HasLineno()`, return `loc.Filename, loc.Line`.

### Stub 10: `CoverNode(coveredID int) (bool, error)` (line 313)
**Python:** `ivy_ui.py:178` — `g.cover(covered_node, covering_node)`.
- Get mark, resolve both state IDs, call `ui.AG.Cover(covered, covering)`. Sync on success.

### Stub 11: `JoinNode(nodeID int) error` (line 323)
**Python:** `ivy_ui.py:190` — `g.join(node1, node2, self.get_alpha())`.
- Get mark, resolve both states, call `ui.AG.Join(state1, state2, abstractor)`. Sync.

### Stub 12: `TryConjecture(nodeID int, conjecture string) error` (line 333)
**Python:** `ivy_ui.py:400` — parses conjecture, dispatches to BMC or concept graph by mode.
- Parse conjecture via `logicparser.ToFormula`. Build `module.DualClauses`. If mode is induction/bounded: call `ui.AG.BMC(state, dual, nil, nil)`. Otherwise: add constraints to concept graph.

### Stub 13: `TryRememberedGraph(nodeID int, goalName string) error` (line 339)
**Python:** `ivy_ui.py:422` — loads saved graph copy, sets parent state.
- Look up `ui.RememberedGraphs[goalName]`, call `sg.Copy()` (exists at `graph_model.go:410`), set parent state and clauses.

### Stub 14: `BMC(nodeID int, errCond string, bound int) (*AnalysisGraphState, error)` (line 374)
**Python:** `ivy_ui.py:550` — `self.g.bmc(state, err_cond, bound=bound)`.
- Parse `errCond` via `logicparser.ToFormula`. Call `ui.AG.BMC(state, fmla, nil, &bound)`. Convert result to `*AnalysisGraphState`.

### Stub 15: `TryProperty(propText string) error` on `IvyUI` (line 408)
**Python:** `ivy_ui.py:578` — creates fresh AG with `top_alpha` initializer, runs BMC.
- Add `Mod *module.Module` field to `IvyUI`. Parse property, build dual clauses, create `art.NewAnalysisGraph(ui.Mod)`, add initial state with top_alpha, call `ag.BMC(ag.States[0], dual, nil, nil)`.

## Phase 3: Update DeleteNode and ArgNodeAction Dispatch

### 3A. Fix `DeleteNode` (`ui_main.go:302`)
Currently only removes from lightweight `G.States`. Must also call `ui.AG.Delete(state)` and sync. Clear `Mark` if deleted node was marked.

### 3B. Wire `ArgNodeAction` dispatch (`session.go:706`)
Replace placeholder case branches with real calls to `s.AGUI.xxx()`. Parse `nodeID` string → int. For edge actions, parse source/target from args. Each case delegates to the corresponding AGUI method and syncs.

## Phase 4: Abstractor Callback

The `AnalysisGraphUI` methods that need an abstractor (stubs 4, 5, 6, 7, 11) can't import `alpha` directly. Add a field `AlphaFn func() art.Abstractor` to `AnalysisGraphUI` that Session sets:
```go
s.AGUI.AlphaFn = func() art.Abstractor { return s.GetAlpha(s.AGUI.Mode) }
```
Then each stub calls `ui.AlphaFn()` to get the abstractor.

## Phase 5: Comprehensive Unit Tests

**New file:** `webui/ui_arg_test.go`

### Test Setup Helper
```go
func loadARGTestSession(t *testing.T) (*Session, *AnalysisGraphUI) {
    cfg := module.NewConfig()
    s := NewSession(cfg, "test-arg")
    s.LoadFileContent("test.ivy", []byte(ivySample))
    drainEvents(s)
    s.AG.AddInitialState(nil, nil)
    s.syncARGToGraph()
    return s, s.AGUI
}
```

### Test Cases (30+ tests)

| Test | Target | Asserts |
|------|--------|---------|
| `TestNodeExecuteCommands` | Stub 1 | Returns non-empty list for initial state with actions |
| `TestNodeExecuteCommandsNilAG` | Stub 1 error | Returns nil when AG is nil |
| `TestNodeExecuteCommandsSorted` | Stub 1 | Action labels are sorted alphabetically |
| `TestCheckLocalSafetySafe` | Stub 2 | Initial state returns `(true, "Node is safe")` |
| `TestCheckLocalSafetyNilAG` | Stub 2 error | Returns false + error message when no AG |
| `TestCheckBoundedSafetySafe` | Stub 3 | Initial state returns `(true, ...)` |
| `TestCheckBoundedSafetyNilAG` | Stub 3 error | Returns false + error when no AG |
| `TestFindExtensionOpen` | Stub 4 | Returns action label, no error; AG gains new state |
| `TestFindExtensionClosed` | Stub 4 | Returns "state N is closed" when fully explored |
| `TestFindExtensionNilAG` | Stub 4 error | Returns error when no AG |
| `TestExecuteAction` | Stub 5 | Succeeds; AG state count increases by 1; transition added |
| `TestExecuteActionBadName` | Stub 5 error | Returns error for unknown action name |
| `TestExecuteActionBadNode` | Stub 5 error | Returns error for invalid nodeID |
| `TestRecalculateAll` | Stub 6 | No panic; state count preserved after recalculation |
| `TestRecalculateAllNilAG` | Stub 6 | No panic when AG is nil |
| `TestRecalculateEdge` | Stub 7 | Execute action, then recalculate its edge — no error |
| `TestRecalculateEdgeBadIDs` | Stub 7 | No crash for nonexistent edge |
| `TestDecomposeEdge` | Stub 8 | Returns sub-graph AnalysisGraphState or nil (action-dependent) |
| `TestDecomposeEdgeBadIDs` | Stub 8 error | Returns error for nonexistent edge |
| `TestViewSourceEdge` | Stub 9 | Returns filename+line or "no source" error |
| `TestViewSourceEdgeNoAction` | Stub 9 error | Returns error for join edge (no Op) |
| `TestCoverNodeSuccess` | Stub 10 | Mark node 0, cover node 1 by it: returns true if subsumption holds |
| `TestCoverNodeNoMark` | Stub 10 error | Returns `(false, "no marked node")` |
| `TestCoverNodeBadID` | Stub 10 error | Returns error for invalid node ID |
| `TestJoinNode` | Stub 11 | Mark node, join with another — AG gains joined state |
| `TestJoinNodeNoMark` | Stub 11 error | Returns "no marked node" error |
| `TestTryConjectureEmpty` | Stub 12 error | Returns "no conjecture specified" |
| `TestTryConjectureBMC` | Stub 12 | In induction mode, runs BMC path |
| `TestTryRememberedGraphNotFound` | Stub 13 error | Returns "no remembered graph named X" |
| `TestTryRememberedGraphEmpty` | Stub 13 | Empty goalName returns nil (list mode) |
| `TestBMCBadNode` | Stub 14 error | Invalid nodeID returns error |
| `TestBMCUnreachable` | Stub 14 | Contradiction errCond returns "unreachable" error |
| `TestTryPropertyEmpty` | Stub 15 error | Returns "no property specified" |
| `TestDeleteNodeDelegates` | Fix 3A | After delete, AG.States shrinks; covering/transitions cleaned |
| `TestArgNodeActionCheckSafety` | Dispatch 3B | Session.ArgNodeAction("state_0","check_safety",nil) returns result |
| `TestArgNodeActionCover` | Dispatch 3B | Session.ArgNodeAction cover dispatches to AGUI.CoverNode |
| `TestStateByIDOutOfBounds` | Helper | Returns error for negative or too-large ID |
| `TestTransitionByEndpointsNotFound` | Helper | Returns error when no matching transition |
| `TestArtToGraphState` | Helper | Converts AG with states+transitions+covering correctly |

## Verification

Run tests via: `cd ~/ivy/goivy && make test`

After implementation, verify:
1. All 30+ new tests pass
2. Existing `webui/*_test.go` tests still pass (no regressions)
3. The `ivySample` fixture creates a session with actionable ARG states

## Critical Files

| File | Changes |
|------|---------|
| `webui/ui_main.go` | All 15 stub implementations, new fields, helpers |
| `webui/session.go` | Add AGUI field, wire in LoadFileContent, GetAlpha, update ArgNodeAction |
| `webui/ui_arg_test.go` | New test file with 30+ tests |
| `art/art.go` | Reference only — all methods already implemented |
| `alpha/alpha.go` | Reference only — Alpha/PredicateAlpha already exist |
| `webui/graph_model.go:410` | Reference only — Graph.Copy() already exists |
