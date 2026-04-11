# Fix LabeledFormula ID & metadata loss in L2S / Ranking / Temporal tactics

Created: 2026-04-11 23:50

## Context

Running `cd parser && go test -v -count=1 -run TestOrdLive` (logged in `~/ivy/goivy/log.red`) produces a Go-vs-Python xtrace divergence at index **254994**:

```
254993  go : XTRACE: l2s.l2sTacticInt goalConc result type=TemporalModels (isTemporalModels=True)
        py : XTRACE: l2s.l2sTacticInt goalConc result type=TemporalModels (isTemporalModels=True)

254994  go : XTRACE: ast.LF.__init__ id=2250 counter=2251          ← Go creates a NEW LF
        py : XTRACE: ast.LF.clone PRESERVE origid=599 counter=2250  ← Python preserves axiom id=599
```

The trigger is a substantive **semantic** bug, not just a trace mismatch:

- **Python** (`ivy_l2s.py:131-132`):
  ```python
  assumed_gprops = [x for x in prover.axioms if not x.explicit and x.temporal and isinstance(x.formula,lg.Globally)]
  model.asms.extend([p.clone([p.label,p.formula.args[0]]) for p in assumed_gprops])
  ```
  `p.clone(...)` invokes `LabeledFormula.clone` (`ivy_ast.py:649-666`) which **preserves the axiom's id** (decrements `lf_counter`, restores `res.id = self.id`) and **copies all metadata** (`temporal`, `explicit`, `definition`, `assumed`, `unprovable`).

- **Go** (`check/l2s.go:341`):
  ```go
  model.Asms = append(model.Asms, m.Cfg.AstCfg.NewLabeledFormula(ax.Label, g.Body))
  ```
  `NewLabeledFormula` (`ast/decl_ast.go:59-69`) allocates a brand-new id **and loses every metadata flag** — the resulting LF has `Temporal=nil`, `Explicit=false`, `IsDefinition=false`, `Assumed=false`, `Unprovable=false`, regardless of what the axiom held. This is the axiom that started as `Temporal=true`, which is now silently flipped to `nil`.

`ast.LabeledFormula.Clone` already exists (`ast/decl_ast.go:87-106`) and correctly mirrors Python: it delegates to `cloneInternal` (which copies all metadata — see `ast/decl_ast.go:110-126`), then decrements `cfg.LfCounter` and restores the original id, emitting `ast.LF.clone PRESERVE`. The fix is to call it at the sites where Python calls `lf.clone(args)`.

The same bug pattern appears in six locations in `check/l2s.go`, eight in `check/ranking.go`, and two in `temporal/temporal.go` (the latter also lose the label, passing `nil` where Python passes the axiom/goal label). All drop at least one of: id, label, or metadata flags.

## Scope

Primary (what unblocks `TestOrdLive`):
1. Fix the six `NewLabeledFormula` misuses in `check/l2s.go` that map to Python `lf.clone(args)` calls in `ivy_l2s.py`.

Expanded (bugs uncovered during investigation, same root cause):
2. Fix the eight analogous misuses in `check/ranking.go` (mirror sites in `ivy_ranking.py`).
3. Fix two misuses in `temporal/temporal.go` that *also* silently drop the label.
4. Add a `ModelPass` helper fix in `check/ranking.go:425,430`.

Out of scope: The three `NewLabeledFormulaFrom` + manual `lf.ID = src.ID` sites in `check/check.go:840`, `vmt/vmt.go:575`, `module/theory.go:277`. These are functionally correct (they preserve both id and metadata) — the agent flagged them as stylistically inferior, but they do the right thing and are not on the hot path. Leave alone.

## Plan

### Step 1 — primary l2s.go fixes (unblocks TestOrdLive divergence)

Critical file: `/Users/jaten/ivy/goivy/check/l2s.go`

