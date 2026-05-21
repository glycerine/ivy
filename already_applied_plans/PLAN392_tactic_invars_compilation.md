# Plan: Fix golden test divergence at line 221057

## Context

The golden test (`make golden-all`) diverges at trace line 221057 during the `ranking` tactic applied to the `sys_live.live` goal. Go emits `temporal.NormalProgramFromModule EXIT` where Python emits `proof.GoalVocab ENTER label=sys_live.live` (a second GoalVocab call).

## Root Cause

In Python's `ivy_ranking.py:l2s_tactic_int` (lines 95–107), after entering the `WithSorts` context, the code:
1. Splits `proof.tactic_decls` into `tactic_invars` and `tactic_defns`
2. For each `tactic_defn`, calls `compile_definition_goal_vocab(defn, goal)` which internally calls `goal_vocab` — emitting the second `GoalVocab ENTER` at trace line 221057
3. Only then uses the model (`conc.model.clone([])`, which has no trace)

In Go's `RankingL2STactic` (`check_ranking.go:171`), after `WithSorts.Enter` (221056), the code jumps directly to `ExtractNormalProgram(m)` at line 200 — emitting `NormalProgramFromModule EXIT` at 221057 — with NO `CompileDefinitionGoalVocab` loop before it. The `tacticInvars` compilation (`compile_with_goal_vocab`, Python line 126) is also absent.

## Fix

**File**: `/Users/jaten/ivy/goivy/check_ranking.go`  
**Function**: `RankingL2STactic` (line 171)

### Step 1 — Add tactic-defns processing before `ExtractNormalProgram` (fixes trace line 221057)

After the `TacticLets` check (line 196) and before `m := cfg.Mod` / `ExtractNormalProgram(m)` (lines 199–200), insert:

```go
// Python ivy_ranking.py:95-107 — split TacticDecls, compile definitions
var tacticDefns []Node
var tacticInvars []*LabeledFormula
if cfg.Proof != nil {
    for _, d := range cfg.Proof.TacticDecls {
        if dd, ok := d.(*DerivedDecl); ok {
            tacticDefns = append(tacticDefns, dd)
        } else if lf, ok := d.(*LabeledFormula); ok {
            tacticInvars = append(tacticInvars, lf)
        }
    }
}
for _, defn := range tacticDefns {
    var compErr error
    goal, compErr = CompileDefinitionGoalVocab(m.Cfg.AstCfg, defn, goal, m)
    if compErr != nil {
        return nil, compErr
    }
}
```

Note: `m` is assigned just before this block (`m := cfg.Mod`), so that line stays in place.

### Step 2 — Add tactic-invars compilation (Python line 126)

After `var invars []*LabeledFormula` / `invars = append(invars, model.Invars...)` (line 252–253), insert:

```go
// Python ivy_ranking.py:126 — compile user-supplied tactic invariants
for _, inv := range tacticInvars {
    compiled := CompileWithGoalVocab(inv, goal, m)
    if compiled == nil {
        continue
    }
    labeled := LabelTemporalNode(compiled, proofLabel).(*LabeledFormula)
    invars = append(invars, labeled)
}
```

## Key functions (already exist, no new code needed)

- `CompileDefinitionGoalVocab` — `proof_goal.go:633` — calls `GoalVocab` internally
- `CompileWithGoalVocab` — `proof_goal.go:620` — calls `GoalVocab` internally
- `LabelTemporalNode` — already used elsewhere in check_ranking.go
- `ExtractNormalProgram` — `check_l2s_shared.go:1017` → `NormalProgramFromModule`

## Trace effect

After the fix, for each `tactic_defn` in `cfg.Proof.TacticDecls`:
- `proof.GoalVocab ENTER label=sys_live.live` fires at 221057 ✓ (matches Python)
- `temporal.NormalProgramFromModule EXIT` fires after the loop ✓ (matches Python)

## Verification

```
cd ~/ivy/goivy && make test
```

Then if clean:
```
cd ~/ivy/goivy && make golden-all 2>&1 | tail -30
```

The divergence at line 221057 should be resolved. Additional divergences may surface (the golden test reports only the first mismatch), to be addressed in follow-on changes.
