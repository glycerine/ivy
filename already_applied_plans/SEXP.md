# Plan: Add Canon() Canonical to All AST Types

**Created:** 2026-03-24 00:15

## Context

The `Canon() Canonical` method is part of the `ast.Node` interface. It must be implemented on every AST type to produce a deterministic, hashable s-expression string that captures all internal state. This enables incremental Merkle hashing to verify Go/Python parser state equivalence. Currently only 11 of ~130+ types implement it. The infrastructure (`Canonical` type, `Canonizer` interface) is already in `ivyutils/canon.go`. Examples exist in `ast/sort.go` and `ast/decl.go`.

For `logic/` package types that already have `Sexp() NodeKey`, we simply type-convert: `return Canonical(t.Sexp())`.

## S-expression Conventions

```
Struct:   (typeName field:"stringval" field2:value)
Slice:    [elem1 elem2 elem3]          -- space separated, NO commas
Map:      (hash key:value key2:value2)  -- sorted keys
String:   "double quoted"
Bool:     true / false
Int:      42
Nil node: nil
```

- Single space between fields, no commas anywhere
- Struct names use lowercase first letter (matching existing examples: `constantSort`, `base`, `declBase`)
- Field names use lowercase first letter, otherwise identical to Go field names

## Implementation Categories

### Category A: DeclBase-only types (40 types) — trivial one-liner pattern

These embed `DeclBase` with no extra fields. Pattern:
```go
func (d *MacroDecl) Canon() iu.Canonical {
    return iu.Canonical(fmt.Sprintf("(macroDecl declBase:%v)", d.DeclBase.Canon()))
}
```

**Types:** MacroDecl, ObjectDecl, ActionDecl, RelationDecl, ConstantDecl, ParameterDecl, DestructorDecl, ConstructorDecl, TypeDecl, VariantDecl, AxiomDecl, ConjectureDecl, ProofDecl, NamedDecl, SchemaDecl, TheoremDecl, DerivedDecl, DefinitionDecl, ProgressDecl, RelyDecl, MixOrdDecl, ConceptDecl, InitDecl, StateDecl, UpdateDecl, AssertDecl, InterpretDecl, MixinDecl, IsolateDecl, ExportDecl, ImportDecl, PrivateDecl, AliasDecl, DelegateDecl, ImplementTypeDecl, NativeDecl, AttributeDecl, InstantiateDecl, AutoInstanceDecl, ScenarioDecl, SubclassDecl

**File:** `ast/decl.go`

### Category B: Embedding wrapper types (9 types) — trivial delegation

These embed another type and just add a wrapper name. Pattern:
```go
func (d *PropertyDecl) Canon() iu.Canonical {
    return iu.Canonical(fmt.Sprintf("(propertyDecl axiomDecl:%v)", d.AxiomDecl.Canon()))
}
```

| Type | Embeds | File |
|------|--------|------|
| PropertyDecl | AxiomDecl | decl.go |
| FreshConstantDecl | ConstantDecl | decl.go |
| GhostTypeDef | TypeDef | decl.go |
| TrustedIsolateDef | IsolateDef | decl.go |
| ExtractDef | IsolateDef | decl.go |
| ProcessDef | ExtractDef | decl.go |
| AssumeGlobalTactic | AssumeTactic | tactic.go |
| IsolateObjectDecl | IsolateDecl | decl.go |
| DefinitionSchema | Definition | formula.go |

### Category C: Core term/formula types with custom fields (26 types in ast.go)

Each needs a hand-written Canon() method. Pattern:
```go
func (s *Symbol) Canon() iu.Canonical {
    return iu.Canonical(fmt.Sprintf("(symbol base:%v rep:%q sort:%v)", s.Base.Canon(), s.Rep, nodeCanon(s.Sort)))
}
```

**File `ast/ast.go`:**
1. NoneAST — `(noneAST base:%v)`
2. Symbol — `(symbol base:%v rep:%q sort:%v)`
3. Atom — `(atom base:%v rep:%q terms:[...] aSort:%v)`
4. App — `(app base:%v rep:%v terms:[...] aSort:%v)`
5. Variable — `(variable base:%v rep:%q vSort:%v)`
6. Old — `(old base:%v term:%v)`
7. This — `(this base:%v)`
8. MethodCall — `(methodCall base:%v obj:%v method:%v)`
9. Literal — `(literal base:%v polarity:%d atom:%v)`
10. Dot — `(dot base:%v left:%v right:%v)`
11. Bracket — `(bracket base:%v left:%v right:%v)`
12. Tuple — `(tuple base:%v elems:[...])`
13. Some — `(some base:%v params:[...] fmla:%v)`
14. SomeMin — `(someMin base:%v params:[...] fmla:%v index:%v)`
15. SomeMax — `(someMax base:%v params:[...] fmla:%v index:%v)`
16. SomeExpr — `(someExpr base:%v param:%v fmla:%v ifValue:%v elseVal:%v)`
17. KeyArg — `(keyArg app:%v)` (embeds *App)
18. DebugItem — `(debugItem base:%v name:%v value:%v)`
19. ThunkAction — `(thunkAction base:%v label:%v action:%v sort:%v body:%v)`
20. CrashAction — `(crashAction base:%v declArgs:[...])`
21. CallAction — `(callAction base:%v elems:[...] uniqueID:%d)`
22. Sequence — `(sequence base:%v stmts:[...])`
23. TemporalModels — `(temporalModels base:%v model:%v fmla:%v)`
24. CompiledNode — `(compiledNode base:%v)` (skip opaque Node field)

