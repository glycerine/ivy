# Plan: Eliminate interface{} Import-Cycle Kludges in goivy

## Context

The Go port uses `interface{}` in many places where concrete types are known but can't be imported due to Go's prohibition on circular imports. Every `interface{}` is a lost compile-time check — bugs that Python catches at runtime (and Go should catch at compile time) slip through silently. This plan catalogs every kludge, groups them by root cause, and proposes fixes.

## Package Dependency Layers (quick reference)

```
Layer 0 (foundation):  ast, lexer, ivyutils
Layer 1:               logic
Layer 2:               clauseops, z3bridge, logicutil, logicparser
Layer 3:               module, solver
Layer 4:               transrel, actions, isolate
Layer 5:               interp, compiler, art
Layer 6:               proof, trace, bmc, temporal
Layer 7:               tactics, webui
Layer 8:               iupdr, check (top-level consumer)
```

Key constraint: arrows point UP only. A package can import anything in a lower layer.
`module` (L3) cannot import `actions` (L4). `compiler` (L5) cannot import `proof` (L6).

---

## Catalog of interface{} Kludges

### Group A: `module.Actions` and related maps — the biggest win

**Root cause:** `module` is L3, `actions.Action` is defined in L4. Module can't import actions.

| Field | File | Should be | Used by |
|-------|------|-----------|---------|
| `module.Actions map[string]interface{}` | module.go:39 | `map[string]Action` | everywhere |
| `module.Mixins map[string][]interface{}` | module.go:40 | `map[string][]MixinDef` | isolate |
| `module.Predicates map[string]interface{}` | module.go:42 | `map[string]lg.Expr` or `map[string]ast.Node` | art, interp |
| `module.InitialActions []interface{}` | module.go:44 | `[]Action` | compiler |
| `module.CompileActionBodyFn` return | module.go:115 | `Action` | actions |
| `art.Actions map[string]interface{}` | art.go:260 | mirrors module | art |
| `art.Predicates map[string]interface{}` | art.go:261 | mirrors module | art |
| `art.Mixins`, `art.Isolates`, etc. | art.go:263-266 | mirrors module | art |

**Fix — Move `Action` interface to `module`**

The `actions.Action` interface depends only on `lg.Expr`, `lg.Symbol`, `ast.Location` — all L0-L1 types. It can live in `module` without any new imports. Then:
- `module.Actions` becomes `map[string]module.Action`
- All 50+ type assertions `x.(actions.Action)` throughout the codebase become compile-time checked
- `actions` package COULD BUT WILL NOT embed/alias `module.Action` for backward compat: `type Action = module.Action`. OMIT THIS, AS WE DO NOT WANT BACKWARDS COMPAT junking up our code base at this point.

Estimated impact: ~15 files touched, ~50 type assertions eliminated.

### Group B: `module.Schemata`, `Theorems`, `Interps`, `Isolates` — AST collections

| Field | Should be |
|-------|-----------|
| `Schemata map[string]interface{}` | `map[string]*ast.LabeledFormula` |
| `Theorems map[string]interface{}` | `map[string]*ast.LabeledFormula` |
| `Interps map[string][]interface{}` | `map[string][]*ast.LabeledFormula` |
| `Isolates map[string]interface{}` | needs `IsolateDefInterface` from isolate pkg |

**Fix for Schemata/Theorems/Interps:** These are all `*ast.LabeledFormula`. The `ast` package is L0 — `module` already imports it. **Just change the types.** No structural work needed.

**Fix for Isolates:** `IsolateDefInterface` is defined in `isolate` (L4). Same pattern as Group A — move the interface to `module`. It only needs `ast.Node` methods.

### Group C: `compiler ↔ proof` cycle — callback injection

| Item | File | Current |
|------|------|---------|
| `module.AdmitDefinitionFn` param `proof` | module.go:119 | `interface{}` (should be `ast.Node`) |
| `compiler.AdmitDefinitionFactory` param `pf` | ivy_compile.go:45 | `interface{}` |
| `check/check.go` init() wiring | check.go:30 | type-asserts `pf.(ast.Node)` |

**Fix:** Change `proof interface{}` to `proof ast.Node` in both `AdmitDefinitionFn` and `AdmitDefinitionFactory` signatures. The actual type IS `ast.Node` (confirmed by the type assertion in check.go:40). No cycle — `ast` is L0. **Just fix the signatures.**

### Group D: `module.CompileActionBodyFn` return type

