# Merge clauseops/ into module/, transrel/ into actions/

**Created:** 2026-04-04 ~16:00 UTC (revised from earlier failed attempt)

## Context

The split of clauseops/ and transrel/ into separate packages forces `interface{}`
annotation hacks and prevents type-safe code. We want to merge them into their
natural homes following the layer architecture (from ~/ivy/README.md:294-298):

- **Layer 2**: logic, ivylogic, ivyutils, logicutil, ast
- **Layer 3**: module, solver, clauseops (→ merge clauseops INTO module)
- **Layer 4**: actions, transrel (→ merge transrel INTO actions)

**clauseops → module** (layer 3 → layer 3): Safe because clauseops only imports
layer 1-2 packages ({ast, ivylogic, ivyutils, logic, logicutil, xtracer}).
Module already imports clauseops, so the merge just inlines it. Solver also
imports clauseops — after merge, solver imports module (already does for other
things, no new cycle).

**transrel → actions** (layer 3→4 → layer 4): Safe because transrel imports
{clauseops(→module), ivylogic, ivyutils, logic, module, solver, xtracer} — all
layer 1-3. Actions already imports transrel. No cycle.

**Why the previous attempt failed**: We tried merging BOTH into actions/.
That created a cycle: actions→solver→clauseops(now actions). The fix is to keep
clauseops at layer 3 by merging it into module instead.

**Critical lesson**: Do NOT use blanket `sed 's/co\.//g'` — it mangles unrelated
code like `s.Context`→`s.ntext`, `destr.CSort`→`desCSort`. Use targeted,
file-by-file edits instead.

## Naming Conflicts

### clauseops → module conflicts:

1. **`canon.go`** filename — both packages have one.
   - `module/canon.go`: Module.Canon(), CanonSnapshot(), ~60 canon helpers
   - `clauseops/canon.go`: Clauses.Canon() (27 lines)
   - **Fix:** Rename clauseops's to `clauses_canon.go` when copying.

2. **`ResortSort`** — different signatures:
   - `module/canonize.go:212`: `func ResortSort(s lg.Sort, rn map[lg.NodeKey]*SortRefinement) lg.Sort`
   - `clauseops/batch16.go:247`: `func ResortSort(s lg.Sort, subs map[lg.NodeKey]lg.Sort) lg.Sort`
   - **Fix:** Rename clauseops's to `resortSortBySort` (unexport — only called internally).

3. **`ResortClauses`** — different signatures:
   - `module/canonize.go:371`: `func ResortClauses(cls *Clauses, rn map[..]*SortRefinement) *Clauses`
   - `clauseops/ops.go:917`: `func ResortClauses(clauses *Clauses, subs map[...]lg.Sort) *Clauses`
   - **Fix:** Rename clauseops's to `resortClausesBySort` (unexport — only called internally).

4. **`defToConstraint`** — duplicate implementations:
   - `module/theory.go:535`: local copy
   - `clauseops/clauses.go:247`: canonical copy (delegates to `il.DefinitionToConstraint`)
   - **Fix:** Delete module/theory.go's copy after merge (clauseops's version will be
     in the same package).

### transrel → actions conflicts:

5. **`BindOldsAction`** — function vs type:
   - `transrel/transrel.go:1531`: `func BindOldsAction(u *Update) *Update`
   - `actions/action.go:960`: `type BindOldsAction struct { ... }`
   - **Fix:** Rename transrel's function to `BindOldsUpdate`.

