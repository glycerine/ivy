# Plan: Post-merge cleanup of `check/` — kill `L2STraceHookData`, dead hooks, and merge artifacts

Created: 2026-04-11 14:00 UTC

## Context

The previous plan merged the `l2s/` and `ranking/` packages into `check/` as a pure structural refactor. With those packages now unified, several workarounds that existed only because of the old package boundary have become obsolete:

- The opaque `L2STraceHookData` struct + `HookKind*` enum + type-assert dispatch was a workaround for `ast`/`module` not being able to import a hook function type defined in the (then-separate) `l2s` package. Now `check` is a single package, the function type lives in it, and the consumer can store a closure directly.
- Two parallel sets of hook implementations exist: one operates on `*trace.TraceBase` and one on `*MatchHandler`. **Verified by grep**: nothing calls the `*trace.TraceBase` versions. They are dead code from an aborted earlier wiring attempt.
- `l2s_shared.go` exports a long list of `Export*` and Capitalized wrappers (`ExportL2sW`, `Forall`, `MakeAnd`, `ApplyNB`, `VarsToNodes`, `SortedSymbols`, `CollectAllNamedBinders`, `DedupeVarBodyPairs`, `StrPtr`, `ExportApplyL2sInit`, …) — these were exported only so the old separate `ranking/` package could call them across the boundary. Now nothing calls them.
- Several files carry stale comments left over from the merge ("Renamed from … during the l2s/+ranking/ → check/ merge", "Phase 7 follow-up", "without a direct import cycle", "Moved from l2s/hooks.go to break the import cycle"). These are noise.
- `MatchHandler.IsCti` is typed `interface{}` but is only ever assigned a `*module.Clauses`. It's also only written, never read — set in anticipation of the GUI consumer (`gui_art`), which is currently a stub.

This plan does the semantic cleanup that the merge unblocked. Scope was widened during investigation to also remove all dead helper exports and stale comments uncovered along the way.

## Investigation findings

### F1. The `*trace.TraceBase`-flavored hook chain is entirely dead

Grep results (full repo, command-by-command):

| Symbol | Definition | Callers |
|---|---|---|
| `L2STraceHook` | `check/l2s_hooks.go:18` | **none** (one comment-only mention at `check.go:602`) |
| `L2SAutoHook` | `check/l2s_hooks.go:66` | **none** |
| `L2SAutoHookConfig` | `check/l2s_hooks.go:57` | only its own `L2SAutoHook` (dead) |
| `L2SRenamingHook` | `check/l2s_hooks.go:47` | only `L2SAutoHook` (dead) and `RankingAutoHook` (dead, see F2) |
| `RankingAutoHook` | `check/ranking_hooks.go:24` | **none** |
| `RankingAutoHookConfig` | `check/ranking_hooks.go:15` | only `RankingAutoHook` (dead) |
| `diagnoseRankingFailure` | `check/ranking_hooks.go:56` | only `RankingAutoHook` (dead) |
| `lfName` (ranking_hooks.go) | `check/ranking_hooks.go:124` | only `RankingAutoHook` (dead); duplicate of `l2sLfName` |

The ONLY hook code actually executed is the `MatchHandler`-flavored pair `applyL2SRenamingToHandler` and `applyL2SAutoDiagnostics`, called from the dispatch at `check/check.go:598-617`.

### F2. Entire `check/ranking_hooks.go` is dead

Confirmed by grep — no function or type from `ranking_hooks.go` has any caller. The whole file (133 lines) is unreachable.

### F3. The `L2STraceHookData` opaque dispatch is unnecessary

Trace data flow:

1. **Construction site** — `check/l2s.go:733-750` (the only place):
   ```go
   if strings.HasPrefix(tacticName, "l2s_auto5") {
       result[0].TraceHook = &L2STraceHookData{Kind: HookKindAuto, Subs: …, Tasks: …, Triggers: …}
   } else if strings.HasPrefix(tacticName, "l2s_auto") {
       result[0].TraceHook = &L2STraceHookData{Kind: HookKindRenaming, Subs: …}
   } else if tacticName == "l2s_full" {
       result[0].TraceHook = &L2STraceHookData{Kind: HookKindFull}
   }
   ```

