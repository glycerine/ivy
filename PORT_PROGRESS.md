# Ivy Python → Go Port: Detailed Status & Progress Report

**Date**: 2026-03-14
**Module path**: `github.com/glycerine/goivy`
**Branch**: `goport` (base: `master`)
**Python source**: `~/pyivy/ivy/ivy/` — 54,099 lines across 68 `.py` files

---

## PART 1: COMPLETED WORK — 23 Chunks

### Summary

| # | Package | Source | Tests | Fuzz | Python Source | Description |
|---|---------|--------|-------|------|---------------|-------------|
| 1-3 | `logic/` | 1,020 | 934 | 4 | `logic.py` (488 lines) | Sort, Node, Term, Formula types |
| 4 | `typeinfer/` | 857 | 250 | 1 | `type_inference.py` (419 lines) | Union-find type inference |
| 5 | `logicutil/` | 671 | 453 | 1 | `logic_util.py` (335 lines) | FreeVars, Substitute, EqualModAlpha |
| 6 | `ivyutils/` | 609 | 368 | 2 | `ivy_utils.py` (773 lines) | Renamer, Parameters, Graph algos |
| 7 | `z3bridge/` | 914 | 419 | **0** | `z3_utils.py` (197) + `ivy_solver.py` (1,716) | Z3 CGo wrapper + Translator |
| 8 | `ast/` | 2,466 | 642 | **0** | `ivy_ast.py` (1,965 lines) | 100+ AST node types |
| 9 | `lexer/` | 793 | 424 | **2** | `ivy_lexer.py` (304 lines) | Hand-written tokenizer |
| 10 | `parser/` | 1,803 | 787 | **2** | `ivy_parser.py` (3,161) + `ivy_logic_parser.py` (655) | Recursive descent parser |
| 11 | `ivylogic/` | 2,590 | 1,150 | 1 | `ivy_logic.py` (1,774 lines) | Higher-level logic IR, Sig, sort inference, EPR |
| 12 | `module/` | 1,050 | 580 | 1 | `ivy_module.py` (412 lines) | Module context, theory gen, canonize types |
| 13 | `actions/` | 1,216 | 647 | 2 | `ivy_actions.py` (1,687 lines) | Action semantics, 20+ action types, annotations |
| 14 | `transrel/` | 1,200 | 1,560 | 5 | `ivy_transrel.py` (669 lines) | Transition relations, composition, images |
| 15 | `proof/` | 1,261 | 846 | 2 | `ivy_proof.py` (1,653 lines) | Proof checker, matching, skolemization, goals |
| 16 | `theory/` | 279 | 337 | 1 | `ivy_theory.py` (181 lines) | Built-in theories (int/nat/bv) |
| 17 | `clauseops/` | 1,314 | 801 | 2 | `ivy_logic_utils.py` (1,635 lines) | Clauses type, clause ops, AST utilities |
| 18 | `compiler/` | 2,012 | 634 | 1 | `ivy_compiler.py` (2,320 lines) | AST→logic IR compilation, DeclInterp |
| 19 | `isolate/` | 939 | 855 | 1 | `ivy_isolate.py` (2,022 lines) | Modular verification, stripping, deps |
| 20 | `interp/` | 1,200 | 1,100 | 2 | `ivy_interp.py` (643 lines) | Symbolic interpreter, State, EvalContext |
| 21 | `art/` | 750 | 900 | 2 | `ivy_art.py` (523 lines) | Analysis graph, Z3-backed cover/BMC |
| 22 | `check/` | 1,050 | 700 | 1 | `ivy_check.py` (1,041 lines) | Top-level verification driver |
| 23 | `mc/` | 2,200 | 1,800 | 3 | `ivy_mc.py` (1,772 lines) | AIGER encoding, model checking engine |
| | **TOTAL** | **25,975** | **16,669** | **39** | | |

**Grand total**: 42,644 Go lines across 103 files, ~1,360 tests + 39 fuzz tests, 21 packages.
All `go vet` clean. All tests passing.

---

### Chunk 1-3: `logic/` — Core Logic Types

**Source**: `logic.py` (488 lines)
**Go files**: `sort.go` (209), `node.go` (82), `term.go` (161), `formula.go` (549), `error.go` (19)
**Tests**: `sort_test.go` (253), `term_test.go` (248), `formula_test.go` (433) — **53 tests, 4 fuzz**

**What's in it**:
- `Sort` interface (embeds `Node`) with seal method; 6 concrete types: `UninterpretedSort`, `BooleanSort`, `FunctionSort`, `EnumeratedSort`, `RangeSort`, `TopSort`
- `Node` interface: `NodeSort()`, `Children()`, `String()`, `Equal(Node)`
- Terms: `Var` (uppercase name validation), `Const`, `Apply` (arity/sort validation)
- 15 formula types: `Eq`, `Ite`, `Not`, `Globally`, `Eventually`, `WhenOperator`, `Cond`, `And`, `Or`, `Implies`, `Iff`, `ForAll`, `Exists`, `Lambda`, `NamedBinder`
- `True = &And{}`, `False = &Or{}`
- Helpers: `SortEqual`, `FirstOrderSort`, `ContainsTopSort`, `IsPolymorphic`, `IsBooleanOrTop`