**File `ast/formula.go` (17 types):**
1. And — `(and base:%v terms:[...])`
2. Or — `(or base:%v terms:[...])`
3. Not — `(not base:%v body:%v)`
4. Implies — `(implies base:%v t1:%v t2:%v)`
5. Iff — `(iff base:%v t1:%v t2:%v)`
6. Ite — `(ite base:%v cond:%v then:%v else:%v)`
7. Forall — `(forall base:%v bounds:[...] body:%v)`
8. Exists — `(exists base:%v bounds:[...] body:%v)`
9. Isa — `(isa base:%v terms:[...])`
10. Globally — `(globally base:%v body:%v)`
11. Eventually — `(eventually base:%v body:%v)`
12. WhenOperator — `(whenOperator base:%v name:%q t1:%v t2:%v)`
13. Let — `(let base:%v defs:[...] body:%v)`
14. Definition — `(definition base:%v lhs:%v rhs:%v)`
15. DefinitionSchema — `(definitionSchema definition:%v)`
16. NamedBinder — `(namedBinder base:%v name:%q bounds:[...] body:%v)`
17. Trigger — `(trigger base:%v pattern:%v terms:[...])`

### Category D: Tactic types (21 types in ast/tactic.go)

**File `ast/tactic.go`:**
1. Tactic — `(tactic base:%v elems:[...])`
2. SchemaInstantiation — `(schemaInstantiation base:%v schemaName:%v ren:%v matches:[...])`
3. AssumeTactic — `(assumeTactic base:%v tLabel:%v schemaName:%v ren:%v matches:[...])`
4. AssumeGlobalTactic — `(assumeGlobalTactic assumeTactic:%v)`
5. UnfoldSpec — `(unfoldSpec base:%v defName:%v renamings:[...])`
6. UnfoldTactic — `(unfoldTactic base:%v tLabel:%v premise:%v unfSpecs:[...])`
7. ForgetTactic — `(forgetTactic base:%v names:[...])`
8. ShowGoalsTactic — `(showGoalsTactic base:%v)`
9. DeferGoalTactic — `(deferGoalTactic base:%v elems:[...])`
10. NullTactic — `(nullTactic base:%v)`
11. LetTactic — `(letTactic base:%v defs:[...])`
12. WitnessTactic — `(witnessTactic base:%v witnesses:[...])`
13. SpoilTactic — `(spoilTactic base:%v target:%v)`
14. IfTactic — `(ifTactic base:%v cond:%v then:%v else:%v)`
15. PropertyTactic — `(propertyTactic base:%v prop:%v pName:%v proof:%v)`
16. FunctionTactic — `(functionTactic base:%v elems:[...])`
17. TacticWith — `(tacticWith base:%v elems:[...])`
18. TacticLets — `(tacticLets base:%v lets:[...])`
19. TacticTactic — `(tacticTactic base:%v tName:%v body:%v proof:%v labels:[...])`
20. ProofTactic — `(proofTactic base:%v tLabel:%v proof:%v)`
21. ComposeTactics — `(composeTactics base:%v tactics:[...])`

### Category E: Decl types with extra fields (non-DeclBase-only)