2. **Propagation** — `check/isolate_check.go:806-808` and `:862-864`:
   ```go
   if goal.TraceHook != nil { fakeMod.TraceHook = goal.TraceHook }
   ```
   (and `ast/decl_ast.go:124` copies it during `LabeledFormula.cloneInternal`)

3. **Consumption** — `check/check.go:598-617` (the only place):
   ```go
   if mod.TraceHook != nil {
       if data, ok := mod.TraceHook.(*L2STraceHookData); ok {
           switch data.Kind {
           case HookKindFull:                          // no-op (with TODO)
           case HookKindRenaming, HookKindAuto:
               if data.Subs != nil {
                   applyL2SRenamingToHandler(handler, data.Subs)
               }
               if data.Kind == HookKindAuto && data.Tasks != nil {
                   applyL2SAutoDiagnostics(handler, ffcs, data)
               }
           }
       }
   }
   ```

The struct + enum + switch can be replaced with a function type whose value captures `Subs`, `Tasks`, `Triggers` by closure. The `HookKindFull` branch is already a documented no-op for `MatchHandler` and stays a no-op.

Constraint reminder: `ast` and `module` cannot import `check` (would create a cycle). So the field on `LabeledFormula` and `Module` must remain typed as `interface{}` — but the **value** stored in it can be a strongly-typed `check.TraceHookFn` that the dispatch site recovers via a single type-assertion. This eliminates the enum, the struct, the switch, and the dispatch ladder.

### F4. Dead exported helper wrappers in `l2s_shared.go`

Verified by grep. Each of these is defined at the cited line in `check/l2s_shared.go`, exists only as a one-line wrapper around a lowercase internal helper, and has zero callers anywhere in the repo:

| Wrapper | Line | Wraps | Callers |
|---|---|---|---|
| `TransformAction` | 75 | `transformAction` | **2** in `ranking.go:352, 356` (alive — see below) |
| `CollectAllNamedBinders` | 80 | `collectAllNamedBinders` | **0** |
| `SortedSymbols` | 85 | `sortedSymbols` | **0** |
| `ApplyNB` | 90 | `applyNB` | **0** |
| `VarsToNodes` | 95 | `varsToNodes` | **0** |
| `Forall` | 100 | `forall` | **0** |
| `MakeAnd` | 105 | `makeAnd` | **0** |
| `SetLineno` | 110 | `setLineno` | **2** in `ranking.go:410, 432` (alive — see below) |
| `ExportL2sW` | 115 | `l2sW` | **0** |
| `ExportL2sS` | 120 | `l2sS` | **0** |
| `ExportL2sG` | 125 | `l2sG` | **0** |
| `ExportOldL2sG` | 130 | `oldL2sG` | **0** |
| `ExportL2sInit` | 135 | `l2sInit` | **0** |
| `ExportL2sWhen` | 140 | `l2sWhen` | **0** |
| `ExportL2sOld` | 145 | `l2sOld` | **0** |
| `StrPtr` | 150 | `strPtr` | **0** |
| `ExportApplyL2sInit` | 155 | `applyL2sInit` | **0** |
| `DedupeVarBodyPairs` | 160 | `dedupeVarBodyPairs` | **0** |

Two wrappers (`TransformAction`, `SetLineno`) have callers in the merged `ranking.go`. Those callers can be redirected to the lowercase internal helpers (now in the same package), letting us delete the wrappers too.

### F5. `MatchHandler.IsCti interface{}` is mistyped and only-written

- Defined at `check/helpers.go:148` as `interface{}`.
- Set exactly once at `check/check.go:623` to `module.FormulaToClauses(...)` (a `*module.Clauses`).
- **Never read** (verified by grep `\.IsCti` — only the assignment matches).

It's set in anticipation of the `gui_art` consumer that's currently a stub at `phase7.go:18` (`GuiArt`). The Python source uses it. We should keep the field but type it correctly so the future GUI implementation gets a typed value. `module.Clauses` is in `module/`, which `check/` already imports — no cycle.

### F6. Stale post-merge comments

