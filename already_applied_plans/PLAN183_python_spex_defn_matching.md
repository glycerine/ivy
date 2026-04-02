# Plan: Comprehensive CanonSnapshot() — trace ALL Module fields

**Created**: 2026-04-02 03:00

## Context

After the golden test XTRACE traces match, `CanonSnapshot()` (Go: `module/canon.go:57`, Python: `ivy_module.py:310`) only traces 11 of ~65 Module data fields. This leaves massive gaps in cross-language conformance testing. The goal is to trace **every** field so `diff out.py.xtrace out.go.xtrace` catches any data divergence.

## Fields already traced (11)

| Field | Helper |
|---|---|
| LabeledAxioms | canonLFSlice |
| Definitions | canonLFSlice |
| LabeledProps | canonLFSlice |
| LabeledInits | canonLFSlice |
| LabeledConjs | canonLFSlice |
| Sig.Sorts | canonSortMap |
| Sig.Interp | canonInterpMap |
| Schemata | canonSchemaMap |
| InitCond | .Canon() |
| Theory | .Canon() |
| Actions | canonActionMap + per-action Sexp |

## Fields to skip (runtime/internal only)

| Go Field | Reason |
|---|---|
| Cfg | Config object |
| CompCfg | Compiler config |
| CompileActionBodyFn | Function callback |
| AdmitDefinitionFn | Function callback |
| Instantiator | Function callback |
| prevModule | Context management |
| oldSig | Context management |
| VPrivates | Go-only (not in Python) |
| Name | Go-only (Python Module has no `self.name`) |
| Macros | Go-only (Python `macros` is on parser, not module) |

## New helper functions needed in Go (`module/canon.go`)

Each Python counterpart is `_snake_case` version in `ivy_module.py`.

### Reusable for multiple fields

| Helper | Signature | Format | Used for |
|---|---|---|---|
| `canonExprSlice` | `(exprs []lg.Expr) string` | `[sexp1 sexp2]` or `[]` | AllRelations, Rely |
| `canonActionSlice` | `(actions []Action) string` | `[sexp1 sexp2]` or `[]` | InitialActions |
| `canonBoolMap` | `(m map[string]bool) string` | `(set k1 k2)` or `(set)` sorted | GhostSorts, FiniteSorts, Privates |
| `canonStringSlice` | `(ss []string) string` | `["s1" "s2"]` or `[]` | SortOrder, Logics |
| `canonConstSlice` | `(cs []*lg.Const) string` | `[sexp1 sexp2]` or `[]` | Params, SymbolOrder |
| `canonNodeSlice` | `(nodes []ast.Node) string` | `[canon1 canon2]` or `[]`; nil→`nil` | Natives, MixOrd, Imports, ParamDefaults |
| `canonInsMapSort` | `(m *iu.InsMap[string, lg.Sort]) string` | `(insMap k:sexp ...)` or `(insMap)` | Relations, Functions |
| `canonInsMapBool` | `(m *iu.InsMap[string, bool]) string` | `(insMap k1 k2)` or `(insMap)` | PublicActions |
| `canonSortSliceMap` | `(m map[string][]lg.Sort) string` | `(hash k:[sexp ...])` sorted | Variants, Supertypes |
| `canonConstSliceMap` | `(m map[string][]*lg.Const) string` | `(hash k:[sexp ...])` sorted | SortDestructors, SortConstructors |
| `canonStringMap` | `(m map[string]string) string` | `(hash k:"v" ...)` sorted | Aliases |
| `canonStringSliceMap` | `(m map[string][]string) string` | `(hash k:["s1" "s2"])` sorted | ConjActions |
| `canonExprMap` | `(m map[string]lg.Expr) string` | `(hash k:sexp ...)` sorted | ExtPreconds |
| `canonInterfaceSlice` | `(items []interface{}) string` | `[canon1 ...]`; Canonizer→Canon(), else %v | Updates, Progress |
| `canonNode` | `(n ast.Node) string` | Canonizer→Canon(), else %v; nil→`nil` | IsolateProof |

### Specialized for specific struct types