**Fuzz tests**:
- `FuzzUninterpretedSort` — arbitrary name strings for sort construction
- `FuzzNewFunctionSort` — variable arity (0-20 args)
- 2 more in sort_test.go and term_test.go

**Design decisions**: `Sort` embeds `Node` (solves `Equal(Sort)` vs `Equal(Node)` conflict). Python `recstruct.__iter__` → Go `Children()`. Python `frozenset(variables)` → sorted display for deterministic `String()`.

---

### Chunk 4: `typeinfer/` — Type Inference

**Source**: `type_inference.py` (419 lines)
**Go files**: `sortvar.go` (52), `unify.go` (195), `infer.go` (610)
**Tests**: `typeinfer_test.go` (250) — **14 tests, 1 fuzz**

**What's in it**:
- `SortOrVar` interface with `SortWrapper` (wraps `logic.Sort`) and `SortVar` (mutable union-find node)
- `Find()` with path compression, `Unify()` with occurs check, `OccursIn()`
- `ConvertFromSortVars()`, `ConvertToSortVars()`, `InsertSortVars()`
- `InferSorts()` handles all 15+ node types, returns `InferResult{Sort, Concretize}`
- `ConcretizeSorts()` — replaces TopSorts with concrete sorts

**Fuzz tests**:
- `FuzzUnify` — unification of sort variables against various SortOrVar combinations

**Note**: `collectNames()` is a simplified stand-in; should eventually use `logicutil.FreeVariables`.

---

### Chunk 5: `logicutil/` — Logic Utilities

**Source**: `logic_util.py` (335 lines)
**Go files**: `logicutil.go` (671)
**Tests**: `logicutil_test.go` (453) — **22 tests, 1 fuzz**

**What's in it**:
- `FreeVariables(t)` — by pointer identity; `FreeVariablesByName(t)` — by name
- `UsedVariables(t)`, `BoundVariables(t)`, `UsedConstants(t)`
- `Substitute(t, subs)` — simultaneous substitution with `CaptureError` detection
- `IsTautologyEquality(t)` — detects `Eq(x,x)`
- `EqualModAlpha(t, u)` — alpha-equivalence using `pushableMap`

**Fuzz tests**:
- `FuzzSubstitute` — variable substitution with boolean formula combinations

---

### Chunk 6: `ivyutils/` — General Utilities

**Source**: `ivy_utils.py` (773 lines)
**Go files**: `renamer.go` (135), `parameter.go` (189), `graph.go` (144), `names.go` (141)
**Tests**: `ivyutils_test.go` (368) — **27 tests, 2 fuzz**

**What's in it**:
- `UniqueRenamer` — unique names with suffix collision avoidance
- `VariableGenerator` — A,B,...,Z,AA,BB,... (skipping O)
- `ConstantNameGenerator()` — yields a-z, a0-z0, a1-z1,...
- `Parameter`, `BooleanParameter`, `EnumeratedParameter` with `Registry`
- `TopologicalSort[T,K]`, `Reachable[T,K]`, `FindCycle[K]` (Go generics)
- `ComposeNames`, `SplitName`, `BaseName`, `ParentChildName`, `ExtractParametersName`

**Fuzz tests**:
- `FuzzUniqueRenamer` — arbitrary prefixes and names
- `FuzzSplitName` — robustness of name parsing

---

### Chunk 7: `z3bridge/` — Z3 Integration

**Source**: `z3_utils.py` (197 lines) + partial `ivy_solver.py` (1,716 lines)
**Go files**: `quantifier.go` (549), `translate.go` (365)
**Tests**: `z3bridge_test.go` (419) — **22 tests, 0 fuzz** ⚠️

**What's in it**:
- Self-contained Z3 CGo wrapper (NOT using go-z3, which lacks quantifier support)
- `Context` with thread-safe mutex, `Sort`, `Expr`, `FuncDecl` with `runtime.SetFinalizer` ref counting
- Boolean ops: `Not`, `And`, `Or`, `Implies`, `Iff`, `Eq`, `Ite`
- **Quantifiers**: `ForAll`, `Exists` via `Z3_mk_forall_const`/`Z3_mk_exists_const`
- `Solver` with `Assert`, `Check`, `Push`, `Pop`, `Model`
- `Translator` — converts all Ivy `logic.Node` types to Z3 `Expr`
- `Implies(f1, f2)`, `IsSat(f)` convenience functions
- CGo flags: `-I/usr/local/opt/z3/include -L/usr/local/opt/z3/lib -lz3`

**⚠️ Missing fuzz tests**: Should fuzz `Translator.Translate()` with random logic trees.

**Design decisions**: Direct CGo (not go-z3) because go-z3 lacks ForAll/Exists. `newExpr`/`newSort`/`newFuncDecl` must be called with lock already held (avoids nested `ctx.do()` deadlock).

---