| File | Line | Comment | Why stale |
|---|---|---|---|
| `check/check.go` | 35 | "without a direct import cycle" | the cycle was fixed long ago |
| `check/check.go` | 588-597 | "Phase 7 follow-up", "MatchHandler<->TraceBase bridge" | bridge is being eliminated, not bridged |
| `check/check.go` | 602-603 | "L2STraceHook marks the loop start … tracked as a TODO" | L2STraceHook (dead) is being deleted; the TODO won't apply |
| `check/l2s_hooks.go` | 198-202 | "MatchHandler counterpart of L2SRenamingHook" | L2SRenamingHook (dead) is being deleted |
| `check/l2s_hooks.go` | 220-224 | "MatchHandler counterpart of L2SAutoHook" | L2SAutoHook (dead) is being deleted |
| `check/l2s_hooks.go` | 245-247 | "Phase 7 follow-up" bridge | bridge is being eliminated |
| `check/l2s.go` | 2 | "(formerly the l2s/ package, merged into check)" | git history is the home for "formerly" notes |
| `check/l2s.go` | 729-732 | "type-asserts it to *l2s.L2STraceHookData" | hook data type is being deleted |
| `check/l2s.go` | 776-777 | "(cloneGoalWithASTConc was deleted …)" | drive-by stale "removed" comment |
| `check/l2s.go` | 1003-1006 | "Renamed from RegisterTactics to avoid colliding … during the l2s/ → check/ package merge" | rename is permanent; collision is gone |
| `check/ranking.go` | 1-3 | "(formerly the ranking/ package, merged into check)" | git history covers this |
| `check/ranking.go` | 168-171 | "strPtr is provided by l2s.go (identical implementation); the ranking definition was deleted during the l2s/+ranking/ → check/ merge" | post-merge archaeology |
| `check/ranking.go` | 218-222 | "Renamed from L2STactic during the l2s/+ranking/ → check/ merge" | rename context |
| `check/ranking.go` | 728-734 | "Renamed from makeAnd during the l2s/+ranking/ → check/ merge" | rename context |
| `check/ranking.go` | 758-762 | "Renamed from Desugar during the l2s/+ranking/ → check/ merge" | rename context |
| `check/ranking_test.go` | 11-14 | "Renamed from testAstCfg during the l2s/+ranking/ → check/ merge" | rename context |
| `check/l2s_shared.go` | 1-2 | "shared between the l2s tactic and the ranking tactic" | both tactics now live in the same package |
| `check/hook_data.go` | 1-7 | "without an import cycle" framing | file is being deleted entirely |
| `check/phase7.go` | 1-2 | "Phase 7 helper functions" | "Phase 7" is internal jargon — describe what it actually does |
| `temporal/temporal.go` | 605, 634 | "(cloneGoalWithASTConc was deleted …)" | same drive-by clean as l2s.go:776 |

### F7. `l2sLfName` and `lfName` are duplicates

`check/l2s_hooks.go:178-186` defines `l2sLfName(lf)` and `check/ranking_hooks.go:124-132` defines `lfName(lf)`. Bodies are identical. After deleting `ranking_hooks.go`, the duplication disappears for free; no extra step needed.

## Approach

Six sequenced edits, each compileable on its own. After each pass: `go build ./check/...`. Final verification at the end runs the full test sweep.

### Pass A — Define `TraceHookFn` and convert the construction site

1. Add a new type in `check/l2s_hooks.go` near the top (or in `helpers.go` next to `MatchHandler`):
   ```go
   // TraceHookFn is the function type stored in ast.LabeledFormula.TraceHook
   // (and propagated to module.Module.TraceHook) by L2S tactics. It is invoked
   // by the trace formatter in check.go after constructing a MatchHandler from
   // the failing checker. Mirrors Python's goal.trace_hook closure
   // (ivy_l2s.py:1311-1313).
   type TraceHookFn func(handler *MatchHandler, fcs []Checker)
   ```