Replace each `NewLabeledFormula(lf.Label, newFormula)` with `lf.Clone([]ast.Node{lf.Label, newFormula}).(*ast.LabeledFormula)`. Existing helper: `ast.LabeledFormula.Clone` at `ast/decl_ast.go:87-106`.

| Line | Current | Python ref | Replace with |
|---|---|---|---|
| 341 | `m.Cfg.AstCfg.NewLabeledFormula(ax.Label, g.Body)` | `ivy_l2s.py:132` `p.clone([p.label,p.formula.args[0]])` | `ax.Clone([]ast.Node{ax.Label, g.Body}).(*ast.LabeledFormula)` |
| 396 | `m.Cfg.AstCfg.NewLabeledFormula(inv.Label, labeled)` | `ivy_l2s.py:175` (`compile_with_goal_vocab` → `LF.clone`, then `label_temporal` → `LF.clone`) | `inv.Clone([]ast.Node{inv.Label, labeled}).(*ast.LabeledFormula)` |
| 440 | `l2sAcfg.NewLabeledFormula(inv.Label, Desugar(expr, proofLabel))` | `ivy_l2s.py:716` `list(map(desugar,invars))` → `expr.clone(...)` | `inv.Clone([]ast.Node{inv.Label, Desugar(expr, proofLabel)}).(*ast.LabeledFormula)` |
| 476 | `l2sAcfg.NewLabeledFormula(inv.Label, transform(inv.Formula.(lg.Expr)))` | `ivy_l2s.py:738` `model.invars = [transform(x) for x in model.invars]` → `LF.clone` via `replace_temporals_by_named_binder_g_ast` line 322 | `inv.Clone([]ast.Node{inv.Label, transform(inv.Formula.(lg.Expr))}).(*ast.LabeledFormula)` |
| 479 | `l2sAcfg.NewLabeledFormula(asm.Label, transform(asm.Formula.(lg.Expr)))` | `ivy_l2s.py:739` `model.asms = [transform(x) for x in model.asms]` | `asm.Clone([]ast.Node{asm.Label, transform(asm.Formula.(lg.Expr))}).(*ast.LabeledFormula)` |
| 495 | `l2sAcfg.NewLabeledFormula(lf.Label, transform(e))` | `ivy_l2s.py:744` `list_transform(prems,transform)` | `lf.Clone([]ast.Node{lf.Label, transform(e)}).(*ast.LabeledFormula)` |

Notes:
- For lines 476/479, the loop variable (`inv`, `asm`) is the slice element at index `i`; re-read it from the slice if the loop variable has been shadowed. Current loop uses `for i, inv := range model.Invars` so `inv` is the right source.
- For line 440, `inv` is `invars[i]` — same story.
- `Clone` returns `Node`; the `.(*ast.LabeledFormula)` assertion is required (package `ast` is already imported).

### Step 2 — ranking.go fixes (same bug pattern, different tactic)

Critical file: `/Users/jaten/ivy/goivy/check/ranking.go`