### Chunk 8: `ast/` — AST Node Types

**Source**: `ivy_ast.py` (1,965 lines, 157 Python classes)
**Go files**: `ast.go` (564), `formula.go` (318), `decl.go` (1,145), `sort.go` (147), `tactic.go` (292)
**Tests**: `ast_test.go` (642) — **51 tests, 0 fuzz** ⚠️

**What's in it**:
- Base types: `Node` interface, `Base` struct, `Location`, `NoneAST`
- Core: `Symbol`, `Atom`, `App`, `Variable`, `Old`, `This`, `Literal`, `Dot`, `Bracket`, `Tuple`
- Formulas: `And`, `Or`, `Not`, `Implies`, `Iff`, `Ite`, `Forall`, `Exists`, `Globally`, `Eventually`, `WhenOperator`, `Let`, `Definition`, `NamedBinder`, `Isa`, `Trigger`
- 40+ declaration types: `ModuleDecl`, `ObjectDecl`, `ActionDecl`, `TypeDecl`, `RelationDecl`, `ConstantDecl`, `AxiomDecl`, `PropertyDecl`, `IsolateDecl`, `MixinDecl`, all variants
- 15+ tactic types for proof scripts
- AST-level sort types: `ConstantSort`, `EnumeratedSort`, `StructSort`, `FunctionSort`, `RelationSort`, `Range`
- Helpers: `IsTrue()`, `IsFalse()`, `HasTemporal()`, `IsEquals()`
- `LabeledFormula` with atomic counter for unique IDs

**⚠️ Missing fuzz tests**: Should fuzz `Clone()` round-trips and `String()` output.

---

### Chunk 9: `lexer/` — Hand-Written Tokenizer

**Source**: `ivy_lexer.py` (304 lines)
**Go files**: `token.go` (241), `lexer.go` (552)
**Tests**: `lexer_test.go` (382) — **35 tests, 0 fuzz** ⚠️

**What's in it**:
- 47 token types + 120+ reserved keywords with version-gating (v1.0 through v2.0)
- Multi-character operator disambiguation (`->` vs `-`, `<->` vs `<=` vs `<`, `...` vs `..` vs `.`)
- Native quote (`<<<...>>>`) support, quoted strings, comment handling
- Unicode temporal operators (□, ◇)
- `Peek()` / `NextToken()` interface, `Tokenize()` convenience function

**Fuzz tests**:
- `FuzzLexer` — Arbitrary byte input with Version{1,7}, loops NextToken() up to 10K times
- `FuzzLexerAllVersions` — Fuzzes both input and version numbers

---

### Chunk 10: `parser/` — Recursive Descent Parser

**Source**: `ivy_parser.py` (3,161 lines) + `ivy_logic_parser.py` (655 lines)
**Go files**: `parser.go` (328), `expr.go` (470), `decl.go` (738), `action.go` (267)
**Tests**: `parser_test.go` (693) — **59 tests, 0 fuzz** ⚠️

**What's in it**:
- Pratt/precedence-climbing expression parser with 15 precedence levels
- All top-level declarations: type, relation, individual, function, axiom, property, conjecture, action, init, module, object, isolate, export, import, mixin, variant, definition, schema, theorem, attribute, etc.
- Action body parsing: if/else, while, for, local, let, var, assignment, call, assume/assert/require/ensure
- Proof body parsing (simplified)

**Fuzz tests**:
- `FuzzParser` — Arbitrary strings to Parse(), 40 seed inputs covering all declaration types
- `FuzzParserExpr` — Expression parsing with 51 seed inputs covering all expression forms

---

### Chunk 11: `ivylogic/` — Higher-Level Logic IR

**Source**: `ivy_logic.py` (1,774 lines)
**Go files**: `sig.go` (381), `ivylogic.go` (382), `classify.go` (240), `formula.go` (247), `poly.go` (114), `util.go` (613)
**Tests**: `ivylogic_test.go` (808) — **54 tests, 1 fuzz**

**What's in it**:
- `Sig` — first-order signature (sorts, symbols, constructors, interpretations)
- `SymbolEntry` with `UnionSort` for polymorphic symbol overloading
- `WithSymbols`, `WithSorts` — scoped symbol/sort additions with Enter/Exit
- Formula classification: `IsQF`, `IsPrenexUniversal`, `IsPrenexExistential`, `IsAE`, `IsEA`
- Formula types: `Some`, `Definition`, `DefinitionSchema`, `Let`, `Literal`, `Predicate`
- Simplification: `SimpAnd`, `SimpOr`, `SimpNot`, `SimpIte`
- Polymorphic symbols table: 21 built-in entries + dynamic `bfe[...]`
- `VariableUniqifier` — alpha-converts formulas for unique variable names
- Utilities: `CloneNode`, `CloneBinder`, `NodeArgs`, `CloseFormula`, `Extensionality`, `PartialFunction`, `ASTMatch`, `LabelTemporal`, `NormalizeOps`, `AlphaAvoid`
- Type predicates: `IsVariable`, `IsConstant`, `IsApp`, `IsAtom`, `IsQuantifier`, `IsBinder`, `IsTemporal`, `HasTemporal`, `IsNumeral`, `IsGprop`, etc.

