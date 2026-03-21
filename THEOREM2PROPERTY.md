# Solo 6: TheoremToProperty — Full Sort/Symbol Renaming

## Context

`TheoremToProperty` (`compiler/ivy_compile.go:1552-1608`) converts a theorem (proved by schema/tactic) into a property. The current Go skeleton handles premise extraction and implication building, but **completely omits the sort/symbol renaming logic** (Python lines 1788-1808). This is the core algorithmic part:

1. Extract vocabulary (sorts, symbols) from the goal's SchemaBody premises
2. For each sort/symbol whose name collides with the global signature, generate a unique renamed version
3. Apply the renaming match to the entire goal before extracting premises
4. Also: definition premises should be appended to `mod.Definitions` (requires `mod` parameter)
5. Also: check that the conclusion is not a `Definition` (raise error if so)

Python source (`ivy_compiler.py:1786-1826`):
```python
def theorem_to_property(prop,pos=True):
    if isinstance(prop.formula,ivy_ast.SchemaBody):
        vocab = ip.goal_vocab(prop)
        match = dict()
        for sort in vocab.sorts:
            if sort.name in ivy_logic.sig.sorts:
                newname = unused_name_with_base(sort.name,ivy_logic.sig.sorts)
                newsort = UninterpretedSort(newname)
                ivy_logic.add_sort(newsort)
                match[sort] = newsort
            else:
                ivy_logic.add_sort(sort)
        for sym in vocab.symbols:
            if sym.name in ivy_logic.sig.symbols:
                newname = unused_name_with_base(sym.name,ivy_logic.sig.symbols)
                newsym = ivy_logic.Symbol(newname,sym.sort)
                newsym = ip.apply_match_func(match,newsym)
                ivy_logic.add_symbol(newsym.name,newsym.sort)
                match[sym] = newsym
            else:
                ivy_logic.add_symbol(sym.name,sym.sort)
        if match:
            prop = ip.apply_match_goal(match,prop,ip.apply_match_alt)
        prems = []
        for x in prop.formula.prems():
            if isinstance(x,ivy_ast.LabeledFormula):
                if x.definition:
                    im.module.definitions.append(prop_to_def(x))
                else:
                    if not x.explicit:
                        prems.append(theorem_to_property(x,False).formula)
        conc = prop.formula.conc()
        if isinstance(conc,ivy_logic.Definition):
            raise iu.IvyError(prop,"definitional subgoal must be discharged")
        fmla = ivy_logic.Implies(ivy_logic.And(*prems),conc) if prems else conc
        res = ivy_ast.LabeledFormula(prop.label,fmla)
        res.lineno = prop.lineno
        return res
    return prop
```

## Bugs / Missing Pieces

### Bug 1: Sort/symbol renaming entirely missing (lines 1788-1808)
The Go skeleton never calls `GoalVocab`, never checks for name collisions in `sig.Sorts`/`sig.Symbols`, and never builds a rename match. This means sort/symbol names from the schema can collide with existing global names, causing incorrect proofs or subtle logic errors.

### Bug 2: No `mod` parameter — can't access sig or append definitions
Python uses `ivy_logic.sig` (global) and `im.module.definitions` (global module). The Go function takes only `*ast.LabeledFormula` and has no access to the signature or module definitions. Must add `mod *module.Module` parameter.

### Bug 3: `GoalVocab` doesn't collect sorts from premises
Python: `sorts = [s for s in prems if isinstance(s, il.UninterpretedSort)]`
Go `GoalVocab` (`proof/goal.go:146-195`) only collects `ConstantDecl` symbols and `LabeledFormula` conclusions — never checks if a premise is an `UninterpretedSort`. Must fix.

### Bug 4: `ApplyMatchGoalNode` doesn't handle sort/ConstantDecl premises
Python `apply_match_goal` handles 3 premise types: `LabeledFormula`, `UninterpretedSort/EnumeratedSort`, and `ConstantDecl`. Go `ApplyMatchGoalNode` (`proof/phase5_matching.go:1064-1082`) only recurses on `LabeledFormula` and passes other premises through unchanged. Sort/symbol rename match won't propagate to these premises.

