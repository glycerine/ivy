# Solo 7: Main `ivy_compile` integration (§6.3 #31)

## Context

This is the final integration task from SECTION6_PLAN.md. The `IvyCompile` function in
`compiler/ivy_compile.go` orchestrates the three-pass compilation pipeline. Most pieces
are already implemented (Batches A–G, Solos 1–6 are done). This task closes the remaining
gaps to make `IvyCompile` faithfully match Python's `ivy_compile` (ivy_compiler.py:2190–2254).

## Gaps Found

### Gap 1: Missing version check for "this" isolate creation (`ivy_compile.go:144–152`)
- **Python** (line 2220–2225): `if not iu.version_le(iu.get_string_version(),"1.6"):` — only creates the default "this" isolate when version > 1.6
- **Go**: Unconditionally creates "this" isolate. Comment says "From version 1.7" but the guard is missing.
- **Fix**: Add `if !iu.VersionLE(iu.GetStringVersion(), "1.6")` guard. Delete the inaccurate "From version 1.7" comment, replace with the Python-faithful version check.

### Gap 2: `mod.type_check()` not called (`ivy_compile.go:135–137`)
- **Python** (line 2211): `mod.type_check()` → calls `type_check_list(self, self.axioms)` + `self.type_check_concepts()`
- **Go**: Comment says "Currently delegated to interp.ModuleTypeCheck (stub)" — but it's never actually called.
- **Existing utilities**:
  - `interp.ModuleTypeCheck(mod)` at `interp/helpers.go:422` — exists but is a **partial stub** (only nil-checks instead of real arity checking)
  - `actions.TypeCheck(domain, node)` at `actions/phase3.go:112` — **fully implemented** arity checker
  - `interp.TypeCheckList(domain, items)` at `interp/phase4.go:24` — **fully implemented** recursive list type checker
- **Fix (two parts)**:
  - **Part A**: Fix `interp.ModuleTypeCheck` to call real `TypeCheckList` on axioms (not just nil-check), and fix `ModuleTypeCheckConcepts` to do real arity checking with concept space relations (matching Python)
  - **Part B**: Call `interp.ModuleTypeCheck(mod)` in `IvyCompile` at line 135

### Gap 3: Python assertions at lines 2216–2218 not ported
- **Python**: After `type_check_action(action, mod)`:
  ```python
  if not hasattr(action,'lineno'):
      print("no lineno: {}".format(name))
  assert hasattr(action,'formal_params'), action
  ```
- **Go**: No equivalent checks.
- **Fix**: Add warning for actions missing lineno, and a panic/error check for actions without FormalParams. These are debugging safety nets that should be ported faithfully.

### Gap 4: `Concept` handler uses `GetRelationSort` without term (`decl.go:1430`)
- **Python** (line 1210): `add_symbol(rel.relname, get_relation_sort(sig, rel.args, c.args[1]))` — passes body as term for joint sort inference
- **Go** (line 1430): `d.Compiler.GetRelationSort(atom.Terms)` — no term, same bug pattern as Progress (fixed in Batch G)
- **Fix**: Use `GetRelationSortWithTerm(atom.Terms, body)` (already implemented in Batch G)

### Gap 5: `ModuleTypeCheckConcepts` is a stub (`interp/helpers.go:436`)
- **Python** (`ivy_interp.py:48–54`): Temporarily extends `self.relations` with concept space arities, then type-checks concept space formulas via `type_check_list`
- **Go**: Only nil-checks concept spaces
- **Fix**: Port the Python logic: temporarily add concept relation arities to `mod.Relations`, run `TypeCheckList` on concept space formulas, restore original relations

## Files to Modify

1. **`compiler/ivy_compile.go`** — Add version guard for "this" isolate; call `ModuleTypeCheck`; add action assertion checks
2. **`compiler/decl.go`** — Fix Concept handler to use `GetRelationSortWithTerm`
3. **`interp/helpers.go`** — Fix `ModuleTypeCheck` and `ModuleTypeCheckConcepts` to do real arity checking
4. **`compiler/solo7_test.go`** — New test file
5. **`interp/solo7_interp_test.go`** — New test file for type check functions