6. **`RenameClauses`** — dead code in transrel (zero callers):
   - `transrel/transrel.go:1272`: `func RenameClauses(node lg.Expr, rn map[string]string) lg.Expr`
   - No conflict after clauseops→module (clauseops's RenameClauses will be in module).
   - **Fix:** Rename to `RenameExprByName` for clarity.

7. **`ShortTypeName`** — wrapper in actions:
   - `actions/helpers.go:17`: `func ShortTypeName(v interface{}) string { return co.ShortTypeName(v) }`
   - After merge, `co.ShortTypeName` becomes `mod.ShortTypeName` — wrapper just changes prefix.
   - **Fix:** Change to `mod.ShortTypeName(v)` during import update. No naming conflict.

8. **`TestBindOldsAction`** — test name conflict:
   - `transrel/impl_test.go:579` — tests the function
   - `actions/actions_test.go:185` — tests the type
   - **Fix:** Rename transrel's to `TestBindOldsUpdate`.

9. **`TestNullUpdate`** — test name conflict:
   - `transrel/transrel_test.go:144`
   - `actions/update_test.go:446`
   - **Fix:** Rename transrel's to `TestNullUpdateTransrel`.

## Execution Plan

### Phase 1: Resolve naming conflicts (in-place, before any moves)

All renames happen while files are still in their original packages.
Build+test after each to confirm correctness.

**In clauseops/:**
1. `batch16.go`: Rename `ResortSort` → `resortSortBySort` (and all internal callers)
2. `ops.go`: Rename `ResortClauses` → `resortClausesBySort` (and all internal callers)

**In transrel/:**
3. `transrel.go`: Rename `RenameClauses` → `RenameExprByName`
4. `transrel.go`: Rename `BindOldsAction` → `BindOldsUpdate` (+ callers in actions/)
5. `impl_test.go`: Rename `TestBindOldsAction` → `TestBindOldsUpdate`
6. `transrel_test.go`: Rename `TestNullUpdate` → `TestNullUpdateTransrel`

**Verify:** `go build ./... && go test ./clauseops/... ./transrel/... ./actions/...`

### Phase 2: Move clauseops/ files into module/

For each file in clauseops/ (8 source + 3 test):
1. Copy to module/ (rename `canon.go` → `clauses_canon.go`)
2. Change `package clauseops` → `package module`
3. No self-import removal needed (clauseops doesn't import itself)

Then in module's existing files:
4. Remove `co "github.com/glycerine/ivy/goivy/clauseops"` import
5. Replace `co.Foo` → `Foo` (now in same package) — do this with targeted
   Edit tool calls, NOT blanket sed
6. Delete `defToConstraint` from `module/theory.go` (duplicate of clauseops version)
7. Add `lu "github.com/glycerine/ivy/goivy/logicutil"` to module files that need it
   (clauseops code uses `lu.CloseEPR` etc.)

**Verify:** `go build ./module/... && go test ./module/...`

### Phase 3: Move transrel/ files into actions/

For each file in transrel/ (3 source + 2 test):
1. Copy to actions/ (no filename conflicts)
2. Change `package transrel` → `package actions`
3. Remove `co "github.com/glycerine/ivy/goivy/clauseops"` import from moved files
4. Replace `co.Foo` → `mod.Foo` in moved files (clauseops is now in module)
5. Ensure moved files import module (they already import it for `*mod.Module`)

Then in actions' existing files:
6. Remove `"github.com/glycerine/ivy/goivy/transrel"` import
7. Replace `transrel.Foo` → `Foo` (now in same package)
8. Replace `co.Foo` → `mod.Foo` where actions files referenced clauseops
9. Remove `co "github.com/glycerine/ivy/goivy/clauseops"` import from actions files
10. Ensure `mod "github.com/glycerine/ivy/goivy/module"` import exists

**Verify:** `go build ./actions/... && go test ./actions/...`

### Phase 4: Update all external importers

Every file outside module/ and actions/ that imports clauseops or transrel:

**clauseops importers** (~15 packages: solver, interp, compiler, fragment, check,
tactics, art, bmc, vmt, alpha, iupdr, l2s, ranking, proof, isolate):
- Change `co "...clauseops"` → `mod "...module"` (or add mod alias if module already imported)
- Replace `co.Foo` → `mod.Foo`
- If file already imports module without alias, add `mod` alias

**transrel importers** (~10 packages: mc, trace, art, tactics, ranking, interp,
vmt, check, bmc):
- Change `tr "...transrel"` / `"...transrel"` → use actions import
- Replace `tr.Foo` → `actions.Foo`
- If file already imports actions, just drop the transrel import

**Critical:** Use targeted Edit tool calls per-file. No sed — it mangles
unrelated code (`s.Context`→`s.ntext`, `destr.CSort`→`desCSort`).

**Verify:** `go build ./...`

### Phase 5: Delete old directories

1. `rm -rf clauseops/`
2. `rm -rf transrel/`

**Verify:** `go build ./... && go test ./...`

### Phase 6 (optional, separate PR): Type-safety cleanup

After everything compiles and tests pass, the `interface{}` annotation hacks in
the merged actions/ package can be replaced with concrete `Annotation` types.
This is a follow-up task, not part of this merge.

## File inventory

### clauseops/ → module/ (8 source + 3 test):
| Source | Destination |
|---|---|
| clauseops/astutil.go | module/astutil.go |
| clauseops/batch16.go | module/batch16.go |
| clauseops/canon.go | module/clauses_canon.go |
| clauseops/clauses.go | module/clauses.go |
| clauseops/litclause.go | module/litclause.go |
| clauseops/ops.go | module/ops.go |
| clauseops/skolem.go | module/skolem.go |
| clauseops/subsume.go | module/subsume.go |
| clauseops/clauseops_test.go | module/clauseops_test.go |
| clauseops/port_audit_test.go | module/port_audit_test.go |
| clauseops/sexp_test.go | module/sexp_test.go |

### transrel/ → actions/ (3 source + 2 test):
| Source | Destination |
|---|---|
| transrel/interpolant.go | actions/interpolant.go |
| transrel/phase4.go | actions/phase4.go |
| transrel/transrel.go | actions/transrel.go |
| transrel/impl_test.go | actions/impl_test.go |
| transrel/transrel_test.go | actions/transrel_test.go |

## Verification

1. `go build ./...` — compiles clean
2. `go test ./module/...` — all former clauseops tests pass
3. `go test ./actions/...` — all former transrel tests pass
4. `go test ./...` — full suite passes
5. `grep -rn 'goivy/clauseops\|goivy/transrel' --include='*.go' .` — no stale imports
