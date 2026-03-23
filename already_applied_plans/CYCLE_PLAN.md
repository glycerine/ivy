# Implementation Plan: Eliminate interface{} Import-Cycle Kludges

## Context

The design in `~/goivy/CYCLE_MOVES.md` catalogs `interface{}` kludges caused by Go import cycle avoidance. This plan specifies the exact mechanical steps to implement each phase. No backward-compat aliases — every reference is updated to the canonical location.

---

## Phase 1: Zero-risk signature fixes (Group C)

**Goal:** Change `proof interface{}` → `proof ast.Node` in the AdmitDefinition callback chain.

### Steps

1. **`module/module.go:119`** — Change:
   ```go
   AdmitDefinitionFn func(defn *ast.LabeledFormula, proof interface{}) error
   ```
   to:
   ```go
   AdmitDefinitionFn func(defn *ast.LabeledFormula, proof ast.Node) error
   ```

2. **`compiler/ivy_compile.go:45`** — Change:
   ```go
   var AdmitDefinitionFactory func(mod *module.Module) func(defn *ast.LabeledFormula, proof interface{}) error
   ```
   to:
   ```go
   var AdmitDefinitionFactory func(mod *module.Module) func(defn *ast.LabeledFormula, proof ast.Node) error
   ```

3. **`check/check.go:30-41`** — Change callback signature and remove type assertion:
   ```go
   // Old:
   return func(defn *ast.LabeledFormula, pf interface{}) error {
       ...
       pfNode, _ := pf.(ast.Node)
       _, err := prover.AdmitDefinition(defn, pfNode)

   // New:
   return func(defn *ast.LabeledFormula, pf ast.Node) error {
       ...
       _, err := prover.AdmitDefinition(defn, pf)
   ```

**Files:** `module/module.go`, `compiler/ivy_compile.go`, `check/check.go`
**Verify:** `go build ./... && go test ./check/... ./compiler/...`

---

## Phase 2: AST collection typing (Group B partial)

**Goal:** Type `Schemata`, `Theorems`, `Interps` with concrete AST types.

### Steps

1. **`module/module.go`** — Change field declarations:
   ```go
   // Old:
   Schemata       map[string]interface{}
   Theorems       map[string]interface{}
   Interps        map[string][]interface{}

   // New:
   Schemata       map[string]*ast.LabeledFormula
   Theorems       map[string]*ast.LabeledFormula
   Interps        map[string][]ast.Node
   ```
   Note: `Interps` stores `ast.Node` (not `*ast.LabeledFormula`) per compiler/decl.go usage.

2. **`module/new.go`** (or wherever `New()` initializes maps) — Update `make()` calls to match new types.

3. **`module/module.go` Clone method** — Update clone logic (around line 289+) to remove type assertions when copying these maps.

4. **`compiler/decl.go`** — Remove type assertions at all Schemata/Theorems/Interps assignment sites (~10 locations). Some may need explicit `ast.Node` → `*ast.LabeledFormula` conversion where the compiler creates LabeledFormula objects.

5. **`compiler/ivy_compile.go`** — Update Schemata assignments (~2 locations).

6. **`compiler/phase6.go`** — Update Schemata assignments (~2 locations).

7. **`check/check.go:32-37`** — Remove the `typedSchemata` conversion loop entirely — `mod.Schemata` is now already `map[string]*ast.LabeledFormula`.

8. **`proof/checker.go`** — Update any Schemata lookups/type assertions (~5 locations).

9. **`actions/extra_actions.go:800`** — Remove `if mlf, ok := schema.(*ast.LabeledFormula)` assertion.

10. **`isolate/helpers.go`** — Update Interps iteration (lines 401-405) to remove `itp.(*ast.LabeledFormula)` assertions.