| Line | Current | Python ref | Replace with |
|---|---|---|---|
| 275 | `acfg.NewLabeledFormula(inv.Label, desugarFn(inv.Formula.(lg.Expr)))` | `ivy_ranking.py:505` `invars = list(map(desugar,invars))` | `inv.Clone([]ast.Node{inv.Label, desugarFn(inv.Formula.(lg.Expr))}).(*ast.LabeledFormula)` |
| 278 | `acfg.NewLabeledFormula(pc.Label, desugarFn(pc.Formula.(lg.Expr)))` | `ivy_ranking.py:505` (postconds) | `pc.Clone([]ast.Node{pc.Label, desugarFn(pc.Formula.(lg.Expr))}).(*ast.LabeledFormula)` |
| 299 | `acfg.NewLabeledFormula(inv.Label, transform(inv.Formula.(lg.Expr)))` | `ivy_ranking.py:526` | `inv.Clone([]ast.Node{inv.Label, transform(inv.Formula.(lg.Expr))}).(*ast.LabeledFormula)` |
| 302 | `acfg.NewLabeledFormula(asm.Label, transform(asm.Formula.(lg.Expr)))` | `ivy_ranking.py:527` | `asm.Clone([]ast.Node{asm.Label, transform(asm.Formula.(lg.Expr))}).(*ast.LabeledFormula)` |
| 312 | `acfg.NewLabeledFormula(inv.Label, transform(inv.Formula.(lg.Expr)))` | `ivy_ranking.py:526` (modPass invars) | `inv.Clone([]ast.Node{inv.Label, transform(inv.Formula.(lg.Expr))}).(*ast.LabeledFormula)` |
| 316 | `acfg.NewLabeledFormula(pc.Label, transform(pc.Formula.(lg.Expr)))` | `ivy_ranking.py:526` (modPass postconds) | `pc.Clone([]ast.Node{pc.Label, transform(pc.Formula.(lg.Expr))}).(*ast.LabeledFormula)` |
| 425 | `inv.Cfg.NewLabeledFormula(inv.Label, transform(inv.Formula.(lg.Expr)))` | `ivy_ranking.py:526` via `ModelPass` helper | `inv.Clone([]ast.Node{inv.Label, transform(inv.Formula.(lg.Expr))}).(*ast.LabeledFormula)` |
| 430 | `asm.Cfg.NewLabeledFormula(asm.Label, transform(asm.Formula.(lg.Expr)))` | `ivy_ranking.py:527` via `ModelPass` | `asm.Clone([]ast.Node{asm.Label, transform(asm.Formula.(lg.Expr))}).(*ast.LabeledFormula)` |

### Step 3 — temporal.go (double bug: id **and** label both wrong)

Critical file: `/Users/jaten/ivy/goivy/temporal/temporal.go`

Both sites currently pass `nil` as the label, even though the Python equivalents use the axiom/goal label.

| Line | Current | Python ref | Replace with |
|---|---|---|---|
| 468 | `pc.GetAstCfg().NewLabeledFormula(nil, invar)` | `ivy_temporal.py:275` `invars.append(ipr.clone_goal(goal,[],invar))` → `goal.clone_with_fresh_id([goal.label, invar])` | `goal.CloneWithFreshID([]ast.Node{goal.Label, invar}).(*ast.LabeledFormula)` — fresh id *is* intentional here (Python uses `clone_with_fresh_id`), but the label must be `goal.Label`, not `nil` |
| 594 | `pc.GetAstCfg().NewLabeledFormula(nil, g.Body)` | `ivy_temporal.py:378` `p.clone([p.label,p.formula.args[0]])` | `ax.Clone([]ast.Node{ax.Label, g.Body}).(*ast.LabeledFormula)` — preserve id + label + metadata |

For line 468: `goal.CloneWithFreshID` is already defined at `ast/decl_ast.go:128-138`. It corresponds exactly to Python's `clone_with_fresh_id`. This is the one case in this plan where fresh id is correct — but the label still has to come from the goal, not be `nil`.

For line 594: the surrounding loop already binds `for _, ax := range pc.GetAxioms()` and narrows to the `*lg.Globally` case via `g`; use `ax` (the outer `*ast.LabeledFormula`) for the clone receiver.

### Step 4 — verification

1. Build: `cd /Users/jaten/ivy/goivy && /usr/local/go/bin/go build ./check/... ./temporal/...` (should compile cleanly)
2. Unit check locally if there are focused tests for l2s clone: `cd check && /usr/local/go/bin/go test -run L2S -count=1 ./...`
3. Golden trace: `cd /Users/jaten/ivy/goivy && if [ -f log.red ]; then mv log.red log.red.prev; fi; cd parser && /usr/local/go/bin/go test -v -count=1 -run TestOrdLive &> ../log.red`
   - Expect: previous divergence at 254994 resolved. Check with `grep -n "divergence\|FAIL" log.red`.
   - If a *new* divergence appears further in the trace, inspect it — it may be another site (e.g. inside `l2s_auto.go`, `liveness.go`, or a compile/label_temporal path) that this plan did not cover. Record it and fix in the same fashion.