### Bug 5: Missing `Definition` check on conclusion
Python: `if isinstance(conc, ivy_logic.Definition): raise IvyError(...)`. Go skeleton silently accepts definition conclusions.

### Bug 6: Definition premises not appended to `mod.Definitions`
Python: `im.module.definitions.append(prop_to_def(x))`. Go skeleton has a comment "skip" but doesn't actually append. `PropToDef` exists at `compiler/phase6.go:1558`.

## Files to Modify

1. **`~/goivy/compiler/ivy_compile.go`** — Rewrite `TheoremToProperty` with full renaming logic
2. **`~/goivy/proof/goal.go`** — Fix `GoalVocab` to collect `UninterpretedSort` premises
3. **`~/goivy/proof/phase5_matching.go`** — Fix `ApplyMatchGoalNode` to handle sort and ConstantDecl premises
4. **`~/goivy/compiler/phase6.go`** — Update `mapTheoremToProperty` to pass `mod`
5. **`~/goivy/check/check.go`** — Update call site to pass `mod`
6. **`~/goivy/check/isolate_check.go`** — Update call site to pass `mod`
7. **`~/goivy/compiler/theorem_to_property_test.go`** — New test file

## Existing Utilities to Reuse

| Utility | Location | Purpose |
|---------|----------|---------|
| `proof.GoalVocab()` | `proof/goal.go:146` | Extract sorts/symbols/variables from goal |
| `proof.Vocab` | `proof/goal.go:12` | `{Sorts, Symbols, Variables}` struct |
| `iu.UnusedNameWithBase()` | `ivyutils/renamer.go:105` | Generate unique name avoiding collisions |
| `proof.ApplyMatchGoalNode()` | `proof/phase5_matching.go:1064` | Apply rename match to goal |
| `proof.ApplyMatchAlt()` | `proof/phase5_matching.go:739` | Apply match to formula |
| `proof.ApplyMatchFunc()` | `proof/match.go:468` | Apply sort match to symbol's sort |
| `PropToDef()` | `compiler/phase6.go:1558` | Convert labeled formula to definition |
| `il.Sig.AddSort()` | `ivylogic/sig.go:300` | Add sort to signature |
| `il.Sig.AddSymbol()` | `ivylogic/sig.go:94` | Add symbol to signature |
| `lg.NewImplies()` | logic package | Build implication |
| `il.NormalizedAnd()` | ivylogic package | Build conjunction |
| `lg.Key()` | `logic/sexp.go:156` | Create `NodeKey` for map keys |
| `lg.UninterpretedSort` | `logic/sort.go:26` | Sort type |
| `lg.Definition` | `logic/definition.go:11` | Definition type |

## Plan

### Step 1: Fix `GoalVocab` to collect sorts (`proof/goal.go:146`)

Add sort collection in the premise loop, matching Python:
```go
// In the premise loop, after the ConstantDecl check:
if s, ok := p.(lg.Sort); ok {
    if _, isUninterp := s.(*lg.UninterpretedSort); isUninterp {
        sorts = append(sorts, s)
    }
}
```

### Step 2: Fix `ApplyMatchGoalNode` to handle sort/ConstantDecl premises (`proof/phase5_matching.go:1064`)