**Files:** `module/module.go`, `module/new.go`, `compiler/decl.go`, `compiler/ivy_compile.go`, `compiler/phase6.go`, `check/check.go`, `proof/checker.go`, `actions/extra_actions.go`, `isolate/helpers.go`
**Verify:** `go build ./... && go test ./module/... ./compiler/... ./check/... ./proof/... ./isolate/...`

---

## Phase 3: Move `Action` interface to `module` (Group A + D)

**Goal:** Move `Action` interface, `ActionBase` struct, `ActionNodeWrapper`, `WrapAction`, `UnwrapAction` from `actions/` to `module/`. Type `module.Actions` as `map[string]Action`. No aliases.

### Step 3a: Create `module/action.go`

New file containing (moved from `actions/action.go`):

- `Action` interface (lines 20-52)
- `ActionBase` struct (lines 56-62) + all methods (lines 64-85)
- `CopyFormalsTo` method
- `ActionNodeWrapper` struct (lines 1187-1198) + methods
- `WrapAction()` func (line 1201)
- `UnwrapAction()` func (line 1207)

These only depend on `lg.Expr`, `lg.Symbol`, `ast.Location`, `ast.Base`, `ast.Node`, `lg.Sort` — all already imported by `module`.

### Step 3b: Remove from `actions/action.go`

Delete the moved types/functions. All concrete action types (Sequence, AssumeAction, etc.) stay in `actions/` and now implement `module.Action`. They embed `module.ActionBase` instead of `actions.ActionBase`.

### Step 3c: Update `module/module.go` fields

```go
// Old:
Actions        map[string]interface{}
InitialActions []interface{}
CompileActionBodyFn func(node ast.Node) (interface{}, error)

// New:
Actions        map[string]Action
InitialActions []Action
CompileActionBodyFn func(node ast.Node) (Action, error)
```

### Step 3d: Update `module/new.go`

Change `make(map[string]interface{})` → `make(map[string]Action)` for Actions map initialization.

### Step 3e: Mass rename `actions.Action` → `module.Action`

Every file referencing `actions.Action` as a type needs updating. From the exploration: **~338 references across 19 packages.** This is the bulk of the work. Key files:

| Package | Files to update |
|---------|----------------|
| `actions/` (8 files) | `action.go`, `extra_actions.go`, `helpers.go`, `match.go`, `phase3.go`, `transforms.go`, `update.go`, `annotation.go` — concrete types now embed `module.ActionBase`, implement `module.Action` |
| `isolate/` (8 files) | `isolate.go`, `helpers.go`, `phase7.go`, `create.go`, `strip.go`, `deps.go`, `iter.go` |
| `compiler/` (6 files) | `phase6.go`, `helpers.go`, `ivy_compile.go`, `action.go`, `decl.go` + tests |
| `interp/` (4 files) | `eval.go`, `interp.go`, `helpers.go`, `phase4.go` |
| `art/` (1 file) | `art.go` — change `Actions map[string]interface{}` → `map[string]module.Action` etc. |
| `check/` (3 files) | `helpers.go`, `isolate_check.go` + tests |
| `trace/` | `trace.go` |
| `bmc/` | `bmc.go` |
| `temporal/` | `temporal.go` + test |
| `tactics/` | `tactics.go` |
| `gogen/`, `cppgen/`, `leangen/`, `mc/`, `l2s/`, `vmt/`, `ranking/`, `printer/`, `fragment/` | 1-3 files each |

### Step 3f: Remove type assertions

Everywhere that does `x.(actions.Action)` on a value from `module.Actions` map — the assertion is now unnecessary since the map is typed. Delete the assertion and use the value directly. This eliminates ~50 type assertions.

### Step 3g: Update `actions/extra_actions.go` CompileActionBodyFn usage

The wrapper at line 754-767 that converts `interface{}` to `Action` via type assertion is no longer needed — the callback now returns `module.Action` directly.

**Files:** `module/action.go` (new), `actions/action.go`, `module/module.go`, `module/new.go`, `art/art.go`, + ~30 files across 19 packages
**Verify:** `go build ./... && go vet ./... && go test ./...`