| Helper | Used for |
|---|---|
| `canonPostcondsMap(m map[string][]*ast.LabeledFormula) string` | Postconds |
| `canonInsMapMixins(m *iu.InsMap[string, []MixinDef]) string` | Mixins |
| `canonInsMapHierarchy(m *iu.InsMap[string, *iu.InsMap[string, bool]]) string` | Hierarchy |
| `canonNodeMapSlice(m map[string][]ast.Node) string` | Interps |
| `canonNamedActionSlice(nas []NamedAction) string` | Initializers |
| `canonInstantiationSlice(insts []Instantiation) string` | Instantiations |
| `canonIsolateMap(m map[string]*ast.IsolateDef) string` | Isolates |
| `canonIsolateInfo(info *IsolateInfo) string` | IsolateInfo |
| `canonExporterSlice(exports []Exporter) string` | Exports |
| `canonDelegatorSlice(delegates []Delegator) string` | Delegates |
| `canonNativeTypeMap(m map[string]*ast.NativeType) string` | NativeTypes |
| `canonConceptSpaceSlice(css []ConceptSpace) string` | ConceptSpaces |
| `canonProofEntrySlice(proofs []ProofEntry) string` | Proofs |
| `canonNamedEntrySlice(named []NamedEntry) string` | Named |
| `canonSubgoalEntrySlice(subs []SubgoalEntry) string` | Subgoals |

### Existing helpers reused for new fields

| Existing helper | New field(s) |
|---|---|
| `canonLFSlice` | Assertions, AssumedInvs, NativeDefinitions, ConjSubgoals |
| `canonSortMap` | DestructorSorts, ConstructorSorts |
| `canonSchemaMap` | Theorems, Predicates, IsolateProofs |
| `canonActionMap` | BeforeExport |
| `canonInterpMap` | Attributes (same `map[string]interface{}` type) |

## CanonSnapshot trace order (Go)

Complete ordered list of all trace lines. `[E]` = existing, `[N]` = new.