**File `ast/decl.go`:**
1. LabeledFormula — `(labeledFormula base:%v label:%v formula:%v id:%d lineno:%d temporal:%v explicit:%v isDefinition:%v assumed:%v unprovable:%v)`
2. ActionDef — `(actionDef base:%v name:%v body:%v formalParams:[...] formalReturns:[...])`
3. TypeDef — `(typeDef base:%v name:%v value:%v finite:%v)`
4. VariantDef — `(variantDef base:%v name:%v vSort:%v)`
5. SchemaBody — `(schemaBody base:%v elems:[...])`
6. Schema — `(schema base:%v defn:%v fresh:[...] instances:[...])`
7. MixinBeforeDef — `(mixinBeforeDef base:%v mixerNode:%v mixeeNode:%v)`
8. MixinImplementDef — `(mixinImplementDef base:%v mixerNode:%v mixeeNode:%v)`
9. MixinAfterDef — `(mixinAfterDef base:%v mixerNode:%v mixeeNode:%v)`
10. IsolateDef — `(isolateDef base:%v elems:[...] withArgs:%d trusted:%v isObject:%v)`
11. ExportDef — `(exportDef base:%v exportedNode:%v scopeNode:%v)`
12. ImportDef — `(importDef base:%v imported:%v scope:%v)`
13. DelegateDef — `(delegateDef base:%v elems:[...])`
14. NativeCode — `(nativeCode base:%v code:%q)`
15. NativeType — `(nativeType base:%v elems:[...])`
16. NativeExpr — `(nativeExpr base:%v elems:[...] aSort:%v)`
17. NativeDef — `(nativeDef base:%v elems:[...])`
18. AttributeDef — `(attributeDef base:%v name:%v value:%v)`
19. Instantiation — `(instantiation base:%v name:%v sort:%v)`
20. StateDef — `(stateDef base:%v stateName:%q rels:[...])`
21. Renaming — `(renaming base:%v lhs:%q rhs:%q)`
22. PlaceList — `(placeList base:%v places:[...])`
23. ScenarioTransition — `(scenarioTransition base:%v from:%v to:%v transition:%v)`
24. ScenarioDef — `(scenarioDef base:%v name:%q places:%v initial:%v transitions:[...] invariant:%v)`
25. PlaceInfo — `(placeInfo base:%v placeName:%q expected:%v facts:[...])`
26. DefineInfo — `(defineInfo base:%v value:%v defs:[...])`
27. ScenarioBeforeMixin — `(scenarioBeforeMixin base:%v scenario:%v mixin:%v)`
28. ScenarioAfterMixin — `(scenarioAfterMixin base:%v scenario:%v mixin:%v)`
29. PrivateDef — `(privateDef base:%v name:%q expr:%v)`
30. ImplementTypeDef — `(implementTypeDef base:%v iType:%v cType:%v interp:%v)`
31. PatternBasedUpdate — `(patternBasedUpdate base:%v rHS:%v patterns:%v)`
32. UpdatePattern — `(updatePattern base:%v pattern:%q rHS:%v)`
33. UpdatePatternList — `(updatePatternList base:%v patterns:[...])`
34. SymbolList — `(symbolList base:%v symbols:[...])`  — note: `[]string`, quote each
35. RME — `(rme base:%v term:%v type:%v)`
36. NamedSpace — `(namedSpace base:%v name:%q)`
37. ProductSpace — `(productSpace base:%v spaces:[...])`
38. SumSpace — `(sumSpace base:%v spaces:[...])`

### Category F: Logic types — type conversion from Sexp()

**File `logic/canon.go` (new file):**

Add a Canon() method to each logic type that already has Sexp():
```go
func (v *Variable) Canon() iu.Canonical { return iu.Canonical(v.Sexp()) }
```

Types: Variable, Symbol, Apply, Eq, Not, And, Or, Implies, Iff, Ite, Globally, Eventually, WhenOperator, Cond, ForAll, Exists, Lambda, NamedBinder, Definition, DefinitionSchema, NativeExpr, UninterpretedSort, BooleanSort, FunctionSort, EnumeratedSort, RangeSort, TopSort

## Helper function needed

Add a `nodeCanon` helper for nil-safe Canon calls on `Node` fields:
```go
func nodeCanon(n Node) iu.Canonical {
    if n == nil {
        return "nil"
    }
    return n.Canon()
}
```

For `*bool` (LabeledFormula.Temporal):
```go
func boolPtrCanon(b *bool) string {
    if b == nil { return "nil" }
    if *b { return "true" }
    return "false"
}
```

For `[]string` (TacticTactic.Labels, SymbolList.Symbols):
```go
func stringSliceCanon(ss []string) string {
    if len(ss) == 0 { return "[]" }
    parts := make([]string, len(ss))
    for i, s := range ss { parts[i] = fmt.Sprintf("%q", s) }
    return "[" + strings.Join(parts, " ") + "]"
}
```

## Files to Modify

| File | Types | Count |
|------|-------|-------|
| `ast/ast.go` | Core terms + helpers | 24 + helpers |
| `ast/formula.go` | Formula types | 17 |
| `ast/tactic.go` | Tactic types | 21 |
| `ast/decl.go` | Decl types | ~70 |
| `logic/canon.go` (NEW) | Logic type wrappers | ~27 |

**Total: ~130 Canon() methods + 3 helpers**

## Implementation Order

1. Add `nodeCanon`, `boolPtrCanon`, `stringSliceCanon` helpers to `ast/ast.go`
2. Core terms in `ast/ast.go` (Category C) — these are referenced by everything else
3. Formula types in `ast/formula.go` (Category C + D partial)
4. DeclBase-only types in `ast/decl.go` (Category A) — bulk of the count, mechanical
5. Embedding wrapper types (Category B) — trivial
6. Decl types with extra fields (Category E)
7. Tactic types in `ast/tactic.go` (Category D)
8. Logic wrappers in `logic/canon.go` (Category F)

## Verification

```bash
cd ~/goivy && go build ./...
```

The code should compile with zero errors since Canon() is already in the Node interface and currently causes implicit interface satisfaction failures for types that don't implement it (they rely on Base.Canon() promotion, which works but gives wrong output). After this change, each type produces its own correct canonical form.