---

## Phase 4: Move isolate interfaces + CompilerConfig to `module` (Groups B remainder, E, F)

### Step 4a: Move interfaces to `module/ifaces.go` (new file)

From `isolate/`:
- `IsolateDefInterface` (iter.go:69-77) — only uses `[]string`, `bool`
- `MixinDef` (isolate.go:167-175) — only uses `string`, `bool`
- `Exporter` (local in 3 files, consolidate) — only uses `string`
- `Delegator` (local in 3 files, consolidate) — only uses `string`

Rename to exported names: `IsolateDefInterface`, `MixinDef`, `Exporter`, `Delegator`.

### Step 4b: Type module fields

```go
// Old:
Isolates      map[string]interface{}
Mixins        map[string][]interface{}
Exports       []interface{}
Delegates     []interface{}

// New:
Isolates      map[string]IsolateDefInterface
Mixins        map[string][]MixinDef
Exports       []Exporter
Delegates     []Delegator
```

### Step 4c: Move `CompilerConfig` to `module/compiler_config.go` (new file)

From `compiler/phase6.go:2213-2238`:
```go
type CompilerConfig struct {
    OptionVerifying bool
    PropIDCounter   int64
    ModCfg          *Config  // was *module.Config, now self-referential
}
```
Plus methods: `NewCompilerConfig`, `FreshPropID`, `SetVerifying`, `GetVerifying`.

Update `module/module.go:109`:
```go
// Old:
CompCfg interface{}

// New:
CompCfg *CompilerConfig
```

### Step 4d: Update isolate/ package

Remove local `exporter`, `delegator`, `MixinDef`, `IsolateDefInterface` definitions. Import from `module`. Remove type assertions on `mod.Isolates`, `mod.Mixins`, `mod.Exports`, `mod.Delegates`.

### Step 4e: Update compiler/ package

Remove `CompilerConfig` from `phase6.go`. Import from `module`. Remove `mod.CompCfg.(*CompilerConfig)` type assertions — use `mod.CompCfg` directly.

**Files:** `module/ifaces.go` (new), `module/compiler_config.go` (new), `module/module.go`, `module/new.go`, `isolate/isolate.go`, `isolate/helpers.go`, `isolate/create.go`, `isolate/iter.go`, `isolate/deps.go`, `isolate/strip.go`, `compiler/phase6.go` + tests
**Verify:** `go build ./... && go vet ./... && go test ./isolate/... ./compiler/...`

---

## Phase 5: Remaining low-value fixes (Groups F remainder, G)

### Step 5a: `ParamDefaults` → `[]ast.Node`

Change in `module/module.go:95`:
```go
// Old:
ParamDefaults []interface{}
// New:
ParamDefaults []ast.Node
```
Update: `module/new.go`, `isolate/strip.go:338`, `compiler/decl.go:1449`, `cppgen/repl.go:334`.

### Step 5b: `art.State.Value` → `*transrel.Update`

Change in `art/art.go:38`:
```go
// Old:
Value    interface{}
// New:
Value    *transrel.Update
```
Update BMC assignment code in `art/art.go`.

### Step 5c: `Predicates` → `map[string]ast.Node`

Change in `module/module.go:42` and `art/art.go` — Predicates stores AST nodes.

**Files:** `module/module.go`, `art/art.go`, `isolate/strip.go`, `compiler/decl.go`, `cppgen/repl.go`
**Verify:** `go build ./... && go test ./...`

---

## Verification (after all phases)

1. `go build ./...` — full project compiles
2. `go vet ./...` — no warnings
3. `go test ./...` — all tests pass
4. `go test -tags web ./webui/...` — webui tests pass
5. Count remaining `interface{}` in module.go: `grep -c 'interface{}' module/module.go`
6. Spot-check: no `.(actions.Action)` type assertions remain on module.Actions values

---

## Package Dependency Layers (reference)

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