**Fuzz tests**:
- `FuzzSigAddSymbol` — arbitrary symbol names and sort names

---

### Chunk 12: `module/` — Module Context

**Source**: `ivy_module.py` (412 lines)
**Go files**: `module.go` (491)
**Tests**: `module_test.go` (268) — **16 tests, 1 fuzz**

**What's in it**:
- `Module` — central container for all module-level state (40+ fields)
- `LabeledFormula` — formula with label, line number, temporal flag
- `IsolateInfo`, `MixinTriple` — isolate metadata
- `Clear()`, `Copy()` — reset and shallow-copy
- `AddToHierarchy()`, `AddObject()` — hierarchy management
- `FindAction()`, `IsVariant()`, `VariantIndex()`, `SortCard()`, `SortDependencies()`
- Helper types: `NamedAction`, `ProofEntry`, `NamedEntry`, `SubgoalEntry`

**Fuzz tests**:
- `FuzzModuleCopy` — copy module with random sorts/symbols, verify isolation

---

### Chunk 13: `actions/` — Action Semantics

**Source**: `ivy_actions.py` (1,687 lines, 49 Python classes)
**Go files**: `action.go` (849), `annotation.go` (191), `helpers.go` (176)
**Tests**: `actions_test.go` (647) — **41 tests, 2 fuzz**

**What's in it**:
- `Action` interface with `Clone`, `String`, `IterCalls`, `IterSubactions`, `GetFormalParams`/`GetFormalReturns`
- 20+ concrete action types: `Sequence`, `AssumeAction`, `AssertAction`, `RequireAction`, `EnsureAction`, `AssignAction`, `HavocAction`, `SetAction`, `IfAction`, `WhileAction`, `ChoiceAction`, `CallAction`, `LocalAction`, `LetAction`, `BindOldsAction`, `NativeAction`, `CrashAction`, `ThunkAction`, `EnvAction`, `ReturnAction`, `IgnoreAction`
- `Schema` — parametrized action schema wrapping a definition
- `RME` — requires-modifies-ensures clause
- Annotation types: `EmptyAnnotation`, `ConjAnnotation`, `ComposeAnnotation`, `RenameAnnotation`, `IteAnnotation`
- Helper functions: `ConcatActions`, `HasCode`, `CallSet`, `PrefixAction`, `PostfixAction`, `ApplyMixin`, `ParamsToStr`, `ActionDefToStr`

**Note**: `IntUpdate()` methods (computing transition relations) are stubbed — they depend on `Clauses` from unported `ivy_logic_utils.py`.

**Fuzz tests**:
- `FuzzActionClone` — clone random action types, verify independence
- `FuzzAnnotation` — construct random annotation trees, verify String()

---

### Chunk 14: `transrel/` — Transition Relations

**Source**: `ivy_transrel.py` (669 lines)
**Go files**: `transrel.go` (459)
**Tests**: `transrel_test.go` (782) — **38 tests, 5 fuzz**

**What's in it**:
- Symbol renaming: `New()`, `IsNew()`, `NewOf()`, `Old()`, `IsOld()`, `OldOf()` — prefix-based symbol versioning
- Skolem detection: `IsSkolem()`, `IsGlobalSkolem()`
- `Update` type: `(Modified []string, TR lg.Node, Pre lg.Node)` — the core transition relation triple
- State operations: `NullUpdate()`, `PureState()`, `IsPureState()`, `TopState()`, `BottomState()`
- Frame operations: `FrameDef()`, `Frame()`, `DiffFrame()`, `UpdatedJoin()`
- Composition stubs: `ComposeUpdates()`, `JoinAction()`, `IteAction()`, `Hide()`, `StateToAction()`, `ActionToState()`, `ForwardImage()`
- Types: `CounterExample`, `ActionFailed`, `History`
- Utility: `ListDiff()`, `ListUnion()`

**Note**: Composition and image operations are stubbed — they require `Clauses` infrastructure from `ivy_logic_utils.py`.

**Fuzz tests**:
- `FuzzNewOld` — round-trip symbol renaming
- `FuzzIsSkolem` — arbitrary strings for skolem detection
- `FuzzIsGlobalSkolem` — arbitrary strings for global skolem detection
- `FuzzUpdatedJoin` — union of modified symbol lists
- `FuzzFrameDef` — frame definition construction

---

### Chunk 15: `proof/` — Proof Checker

**Source**: `ivy_proof.py` (1,653 lines)
**Go files**: `errors.go` (72), `match.go` (578), `goal.go` (338), `checker.go` (138), `skolem.go` (289)
**Tests**: `proof_test.go` (846) — **48 tests, 2 fuzz**