## Existing Utilities to Reuse

| Utility | Location | Purpose |
|---------|----------|---------|
| `iu.VersionLE()` | `ivyutils/names.go:223` | Version comparison |
| `iu.GetStringVersion()` | `ivyutils/names.go:135` | Get current Ivy language version |
| `actions.TypeCheck()` | `actions/phase3.go:112` | Arity-check all apps in an expression |
| `interp.TypeCheckList()` | `interp/phase4.go:24` | Recursively type-check a list |
| `interp.ModuleTypeCheck()` | `interp/helpers.go:422` | Module-level type check (needs fixing) |
| `GetRelationSortWithTerm()` | `compiler/phase6.go:363` | Relation sort with term-based inference |
| `CheckInstantiations()` | `compiler/phase6.go:1557` | Already working, no changes needed |
| `AddToHierarchy()` | `module/module.go:334` | Already working, no changes needed |

## Plan

### Step 1: Fix version guard for "this" isolate (`compiler/ivy_compile.go:144–152`)

```go
// Python: if not iu.version_le(iu.get_string_version(),"1.6"):
if !iu.VersionLE(iu.GetStringVersion(), "1.6") {
    if _, ok := mod.Isolates["this"]; !ok {
        isol := &ast.IsolateDef{
            Elems:    []ast.Node{ast.NewAtom("this"), ast.NewAtom("this")},
            WithArgs: 0,
        }
        mod.Isolates["this"] = isol
    }
}
```

### Step 2: Fix `ModuleTypeCheck` to do real arity checking (`interp/helpers.go:422`)

Python's `module_type_check`:
```python
def module_type_check(self):
    type_check_list(self, self.axioms)
    self.type_check_concepts()
```

Go fix — use `TypeCheckList` on axioms (converting `[]*ast.LabeledFormula` formulas to `[]interface{}`):

```go
func ModuleTypeCheck(mod *module.Module) error {
    // Python: type_check_list(self, self.axioms)
    axiomExprs := make([]interface{}, 0, len(mod.LabeledAxioms))
    for _, ax := range mod.LabeledAxioms {
        if ax != nil && ax.Formula != nil {
            if expr, ok := ax.Formula.(lg.Expr); ok {
                axiomExprs = append(axiomExprs, expr)
            }
        }
    }
    if err := TypeCheckList(mod, axiomExprs); err != nil {
        return err
    }
    return ModuleTypeCheckConcepts(mod)
}
```

### Step 3: Fix `ModuleTypeCheckConcepts` to do real type checking (`interp/helpers.go:436`)

Python's `module_type_check_concepts`:
```python
def module_type_check_concepts(self):
    relations = self.relations
    self.relations = dict(iter(relations.items()))
    self.relations.update((x.rep, len(x.args)) for x, y in self.concept_spaces)
    type_check_list(self, [y for x, y in self.concept_spaces])
    self.relations = relations
```

Go fix — temporarily extend Relations, type-check concept space formulas, restore:

```go
func ModuleTypeCheckConcepts(mod *module.Module) error {
    if len(mod.ConceptSpaces) == 0 {
        return nil
    }
    // Save original relations
    origRelations := mod.Relations
    // Copy and extend with concept space arities
    newRelations := make(map[string]lg.Sort, len(origRelations))
    for k, v := range origRelations {
        newRelations[k] = v
    }
    // Extract concept space formulas and add relation arities
    var formulas []interface{}
    for _, cs := range mod.ConceptSpaces {
        // concept_spaces entries are (rel, body) pairs
        if pair, ok := cs.([2]interface{}); ok {
            if atom, ok := pair[0].(interface{ Rep() string; Args() int }); ok {
                // add (x.rep, len(x.args)) to relations
                // ... adapt to actual concept space pair type
            }
            formulas = append(formulas, pair[1])
        }
    }
    mod.Relations = newRelations
    err := TypeCheckList(mod, formulas)
    mod.Relations = origRelations
    return err
}
```