2. Rewrite `check/l2s.go:733-750` to build closures instead of struct literals:
   ```go
   if len(result) > 0 && result[0] != nil {
       subs := cfg.Subs
       tasks := cfg.Tasks
       triggers := cfg.Triggers
       switch {
       case strings.HasPrefix(tacticName, "l2s_auto5"):
           result[0].TraceHook = TraceHookFn(func(handler *MatchHandler, fcs []Checker) {
               if subs != nil {
                   applyRenamingToHandler(handler, subs)
               }
               applyAutoDiagnosticsToHandler(handler, fcs, tasks, triggers)
           })
       case strings.HasPrefix(tacticName, "l2s_auto"):
           result[0].TraceHook = TraceHookFn(func(handler *MatchHandler, fcs []Checker) {
               if subs != nil {
                   applyRenamingToHandler(handler, subs)
               }
           })
       case tacticName == "l2s_full":
           // No MatchHandler-side hook for l2s_full — Python's trace_hook
           // marks loop_start on a trace.TraceBase, which we don't build
           // here. Leave TraceHook nil.
       }
   }
   ```
   The renamed helpers (`applyRenamingToHandler`, `applyAutoDiagnosticsToHandler`) are introduced atomically in Pass B; the simplest sequencing is to do Passes A and B as one combined edit so the build passes after each save.

3. Rewrite `check/check.go:598-617` to a single closure call:
   ```go
   // Apply the trace hook attached by L2S tactics, if any. The hook is
   // stored as TraceHookFn (boxed in interface{} on Module/LabeledFormula
   // because ast and module cannot import check).
   if mod.TraceHook != nil {
       if hook, ok := mod.TraceHook.(TraceHookFn); ok {
           hook(handler, ffcs)
       }
   }
   ```
   The pre-existing comment block at lines 588-597 is replaced with the shorter comment above.

### Pass B — Delete dead hook code and rename helpers

In `check/l2s_hooks.go`:
- Delete `L2STraceHook` (lines 15-42).
- Delete `L2SRenamingHook` (lines 44-54).
- Delete `L2SAutoHookConfig` (lines 56-61).
- Delete `L2SAutoHook` (lines 63-95) including its body. Note: the body assigns `tr.PP = L2SGToGlobally`. After this delete, re-grep `L2SGToGlobally` — if it has zero remaining callers, also delete the function it refers to.
- Rename `applyL2SRenamingToHandler` → `applyRenamingToHandler` (the `L2S` prefix was disambiguating from a deleted function).
- Rewrite `applyL2SAutoDiagnostics` → `applyAutoDiagnosticsToHandler` with a flatter signature that takes the maps directly instead of a `*L2STraceHookData`:
  ```go
  func applyAutoDiagnosticsToHandler(
      handler *MatchHandler,
      fcs []Checker,
      tasks map[string]map[string]*lg.Eq,
      triggers map[string]map[string]*lg.Eq,
  ) { … }
  ```
