# Plan: Merge `l2s/` and `ranking/` into the `check/` package

Created: 2026-04-11 12:30 UTC

## Context

The Go port created too many small packages. `l2s/` and `ranking/` are tactic-implementation packages that are only consumed by `check/` (and indirectly by `cmd/goivy_check`). The split is purely structural — there are no other consumers, and a previous import cycle (`l2s → check → l2s`) that motivated separation has been gone for a while.

The smoking gun is `check/l2s_hooks.go`, whose top comment says it was "moved from l2s/hooks.go to break the import cycle l2s → check → l2s". With the cycle gone, the artifacts of that workaround remain — split files, opaque dispatch via `interface{}` fields, and a layer of indirection between l2s tactic state and check's diagnostic plumbing.

This plan merges `l2s/` and `ranking/` into `check/` as a **purely structural refactor**:
- Move files wholesale (no merging or splitting of file contents).
- Change `package l2s` / `package ranking` → `package check` at the top.
- Strip the `l2s.` and `check.` qualifiers inside the moved files.
- Resolve the small set of symbol-name collisions by renaming.

After this plan, follow-up plans can simplify the trace_hook plumbing (delete `L2STraceHookData`, switch to direct closures, etc.) — but that semantic cleanup is **out of scope here**. This plan only does the move.

## Investigation findings

### Import graph (verified)

- `check` imports `l2s` (in `check/check.go` and `check/l2s_hooks.go`)
- `ranking` imports `l2s` (in `ranking/ranking.go` and `ranking/hooks.go`)
- `ranking` imports `check` (in `ranking/hooks.go`)
- **No other package** in goivy imports `l2s` or `ranking` (verified via `grep -rln "github.com/glycerine/ivy/goivy/(l2s|ranking)\b"`).
- `cmd/goivy_check/main.go` only imports `check` — never `l2s` or `ranking` directly.

So the merge has exactly four import sites to update (the four files listed above), plus the package declarations and qualifier strips inside the moved files themselves.

### Files to move (file inventory)

**From `l2s/` (4 files):**
- `l2s/hook_data.go` — opaque trace-hook payload struct (Phase 7 artifact)
- `l2s/l2s.go` — main `l2sTacticInt` and entry points (`L2STactic`, `L2STacticFull`, `L2STacticAuto`)
- `l2s/l2s_auto.go` — `l2sAutoInvariants` and per-task closures
- `l2s/shared.go` — `InstrumentationConfig`, `SharedStep1..12`, helpers used by both l2s and ranking

**From `ranking/` (4 files):**
- `ranking/hooks.go` — `RankingAutoHook`, `RankingAutoHookConfig`
- `ranking/ranking.go` — main ranking pipeline, `L2STactic(*L2STacticConfig)` entry, `Desugar`, `RegisteredTactics` registry
- `ranking/ranking_test.go` — tests
- `ranking/tactic.go` — `TacticFunc` type and supporting tactic glue

**Total: 8 files moved into `check/`.**

### Filename collisions (none)

`check/` has: `check.go`, `check_port_test.go`, `check_test.go`, `helpers.go`, `isolate_check.go`, `l2s_hooks.go`, `phase7.go`, `regression_test.go`, `vprint.go`.

None of the eight incoming filenames literally collide. However, `shared.go`, `hooks.go`, and `tactic.go` are generic enough to be confusing inside the merged package. We **rename three for clarity** (no functional change):

| Source | Destination |
|---|---|
| `l2s/hook_data.go` | `check/hook_data.go` |
| `l2s/l2s.go` | `check/l2s.go` |
| `l2s/l2s_auto.go` | `check/l2s_auto.go` |
| `l2s/shared.go` | `check/l2s_shared.go` |
| `ranking/hooks.go` | `check/ranking_hooks.go` |
| `ranking/ranking.go` | `check/ranking.go` |
| `ranking/ranking_test.go` | `check/ranking_test.go` |
| `ranking/tactic.go` | `check/ranking_tactic.go` |

### Symbol-name collisions (verified by `grep` and counting)

Six real collisions (and only six). All are package-level decls; no method-name collisions.