```
// ENTER
"module.CanonSnapshot ENTER label=%s"                                          [E]

// Group 1: Declarations (labeled formula slices)
"module.CanonSnapshot %s labeledAxioms=%s"     canonLFSlice(LabeledAxioms)     [E]
"module.CanonSnapshot %s definitions=%s"       canonLFSlice(Definitions)       [E]
"module.CanonSnapshot %s labeledProps=%s"      canonLFSlice(LabeledProps)      [E]
"module.CanonSnapshot %s labeledInits=%s"      canonLFSlice(LabeledInits)      [E]
"module.CanonSnapshot %s labeledConjs=%s"      canonLFSlice(LabeledConjs)      [E]
"module.CanonSnapshot %s assertions=%s"        canonLFSlice(Assertions)        [N]
"module.CanonSnapshot %s assumedInvs=%s"       canonLFSlice(AssumedInvs)       [N]
"module.CanonSnapshot %s nativeDefinitions=%s" canonLFSlice(NativeDefinitions) [N]
"module.CanonSnapshot %s conjSubgoals=%s"      canonLFSlice(ConjSubgoals)      [N] nil-guarded

// Group 2: All relations
"module.CanonSnapshot %s allRelations=%s"      canonExprSlice(AllRelations)    [N]

// Group 3: Signature
"module.CanonSnapshot %s sig.sorts=%s"         canonSortMap(Sig.Sorts)         [E]
"module.CanonSnapshot %s sig.interp=%s"        canonInterpMap(Sig.Interp)      [E]

// Group 4: Relations and functions
"module.CanonSnapshot %s relations=%s"         canonInsMapSort(Relations)      [N]
"module.CanonSnapshot %s functions=%s"         canonInsMapSort(Functions)      [N]

// Group 5: Schemata, theorems, predicates
"module.CanonSnapshot %s schemata=%s"          canonSchemaMap(Schemata)        [E]
"module.CanonSnapshot %s theorems=%s"          canonSchemaMap(Theorems)        [N]
"module.CanonSnapshot %s predicates=%s"        canonSchemaMap(Predicates)      [N]

// Group 6: InitCond and Theory
"module.CanonSnapshot %s initCond=%s"          InitCond.Canon()                [E]
"module.CanonSnapshot %s theory=%s"            Theory.Canon()                  [E]

// Group 7: Actions
"module.CanonSnapshot %s actions.keys=%d keys=%s" + per-action lines           [E]
"module.CanonSnapshot %s beforeExport=%s"      canonActionMap(BeforeExport)    [N]

// Group 8: Mixins, PublicActions
"module.CanonSnapshot %s mixins=%s"            canonInsMapMixins(Mixins)       [N]
"module.CanonSnapshot %s publicActions=%s"     canonInsMapBool(PublicActions)  [N]

// Group 9: Postconds
"module.CanonSnapshot %s postconds=%s"         canonPostcondsMap(Postconds)    [N]

// Group 10: Initializers
"module.CanonSnapshot %s initializers=%s"      canonNamedActionSlice(Initializers) [N]
"module.CanonSnapshot %s initialActions=%s"    canonActionSlice(InitialActions)    [N]

// Group 11: Hierarchy
"module.CanonSnapshot %s hierarchy=%s"         canonInsMapHierarchy(Hierarchy) [N]

// Group 12: Updates, Instantiations
"module.CanonSnapshot %s updates=%s"           canonInterfaceSlice(Updates)    [N]
"module.CanonSnapshot %s instantiations=%s"    canonInstantiationSlice(Instantiations) [N]

// Group 13: Isolates
"module.CanonSnapshot %s isolates=%s"          canonIsolateMap(Isolates)       [N]
"module.CanonSnapshot %s isolateInfo=%s"       canonIsolateInfo(IsolateInfo)   [N] nil-guarded
"module.CanonSnapshot %s isolateProofs=%s"     canonSchemaMap(IsolateProofs)   [N]
"module.CanonSnapshot %s isolateProof=%s"      canonNode(IsolateProof)         [N] nil-guarded

// Group 14: Exports, imports, delegates
"module.CanonSnapshot %s exports=%s"           canonExporterSlice(Exports)     [N]
"module.CanonSnapshot %s imports=%s"           canonNodeSlice(Imports)         [N]
"module.CanonSnapshot %s delegates=%s"         canonDelegatorSlice(Delegates)  [N]

// Group 15: Sorts and destructors
"module.CanonSnapshot %s destructorSorts=%s"   canonSortMap(DestructorSorts)   [N]
"module.CanonSnapshot %s sortDestructors=%s"   canonConstSliceMap(SortDestructors) [N]
"module.CanonSnapshot %s constructorSorts=%s"  canonSortMap(ConstructorSorts)  [N]
"module.CanonSnapshot %s sortConstructors=%s"  canonConstSliceMap(SortConstructors) [N]
"module.CanonSnapshot %s ghostSorts=%s"        canonBoolMap(GhostSorts)        [N]
"module.CanonSnapshot %s sortOrder=%s"         canonStringSlice(SortOrder)     [N]
"module.CanonSnapshot %s symbolOrder=%s"       canonConstSlice(SymbolOrder)    [N]
"module.CanonSnapshot %s variants=%s"          canonSortSliceMap(Variants)     [N]
"module.CanonSnapshot %s supertypes=%s"        canonSortSliceMap(Supertypes)   [N]
"module.CanonSnapshot %s finiteSorts=%s"       canonBoolMap(FiniteSorts)       [N]

// Group 16: Interpretations and natives
"module.CanonSnapshot %s interps=%s"           canonNodeMapSlice(Interps)      [N]
"module.CanonSnapshot %s natives=%s"           canonNodeSlice(Natives)         [N]
"module.CanonSnapshot %s nativeTypes=%s"       canonNativeTypeMap(NativeTypes) [N]

// Group 17: Properties and proofs
"module.CanonSnapshot %s progress=%s"          canonInterfaceSlice(Progress)   [N]
"module.CanonSnapshot %s rely=%s"              canonExprSlice(Rely)            [N]
"module.CanonSnapshot %s mixOrd=%s"            canonNodeSlice(MixOrd)          [N]
"module.CanonSnapshot %s privates=%s"          canonBoolMap(Privates)          [N]
"module.CanonSnapshot %s proofs=%s"            canonProofEntrySlice(Proofs)    [N]
"module.CanonSnapshot %s named=%s"             canonNamedEntrySlice(Named)     [N]
"module.CanonSnapshot %s subgoals=%s"          canonSubgoalEntrySlice(Subgoals) [N]
"module.CanonSnapshot %s conjActions=%s"       canonStringSliceMap(ConjActions) [N]

// Group 18: Parameters
"module.CanonSnapshot %s params=%s"            canonConstSlice(Params)         [N]
"module.CanonSnapshot %s paramDefaults=%s"     canonNodeSlice(ParamDefaults)   [N]

// Group 19: Other
"module.CanonSnapshot %s aliases=%s"           canonStringMap(Aliases)         [N]
"module.CanonSnapshot %s attributes=%s"        canonInterpMap(Attributes)      [N]
"module.CanonSnapshot %s extPreconds=%s"       canonExprMap(ExtPreconds)       [N]
"module.CanonSnapshot %s conceptSpaces=%s"     canonConceptSpaceSlice(ConceptSpaces) [N]
"module.CanonSnapshot %s logics=%s"            canonStringSlice(Logics)        [N]

// EXIT
"module.CanonSnapshot EXIT label=%s"                                           [E]
```

## Python differences