In the premise loop, add cases matching Python's `apply_match_goal`:
```go
for _, p := range prems {
    if lf, ok := p.(*ast.LabeledFormula); ok {
        newPrems = append(newPrems, ApplyMatchGoalNode(match, lf))
    } else if s, ok := p.(lg.Sort); ok {
        // Apply sort renaming: match[sort] → newSort
        key := lg.Key(s)
        if rep, ok := match[key]; ok {
            if rs, ok := rep.(lg.Sort); ok {
                newPrems = append(newPrems, rs)
                continue
            }
        }
        newPrems = append(newPrems, p)
    } else if cd, ok := p.(*ast.ConstantDecl); ok {
        // Apply symbol renaming via ApplyMatchFuncAlt
        args := cd.Args()
        if len(args) > 0 {
            if sym, ok := args[0].(*lg.Symbol); ok {
                newSym := ApplyMatchFunc(match, sym)
                // Check if newSym itself is in match
                if rep, ok := match[lg.Key(newSym)]; ok {
                    newPrems = append(newPrems, cd.Clone([]ast.Node{rep.(ast.Node)}))
                } else {
                    newPrems = append(newPrems, cd.Clone([]ast.Node{newSym}))
                }
            } else {
                newPrems = append(newPrems, p)
            }
        } else {
            newPrems = append(newPrems, p)
        }
    } else {
        newPrems = append(newPrems, p)
    }
}
```

### Step 3: Change `TheoremToProperty` signature (`compiler/ivy_compile.go`)

Change from `func TheoremToProperty(goal *ast.LabeledFormula)` to:
```go
func TheoremToProperty(goal *ast.LabeledFormula, mod *module.Module) *ast.LabeledFormula
```

### Step 4: Rewrite `TheoremToProperty` body

Full implementation matching Python line-by-line:

```go
func TheoremToProperty(goal *ast.LabeledFormula, mod *module.Module) *ast.LabeledFormula {
    if goal == nil {
        return nil
    }
    sb, ok := goal.Formula.(*ast.SchemaBody)
    if !ok {
        return goal
    }

    // Step A: Extract vocabulary and build rename match
    sig := mod.Sig
    vocab := proof.GoalVocab(goal)
    match := make(map[lg.NodeKey]lg.Expr)

    for _, sort := range vocab.Sorts {
        name := il.SortName(sort)
        if _, exists := sig.Sorts[name]; exists {
            // Name collision — generate unique name
            usedNames := make(map[string]struct{}, len(sig.Sorts))
            for k := range sig.Sorts {
                usedNames[k] = struct{}{}
            }
            newname := iu.UnusedNameWithBase(name, usedNames)
            newsort := &lg.UninterpretedSort{Name: newname}
            sig.AddSort(newsort)
            match[lg.Key(sort)] = newsort
        } else {
            sig.AddSort(sort)
        }
    }

    for _, sym := range vocab.Symbols {
        if _, exists := sig.Symbols[sym.Name]; exists {
            usedNames := make(map[string]struct{}, len(sig.Symbols))
            for k := range sig.Symbols {
                usedNames[k] = struct{}{}
            }
            newname := iu.UnusedNameWithBase(sym.Name, usedNames)
            newsym := lg.NewSymbol(newname, sym.CSort)
            newsym = proof.ApplyMatchFunc(match, newsym)
            sig.AddSymbol(newsym.Name, newsym.CSort)
            match[lg.Key(sym)] = newsym
        } else {
            sig.AddSymbol(sym.Name, sym.CSort)
        }
    }

    // Step B: Apply rename match to entire goal
    prop := goal
    if len(match) > 0 {
        prop = proof.ApplyMatchGoalNode(match, goal)
    }

    // Step C: Process premises
    sb = prop.Formula.(*ast.SchemaBody) // re-extract after match
    var prems []ast.Node
    for _, x := range sb.Prems() {
        if lf, ok := x.(*ast.LabeledFormula); ok {
            if lf.IsDefinition {
                mod.Definitions = append(mod.Definitions, PropToDef(lf).(*ast.LabeledFormula))
            } else if !lf.Explicit {
                sub := TheoremToProperty(lf, mod)
                if sub != nil {
                    prems = append(prems, sub.Formula)
                }
            }
        }
    }

    // Step D: Check conclusion is not a Definition
    conc := sb.Conc()
    if _, isDef := conc.(*lg.Definition); isDef {
        panic(fmt.Sprintf("definitional subgoal must be discharged"))
        // Or use iu.IvyError if available
    }

    // Step E: Build formula
    var fmla ast.Node
    if len(prems) > 0 {
        premExprs := exprSlice(prems)
        concExpr := nodeToExpr(conc)
        if len(premExprs) > 0 {
            antecedent := il.NormalizedAnd(premExprs...)
            impl, err := lg.NewImplies(antecedent, concExpr)
            if err != nil {
                fmla = conc
            } else {
                fmla = impl
            }
        } else {
            fmla = conc
        }
    } else {
        fmla = conc
    }

    result := &ast.LabeledFormula{
        Label:   prop.Label,
        Formula: fmla,
        Lineno:  prop.Lineno,
    }
    return result
}
```