| Item | File | Current |
|------|------|---------|
| `CompileActionBodyFn` return | module.go:115 | `(interface{}, error)` should be `(Action, error)` |

**Fix:** Once Group A is done (Action interface in module), change return to `(Action, error)`. The type assertion in `actions/extra_actions.go:760` becomes unnecessary.

### Group E: `module.CompCfg` — compiler config

| Item | File | Current |
|------|------|---------|
| `CompCfg` | module.go:109 | `interface{}` (should be `*compiler.CompilerConfig`) |

**Fix:** Move `CompilerConfig` struct to `module`. It's a config struct with no compiler-specific dependencies.

### Group F: Remaining `interface{}` collections in module

These are heterogeneous or rarely-accessed:

| Field | Assessment |
|-------|-----------|
| `Exports []interface{}` | Has `exporter` interface in isolate — move to module |
| `Imports []interface{}` | Similarly typed — move interface to module |
| `Delegates []interface{}` | Has `delegator` interface in isolate — move to module |
| `Updates []interface{}` | Mixed AST types — leave as-is |
| `Natives`, `NativeDefinitions` | Compiler-specific — leave as-is |
| `Progress`, `Rely`, `MixOrd` | Temporal logic — leave as-is |
| `ConceptSpaces`, `AbstrPreds` | Rarely used — leave as-is |
| `Macros`, `Attributes`, `BeforeExport` | Generic maps — leave as-is |
| `ParamDefaults` | `ast.Node` or nil : convert to `[]ast.Node` |

### Group G: art.State fields

| Field | Assessment |
|-------|-----------|
| `State.Value interface{}` | Assigned `*transrel.Update` during BMC — could type it |
| `State.Universe interface{}` | Solver-specific map — leave as-is |
| `Counterexample.Conc interface{}` | Varies by checker — leave as-is |

---

## Recommended Execution Order

### Phase 1: Zero-risk signature fixes (Group C)
- Change `AdmitDefinitionFn` and `AdmitDefinitionFactory` param from `interface{}` to `ast.Node`
- Files: `module/module.go`, `compiler/ivy_compile.go`, `check/check.go`
- ~5 lines changed, removes 1 type assertion

### Phase 2: AST collection typing (Group B, partial)
- Change `Schemata`, `Theorems`, `Interps` from `interface{}` to `*ast.LabeledFormula`
- Files: `module/module.go`, `module/new.go`, `compiler/`, `isolate/`, `check/`
- Removes ~10 type assertions

### Phase 3: Move `Action` interface to `module` (Group A)
- Move `Action` interface + `ActionBase` struct to `module/action.go`
- In `actions/action.go`: `type Action = module.Action`, `type ActionBase = module.ActionBase`
- Change `module.Actions` to `map[string]module.Action`
- Change `module.CompileActionBodyFn` return to `(module.Action, error)` (Group D)
- Update `art.Actions`, `art.Predicates`, etc.
- Removes ~50 type assertions across the codebase

### Phase 4: Move isolate interfaces to `module` (Group B remainder + F)
- Move `IsolateDefInterface`, `MixinDef`, `exporter`, `delegator` interfaces to module
- Type `module.Isolates`, `module.Mixins`, `module.Exports`, `module.Delegates`

### Phase 5: Remaining fixes
- `module.ParamDefaults` → `[]ast.Node`
- `art.State.Value` → `*transrel.Update`
- Others as warranted

---

## Verification

After each phase:
1. `go build ./...` — compiles
2. `go vet ./...` — no warnings
3. `go test ./art/... ./actions/... ./module/... ./compiler/... ./check/... ./isolate/...` — passes
4. `go test -tags web ./webui/...` — passes
5. Grep for remaining `interface{}` to track progress: `grep -rn 'interface{}' module/module.go | wc -l`

---

## Files to Modify (by phase)

**Phase 1:** `module/module.go`, `compiler/ivy_compile.go`, `check/check.go`
**Phase 2:** `module/module.go`, `module/new.go`, `compiler/phase6.go`, `isolate/helpers.go`, `check/check.go`
**Phase 3:** New `module/action.go`, `actions/action.go`, `module/module.go`, `art/art.go`, `isolate/strip.go`, `trace/trace.go`, `check/isolate_check.go`, `bmc/bmc.go`, + ~10 more files with type assertions
**Phase 4:** `module/module.go`, `isolate/isolate.go`, `isolate/helpers.go`, `isolate/create.go`