| Go | Python | Note |
|---|---|---|
| `map[string]bool` | `set()` | GhostSorts, FiniteSorts, Privates are Python sets; iterate `sorted(s)` |
| `*iu.InsMap[K,V]` | `dict` | Python dicts are insertion-ordered since 3.7 |
| `[]MixinDef` | `list` of mixin objects | MixinDef has `.canon()` from canon_ast.py |
| `ConjSubgoals` nil | `conj_subgoals` None | Guard with `if self.conj_subgoals is not None` |
| `Name`, `Macros` | not on Python Module | Skip these traces in Python |
| `abstraction_predicates` | exists in Python | Commented out in Go (`AbstrPreds`); skip for now |

## Implementation steps

### Step 1: Go helpers (`module/canon.go`)

Add ~30 new helper functions after the existing helpers. Each is small (5-15 lines). Group:
- Generic: canonExprSlice, canonActionSlice, canonBoolMap, canonStringSlice, canonConstSlice, canonNodeSlice, canonNode, canonInterfaceSlice
- Map variants: canonInsMapSort, canonInsMapBool, canonInsMapMixins, canonInsMapHierarchy, canonPostcondsMap, canonSortSliceMap, canonConstSliceMap, canonStringMap, canonStringSliceMap, canonNodeMapSlice, canonExprMap, canonNativeTypeMap, canonIsolateMap
- Struct-specific: canonNamedActionSlice, canonInstantiationSlice, canonIsolateInfo, canonExporterSlice, canonDelegatorSlice, canonConceptSpaceSlice, canonProofEntrySlice, canonNamedEntrySlice, canonSubgoalEntrySlice

### Step 2: Go CanonSnapshot expansion (`module/canon.go`)

Replace the existing CanonSnapshot with the full version shown above.

### Step 3: Python helpers (`ivy_module.py`)

Add ~28 matching `_snake_case` functions after the existing `_canon_action_map` (around line 427).

### Step 4: Python canon_snapshot expansion (`ivy_module.py`)

Expand `canon_snapshot()` with all matching trace lines.

### Step 5: Also update `Module.Canon()` (`module/canon.go:17`)

The `Canon()` method should also be expanded to include all fields, not just the original 11, so that Merkle-chaining covers everything.

## Key files to modify

- `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/module/canon.go` — all Go helpers + CanonSnapshot + Canon
- `/Users/jaten/ivy/pyivy/ivy/ivy/ivy_module.py` — all Python helpers + canon_snapshot + canon

## Verification

```bash
cd ~/ivy/goivy && make test && make golden
diff out.py.xtrace out.go.xtrace
```

The CanonSnapshot traces in both outputs should match exactly for every field. Any divergence reveals a data conformance bug.

---

## Bug Fix: Hierarchy data divergence (golden test line 102790)

**Created**: 2026-04-02

### Symptom

`make golden` fails at line 102790 — Go's `hierarchy["this"]` InsMap contains extra entries (`base step fun p x n z y f1 f2 x1 x2 x3 x4`) that Python's `hierarchy["this"]` dict does not have.

### Root Cause

Go's `Compiler.AddSymbol()` (`compiler/compiler.go:1138`) calls `c.Module.AddToHierarchy(name)` after adding a symbol to the signature. Python's `add_symbol()` (`ivy_logic.py:378-380`) does NOT — it only calls `sig.add_symbol()`.

This means in Go, every symbol added via `AddSymbol` (including quantified variables like `base`, `step`, `fun`, `p`, `x`, `n`, `z`, `y`, `f1`, `f2`, `x1`-`x4`) gets erroneously inserted into the hierarchy. In Python, only names from `decls.defined` (processed in `ivy_compile()` at line 2494) go into the hierarchy.

### Call site comparison

| | Go | Python |
|---|---|---|
| `AddSymbol`/`add_symbol` | calls `AddToHierarchy` ← **BUG** | does NOT call `add_to_hierarchy` |
| `ivy_compile` main loop | calls `AddToHierarchy` for `decls.defined` | calls `add_to_hierarchy` for `decls.defined` |

### Fix

**File**: `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/compiler/compiler.go`
**Line**: 1138

Remove the erroneous `c.Module.AddToHierarchy(name)` call from `AddSymbol()`:

```go
// BEFORE (buggy):
func (c *Compiler) AddSymbol(name string, sort lg.Sort, sig *il.Sig) (*lg.Const, error) {
    sym, err := sig.AddSymbol(name, sort)
    if err != nil {
        return nil, err
    }
    c.Module.AddToHierarchy(name)  // ← NOT in Python, causes extra hierarchy entries
    return sym, nil
}

// AFTER (matches Python):
func (c *Compiler) AddSymbol(name string, sort lg.Sort, sig *il.Sig) (*lg.Const, error) {
    sym, err := sig.AddSymbol(name, sort)
    if err != nil {
        return nil, err
    }
    return sym, nil
}
```

