# Plan: WebUI Session Proof Stack Wiring

**Created:** 2026-05-02 16:12 UTC

## Context

The proof stack shows interactive proof goals in the web UI — conjectures to prove, applied tactics, remaining subgoals. The infrastructure exists on both ends but nothing connects them:

- **proof/ package**: `ProofChecker`, `ProofGoalStack`, `ProofManager` — all ported from Python
- **webui/ package**: `ProofStack`/`ProofGoal` rendering types, `RenderProofStack()`, HTTP endpoints (`/api/session/{id}/proof`), JS API (`getProofGraph()`) — all ready
- **Gap**: `Session.ProofStack` is always nil, `ProofStackData()` returns nil, `ProofGoalAction()` is stubbed, `proof.RegisterFactories` is never called before compilation

## Changes

### 1. Add `ProofMgr` field to Session — `webui/session.go:46`

Add `ProofMgr *proof.ProofManager` to the `Session` struct (after the `ProofStack` field). Add imports for `proof`, `tactics`, `temporal`.

### 2. Wire `proof.RegisterFactories` before compilation — `webui/session.go:114-118`

In `LoadFileContent`, between `mod := module.New()` and `compiler.IvyCompile(...)`, add:
```go
proofCfg := module.TacticNewConfig()
proof.RegisterFactories(mod.Cfg, proofCfg)
tactics.RegisterProofTactics(proofCfg)
temporal.RegisterTactics(proofCfg)
```
This enables the ProofChecker during phase6 `attach_proofs`. Without this, proofs are never checked during webui compilation. Matches `check.Start()` at `check/check.go:1309-1310`.

We skip `check.RegisterTactics` (which adds `mc`/`vmt` tactics) because importing `check` from `webui` would create an import cycle. The `proof`, `tactics`, and `temporal` packages have no such restriction.

**Reuses:** `proof.RegisterFactories` at `proof/register.go:12`, `tactics.RegisterProofTactics` at `tactics/ivy_tactics.go:508`, `temporal.RegisterTactics` at `temporal/temporal.go:857`.

### 3. Populate ProofManager from conjectures — `webui/session.go:197`

After Step 6 (Store compiled module), create ProofManager and populate initial goals from `mod.LabeledConjs`:
```go
s.ProofMgr = proof.NewProofManager()
for _, conj := range mod.LabeledConjs {
    if conj.Formula != nil {
        if fmla, ok := conj.Formula.(logic.Expr); ok {
            s.ProofMgr.Goals.Push(&proof.ProofGoal{Formula: fmla})
        }
    }
}
s.syncProofStack()
```
The `logic.Expr` type assertion guard handles `LabeledFormula.Formula` values that aren't `logic.Expr` (e.g., `*ast.TemporalModels`).

### 4. Add `syncProofStack()` — `webui/session.go` (new method, after `syncARGToGraph`)

Converts live `proof.ProofGoalStack` → lightweight `webui.ProofStack` for rendering. Follows the `syncARGToGraph()` pattern:
- Iterate `s.ProofMgr.Goals.Stack`
- For each `proof.ProofGoal`, create `webui.ProofGoal` with ID, Label, Info (from `logic.PrettyFmla`), ParentID
- Store result in `s.ProofStack`

**Reuses:** `logic.PrettyFmla` at `logic/pretty.go:14`.

### 5. Update `ProofStackData()` — `webui/session.go:670`

Change from `return s.ProofStack` to call `syncProofStack()` first, then return `s.ProofStack`.

### 6. Update `GetProof` — `webui/backend_go.go:507`

Change `RenderProofStack(sess.ProofStack)` to `RenderProofStack(sess.ProofStackData())` so it triggers sync before rendering.

### 7. Wire `ProofGoalAction` — `webui/session.go:849-876`

Replace the stub with real goal lookup and actions:
- **Goal lookup**: Parse goal ID from `"goal_N"` or bare `"N"` format, find in `ProofMgr.Goals.Stack`
- **"view"**: Return actual formula via `logic.PrettyFmla(goal.Formula)`
- **"refute"**: Remove goal from stack via `ProofMgr.Goals.Remove(goal)`, sync, emit `proof_updated`
- **"push"**: Push new sub-goal, sync, emit `proof_updated`
- **"pop"**: Pop top goal, sync, emit `proof_updated`
- **"apply_tactic"**: Status message noting infrastructure is wired but Z3 is required for full execution

### 8. Comprehensive tests — `webui/ui_test.go`

Add tests with `//go:build web` tag:

| Test | What it verifies |
|------|-----------------|
| `TestProofManagerCreatedAfterLoad` | `s.ProofMgr` non-nil after `LoadFileContent` |
| `TestProofStackPopulatedFromConjectures` | `ivySample` (1 conjecture) → ProofStack has 1 goal |
| `TestProofStackDataNonNilAfterLoad` | `ProofStackData()` returns non-nil `*ProofStack` |
| `TestProofStackRenderFromSync` | `RenderProofStack(s.ProofStackData())` produces CyElements with nodes |
| `TestProofGoalActionView` | "view" returns formula text in `result["info"]` |
| `TestProofGoalActionRefute` | "refute" removes goal, stack shrinks |
| `TestProofGoalActionPushPop` | Push/pop manipulate stack correctly |
| `TestEmptyModuleEmptyProofStack` | No conjectures → 0 goals, no crash |
| `TestProofGoalParentRelationships` | Pushed goals get correct parent IDs in synced ProofStack |
| `TestGetProofNonEmptyJSON` | Backend `GetProof` returns JSON with elements after load |

Tests reference `ivySample` from `backend_conform_test.go:17-26`.

## Verification

Run: `cd ~/ivy/goivy && make test-web`

Or targeted: `cd ~/ivy/goivy && XTRACE_OFF=1 go test -v ./webui -count=1 -tags xtrace_off,web -run TestProof`

## Files to modify

- `webui/session.go` — ProofMgr field, imports, RegisterFactories call, conjecture population, syncProofStack, ProofStackData, ProofGoalAction
- `webui/backend_go.go` — one-line change in GetProof (line 507)
- `webui/ui_test.go` — 10 new test functions