### Step 5: Update all call sites

**`compiler/phase6.go:1935-1941`** — `mapTheoremToProperty`:
```go
func mapTheoremToProperty(goals []*ast.LabeledFormula, mod *module.Module) []*ast.LabeledFormula {
    result := make([]*ast.LabeledFormula, len(goals))
    for i, g := range goals {
        result[i] = TheoremToProperty(g, mod)
    }
    return result
}
```
Update callers at lines 1878 and 2058 to pass `mod`.

**`check/check.go:403`**:
```go
modSG = compiler.TheoremToProperty(modSG, mod)  // pass mod
```
(Need to verify what `mod` variable is in scope at this call site.)

**`check/isolate_check.go:668`**:
```go
pgoal := compiler.TheoremToProperty(AstLFToModuleLF(goal), mod)
```

### Step 6: Write tests (`compiler/theorem_to_property_test.go`)

**Unit Tests:**

1. **TestTheoremToProperty_NonSchemaBody** — Non-SchemaBody formula returned unchanged.

2. **TestTheoremToProperty_SimpleSchemaBody** — SchemaBody with no name collisions: sorts/symbols added to sig, conclusion returned as property.

3. **TestTheoremToProperty_SortRenaming** — Sort name collides with sig → renamed with unique suffix, formula updated.

4. **TestTheoremToProperty_SymbolRenaming** — Symbol name collides with sig → renamed, sort mappings applied.

5. **TestTheoremToProperty_BothSortAndSymbolRename** — Both sort and symbol collide → both renamed, symbol's sort updated via `ApplyMatchFunc`.

6. **TestTheoremToProperty_DefinitionPremise** — Definition premise appended to `mod.Definitions`, not included in `prems`.

7. **TestTheoremToProperty_ExplicitPremiseSkipped** — Explicit premise neither added to prems nor definitions.

8. **TestTheoremToProperty_DefinitionConclusion** — Conclusion is `lg.Definition` → panics with error.

9. **TestTheoremToProperty_ImplicationBuilt** — Multiple non-explicit premises → `Implies(And(prems), conc)` structure.

10. **TestTheoremToProperty_NilGoal** — Nil input returns nil.

11. **TestTheoremToProperty_RecursiveSchema** — Nested SchemaBody in premise → recursively converted.

12. **TestGoalVocab_CollectsSorts** — Verify `GoalVocab` now collects `UninterpretedSort` from premises.

13. **TestApplyMatchGoalNode_SortPremises** — Verify sort premises get renamed by match.

**Fuzz Test:**

14. **FuzzTheoremToProperty** — Random schemas with 0-3 sort premises, 0-3 symbol premises, 0-3 labeled formula premises, random name collisions. Verify no panics and result is a valid `LabeledFormula`.

### Step 7: Verify
Run `make test` to ensure no regressions.

## Key Design Notes

- `UnusedNameWithBase` takes `map[string]struct{}` but `sig.Sorts` is `map[string]lg.Sort`. Build a temp `map[string]struct{}` from the keys. This is a minor conversion.
- The `match` map uses `lg.NodeKey` (string) as keys. `lg.Key(sort)` and `lg.Key(sym)` produce the S-expression strings that identify nodes.
- `proof.ApplyMatchFunc(match, sym)` already handles `map[lg.NodeKey]lg.Expr` — it checks `match[lg.Key(s)]` for each sort in the symbol's sort signature.
- The `pos` parameter from Python (`theorem_to_property(prop, pos=True)`) is unused in the Python code (never read), so we omit it.
