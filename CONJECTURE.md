# Batch F — Green Phase: Implement Missing Features

## Batch F — Conjecture/proof pipeline

- [ ] §6.2 #5: `CheckDefinitions` — separate defs from props, check redefinition, detect cycles
- [ ] §6.2 #7: `CreateConjActions` — determine which actions must preserve each conjecture
- [ ] §6.3 #29: `ConjSetup` pass 2 — implement real conjecture setup (currently all no-op)

All part of the conjecture processing pipeline.

## Context

Red phase is complete. 25 tests exist in `compiler/batch_f_test.go`. Results:
- **11 FAIL** — gaps to fix
- **13 PASS** — already-working sanity checks
- **1 SKIP** — interference detection (needs signature change)

Goal: make all 11 failing tests pass by implementing the missing Python logic in Go.

## Files to Modify

| File | What changes |
|------|-------------|
| `compiler/ivy_compile.go` :1202-1276 | **CheckDefinitions** — add redefinition error, NativeDefinitions/Named checks, self-loop error, action interference (version-gated), interpreted symbol check |
| `compiler/ivy_compile.go` :1347-1366 | **CreateConjActions** — add version gate, parent_child_name walk, isolate-scoped exports, call graph integration |
| `compiler/ivy_compile.go` :347-398 | **ConjSetup.ProcessDecls** — fix proof compilation: use `CompileTactic` instead of `CompileNode` (proofs are tactic AST nodes, not formulas) |
| `compiler/batch_f_test.go` | Update tests 9, 10 to use real `actions.AssignAction` instead of `ast.NewAtom` placeholders; update tests 14, 15 to use real `ast.IsolateDef` |

## Existing Code to Reuse