**What's in it**:
- Error types: `Redefinition`, `Circular`, `NoMatch`, `ProofError`, `CaptureError`
- `MatchProblem` — schema-instantiation matching problem
- `ProofChecker` — main proof verification engine with axioms, definitions, schemata
- Match/unification: `Match()`, `MatchQuants()`, `FOMatch()`, `HeadsMatch()`, `FuncsMatch()`, `MergeMatches()`, `EquivAlpha()`
- Match operations: `ApplyMatch()`, `ApplyMatchSym()`, `ApplyMatchFunc()`, `ApplyMatchFreesyms()`, `ComposeMatches()`, `ExtractTerms()`
- Goal operations: `GoalPrems()`, `GoalConc()`, `CloneGoal()`, `NormalizeGoal()`, `GoalVocab()`, `GoalFree()`, `TrivialGoal()`, `CheckConcsMatch()`
- Skolemization: `SkolemizeGoal()`, `SkolemizeFmla()` — converts goals to skolem normal form
- Symbol scope: `AddSymbols`, `RemoveSymbols` — temporary set modification with `Restore()`
- Tactic registry: `RegisteredTactics`, `RegisterTactic()`
- Helper types: `Vocab` (sorts, symbols, variables)

**Fuzz tests**:
- `FuzzMergeMatches` — random match merging with conflict detection
- `FuzzMatch` — pattern matching with free/bound variable combinations

---

## PART 2: FUZZ TEST GAP ANALYSIS

### Current State

| Package | Tests | Fuzz | Status |
|---------|-------|------|--------|
| `logic/` | 53 | 4 | ✅ Good coverage |
| `typeinfer/` | 14 | 1 | ✅ Adequate |
| `logicutil/` | 22 | 1 | ✅ Adequate |
| `ivyutils/` | 27 | 2 | ✅ Good coverage |
| `z3bridge/` | 22 | **0** | ⚠️ Missing |
| `ast/` | 51 | **0** | ⚠️ Missing |
| `lexer/` | 37 | **2** | ✅ Covered |
| `parser/` | 61 | **2** | ✅ Covered |
| `ivylogic/` | 54 | 1 | ✅ Covered |
| `module/` | 16 | 1 | ✅ Covered |
| `actions/` | 41 | 2 | ✅ Covered |
| `transrel/` | 38 | 5 | ✅ Good coverage |
| `proof/` | 48 | 2 | ✅ Covered |

### Recommended Fuzz Tests to Add

**Priority 1 — Parser & Lexer** (crash-prone, arbitrary input):
1. **`FuzzLexer`** — Feed arbitrary `[]byte` to `Tokenize()`. Verify no panics, no infinite loops, always terminates with EOF or ERROR.
2. **`FuzzParser`** — Feed arbitrary strings to `Parse()`. Verify no panics. Optionally verify that if `Parse` succeeds, every returned node satisfies the `Node` interface.
3. **`FuzzParserRoundTrip`** — Parse valid Ivy snippets, `String()` the AST, re-parse. Check structural equivalence.

