# Merge clauseops/ and transrel/ into actions/

**Created:** 2026-04-04 ~14:00 UTC

## Context

The split of clauseops/ and transrel/ from actions/ was a poor design decision that
forces `interface{}` annotations everywhere and prevents type-safe code. The dependency
graph is clean — neither clauseops nor transrel imports actions — so merging is safe.

**Goal:** Move all .go files from `clauseops/` and `transrel/` into `actions/`, keeping
the same filenames. Change package declarations, fix imports across the codebase, resolve
naming conflicts, then delete the empty directories.

**Prior work:** The port audit fixes (bugs 1-5 from the previous plan) were already
implemented in clauseops/ops.go. Those fixes move with the files.

## Naming Conflicts to Resolve BEFORE Moving

### Exported function conflicts (5):

1. **`RenameClauses`** — two different functions:
   - `clauseops/ops.go:571`: `func RenameClauses(clauses *Clauses, subs map[lg.NodeKey]*lg.Const) *Clauses`
   - `transrel/transrel.go:1272`: `func RenameClauses(node lg.Expr, rn map[string]string) lg.Expr`
   - **Fix:** Rename transrel's to `RenameExprByName` (it renames a single Expr by string name map).
     Update all callers of the transrel version.

2. **`BindOldsAction`** — function vs type:
   - `transrel/transrel.go:1531`: `func BindOldsAction(u *Update) *Update`
   - `actions/action.go:960`: `type BindOldsAction struct { ... }`
   - **Fix:** Rename transrel's function to `BindOldsUpdate` (it operates on Updates, not Actions).
     Update all callers.

3. **`ShortTypeName`** — trivial wrapper:
   - `clauseops/astutil.go:105`: actual implementation
   - `actions/helpers.go:17`: `func ShortTypeName(v interface{}) string { return co.ShortTypeName(v) }`
   - **Fix:** Delete the wrapper in actions/helpers.go (it just delegates to clauseops).

### Unexported function conflicts (2):

4. **`isTrue`** — duplicate implementations:
   - `clauseops/ops.go:1346`
   - `actions/update.go:163`
   - **Fix:** Delete the one in actions/update.go (clauseops version will be in the same package).

5. **`isFalse`** — duplicate implementations:
   - `clauseops/ops.go:1354`
   - `actions/update.go:167`
   - **Fix:** Delete the one in actions/update.go.

### Test function conflicts (2):

6. **`TestBindOldsAction`** — both test different things:
   - `transrel/impl_test.go:579` — tests the function (now `BindOldsUpdate`)
   - `actions/actions_test.go:185` — tests the type
   - **Fix:** Rename transrel's to `TestBindOldsUpdate`.

7. **`TestNullUpdate`** — both test NullUpdate:
   - `transrel/transrel_test.go:144`
   - `actions/update_test.go:446`
   - **Fix:** Rename transrel's to `TestNullUpdateTransrel`.

## Post-merge cleanup: Remove `interface{}` hacks

After the merge, these can be replaced with direct types:

- `AnnotConjoiner` interface → use `Annotation` directly
- `AnnotRenamer` interface → use `Annotation.Rename()` directly
- `AnnotIteFunc` callback → use `Annotation.Ite()` directly
- `AnnotOp func(annots ...interface{})` → use `func(annots ...Annotation)`
- `Clauses.Annot interface{}` → `Clauses.Annot Annotation`
- `AnnotConjoiner.ConjWith(other interface{})` → `Annotation.Conj(other Annotation)`

## File inventory

### Files to move (keep same basename):

**From clauseops/ (8 source + 3 test):**
- astutil.go, batch16.go, canon.go, clauses.go, litclause.go, ops.go, skolem.go, subsume.go
- clauseops_test.go, port_audit_test.go, sexp_test.go

**From transrel/ (3 source + 2 test):**
- interpolant.go, phase4.go, transrel.go
- impl_test.go, transrel_test.go

No base-name conflicts with existing actions/ files.

### External importers to update (~40 files):

Packages that import `clauseops`:
  solver, module, interp, compiler, fragment, check, tactics, art, bmc, vmt,
  alpha, iupdr, l2s, ranking

Packages that import `transrel`:
  mc, trace, art, tactics, ranking, interp, vmt, check, bmc

Many already import `actions` alongside one or both, so those just drop the extra import.

## Execution Plan

### Phase 1: Resolve naming conflicts (in-place, before any moves)

Do these renames while files are still in their original packages, so callers
are easy to find and update.

1. In `transrel/transrel.go`: rename `RenameClauses` → `RenameExprByName`.
   Update all callers across the codebase.

2. In `transrel/transrel.go`: rename `BindOldsAction` → `BindOldsUpdate`.
   Update all callers (actions/update.go, actions/update_test.go, etc.).

3. In `actions/helpers.go`: delete the `ShortTypeName` wrapper.
   Update any callers to use `co.ShortTypeName` (which will become just
   `ShortTypeName` after the merge).

4. In `actions/update.go`: delete `isTrue` and `isFalse` (lines 163-171).
   These are identical to clauseops versions which will be in the same package.

5. In `transrel/impl_test.go`: rename `TestBindOldsAction` → `TestBindOldsUpdate`.

6. In `transrel/transrel_test.go`: rename `TestNullUpdate` → `TestNullUpdateTransrel`.

7. Build and test to confirm renames are clean: `go build ./... && go test ./clauseops/... ./transrel/... ./actions/...`

### Phase 2: Move files

For each file in clauseops/ and transrel/:

1. Copy file to actions/ (same basename)
2. Change `package clauseops` or `package transrel` → `package actions`
3. Remove self-imports (e.g., clauseops files importing clauseops, transrel importing clauseops)
4. Remove the `co "clauseops"` and `tr "transrel"` import aliases, replace
   `co.Foo` → `Foo` and `tr.Foo` → `Foo` within the moved files

### Phase 3: Update external importers

For every file outside actions/ that imports clauseops or transrel:

1. Remove `co "github.com/glycerine/ivy/goivy/clauseops"` import
2. Remove `tr "github.com/glycerine/ivy/goivy/transrel"` import
3. Add `"github.com/glycerine/ivy/goivy/actions"` if not already present
4. Replace `co.Foo` → `actions.Foo` (or just `Foo` if already importing actions with no alias)
5. Replace `tr.Foo` → `actions.Foo`
6. Handle files that already import actions — just drop the co/tr aliases and prefixes

### Phase 4: Delete old directories

1. Remove `clauseops/` directory
2. Remove `transrel/` directory

### Phase 5: Type-safety cleanup

Replace `interface{}` with concrete `Annotation` type throughout the merged package:

1. `Clauses.Annot interface{}` → `Clauses.Annot Annotation` (may need nil-Annotation sentinel)
2. Delete `AnnotConjoiner`, `AnnotRenamer`, `AnnotIteFunc` interfaces/callbacks
3. Use `Annotation.Conj()`, `Annotation.Rename()`, `Annotation.Ite()` directly
4. Fix `AnnotOp` type to use `Annotation` instead of `interface{}`

### Phase 6: Build and test

1. `go build ./...`
2. `go test ./actions/...` — all moved tests pass
3. `go test ./...` — full test suite
4. Verify no remaining references to clauseops or transrel packages

## Verification

- `go build ./...` compiles clean
- `go test ./actions/...` passes (includes all former clauseops + transrel tests)
- `go test ./...` full suite passes
- `grep -r "clauseops\|transrel" --include="*.go" goivy/` returns nothing outside of
  comments/strings (no stale imports)