- `GetSymbolDependencies` (`compiler/phase6.go:2327`) — transitive symbol dep collection
- `collectFormulaSymbols` (`compiler/phase6.go:2292`) — formula→symbols extraction
- `actions.Modifies` (`actions/transforms.go:86`) — returns modified symbol names from an Action
- `CompileTactic` (`compiler/compiler.go:1102`) — compiles tactic AST nodes; default case returns node unchanged (matching Python's `pf.compile()`)
- `iu.ParentChildName` (`ivyutils/names.go:66`) — splits "a.b.c" → ["a.b", "c"]
- `isolate.GetIsolateExports` (`isolate/iter.go`) — gets exports for an isolate
- `module.CallGraph` (`module/module.go:416`) — stub, returns empty map (acceptable for now)
- `iu.GetStringVersion` / `iu.SetStringVersion` (`ivyutils/names.go:134,141`)
- `TarjanArcs` (`compiler/phase6.go:1476`) — already used for cycle detection
- `proof.NewProofChecker` (`proof/checker.go:39`) — exists but has no `AdmitDefinition` method yet
- `IsolateDefInterface` (`isolate/isolate.go:72`) — `VerifiedNames()`, `PresentNames()`, `IsExtract()`
- `ast.IsolateDef` (`ast/decl.go:968`) — Elems[0]=name, Elems[1:end-WithArgs]=verified, end=present

All needed imports already exist in `ivy_compile.go`: `actions`, `isolate`, `iu`, `lg`, `lu`.

---

## Step-by-step Implementation

### Step 1: Fix CheckDefinitions (tests 3, 4, 5, 7)

**File:** `compiler/ivy_compile.go` — `CheckDefinitions` function (~line 1202)

Changes to the redefinition check (currently at line 1240-1249):

1a. **Return error on redefinition** (test 3): Change `pp(...)` to `return &lg.IvyError{...}`:
```go
// Replace pp() with:
if prev, exists := defs[sym]; exists {
    return &lg.IvyError{Msg: fmt.Sprintf("redefinition of %s\n%d from here", sym, prev.Lineno)}
}
```

1b. **Add NativeDefinitions check** (test 4): After the `mod.Definitions` loop, add:
```go
for _, nd := range mod.NativeDefinitions {
    if ldf, ok := nd.(*ast.LabeledFormula); ok {
        if logicDef, ok := ldf.Formula.(*lg.Definition); ok {
            sym := definesName(logicDef)
            if _, exists := defs[sym]; exists {
                return &lg.IvyError{Msg: fmt.Sprintf("redefinition of %s", sym)}
            }
            defs[sym] = ldf
        }
    }
}
```

1c. **Add Named check** (test 5): After NativeDefinitions loop:
```go
for _, ne := range mod.Named {
    if sym, ok := ne.Name.(*lg.Symbol); ok {
        if _, exists := defs[sym.Name]; exists {
            return &lg.IvyError{Msg: fmt.Sprintf("redefinition of %s", sym.Name)}
        }
        defs[sym.Name] = ne.Formula
    }
}
```

1d. **Return error for self-loop without proof** (test 7): In the singleton SCC check (~line 1269):
```go
// Current code just logs; change to:
if d, ok := dmap[defName]; ok {
    if _, hasProof := pmap[d.ID]; !hasProof {
        return &lg.IvyError{Msg: fmt.Sprintf("definition of %s requires a recursion schema", defName)}
    }
    // TODO: call prover.AdmitDefinition when it exists
}
```
This also requires building `pmap` (proof map by ID) before the SCC loop. The `dmap` (definition name → LabeledFormula) is needed too. Both match what Python does.

### Step 2: Fix CheckDefinitions — action interference (tests 9, 10)

**File:** `compiler/ivy_compile.go` — add version-gated interference check inside `CheckDefinitions`, after the cycle/SCC check.

Python gates on `iu.version_le("1.7", iu.get_string_version())` meaning "runs when version >= 1.7".

The logic is similar to `CheckMutax` (phase6.go:2242-2287), BUT there's a type mismatch: `CheckMutax` checks `*ast.Definition` while `CheckDefinitions` works with `*lg.Definition`. They are different types.

**Inline the interference check** in `CheckDefinitions` using `*lg.Definition`:

```go
// Action interference check (v1.7+) — matches Python check_definitions lines 1738-1759
ver := iu.GetStringVersion()
if ver >= "1.7" {
    // Collect all symbols modified by actions
    modified := make(map[string]bool)
    for _, actVal := range mod.Actions {
        if act, ok := actVal.(actions.Action); ok {
            for sym := range actions.Modifies(act) {
                modified[sym] = true
            }
        }
    }
    // Build definition map for transitive dep lookup
    defMap := make(map[string]interface{})
    for _, lf := range mod.Definitions {
        if def, ok := lf.Formula.(*lg.Definition); ok {
            defMap[definesName(def)] = def.Rhs
        }
    }
    // Check axioms
    for _, lf := range mod.LabeledAxioms {
        if !lf.Temporal {
            deps := make(map[string]bool)
            GetSymbolDependencies(defMap, deps, lf.Formula)
            for sym := range deps {
                if modified[sym] {
                    return &lg.IvyError{Msg: fmt.Sprintf("immutable symbol assigned: %s", sym)}
                }
            }
        }
    }
    // Check definitions: LHS must not be modified
    for _, lf := range mod.Definitions {
        if def, ok := lf.Formula.(*lg.Definition); ok {
            name := definesName(def)
            if modified[name] {
                return &lg.IvyError{Msg: fmt.Sprintf("immutable symbol assigned: %s", name)}
            }
        }
    }
}
```

**Also fix tests 9 and 10** in `batch_f_test.go`: replace `ast.NewAtom("assign_f")` placeholders with real `actions.AssignAction` that modifies symbol "f":
```go
fSym := lg.NewSymbol("f", lg.Boolean)
assignAction := actions.NewAssignAction(fSym, lg.NewSymbol("true_val", lg.Boolean))
mod.Actions["act1"] = assignAction
```

### Step 3: Fix CheckDefinitions — self-loop with proof (test 8)

Test 8 currently passes! The self-recursive def has a proof, so it goes into LabeledProps (stale), never enters mod.Definitions, never enters the SCC check. No change needed.

### Step 4: Fix CreateConjActions (tests 12, 14, 15)

**File:** `compiler/ivy_compile.go` — rewrite `CreateConjActions` (~line 1347)

```go
func CreateConjActions(mod *module.Module) {
    // Version gate (test 12)
    if iu.GetStringVersion() <= "1.6" {
        return
    }

    if mod.ConjActions == nil {
        mod.ConjActions = make(map[string][]string)
    }

    // Build isolate exports and object→isolate mapping
    type isoEntry struct {
        name string
        def  isolate.IsolateDefInterface
    }
    myexports := make(map[string]map[string]bool) // iso name → exported actions
    objects := make(map[string][]isoEntry)          // verified object → isolates
    cg := mod.CallGraph()

    for isoName, isoVal := range mod.Isolates {
        if isol, ok := isoVal.(isolate.IsolateDefInterface); ok {
            myexports[isoName] = isolate.GetIsolateExports(mod, cg, isol)
            for _, v := range isol.VerifiedNames() {
                objects[v] = append(objects[v], isoEntry{isoName, isol})
            }
        }
    }

    for _, conj := range mod.LabeledConjs {
        if conj.Label == nil {
            continue
        }
        lbl := labelName(conj.Label)
        origLbl := lbl

        // Walk up name hierarchy (test 15)
        for lbl != "this" {
            if _, found := objects[lbl]; found {
                break
            }
            parts := iu.ParentChildName(lbl)
            lbl = parts[0]
        }

        var actionSet map[string]bool
        if lbl == "this" {
            // Top-level: all exported actions
            actionSet = make(map[string]bool)
            for _, exp := range mod.Exports {
                if expDef, ok := exp.(*ast.ExportDef); ok {
                    actionSet[expDef.Exported()] = true
                }
            }
        } else {
            // Isolate-scoped (test 14)
            actionSet = make(map[string]bool)
            for _, entry := range objects[lbl] {
                for act := range myexports[entry.name] {
                    actionSet[act] = true
                }
            }
        }

        actionNames := make([]string, 0, len(actionSet))
        for act := range actionSet {
            actionNames = append(actionNames, act)
        }
        mod.ConjActions[origLbl] = actionNames
    }
}
```

Note: version comparison `ver <= "1.6"` works as lexicographic string comparison for "1.6" vs "1.7" etc. This is correct for single-digit version parts.

### Step 5: Fix ConjSetup proof compilation (tests 19, 22)

Tests 19 and 22 fail because the Go code calls `cs.Compiler.CompileNode(arg)` which tries to compile proof bodies as logic expressions. But proofs are **tactic AST nodes**, not formulas.

**Python behavior:** `pf.compile()` calls `AST.cmpl()` which for most proof nodes (tactics, atoms) just clones recursively. For a plain Atom with no args, it returns itself unchanged — always succeeds.

**Go has `CompileTactic`** (`compiler/compiler.go:1102`) which returns `(ast.Node, error)` and handles all tactic types. Its default case returns the node unchanged with no error — matching Python exactly.

**Fix:** In `ConjSetup.ProcessDecls` (`ivy_compile.go:385`), change `CompileNode` to `CompileTactic`:

```go
// BEFORE (wrong — treats proof as formula):
compiled, err := cs.Compiler.CompileNode(arg)

// AFTER (correct — treats proof as tactic, matching Python pf.compile()):
compiled, err := cs.Compiler.CompileTactic(arg)
```

Also change the `Proof` field assignment from `lg.Expr` to `ast.Node`:
```go
cs.Compiler.Module.Proofs = append(cs.Compiler.Module.Proofs, module.ProofEntry{
    Formula: cs.lastFact,
    Proof:   compiled,  // ast.Node, not lg.Expr — ProofEntry.Proof is interface{}, so this works
})
```

No test changes needed — the tests are correct as-is.

### Step 6: Fix test 17 (interference detection)

Currently skipped. The interference check in `create_conj_actions` is a separate concern from the basic isolate scoping. Once step 4 is done and isolate scoping works, we can implement the interference check or leave it as a follow-up.

**Recommendation:** Keep test 17 skipped for this batch. The interference detection in `create_conj_actions` requires a working `CallGraph()` (currently stub) and `iso.get_isolate_conjs`. It's a deeper feature that should be its own batch.

### Step 7: Version comparison helper

Python has `iu.version_le("1.7", ver)` which does numeric comparison. For Go, lexicographic string comparison suffices for "1.6" vs "1.7" (single-digit parts). `"1.6" <= "1.6"` is true, `"1.7" <= "1.6"` is false.

---

## Test Fixes in batch_f_test.go

| Test | Fix |
|------|-----|
| 9, 10 | Use real `actions.NewAssignAction(fSym, rhsSym)` instead of `ast.NewAtom` placeholder; add `"github.com/glycerine/goivy/actions"` import |
| 14, 15 | Use real `ast.IsolateDef` with proper `Elems` and `WithArgs` instead of `ast.NewAtom` placeholders. Constructor: `&ast.IsolateDef{Elems: []ast.Node{ast.NewAtom("iso_a"), ast.NewAtom("obj_a")}, WithArgs: 0}` — Elems[0]=name, Elems[1:]=verified objects |
| 19, 22 | No test change — fix is in `ConjSetup.ProcessDecls`: use `CompileTactic` instead of `CompileNode` |
| 17 | Keep skipped (CallGraph stub blocks it) |

---

## Verification

After all changes:
```bash
cd /Users/jaten/go/src/github.com/glycerine/goivy && go test -run "TestCheckDefinitions|TestCreateConj|TestConjSetup" -count=1 -v ./compiler/ 2>&1
```

Expected: 24 PASS, 1 SKIP (test 17). All 11 previously-failing tests should now pass.

Also run full test suite to check for regressions:
```bash
cd /Users/jaten/go/src/github.com/glycerine/goivy && go test ./compiler/ 2>&1 | tail -5
```