### Verification

```bash
cd ~/ivy/goivy && make test && make golden
```

The hierarchy line at 102790 should now match. Any remaining divergences in subsequent CanonSnapshot fields will appear as the next `log.red` failure.

---

## Bug Fix: Updates field divergence (golden test line 102792)

**Created**: 2026-04-02

### Symptom

`make golden` fails at line 102792 (0-indexed: 102791). Go emits structured s-expressions for `updates=` like `(DerivedUpdate symbol:(Symbol ...) defn:(Def ...))`. Python emits human-readable `str()` output like `index.max2(X,Y) = (Y:index if (X <= Y) else X)`.

### Root Cause

Python's `DerivedUpdate` and `NamedUpdate` (in `ivy_actions.py`) lack `canon()` methods. When `_canon_interface_slice()` processes the `updates` list, it falls through to `str()` for these types.

Both Go and Python already have the necessary sexp/canon machinery at the logic level:
- `ivy_logic.Definition.sexp()` → `(Def lhs:... rhs:...)` (Python: `logic_sexp.py:113`)
- `lg.Definition.Sexp()` → `(Def lhs:... rhs:...)` (Go: `logic/sexp.go:170`)

These formats match exactly. The issue is just that Python's wrapper types don't call into this machinery.

### Structural difference between Go and Python

| | Go `DerivedUpdate` | Python `DerivedUpdate` |
|---|---|---|
| Fields | `Symbol lg.Expr` + `Defn lg.Expr` (two separate fields) | `self.defn` only (one `ivy_logic.Definition`) |
| Source | `module/derivedupdate.go` | `ivy_actions.py:166` |
| Note | Symbol is redundant — already in `Defn.Lhs` | Python is authoritative; stores just the definition |

| | Go `NamedUpdate` | Python `NamedUpdate` |
|---|---|---|
| Fields | `UpdateName string` + `Body lg.Expr` | `self.sym` (string) + `self.dependencies` (computed) |
| Source | `actions/action.go:1940` | `ivy_actions.py:178` |
| Note | Go stores `Body`; Python discards `fmla` after computing deps | Canon should match Python's stored data |

### Fix — 4 changes

#### 1. Add `canon()` to Python `DerivedUpdate` (in `canon_ast.py`)

The `defn` field is `ivy_logic.Definition` which has `.sexp()` from `logic_sexp.py:208`.

```python
# In canon_ast.py install() function:
def _derivedupdate_canon(self):
    defn_str = self.defn.sexp() if hasattr(self.defn, 'sexp') else str(self.defn)
    return '(DerivedUpdate defn:%s)' % defn_str
act.DerivedUpdate.canon = _derivedupdate_canon
```

#### 2. Add `canon()` to Python `NamedUpdate` (in `canon_ast.py`)

```python
def _namedupdate_canon(self):
    return '(NamedUpdate sym:"%s")' % str(self.sym)
act.NamedUpdate.canon = _namedupdate_canon
```

#### 3. Change Go `DerivedUpdate.Sexp()` to match Python format

**File**: `module/derivedupdate.go:116-124`

Drop the redundant `symbol:` field (Python doesn't have it), emit only `defn:`:

```go
func (a *DerivedUpdate) Sexp() lg.NodeKey {
    defn := "nil"
    if a.Defn != nil {
        defn = string(a.Defn.Sexp())
    }
    return lg.NodeKey(fmt.Sprintf("(DerivedUpdate defn:%v)", defn))
}
```

#### 4. Change Go `NamedUpdate.Sexp()` to match Python format

**File**: `actions/action_expr.go:827-829`

Drop the `ActionBase.CanonFields()` and `elems:` (Python NamedUpdate has no ActionBase and discards the body):

```go
func (a *NamedUpdate) Sexp() lg.NodeKey {
    return lg.NodeKey(fmt.Sprintf("(NamedUpdate sym:\"%s\")", a.UpdateName))
}
```

### Key files to modify

- `/Users/jaten/ivy/pyivy/ivy/ivy/canon_ast.py` — add `DerivedUpdate.canon` and `NamedUpdate.canon`
- `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/module/derivedupdate.go` — simplify `Sexp()`
- `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/actions/action_expr.go` — simplify `NamedUpdate.Sexp()`

### Verification

```bash
cd ~/ivy/goivy && make test && make golden
```

The updates line at 102792 should now match.