Note: The exact type of concept space entries depends on how `DomainSetup.Concept` stores them. Need to check the actual storage format. The concept pairs may be `[2]lg.Expr` or similar.

### Step 4: Call `ModuleTypeCheck` in `IvyCompile` (`ivy_compile.go:135`)

Replace the comment-only stub with an actual call:

```go
// Python line 2211: mod.type_check()
if err := interp.ModuleTypeCheck(mod); err != nil {
    return fmt.Errorf("type check: %w", err)
}
```

Need to add `"github.com/glycerine/goivy/interp"` import to ivy_compile.go.

### Step 5: Add action assertion checks (`ivy_compile.go:140–142`)

After the `TypeCheckAction` loop, add Python's assertion checks. The `actions.Action`
interface has `GetLineno() ast.Location` and `GetFormalParams() []*lg.Symbol`.
`mod.Actions` stores `interface{}`, so type-assert to `actions.Action`:

```go
for name, action := range mod.Actions {
    TypeCheckAction(action, mod)
    // Python: if not hasattr(action,'lineno'): print("no lineno: {}".format(name))
    // Python: assert hasattr(action,'formal_params'), action
    if act, ok := action.(actions.Action); ok {
        if act.GetLineno().Line == 0 {
            pp("no lineno: %s", name)
        }
        if act.GetFormalParams() == nil {
            pp("warning: action %s has no formal_params", name)
        }
    }
}
```

Python uses `assert` which would crash — Go should warn since the port is still maturing.

### Step 6: Fix Concept handler to use term (`compiler/decl.go:1430`)

Change:
```go
relSort, err := d.Compiler.GetRelationSort(atom.Terms)
```
To:
```go
relSort, err := d.Compiler.GetRelationSortWithTerm(atom.Terms, body)
```

Where `body` is `lf.Formula` (the concept body, matching Python's `c.args[1]`).

### Step 7: Write tests (`compiler/solo7_test.go`)

1. **TestIvyCompile_VersionGuard_ThisIsolate** — Version <= 1.6: no "this" isolate created. Version > 1.6: "this" isolate created.
2. **TestIvyCompile_TypeCheckCalled** — Axioms with wrong arity trigger error from type check
3. **TestIvyCompile_ProgressSymbolRemoval** — Progress symbols are removed from sig after processing
4. **TestIvyCompile_CheckInstantiations** — Undefined instantiation returns error
5. **TestIvyCompile_AddToHierarchy** — Dotted names properly added to hierarchy
6. **TestIvyCompile_GlobalObjectsToIsolates** — Global attributes are added to isolate "with" lists
7. **TestConcept_UsesTermForSortInference** — Concept handler uses body for sort inference

### Step 8: Write interp tests (`interp/solo7_interp_test.go`)

1. **TestModuleTypeCheck_ArityError** — Axiom with wrong arity caught
2. **TestModuleTypeCheck_Valid** — Valid axioms pass
3. **TestModuleTypeCheckConcepts_Valid** — Concept spaces pass
4. **TestModuleTypeCheckConcepts_ArityError** — Bad concept space caught

### Step 9: Fuzz tests

1. **FuzzIvyCompile** — Fuzz with random declaration lists, verify no panics
2. **FuzzModuleTypeCheck** — Fuzz with random axiom lists

### Step 10: Verify

Run `make test` to ensure no regressions.

## Key Design Notes

- `type_check_action` is intentionally a no-op in BOTH Python and Go (Python returns immediately at line 1330). No change needed.
- `check_instantiations` and `add_to_hierarchy` are already fully implemented and called correctly.
- Progress symbol removal is already implemented and correct.
- The `compiler` → `interp` import is safe (verified: no cycle — `interp` imports `actions`, `clauseops`, `logic`, `module`, `solver`, `transrel`, `z3bridge` but NOT `compiler`).
- For the action assertions (Step 5), Python uses `assert` (crashes on failure). Go should warn rather than crash, since the port is still in progress and some actions may not yet have all fields.
- The Concept `GetRelationSortWithTerm` fix is the same pattern as the Progress fix from Batch G — uses joint sort inference so untyped relation args can be inferred from the body.
