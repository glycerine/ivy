# Plan: Unify LabeledFormula — eliminate module.LabeledFormula

## Context

Python has ONE `LabeledFormula` class (in `ivy_ast.py:621`). The Go port created two:
- `ast.LabeledFormula` — Formula field is `ast.Node`
- `module.LabeledFormula` — Formula field is `lg.Node`

These can't interoperate because `lg.Node` and `ast.Node` are separate interfaces. This forces adapter wrappers and lossy conversion functions throughout the codebase.

**Goal:** Delete `module.LabeledFormula`, use `ast.LabeledFormula` everywhere.

**Prerequisite:** `lg.Node` types must satisfy `ast.Node` first, so `ast.LabeledFormula.Formula` (type `ast.Node`) can hold `lg.Node` values directly.

## Phase 1: Make lg.Node types satisfy ast.Node

*(Full details in LOGIC_NODE_EMBEDS_AST.md)*

Add 4 methods to 28 types (25 in logic/, 3 in ivylogic/):
- Embed `ast.Base` → free `GetLineno()`/`SetLineno()`
- Add `Args() []ast.Node` per type (wraps `Children()`)
- Add `Clone([]ast.Node) ast.Node` per type (from existing `CloneNode` switch)

New files: `logic/ast_compat.go`, `ivylogic/ast_compat.go`

After this, any `lg.Node` value IS an `ast.Node` — no adapter needed.

## Phase 2: Fix ast.LabeledFormula.Temporal field

**Python stores `temporal` as `None` or `True` (a bool).** The Go `ast.LabeledFormula` has `Temporal Node` — a porting mistake. The `module.LabeledFormula` got it right with `Temporal bool`.

**Change:** `ast/decl.go` — change `Temporal Node` → `Temporal bool`

Also add `Lineno int` field to ast.LabeledFormula (module.LF has it; Python stores it via `.lineno` attribute on the AST base).

ast.LabeledFormula already has `Base` which provides `GetLineno()` returning `Location`. The explicit `Lineno int` field is used by code that reads `lf.Lineno` directly (simpler than `lf.GetLineno().Line`). Keep both: `Base.Loc` for ast.Node interface compliance, `Lineno` for direct access matching Python's `lf.lineno`.

### Field reconciliation

Final `ast.LabeledFormula`:
```go
type LabeledFormula struct {
    Base                    // GetLineno/SetLineno for ast.Node
    Label        Node       // ast.Node (lg.Node values work after Phase 1)
    Formula      Node       // ast.Node (lg.Node values work after Phase 1)
    ID           int64
    Lineno       int        // direct line number (matches Python lf.lineno)
    Temporal     bool       // was Node, Python uses bool — FIXED
    Explicit     bool
    IsDefinition bool       // already in ast.LF, absent from module.LF
    Assumed      bool
    Unprovable   bool
}
```

### Callers of ast.LF.Temporal that treat it as Node

Search for code that does anything other than nil-check on `Temporal`. If any code inspects the node value, we need to handle that. From Python analysis: it's only ever `None` or `True`, so bool is correct.

## Phase 3: Delete module.LabeledFormula, replace with ast.LabeledFormula

### 3a: Change Module struct fields

In `module/module.go`, change all `[]*LabeledFormula` fields to `[]*ast.LabeledFormula`:
- `LabeledAxioms`, `LabeledProps`, `LabeledConjs`, `LabeledInits`
- `AssumedInvs`, `ConjSubgoals`, `Definitions`
- `Subgoals []SubgoalEntry` (contains `*LabeledFormula`)
- `Proofs []ProofEntry` (contains `*LabeledFormula`)
- `Postconds map[string][]*LabeledFormula`

module/ must add `import "github.com/glycerine/goivy/ast"`. Confirmed: no circular dependency (ast/ doesn't import module/).

### 3b: Delete module.LabeledFormula struct and helpers

Remove from `module/module.go`:
- `type LabeledFormula struct`
- `copyLFSlice()`, `copyMapLF()`

### 3c: Update all 31 consumer files across 12 packages

Mechanical replacement: `module.LabeledFormula` → `ast.LabeledFormula` (or just `*ast.LabeledFormula`).

**Packages to update:**
- isolate/ (7 files)
- ranking/ (4 files) — already uses both types
- check/ (4 files) — already uses both types
- compiler/ (3 files) — already uses both types
- temporal/ (2 files)
- printer/ (2 files)
- mc/ (2 files)
- interp/ (2 files)
- art/ (2 files)
- l2s/ (1 file)
- bmc/ (1 file)
- actions/ (1 file)

Most changes are just type name substitution. Code that accesses `lf.Temporal` as bool stays the same (it's still bool). Code that accesses `lf.Formula` as `lg.Node` needs a type assertion: `lf.Formula.(lg.Node)` — but since lg.Node satisfies ast.Node after Phase 1, storing works directly.

### 3d: Remove conversion helpers in check/helpers.go

Delete `ModuleLFToAstLF`, `AstLFToModuleLF`, `ModuleSchemataToAst` — no longer needed.

### 3e: Remove adapter duplicates (Phase 1 cleanup)

Same as LOGIC_NODE_EMBEDS_AST.md Step 5 — remove 4 duplicate `logicNodeAdapter`/`logicASTAdapter` wrappers.

## Implementation order

| Step | What | Files | Risk |
|------|------|-------|------|
| 1 | Embed ast.Base in 28 logic types | logic/*.go, ivylogic/formula.go | Low — additive |
| 2 | Add Args()/Clone() to 28 types | logic/ast_compat.go, ivylogic/ast_compat.go (NEW) | Low — additive |
| 3 | Fix ast.LF.Temporal: Node → bool | ast/decl.go + callers | Medium — semantic change |
| 4 | Add Lineno field to ast.LF | ast/decl.go | Low — additive |
| 5 | Change Module fields to *ast.LF | module/module.go | Medium — type change |
| 6 | Delete module.LabeledFormula | module/module.go | Low — cleanup |
| 7 | Update 31 consumer files | 12 packages | Medium — mechanical but wide |
| 8 | Remove adapters + conversion helpers | proof/, l2s/, temporal/, check/ | Low — cleanup |

Build and test after each step. Steps 1-2 are fully independent. Steps 3-4 can be done in parallel. Steps 5-8 must be sequential.

## Verification

1. `go build ./...` after each step
2. `go vet ./...` after each step
3. `var _ ast.Node = (*logic.Eq)(nil)` compiles (Phase 1 check)
4. `DYLD_LIBRARY_PATH=z3/build go test ./logic/ ./ivylogic/ ./ast/ ./module/ ./proof/ ./check/ ./compiler/ ./isolate/ ./temporal/ ./l2s/ ./ranking/`
5. `DYLD_LIBRARY_PATH=z3/build go test -tags web ./webui/`
6. Grep for `module.LabeledFormula` — should return 0 matches when done