| Symbol | Locations | Resolution |
|---|---|---|
| **`L2STactic`** (exported func) | `l2s/l2s.go:213` (proof tactic, takes `ProofCheckerInterface`) vs `ranking/ranking.go:221` (takes `*L2STacticConfig`) | Keep `L2STactic` for the l2s-package version (it's the proof tactic registered by name). Rename ranking's to **`RankingL2STactic`**. Update its single call site in `ranking/ranking.go:853`. |
| **`Desugar`** (exported func) | `l2s/l2s.go:941` (`Desugar(expr, proofLabel)`) vs `ranking/ranking.go:753` (`Desugar(expr, proofLabel, l2sSaved)`) | Keep `Desugar` for the l2s version. Rename ranking's to **`RankingDesugar`**. Update its callers within the moved ranking files. |
| **`RegisterTactics`** (exported func) | `l2s/l2s.go:1004` (`RegisterTactics(proofCfg)`) vs `check/check.go:1162` (`RegisterTactics(proofCfg, mod)`) | Rename l2s's to **`RegisterL2STactics`**. Update the single call at `check/check.go:1173` from `l2s.RegisterTactics(proofCfg)` to `RegisterL2STactics(proofCfg)`. |
| **`makeAnd`** (unexported func) | `l2s/l2s.go:141` (`&lg.And{Terms: terms}` direct) vs `ranking/ranking.go:728` (uses `lg.NewAnd` constructor with error path) | The implementations differ. Rename ranking's to **`rankingMakeAnd`** and update its callers within the moved ranking files. (Do NOT silently merge — different semantics under construction errors.) |
| **`strPtr`** (unexported func) | `l2s/l2s.go:139` vs `ranking/ranking.go:169` | Implementations are byte-identical. Delete ranking's `strPtr` and let it use the unified one. |
| **`testAstCfg`** (unexported test var) | `ranking/ranking_test.go:12` vs `check/check_port_test.go:18` | Rename ranking's to **`rankingTestAstCfg`** and update references inside `ranking_test.go`. |

No type, var, or const collisions besides `testAstCfg`.

### Cross-package qualifier strips (verified)

Within the moved files, these qualifier-stripped references must be updated:

**Files moved from `l2s/` reference these l2s-package symbols** (which become bare names):
- None — l2s files don't import their own package, so there's nothing to strip inside them.

**Files moved from `ranking/` reference these `l2s.X` symbols** (need stripping):
- `ranking/hooks.go:33`: `l2s.L2SGToGlobally` → `L2SGToGlobally`
- `ranking/ranking.go`: any `l2s.InstrumentationConfig`, `l2s.BuildDefnDeps`, `l2s.BuildDependenciesFunc`, `l2s.SharedStep*`, `l2s.NewWithSorts` (etc.) — strip to bare names. (Full list comes from a single `grep -n "l2s\\." ranking/ranking.go` after moving.)

**Files moved from `ranking/` reference these `check.X` symbols** (need stripping):
- `ranking/hooks.go:26`: `check.Checker` → `Checker`
- `ranking/hooks.go:32`: `check.L2SRenamingHook` → `L2SRenamingHook`
- `ranking/hooks.go:36`: `check.Checker` → `Checker`

**Files staying in `check/` that currently reference `l2s.X`** (need stripping):
- `check/check.go:600`: `*l2s.L2STraceHookData` → `*L2STraceHookData`
- `check/check.go:602`: `l2s.HookKindFull` → `HookKindFull`
- `check/check.go:605`: `l2s.HookKindRenaming, l2s.HookKindAuto` → `HookKindRenaming, HookKindAuto`
- `check/check.go:613`: `l2s.HookKindAuto` → `HookKindAuto`
- `check/check.go:1135`: `l2s.L2STacticFull` → `L2STacticFull`
- `check/check.go:1173`: `l2s.RegisterTactics` → `RegisterL2STactics` (also renamed per collision table)
- `check/l2s_hooks.go:75`: `l2s.L2SGToGlobally` → `L2SGToGlobally`
- `check/l2s_hooks.go:227`: `*l2s.L2STraceHookData` → `*L2STraceHookData`

The `"github.com/glycerine/ivy/goivy/l2s"` import line must be deleted from `check/check.go` and `check/l2s_hooks.go`.

## Approach

Pure file-relocation refactor in three passes. Each pass leaves the build broken until the next; I'll do them sequentially and verify with `go build` only at the end.

### Pass 1 — Move and re-package the eight files

For each file in the inventory above:
1. Move (or copy + delete) the file to its new path under `check/`.
2. Change the top `package l2s` or `package ranking` declaration to `package check`.
3. Remove the import line for `"github.com/glycerine/ivy/goivy/l2s"` and `"github.com/glycerine/ivy/goivy/check"` if present.
4. Strip `l2s.` and `check.` qualifiers from in-file references.
5. Apply the symbol renames from the collision table to definitions and call sites within the file.

The `temporal` import in `l2s/l2s.go` (which currently imports `temporal`) stays — temporal is a separate downstream package and is not part of this merge.

### Pass 2 — Update `check/check.go` and `check/l2s_hooks.go`

These two files stay in place but reference symbols from the moved packages. Delete the `l2s` import from each and strip the `l2s.` qualifier from each reference site (~10 sites total, listed above). Apply the `l2s.RegisterTactics` → `RegisterL2STactics` rename at the single call site.

### Pass 3 — Delete the empty `l2s/` and `ranking/` directories

After all files are moved, `l2s/` and `ranking/` should contain no `.go` files. Remove the directories. Verify no other package still imports them via `grep -rln "github.com/glycerine/ivy/goivy/(l2s|ranking)\b" .`.

## Files to modify

### Move (8 files)

| From | To | Changes inside |
|---|---|---|
| `l2s/hook_data.go` | `check/hook_data.go` | `package l2s` → `package check` |
| `l2s/l2s.go` | `check/l2s.go` | package decl; remove `"github.com/glycerine/ivy/goivy/temporal"` only if check already provides it (verify — it doesn't, so the import stays); rename `RegisterTactics` → `RegisterL2STactics` |
| `l2s/l2s_auto.go` | `check/l2s_auto.go` | package decl |
| `l2s/shared.go` | `check/l2s_shared.go` | package decl |
| `ranking/hooks.go` | `check/ranking_hooks.go` | package decl; remove `l2s` and `check` imports; strip `l2s.` and `check.` qualifiers |
| `ranking/ranking.go` | `check/ranking.go` | package decl; remove `l2s` import; strip `l2s.` qualifiers; rename `L2STactic` → `RankingL2STactic`; rename `Desugar` → `RankingDesugar`; rename `makeAnd` → `rankingMakeAnd`; delete `strPtr` |
| `ranking/ranking_test.go` | `check/ranking_test.go` | package decl; rename `testAstCfg` → `rankingTestAstCfg` and update references; rename references to `L2STactic`/`Desugar`/`makeAnd`/`strPtr` |
| `ranking/tactic.go` | `check/ranking_tactic.go` | package decl |

### Edit in place (2 files)

| File | Changes |
|---|---|
| `check/check.go` | Delete `"github.com/glycerine/ivy/goivy/l2s"` import. Strip `l2s.` from 7 reference sites. Update `l2s.RegisterTactics(proofCfg)` → `RegisterL2STactics(proofCfg)` at line 1173. |
| `check/l2s_hooks.go` | Delete `"github.com/glycerine/ivy/goivy/l2s"` import. Strip `l2s.` from 2 reference sites (lines 75 and 227). Update the file's top comment to remove the now-stale "Moved from l2s/hooks.go to break the import cycle" note. |

### Delete (2 directories)

After successful build:
- `l2s/` (empty after the move)
- `ranking/` (empty after the move)

## Verification

1. **Import-graph sanity** — confirm no cycles introduced:
   ```sh
   cd ~/ivy/goivy && go list -f '{{.ImportPath}}: {{.Imports}}' ./check/...
   ```
   Should NOT mention `goivy/l2s` or `goivy/ranking`.

2. **Build**:
   ```sh
   cd ~/ivy/goivy && go build ./...
   ```
   Must be clean.

3. **Vet** (just the affected package — pre-existing vet warnings in other packages are unrelated):
   ```sh
   cd ~/ivy/goivy && go vet ./check/...
   ```

4. **Tests** for affected and downstream packages:
   ```sh
   cd ~/ivy/goivy && go test ./check/... ./compiler/... ./proof/... ./module/... ./tactics/... ./temporal/...
   ```
   All previously-passing tests should still pass. The renamed `testAstCfg` still works because it's referenced by name within `ranking_test.go` only.

5. **Stale-reference grep** — must come up empty:
   ```sh
   grep -rln "github.com/glycerine/ivy/goivy/l2s\b" ~/ivy/goivy
   grep -rln "github.com/glycerine/ivy/goivy/ranking\b" ~/ivy/goivy
   ```

6. **`cmd/goivy_check` build**:
   ```sh
   cd ~/ivy/goivy && go build ./cmd/goivy_check/...
   ```

## Out of scope

- **Eliminating `L2STraceHookData` opaque struct, the `HookKind*` enum, and the `interface{}` fields on `ast.LabeledFormula.TraceHook` / `module.Module.TraceHook`** — these are still load-bearing because `ast` and `module` cannot import `check`. A follow-up plan can replace this with a side-map keyed by `LF.ID` stored in `module.Module`, eliminating the opaque payloads. Not done here.
- **Refactoring `check/check.go`'s mega-function** or any internal restructuring of the merged package. Pure move.
- **Merging `temporal/`** into check. Temporal is used by more than just check (proof, tactics, l2s/l2s.go). Out of scope.
- **Deleting `ranking_test.go`'s `RegisterTactic`/`GetTactic` test cases** — they exercise ranking's per-tactic registry which is different from `module.ProofConfig.RegisterTactic`. They stay as-is post-rename.

## Risks and rollback

- **Risk: missed `l2s.X` reference inside a moved file.** The build will fail with "undefined: l2s" — easy to catch and fix.
- **Risk: missed `ranking.X` reference somewhere outside the merged files.** Verified to not exist via the import-site scan above; only `check/check.go` and `check/l2s_hooks.go` reference these packages. But still: a final `grep` step in verification catches stragglers.
- **Risk: `temporal` import in `l2s/l2s.go` becomes redundant.** Verify and either keep or drop based on whether the moved code still uses temporal symbols.
- **Risk: `ranking_test.go` references symbols (e.g., `L2STacticConfig`) that I haven't audited.** The test file uses ranking's local types — they're now in the same package, so qualifiers go away naturally. Test compile errors will reveal any missed renames.
- **Rollback**: each file move is independently reversible via `git checkout` of the affected files. No on-disk format change. No semantic change.

## Implementation checklist

### Pass 1 — Move and re-package
- [ ] Move `l2s/hook_data.go` → `check/hook_data.go`; change package decl.
- [ ] Move `l2s/l2s.go` → `check/l2s.go`; change package decl; rename `RegisterTactics` → `RegisterL2STactics`.
- [ ] Move `l2s/l2s_auto.go` → `check/l2s_auto.go`; change package decl.
- [ ] Move `l2s/shared.go` → `check/l2s_shared.go`; change package decl.
- [ ] Move `ranking/hooks.go` → `check/ranking_hooks.go`; change package decl; remove `l2s`/`check` imports; strip qualifiers.
- [ ] Move `ranking/ranking.go` → `check/ranking.go`; change package decl; remove `l2s` import; strip `l2s.` qualifiers; rename `L2STactic` → `RankingL2STactic`, `Desugar` → `RankingDesugar`, `makeAnd` → `rankingMakeAnd`; delete `strPtr`.
- [ ] Move `ranking/tactic.go` → `check/ranking_tactic.go`; change package decl.
- [ ] Move `ranking/ranking_test.go` → `check/ranking_test.go`; change package decl; rename `testAstCfg` → `rankingTestAstCfg` and update references; update any renamed-symbol references.

### Pass 2 — Update remaining check/ files
- [ ] `check/check.go`: delete `l2s` import; strip `l2s.` from the 7 sites; rename `l2s.RegisterTactics(proofCfg)` → `RegisterL2STactics(proofCfg)`.
- [ ] `check/l2s_hooks.go`: delete `l2s` import; strip `l2s.` from the 2 sites; remove the stale top comment about import cycle.

### Pass 3 — Delete empty directories
- [ ] Verify `l2s/` contains no `.go` files; delete directory.
- [ ] Verify `ranking/` contains no `.go` files; delete directory.

### Verification
- [ ] `go build ./...` clean.
- [ ] `go vet ./check/...` clean.
- [ ] `go test ./check/... ./compiler/... ./proof/... ./module/... ./tactics/... ./temporal/...` green.
- [ ] `grep -rln "github.com/glycerine/ivy/goivy/(l2s|ranking)\b" ~/ivy/goivy` returns no matches.
- [ ] `go build ./cmd/goivy_check/...` clean.

## Critical files

- `/Users/jaten/ivy/goivy/l2s/{hook_data,l2s,l2s_auto,shared}.go` — moved
- `/Users/jaten/ivy/goivy/ranking/{hooks,ranking,ranking_test,tactic}.go` — moved
- `/Users/jaten/ivy/goivy/check/check.go` — edited in place
- `/Users/jaten/ivy/goivy/check/l2s_hooks.go` — edited in place
- `/Users/jaten/ivy/goivy/cmd/goivy_check/main.go` — verified to need no changes (only imports `check`)