- The body of `applyAutoDiagnosticsToHandler` retains the failed-checker scan and call to `l2sDiagnoseAutoFailure(name, tasks, triggers, lf, nil)`. Rename `l2sDiagnoseAutoFailure` → `diagnoseAutoFailure` (no other `diagnoseAutoFailure` will exist after Pass C deletes the ranking duplicate).
- Rename `l2sLfName` → `lfName` (no longer needs disambiguation because `ranking_hooks.go`'s duplicate is being deleted in Pass C).
- Re-grep `L2STemporalAndL2SFilter` — it was set on `tr.HiddenSymbols` in the dead branches. If it has zero remaining callers after the delete, delete it too.

In `check/check.go`:
- Comment block at lines 588-597 already replaced in Pass A.
- TODO comment at line 603 already removed by the dispatch rewrite.

### Pass C — Delete `check/ranking_hooks.go` entirely

The whole file is dead. Delete it.

After deletion, `go build ./check/...` should still pass since no caller in the package referenced any of `RankingAutoHook`, `RankingAutoHookConfig`, `diagnoseRankingFailure`, or `lfName` from this file.

### Pass D — Delete dead exports in `check/l2s_shared.go` and demote the two live ones

1. Redirect the two live wrapper callers, then delete those wrappers:
   - `check/ranking.go:352` and `:356`: change `TransformAction(...)` → `transformAction(...)`
   - `check/ranking.go:410` and `:432`: change `SetLineno(...)` → `setLineno(...)`

2. Delete the unused exported wrappers (zero-caller list from F4) plus the two we just orphaned:
   - `TransformAction` (line 75)
   - `CollectAllNamedBinders` (line 80)
   - `SortedSymbols` (line 85)
   - `ApplyNB` (line 90)
   - `VarsToNodes` (line 95)
   - `Forall` (line 100)
   - `MakeAnd` (line 105)
   - `SetLineno` (line 110)
   - `ExportL2sW` (line 115)
   - `ExportL2sS` (line 120)
   - `ExportL2sG` (line 125)
   - `ExportOldL2sG` (line 130)
   - `ExportL2sInit` (line 135)
   - `ExportL2sWhen` (line 140)
   - `ExportL2sOld` (line 145)
   - `StrPtr` (line 150)
   - `ExportApplyL2sInit` (line 155)
   - `DedupeVarBodyPairs` (line 160)

3. Update the doc comment at the top of `l2s_shared.go` (lines 1-6) to drop the "shared between the l2s tactic and the ranking tactic" framing. New text:
   ```go
   // l2s_shared.go contains the L2S instrumentation pipeline steps used by
   // the L2S and ranking tactics. Each SharedStep* function corresponds to a
   // numbered step in l2sTacticInt. The InstrumentationConfig struct carries
   // all state between steps.
   ```

### Pass E — Delete `check/hook_data.go` and clean up stale comments

1. Delete `check/hook_data.go` entirely. (`L2STraceHookKind`, `HookKindNone/Full/Renaming/Auto`, `L2STraceHookData` — all obsolete after Pass A.)
2. Verify no callers remain via grep on each symbol. Build.
3. Strip all stale post-merge comments listed in F6:
   - `check/check.go:35` — remove "without a direct import cycle" framing.
   - `check/l2s.go:1` — drop "(formerly the l2s/ package, merged into check)" parenthetical.
   - `check/l2s.go:776-777` — delete the "(cloneGoalWithASTConc was deleted …)" comment.
   - Same removal at `temporal/temporal.go:605` and `:634`.
   - `check/l2s.go:1003-1006` — replace "Renamed from RegisterTactics …" with a one-line summary of what `RegisterL2STactics` does.
   - `check/ranking.go:1-3` — drop "(formerly the ranking/ package, merged into check)" parenthetical.
   - `check/ranking.go:168-171` — delete the "strPtr is provided by l2s.go (identical implementation); the ranking definition was deleted …" block.
   - `check/ranking.go:218-222`, `:728-734`, `:758-762` — delete the "Renamed from … during the l2s/+ranking/ → check/ merge" sentences from the doc comments on `RankingL2STactic`, `rankingMakeAnd`, and `RankingDesugar`. Keep the rest of each doc comment.
   - `check/ranking_test.go:11-14` — delete the "Renamed from testAstCfg during the l2s/+ranking/ → check/ merge" sentence.
   - `check/phase7.go:1-2` — replace "Phase 7 helper functions for the check package" with a description of what the file actually contains (currently `GuiArt`, the analysis-graph GUI stub).

4. Update doc comments on the `interface{}` fields that previously mentioned `L2STraceHookData`:

   `ast/decl_ast.go:39-43` becomes:
   ```go
   // TraceHook is a diagnostic hook closure attached by tactics (e.g. l2s).
   // The concrete type is check.TraceHookFn; the field is interface{} only
   // because ast cannot import check (cycle). Mirrors Python's
   // dynamically-attached lf.trace_hook attribute (ivy_l2s.py:88, 1311, 1313).
   TraceHook interface{}
   ```

   `module/module.go:142-146` — analogous update: replace "opaque diagnostic hook" with "a check.TraceHookFn closure" and explain the interface{} is for the import-cycle constraint only.

### Pass F — Type fix for `MatchHandler.IsCti`

1. In `check/helpers.go:148`, change
   ```go
   IsCti interface{}
   ```
   to
   ```go
   // IsCti holds the failing conjecture clauses, if any. Set by the trace
   // formatter and consumed by the GUI analysis graph (currently a stub in
   // phase7.go GuiArt). Mirrors Python handler.is_cti (ivy_check.py:401-403).
   IsCti *module.Clauses
   ```
2. The single setter at `check/check.go:623` already produces a `*module.Clauses` from `module.FormulaToClauses`; no change needed there.
3. Verify build.

## Files to modify

| File | Change | Pass |
|---|---|---|
| `check/l2s_hooks.go` | Add `TraceHookFn` type; delete dead hooks `L2STraceHook`, `L2SRenamingHook`, `L2SAutoHook`, `L2SAutoHookConfig`; rename `applyL2SRenamingToHandler`→`applyRenamingToHandler`; rewrite signature of `applyL2SAutoDiagnostics`→`applyAutoDiagnosticsToHandler`; rename `l2sDiagnoseAutoFailure`→`diagnoseAutoFailure`; rename `l2sLfName`→`lfName`; remove stale "MatchHandler counterpart of …" and "Phase 7" comments. Optionally remove `L2STemporalAndL2SFilter` and `L2SGToGlobally` if their last callers are deleted (verify post-edit). | A, B |
| `check/l2s.go` | Replace struct-construction at lines 733-750 with closure construction; drop top-of-file "(formerly the l2s/ package …)" parenthetical; delete the stale "(cloneGoalWithASTConc was deleted …)" comment at line 776; simplify `RegisterL2STactics` doc comment at lines 1003-1006. | A, E |
| `check/check.go` | Replace dispatch at lines 588-617 with single closure call; remove "without a direct import cycle" comment at line 35. | A, E |
| `check/ranking_hooks.go` | **Delete file.** | C |
| `check/hook_data.go` | **Delete file.** | E |
| `check/l2s_shared.go` | Delete dead exported wrappers (18 functions, lines 75-162); update top-of-file comment to drop "shared between the l2s tactic and the ranking tactic" framing. | D |
| `check/ranking.go` | Update the 4 wrapper call sites: `TransformAction`→`transformAction` (lines 352, 356) and `SetLineno`→`setLineno` (lines 410, 432); drop "(formerly the ranking/ package …)" parenthetical at lines 1-3; delete "strPtr is provided by l2s.go" block at 168-171; remove "Renamed from … during the merge" sentences from the doc comments on `RankingL2STactic` (218-222), `rankingMakeAnd` (728-734), and `RankingDesugar` (758-762). | D, E |
| `check/ranking_test.go` | Drop "Renamed from testAstCfg during the merge" comment at lines 11-14. | E |
| `check/phase7.go` | Replace "Phase 7 helper functions" docstring with a description of what the file actually contains. | E |
| `check/helpers.go` | Change `MatchHandler.IsCti interface{}` → `*module.Clauses`; add explanatory comment. | F |
| `ast/decl_ast.go` | Update doc comment on `LabeledFormula.TraceHook` to reference `check.TraceHookFn`. (Field type stays `interface{}`.) | E |
| `module/module.go` | Update doc comment on `Module.TraceHook` to reference `check.TraceHookFn`. (Field type stays `interface{}`.) | E |
| `temporal/temporal.go` | Delete the stale "(cloneGoalWithASTConc was deleted …)" comments at lines 605 and 634. | E |

### Summary counts

- **Files deleted**: 2 (`check/ranking_hooks.go`, `check/hook_data.go`)
- **Functions/types deleted**: ~22 (4 dead hooks + 1 dead config type in `l2s_hooks.go`; 18 dead exports in `l2s_shared.go`; the entire `ranking_hooks.go` content; possibly `L2STemporalAndL2SFilter` and `L2SGToGlobally`)
- **Functions renamed**: 4 (`applyL2SRenamingToHandler`, `applyL2SAutoDiagnostics`, `l2sDiagnoseAutoFailure`, `l2sLfName`)
- **Stale comments removed/rewritten**: ~14 sites
- **Field type fixes**: 1 (`MatchHandler.IsCti`)

## Verification

After each pass, rebuild just the affected package:

```sh
cd ~/ivy/goivy && go build ./check/...
```

Final verification (run after all passes):

1. **Build everything clean**:
   ```sh
   cd ~/ivy/goivy && go build ./...
   ```

2. **Vet the affected package**:
   ```sh
   cd ~/ivy/goivy && go vet ./check/...
   ```

3. **Run tests for affected and downstream packages** (same set as the previous merge plan):
   ```sh
   cd ~/ivy/goivy && go test ./check/... ./compiler/... ./proof/... ./module/... ./tactics/... ./temporal/...
   ```

4. **Stale-symbol grep** — must come up empty:
   ```sh
   grep -rn "L2STraceHookData\|L2STraceHookKind\|HookKindFull\|HookKindRenaming\|HookKindAuto\|HookKindNone\|L2STraceHook\b\|L2SRenamingHook\|L2SAutoHook\|L2SAutoHookConfig\|RankingAutoHook\|RankingAutoHookConfig" ~/ivy/goivy/check ~/ivy/goivy/cmd ~/ivy/goivy/ast ~/ivy/goivy/module
   ```

5. **Stale-comment grep** — must come up empty:
   ```sh
   grep -rn "during the l2s/+ranking/ → check/ merge\|Phase 7\|MatchHandler<->TraceBase bridge\|cloneGoalWithASTConc was deleted\|formerly the l2s/ package\|formerly the ranking/ package" ~/ivy/goivy
   ```

6. **`cmd/goivy_check` build**:
   ```sh
   cd ~/ivy/goivy && go build ./cmd/goivy_check/...
   ```

7. **Run an end-to-end Ivy proof** that exercises the trace_hook path if one exists in the test corpus, to confirm the closure refactor preserves the diagnostic behavior the original dispatch produced. Otherwise rely on the unit tests in step 3.

## Out of scope

- **Merging the duplicate `L2sW` / `L2sG` / `L2sS` / `L2sInit` / `L2sWhen` / `L2sOld` / `OldL2sG` signatures** between `check/l2s.go` (lowercase, `*string` environ) and `check/ranking.go` (uppercase, `string` environ). They have different signatures and both are alive; collapsing them needs a separate audit. **Not done here.**
- **Implementing `GuiArt`** in `phase7.go`. The `IsCti` field type fix is preparation for it; the full implementation is its own project.
- **Implementing the BMC flag** wired up in `check/check.go:1266` (the `_ = someBounded // TODO: wire BMC flag` line). Standalone task.
- **Refactoring `check/check.go`'s mega-functions** or reorganizing files. Pure cleanup pass.
- **Side-map approach for TraceHook** (storing closures in a `module.Module` map keyed by LF identity instead of an `interface{}` field). The closure-via-`interface{}` approach in this plan is a strict improvement over the current enum-dispatched struct and doesn't pre-empt the side-map approach if it's wanted later.

## Risks and rollback

- **Risk: deleting `L2SGToGlobally` or `L2STemporalAndL2SFilter` accidentally.** Both are referenced *only* inside the dead branches we're removing. After removing the dead hook functions, re-grep each name; if zero callers remain, also delete the function. If at least one caller remains (e.g., a test path I missed), keep it.
- **Risk: a comment-only mention of a deleted symbol leaves a dangling reference.** Mitigated by step 4 of verification (stale-symbol grep).
- **Risk: the closure approach changes the timing of `Tasks`/`Triggers`/`Subs` capture.** The current code stores the maps at goal-construction time and reads them at trace-formatting time. The closure approach captures by reference at the same point (goal construction), so the values are the same.
- **Risk: `Module.TraceHook` propagation through `isolate_check.go` still works.** The propagation copies the field value as `interface{}`; the closure is just a different concrete type stored in that interface, so the copy still works. No code change needed in `isolate_check.go`.
- **Risk: `LabeledFormula.cloneInternal` copies the closure shallowly.** Function values are reference types; the same closure ends up referenced in both LFs. This matches the previous behavior (the same `*L2STraceHookData` pointer was shared). Acceptable.
- **Risk: redirecting `TransformAction`/`SetLineno` callers to lowercase might change semantics.** Both wrappers are pure pass-throughs (verified by reading the wrapper bodies — single return statements). Identical semantics.
- **Rollback**: each Pass is independently revertable by undoing the file edits since the previous green build. Pass A and Pass B are most safely done as one combined edit because the new closures reference helpers that Pass B renames.

## Implementation checklist

### Pass A — Closures and dispatch (combine with Pass B for atomicity)
- [ ] Add `TraceHookFn` type (in `check/l2s_hooks.go` near top, or `check/helpers.go` next to `MatchHandler`).
- [ ] In `check/l2s.go:733-750`: replace 3 struct constructions with 2 closure constructions (`l2s_auto5`, `l2s_auto*`); leave `l2s_full` branch as a no-op (no closure).
- [ ] In `check/check.go:598-617`: replace the type-assert + switch with a single `if hook, ok := mod.TraceHook.(TraceHookFn); ok { hook(handler, ffcs) }`.
- [ ] In `check/check.go`: rewrite the comment block at lines 588-597 to a 2-line summary.

### Pass B — Delete dead hooks, rename helpers
- [ ] Delete `L2STraceHook`, `L2SRenamingHook`, `L2SAutoHook`, `L2SAutoHookConfig` from `check/l2s_hooks.go`.
- [ ] Rename `applyL2SRenamingToHandler` → `applyRenamingToHandler` and update Pass-A construction site.
- [ ] Rewrite `applyL2SAutoDiagnostics` → `applyAutoDiagnosticsToHandler` with flat `(handler, fcs, tasks, triggers)` signature; delete the `*L2STraceHookData` parameter.
- [ ] Update Pass-A construction site to call `applyAutoDiagnosticsToHandler`.
- [ ] Rename `l2sDiagnoseAutoFailure` → `diagnoseAutoFailure`.
- [ ] Rename `l2sLfName` → `lfName`.
- [ ] Re-grep `L2SGToGlobally` and `L2STemporalAndL2SFilter`; delete each if zero callers remain.
- [ ] Build: `go build ./check/...` clean.

### Pass C — Delete `ranking_hooks.go`
- [ ] Delete `check/ranking_hooks.go`.
- [ ] Build: `go build ./check/...` clean.

### Pass D — Delete dead exports in `l2s_shared.go`
- [ ] Update 4 callers in `check/ranking.go`: `TransformAction`→`transformAction` (lines 352, 356) and `SetLineno`→`setLineno` (lines 410, 432).
- [ ] Delete the 18 wrapper functions in `check/l2s_shared.go` (lines 75-162). The 16 zero-caller ones plus the 2 newly orphaned (`TransformAction`, `SetLineno`).
- [ ] Update doc comment at top of `l2s_shared.go` (lines 1-6).
- [ ] Build: `go build ./check/...` clean.

### Pass E — Delete `hook_data.go` and stale comments
- [ ] Delete `check/hook_data.go`.
- [ ] Update doc comments on `ast/decl_ast.go:39-43` and `module/module.go:142-146` to reference `check.TraceHookFn`.
- [ ] Strip stale comments per F6 in: `check/check.go:35`, `check/l2s.go:1, 776, 1003-1006`, `check/ranking.go:1-3, 168-171, 218-222, 728-734, 758-762`, `check/ranking_test.go:11-14`, `check/phase7.go:1-2`, `temporal/temporal.go:605, 634`.
- [ ] Build: `go build ./...` clean.

### Pass F — Type fix for `MatchHandler.IsCti`
- [ ] In `check/helpers.go:148`: change `IsCti interface{}` → `IsCti *module.Clauses`.
- [ ] Add doc comment explaining the future GUI consumer.
- [ ] Build: `go build ./...` clean.

### Final verification
- [ ] `go build ./...` clean.
- [ ] `go vet ./check/...` clean.
- [ ] `go test ./check/... ./compiler/... ./proof/... ./module/... ./tactics/... ./temporal/...` green.
- [ ] Stale-symbol grep (verification step 4) returns no matches.
- [ ] Stale-comment grep (verification step 5) returns no matches.
- [ ] `go build ./cmd/goivy_check/...` clean.

## Critical files

- `/Users/jaten/ivy/goivy/check/l2s.go` — closure construction site (Pass A)
- `/Users/jaten/ivy/goivy/check/l2s_hooks.go` — `TraceHookFn` type, delete-dead-hooks, rename helpers (Passes A, B)
- `/Users/jaten/ivy/goivy/check/check.go` — dispatch rewrite (Pass A), comment cleanup (Pass E)
- `/Users/jaten/ivy/goivy/check/ranking_hooks.go` — **delete entirely** (Pass C)
- `/Users/jaten/ivy/goivy/check/hook_data.go` — **delete entirely** (Pass E)
- `/Users/jaten/ivy/goivy/check/l2s_shared.go` — delete dead exports (Pass D)
- `/Users/jaten/ivy/goivy/check/ranking.go` — redirect 4 wrapper callers (Pass D), comment cleanup (Pass E)
- `/Users/jaten/ivy/goivy/check/helpers.go` — `MatchHandler.IsCti` type fix (Pass F)
- `/Users/jaten/ivy/goivy/ast/decl_ast.go` — doc comment update on `LabeledFormula.TraceHook` (Pass E)
- `/Users/jaten/ivy/goivy/module/module.go` — doc comment update on `Module.TraceHook` (Pass E)