4. Make sure nothing regresses on the existing golden comparison suite: `make golden` (or the project's standard goivy-parity check) — the instruction in `log.parser.red` shows `make golden` is the driver.
5. Sanity-scan the test log for orphaned `__init__` emissions around l2s tactic entry (`grep -nA1 "l2s.l2sTacticInt goalConc" log.red`) — after the fix, the next event for both sides should be a matching `clone PRESERVE`.

## Critical files

- `/Users/jaten/ivy/goivy/check/l2s.go` — primary fix sites (lines 341, 396, 440, 476, 479, 495)
- `/Users/jaten/ivy/goivy/check/ranking.go` — same bug pattern (lines 275, 278, 299, 302, 312, 316, 425, 430)
- `/Users/jaten/ivy/goivy/temporal/temporal.go` — label + id both wrong (lines 468, 594)
- `/Users/jaten/ivy/goivy/ast/decl_ast.go` — home of `Clone` (line 87), `cloneInternal` (line 110), and `CloneWithFreshID` (line 128); **no edits needed here**, the helpers are already correct
- `/Users/jaten/ivy/pyivy/ivy/ivy/ivy_l2s.py` — Python reference (lines 132, 175, 716, 738, 739, 744)
- `/Users/jaten/ivy/pyivy/ivy/ivy/ivy_ranking.py` — Python reference (lines 73, 505, 526, 527)
- `/Users/jaten/ivy/pyivy/ivy/ivy/ivy_temporal.py` — Python reference (lines 275, 378)
- `/Users/jaten/ivy/pyivy/ivy/ivy/ivy_ast.py` — `LabeledFormula.clone` (lines 649-666) showing the id-preservation + metadata-copy contract

## Risks & considerations

1. **Metadata on freshly-created LFs elsewhere**: `NewLabeledFormula` remains the right constructor when the new LF is genuinely fresh (schema goals, auto-generated invariants, subgoals). The audit already separated ~15 legitimate uses. This plan only replaces the uses that mirror Python `clone`.

2. **Trace multiplicity at line 396**: Python emits *two* `clone PRESERVE` events per tactic invariant (one from `inv.compile()`, one from `label_temporal(..., proof_label)` at `ivy_logic.py:1773`). Go's `proof.CompileWithGoalVocab` (`proof/goal.go:382-402`) is a thin unwrapper — it returns the formula directly without calling any clone, so Go emits only one `clone PRESERVE` after the fix. Verification step 3 will reveal whether this causes a second divergence further in the trace; if so, the follow-up is to make `CompileWithGoalVocab` (or a new wrapper) call `lf.Clone` when given an LF, mirroring Python's `compile_expr_vocab_ext(lf, vocab)` → `lf.compile()` → `self.clone(...)` chain. Keep that as a Step-4b follow-up rather than pre-implementing blindly.

3. **Aliasing of `model.Asms` with `mod.AssumedInvs`**: `NormalProgramFromModule` (`temporal/temporal.go:261-267`) assigns `Asms: mod.AssumedInvs` by reference, so `append` at `l2s.go:341` may mutate the module's slice in-place if capacity allows. Python has the same aliasing behavior (`normal_program_from_module` in `ivy_temporal.py:217-223` and `NormalProgram.clone` in `ivy_temporal.py:179-183` both share the list), so this is matched and not a new bug. No action needed now, but mention this in the fix commit in case a downstream test later surfaces module-state corruption.

4. **`inv` loop variable shadowing**: in Go 1.22+, the loop variable is per-iteration, so the `Clone` calls are safe. Confirm `go.mod` targets >= 1.22 (it is — this project uses recent Go).

5. **Type assertion panic safety**: `lf.Clone(args).(*ast.LabeledFormula)` panics if `Clone` ever returns nil. Clone is defined to return `c` (a `*LabeledFormula`), not nil, so the assertion is safe. No need for the comma-ok form.