**Priority 2 — AST** (data structure integrity):
4. **`FuzzASTClone`** — Build random AST nodes, clone them, verify the clone is independent (modifying the clone doesn't affect the original).
5. **`FuzzASTString`** — Build random AST nodes, call `String()`, verify no panics.

**Priority 3 — Z3 Bridge** (CGo boundary):
6. **`FuzzTranslator`** — Build random `logic.Node` trees (using the existing logic types), translate to Z3 via `Translator.Translate()`. Verify no panics or Z3 crashes.

---

## PART 3: REMAINING PYTHON CODEBASE

### Python File Inventory (54,099 total lines)

**Already ported** (13,971 lines of Python → 14,537 lines of Go source):

| Python File | Lines | Go Package |
|-------------|-------|------------|
| `logic.py` | 488 | `logic/` |
| `type_inference.py` | 419 | `typeinfer/` |
| `logic_util.py` | 335 | `logicutil/` |
| `ivy_utils.py` | 773 | `ivyutils/` |
| `z3_utils.py` | 197 | `z3bridge/` |
| `ivy_solver.py` | 1,716 | `z3bridge/` (partial) |
| `ivy_ast.py` | 1,965 | `ast/` |
| `ivy_lexer.py` | 304 | `lexer/` |
| `ivy_parser.py` | 3,161 | `parser/` |
| `ivy_logic_parser.py` | 655 | `parser/` |
| `ivy_logic.py` | 1,774 | `ivylogic/` |
| `ivy_module.py` | 412 | `module/` |
| `ivy_actions.py` | 1,687 | `actions/` |
| `ivy_transrel.py` | 669 | `transrel/` |
| `ivy_proof.py` | 1,653 | `proof/` |

**Not yet ported** — organized by tier:

**Note**: `ivy_logic.py` (1,774 lines) and `ivy_transrel.py` (669 lines) are *partially* ported — the core data types are in `logic/` and `logicutil/`, but the higher-level wrappers (Sig management, state versioning, interpolation) still need porting.

#### Tier 2: Compiler & Verification (7,035 lines)

| File | Lines | Role | Key Dependencies |
|------|-------|------|------------------|
| `ivy_compiler.py` | 2,320 | AST → logic IR compilation | parser, actions, logic, module, theory, isolate |
| `ivy_isolate.py` | 2,022 | Modular verification — isolate extraction | logic, actions, module, solver |
| `ivy_check.py` | 1,040 | Top-level verification checker | compiler, isolate, proof, actions, module |
| `ivy_proof.py` | 1,653 | Proof state management and tactics | logic, utils, ast |

#### Tier 3: Symbolic Execution & Model Checking (3,118 lines)

| File | Lines | Role | Key Dependencies |
|------|-------|------|------------------|
| `ivy_interp.py` | 642 | Symbolic interpreter | logic, solver, transrel, actions |
| `ivy_art.py` | 523 | Abstract Reachability Tree (CEGAR) | actions, interp, module |
| `ivy_mc.py` | 1,772 | Model checking engine + AIGER encoding | module, actions, logic, transrel, art |
| `ivy_theory.py` | 181 | Built-in theories (int, nat, bv[n]) | actions, module, logic |

#### Tier 4: Support Libraries (5,577 lines)

| File | Lines | Role |
|------|-------|------|
| `ivy_logic_utils.py` | 1,635 | Clause operations, formula manipulation (partially in logicutil/) |
| `ivy_fragment.py` | 591 | EPR/decidable fragment detection |
| `ivy_unitres.py` | 546 | Unit resolution |
| `ivy_ranking.py` | 1,195 | Termination/ranking analysis |
| `ivy_l2s.py` | 1,522 | Liveness-to-safety reduction |
| `ivy_temporal.py` | 436 | Temporal property handling |
| `ivy_trace.py` | 427 | Counterexample trace generation |
| `ivy_alpha.py` | 225 | Alpha-renaming utilities |

#### Tier 5: Code Generation (8,375 lines) — Lower priority

| File | Lines | Role |
|------|-------|------|
| `ivy_to_cpp.py` | 6,715 | C++ code generation |
| `ivy_cpp_types.py` | 525 | C++ type system |
| `ivy_cpp.py` | 414 | C++ compilation utilities |
| `ivy_dafny_compiler.py` | 478 | Dafny output |
| `ivy_to_md.py` | 47 | Markdown output |
| Others | ~196 | Lean, SMT-LIB, VMT outputs |

#### Tier 6: UI/Visualization (~7,000 lines) — Not a porting priority

| File | Lines | Role |
|------|-------|------|
| `widget_analysis_session.py` | 1,648 | Analysis UI widget |
| `concept.py` | 865 | Concept visualization |
| `ivy_graph.py` | 783 | Graph display |
| `ivy_ui_cti.py` | 722 | CTI UI |
| `ivy_ui.py` | 622 | Main UI framework |
| `ivy_graph_ui.py` | 559 | Graph UI |
| `tk_graph_ui.py` | 430 | Tk graph widgets |
| Others | ~1,400 | Misc UI components |

### Totals by Tier

| Tier | Lines | Status |
|------|-------|--------|
| Already ported | 9,962 | ✅ Done (9,133 Go + 4,141 test) |
| Tier 1: Core Semantics | 4,542 | Next up |
| Tier 2: Compiler & Verification | 7,035 | Pending |
| Tier 3: Symbolic Exec & MC | 3,118 | Pending |
| Tier 4: Support Libraries | 5,577 | Pending |
| Tier 5: Code Generation | 8,375 | Lower priority |
| Tier 6: UI/Visualization | ~7,000 | Not planned |
| Small utilities | ~1,490 | As needed |

---

## PART 4: DETAILED PLANS FOR NEXT CHUNKS

### Chunk 11: `ivylogic/` — Higher-Level Logic IR

**Source**: `ivy_logic.py` (1,774 lines) — the parts NOT already in `logic/`

The `logic/` package has the core data types. `ivy_logic.py` adds:
- `Sig` — signature maintaining all symbols, sorts, and their relationships
- `Symbol` — higher-level constant/relation with name mangling
- Sort inference and resolution: `find_sort()`, `add_sort()`, `find_symbol()`
- Formula classification: `IsQF()`, `IsPrenexUniversal()`, etc.
- `WithSymbols`, `WithSorts` — context-like sort/symbol scoping
- `VariableUniqifier` — unique variable renaming
- `PolySymsDict` — polymorphic symbol resolution
- `extensionality()`, `exclusivity()` — axiom generators for enumerated sorts

**Estimated**: ~800 lines Go + 300 test

**Fuzz tests needed**:
- `FuzzSigAddSymbol` — add random symbols to a Sig, verify consistency

---

### Chunk 12: `module/` — Module Context

**Source**: `ivy_module.py` (412 lines)

Central container for all module-level state:
- `Module` class — holds declarations, relations, definitions, actions, axioms, schemata
- `clear()`, `copy()`, `get_axioms()`, `background_theory()`
- `call_graph()` — action call graph construction
- `sort_card()` — sort cardinality
- `IsolateInfo` — metadata about isolates

**Estimated**: ~400 lines Go + 200 test

**Fuzz tests needed**:
- `FuzzModuleCopy` — copy module, modify copy, verify original unchanged

---

### Chunk 13: `actions/` — Action Semantics

**Source**: `ivy_actions.py` (1,687 lines, 49 classes)

Action class hierarchy (semantic level, post-compilation):
- `Action` base with `UpdateAction`, `AssumeAction`, `AssertAction`
- `AssignAction`, `HavocAction`, `SetAction`
- `IfAction`, `WhileAction`, `ChoiceAction`, `Sequence`
- `CallAction`, `LocalAction`, `LetAction`
- `NativeAction`, `CrashAction`, `ThunkAction`
- `Schema` — parametrized action schemas
- `ActionContext`, `UnrollContext` — evaluation contexts
- Type checking: `type_check()`, `type_ast()`

**Dependencies**: logic, logicutil, ivyutils, transrel (partial), module, ast

**Estimated**: ~1,200 lines Go + 400 test

**Fuzz tests needed**:
- `FuzzActionTypeCheck` — random action trees, verify type checking doesn't panic

---

### Chunk 14: `transrel/` — Transition Relations (remaining)

**Source**: `ivy_transrel.py` (669 lines) — the parts NOT already in `logicutil/`

The remaining functions:
- Symbol renaming: `new()`, `is_new()`, `old()`, `is_old()`, `rename()`
- State operations: `null_update()`, `pure_state()`, `state_to_action()`, `action_to_state()`
- Update composition: `compose_updates()`, `join_action()`, `ite_action()`
- Interpolation: `History` class, `forward_image()`, `interpolant()`
- Constraint handling: `frame()`, `hide()`, `constrain_state()`

**Estimated**: ~500 lines Go + 200 test

---

### Chunk 15: `proof/` — Proof Checking

**Source**: `ivy_proof.py` (1,653 lines)

Self-contained proof engine:
- `ProofChecker` — main proof verification engine
- Schema instantiation: `inst_schema()`, `unify_with_goal()`, `match_goal()`
- Goal management: `make_goal()`, `traverse_goals()`, `list_goals()`
- `MatchProblem` — matching premises to goals
- Error types: `Redefinition`, `Circular`, `NoMatch`, `ProofError`, `CaptureError`

**Dependencies**: logic, logicutil, ast, ivyutils (relatively self-contained)

**Estimated**: ~900 lines Go + 400 test

**Fuzz tests needed**:
- `FuzzProofCheckerMatch` — random formula matching

---

### Chunk 16: `compiler/` — AST → Logic IR

**Source**: `ivy_compiler.py` (2,320 lines)

The big one:
- `IvyDeclInterp` — declaration interpreter
- Context classes: `Context`, `ReturnContext`, `ExprContext`
- Expression compilation: `compile_app()`, `compile_inline_call()`, `compile_method_call()`
- Declaration compilation: `ConstantDecl_cmpl()`, `compile_local()`
- Sort inference during compilation
- Module instantiation with parameter substitution

**Dependencies**: parser, actions, logic, module, theory, isolate — nearly everything

**Estimated**: ~1,500 lines Go + 600 test

---

### Chunk 17: `isolate/` — Modular Verification

**Source**: `ivy_isolate.py` (2,022 lines)

Isolate extraction and analysis:
- `lookup_action()`, `add_mixins()`, `summarize_action()`
- Strip/abstraction: `strip_sort()`, `strip_action()`, `strip_isolate()`
- Side effect analysis: `has_side_effect()`
- Dependency analysis: `get_calls_mods()`, `check_interference()`
- Path analysis utilities

**Dependencies**: logic, actions, module, solver, ast

**Estimated**: ~1,200 lines Go + 400 test

---

### Chunk 18: `theory/` — Built-in Theories

**Source**: `ivy_theory.py` (181 lines) — smallest file

- `Theory` base, `IntegerTheory`, `NaturalTheory`, `BitVectorTheory`
- `theories()`, `parse_theory()`, `get_sort_theory()`, `has_integer_interp()`

**Estimated**: ~150 lines Go + 100 test

---

### Chunk 19: `interp/` — Symbolic Interpreter

**Source**: `ivy_interp.py` (642 lines)

- `State` — abstract state with clauses and preconditions
- `EvalContext` — evaluation context
- State operations: `new_state()`, `concrete_post()`, `concrete_join()`
- `apply_action()` — symbolic action application

**Estimated**: ~500 lines Go + 200 test

---

### Chunk 20: `art/` — Abstract Reachability Tree

**Source**: `ivy_art.py` (523 lines)

- `AnalysisGraph` — CEGAR-style reachability analysis
- State tracking: `add()`, `push_state()`, `pop_state()`
- `apply_action()`, `join_states()`, `refine()`
- Witness management

**Estimated**: ~400 lines Go + 200 test

---

### Chunk 21: `check/` — Verification Driver

**Source**: `ivy_check.py` (1,040 lines)

Top-level verification:
- `Checker`, `ConjChecker` — proof checking drivers
- `check_conjectures()`, `check_temporals()`, `check_properties()`
- Counterexample display

**Dependencies**: Nearly everything

**Estimated**: ~700 lines Go + 300 test

---

### Chunk 22: `mc/` — Model Checking

**Source**: `ivy_mc.py` (1,772 lines)

- `Aiger` class — AIGER circuit format
- `Encoder` — formula-to-circuit encoding
- Schema/axiom instantiation for bounded checking
- Quantifier elimination: `Qelim` class

**Estimated**: ~1,000 lines Go + 400 test

---

## PART 5: IMPLEMENTATION STRATEGY

### Recommended Order

```
DONE: logic → typeinfer → logicutil → ivyutils → z3bridge → ast → lexer → parser
                                                                              |
DONE: ─── ivylogic ─── module ─── actions ── transrel ─── proof              |
                                                              |               |
NEXT: ─── theory ─── compiler ───────────────────────┐        |               |
                          |                          |        |               |
                     interp ── art                   |        |               |
                          |        \                 |        |               |
                     isolate ──────────              |        |               |
                          |                          |        |               |
                     check ──────────────────────────┘        |               |
                          |                                   |               |
                       mc ────────────────────────────────────┘               |
```

### Immediate Next Steps (Priority Order)

1. ~~Add fuzz tests to lexer and parser~~ ✅ Done
2. **Add fuzz test to z3bridge** — CGo boundary is crash-prone.
3. ~~Port Chunk 11: `ivylogic/`~~ ✅ Done
4. ~~Port Chunk 12: `module/`~~ ✅ Done
5. ~~Port Chunk 13: `actions/`~~ ✅ Done
6. ~~Port Chunk 14: `transrel/`~~ ✅ Done
7. ~~Port Chunk 15: `proof/`~~ ✅ Done
8. **Port `ivy_logic_utils.py`** — Clauses type and clause operations (needed to fill in transrel/actions stubs).
9. **Port Chunk 16: `compiler/`** — AST → logic IR compilation.
10. **Port Chunk 17: `isolate/`** — modular verification.
11. **Port Chunk 18: `theory/`** — built-in theories.

### Verification After Each Chunk

1. `go build ./...` — compiles
2. `go test ./... -v` — all unit tests pass
3. `go test ./... -fuzz=. -fuzztime=30s` — fuzz tests find no crashes
4. `go vet ./...` — no warnings
5. Cross-reference with Python where applicable

### Estimated Remaining Work

| Phase | Chunks | Estimated Go Lines | Estimated Tests |
|-------|--------|-------------------|-----------------|
| ~~Fuzz gap fill~~ | ✅ | ~~+300~~ done | +4 fuzz added |
| ~~Core Semantics pt1~~ | ✅ 11-12 | ~~~1,200~~ done | ~1,076 done |
| ~~Core Semantics pt2~~ | ✅ 13-15 | ~~2,936~~ done | ~2,275 done |
| Compiler & Verification (Tier 2) | 16-17 | ~2,700 | ~1,000 |
| Symbolic Exec & MC (Tier 3) | 18-22 | ~2,750 | ~1,200 |
| **Subtotal (remaining core)** | | **~5,450** | **~2,200** |
| Support Libraries (Tier 4) | — | ~3,000 | ~1,200 |
| Code Generation (Tier 5) | — | ~5,000 | ~1,500 |
| **Full remaining total** | | **~13,450** | **~4,900** |

Combined with existing 22,165 lines, the full core port (through Tier 3) would be ~27,615 lines of Go.

---

## PART 6: KEY PYTHON SOURCE FILE PATHS

For reference during implementation:

```
/Users/jaten/pyivy/ivy/ivy/ivy_logic.py         — higher-level logic IR (1,774 lines)
/Users/jaten/pyivy/ivy/ivy/ivy_module.py         — module context (412 lines)
/Users/jaten/pyivy/ivy/ivy/ivy_actions.py        — action semantics (1,687 lines)
/Users/jaten/pyivy/ivy/ivy/ivy_transrel.py       — transition relations (669 lines)
/Users/jaten/pyivy/ivy/ivy/ivy_proof.py          — proof checking (1,653 lines)
/Users/jaten/pyivy/ivy/ivy/ivy_compiler.py       — AST → logic compilation (2,320 lines)
/Users/jaten/pyivy/ivy/ivy/ivy_isolate.py        — modular verification (2,022 lines)
/Users/jaten/pyivy/ivy/ivy/ivy_theory.py         — built-in theories (181 lines)
/Users/jaten/pyivy/ivy/ivy/ivy_interp.py         — symbolic interpreter (642 lines)
/Users/jaten/pyivy/ivy/ivy/ivy_art.py            — abstract reachability (523 lines)
/Users/jaten/pyivy/ivy/ivy/ivy_check.py          — verification driver (1,040 lines)
/Users/jaten/pyivy/ivy/ivy/ivy_mc.py             — model checking (1,772 lines)
/Users/jaten/pyivy/ivy/ivy/ivy_logic_utils.py    — clause operations (1,635 lines)
/Users/jaten/pyivy/ivy/ivy/ivy_solver.py         — Z3 solver bridge (1,716 lines)
```
