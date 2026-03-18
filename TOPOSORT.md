# Topological Sort of FIXME.md Porting Tasks

## Context

The 55 incomplete porting tasks in `FIXME.md` need to be ordered so that no task is blocked by an unfinished dependency. Key infrastructure packages (solver, typeinfer, parser, proof/ProofChecker, actions) are already substantially implemented in Go, which unblocks most items.

## Dependency Graph (FIXME→FIXME edges)

```
30 → 11          (CompileMatchList needs CompileExprVocab)
15 → 30          (CompileMatchFull needs CompileMatchList)
13 → 14 → 16    (TransformDefnMatch → ParameterizeSchema → MakeDistinctVars)
20 → 21          (CheckProperties needs ReorderProps)
34 → 33          (CompileTheories needs CompileTheory)
43 → 23 AND 33   (IvyCompileTheoryFromString needs ReadModule + CompileTheory)
28 → 27 AND 41   (AigerWitnessToIvyTrace needs MatchAnnotation + MatchHandler.Eval)
47 → 27          (MatchAnnotationMC needs MatchAnnotation)
9 → 5            (DecomposeActionApp needs SmallModelClauses via History.satisfy)
7 → solver.Interpolant (⚠️ Z3 interpolation API unavailable in this build)
```

---

## Implementation Order

### Session 1 — Trivial leaf fixes in mc/phase7.go
**Items: 25, 26, 42**
| # | Function | What to do |
|---|----------|------------|
| 25 | `Checked` | Add `CheckedAssert` parameter; compare `action.GetLineno()` |
| 26 | `Badwit` | Return error "model checker returned mis-formatted witness" |
| 42 | `CloneNormal` | Filter tautological equalities, canonicalize Eq arg order |

### Session 2 — Trivial leaf fixes in compiler/phase6.go (batch A)
**Items: 40, 48, 49, 52, 35, 39**
| # | Function | What to do |
|---|----------|------------|
| 40 | `CompileCrashAction` | Change `NewHavocAction` → `NewCrashAction` |
| 48 | `SetVerifying` | Expose `optionVerifying` via getter |
| 49 | `CompileIfTactic` | Propagate errors from branch compilation |
| 52 | `OtherThing` | `isSortInferRoot` returns true for Assign/Set/Havoc/Assume/AssertAction |
| 35 | `AddLabelsToProof` | Recursive walk propagating labels through ComposeTactics/IfTactic |
| 39 | `PropToDef` | Handle LabeledFormula: extract formula, call `il.DropUniversals`, check Definition |

### Session 3 — Small standalone fixes (actions + transrel)
**Items: 2, 31**
| # | Function | File | What to do |
|---|----------|------|------------|
| 2 | `AssignRefs` | actions/phase3.go | Replace `collectSymbols` with `co.UsedSymbolsAST`; add destructor symbol |
| 31 | `Implies` | transrel/phase4.go | Add type-switch: raw formula → `IsPrenexUniversal` + `ClausesImplyFormulaCex`; Clauses → `ClausesImply` |

### Session 4 — Compiler leaf fixes (batch B)
**Items: 36, 32, 53, 54, 55**
| # | Function | What to do |
|---|----------|------------|
| 36 | `CompileDebugAction` | Use `NewDebugAction(compiledExpr, withExprs...)` |
| 32 | `CheckInstantiations` | Iterate `mod.Instantiations`, validate schema names in `mod.Schemata` |
| 53 | `CompileRootArgs` | For bare Atom names, look up in `c.Sig.Symbols` via `il.FindSymbol` |
| 54 | `CompileNativeArg` | Check if atom name is in `mod.Actions`, resolve alias |
| 55 | `CompileNativeSymbol` | Add `mod.DestructorSorts` check + polymorphic `PolySymsDict` wrapper |

### Session 5 — Schema compilation pipeline
**Items: 44, 45, 46, 50, 51**
| # | Function | What to do |
|---|----------|------------|
| 44 | `CompileSchemaPrem` | Add DerivedDecl + PropertyDecl cases; add compiled symbols to `c.Sig` |
| 45 | `CompileSchemaBody` | Preserve label, temporal, explicit, assumed attributes via `CloneWithFreshID` |
| 46 | `CompileSchemaConc` | Add TemporalModels case; use `il.NewWithSymbols` for sort context |
| 50 | `CompilePropertyTactic` | Compile `pt.Prop` via `c.Sortify`; handle definition with `c.CompileDefn` |
| 51 | `CompileProofTactic` | Compile `pt.TLabel` via `c.Sortify` before recursing |

### Session 6 — Sort inference hint pair
**Items: 37, 38** (must be done together — same fix pattern)
| # | Function | What to do |
|---|----------|------------|
| 37 | `SortInferCovariant` | Call `SortInfer(term, sort)` with hint first; fallback to `SortInfer(term)` |
| 38 | `SortInferContravariant` | Same as 37 but reversed arg order in `IsVariant(sort, termSort)` |

### Session 7 — Solver-integration wiring (transrel + interp)
**Items: 4, 5, 6, 8, 10**
| # | Function | File | What to do |
|---|----------|------|------------|
| 4 | `ExtractPrePostModel` | transrel/phase4.go | Wire `solver.ClausesModelToClauses(clauses, ignore, model, numerals)` |
| 5 | `SmallModelClauses` | transrel/phase4.go | Pass `il.Sig.Sorts` as sortsToMinimize; thread `finalCond` + `shrink` to `GetSmallModelWithCond` |
| 6 | `GetCore` | interp/phase4.go | Call `solver.UnsatCore` with conjoined state+axioms and negated clause |
| 8 | `UnderapproximateState` | interp/phase4.go | Wire `solver.ClausesModelToClauses` with `transrel.IsSkolem` filter |
| 10 | `EvalAssertRhs` | interp/phase4.go | Wrap non-RME in `RME(And(), nil, rhs)`, call `EvalState` within `ActionContext` |

### Session 8 — Actions destructor chain (complex)
**Items: 1, 3**
| # | Function | What to do |
|---|----------|------------|
| 1 | `DestrAsgnVal` | Port skolem nondeterministic value (`sym.Skolem()`), destructor frame conditions loop over `sort_destructors`, equality guard formulas (`eqs` + `Or(And(*eqs), equiv_ast(dlhs, drhs))`) |
| 3 | `MakeFieldUpdate` | Construct `AssignAction(f(l,v), r(v))`, call `aa.ActionUpdate(domain, pvars)`, return `*transrel.Update` |

### Session 9 — Proof matching leaves
**Items: 16, 12, 17** (all in proof/phase5_matching.go, no mutual deps)
| # | Function | What to do |
|---|----------|------------|
| 16 | `MakeDistinctVars` | After creating V0,V1,..., call `co.RenameVariablesDistinctAsts(vars, asts)` |
| 12 | `RemoveVarsMatch` | Classify by `il.IsUISort` vs `il.IsConstant`; call `il.RenameVarsNoClash` on symbol-match values |
| 17 | `RenameGoal` | Implement recursive `rec_goal`: GoalDefns → match → CheckAlphaCapture → ApplyMatchGoalNode → AlphaRename → Rename |

### Session 10 — Proof matching: CompileExprVocab
**Items: 11** (unblocks items 30 and 15)
| # | Function | What to do |
|---|----------|------------|
| 11 | `CompileExprVocab` | Push vocab symbols/sorts onto sig, compile expr, call `typeinfer.InferSorts` (already working in typeinfer/infer.go), return inferred result |

### Session 11 — Proof matching: mid-chain
**Items: 30, 14** (30 depends on 11; 14 depends on 16)
| # | Function | What to do |
|---|----------|------------|
| 30 | `CompileMatchList` | For each match Definition, compile LHS via `CompileExprVocab(leftVocab)` and RHS via `CompileExprVocab(rightVocab)` |
| 14 | `ParameterizeSchema` | Create fresh vars via `MakeDistinctVars`; extend symbol sorts with `il.FuncConstSort`; build Lambda match; apply to conclusion |

### Session 12 — Proof matching: top of chain
**Items: 15, 13** (15 depends on 30; 13 depends on 14)
| # | Function | What to do |
|---|----------|------------|
| 15 | `CompileMatchFull` | Iterate compiled match list, call `CompileOneMatch` for each, collect results, call `MergeMatches` |
| 13 | `TransformDefnMatch` | Implement all 8 steps: extract decl/conc syms, build vmap, apply vmap, build dmatch, apply dmatch, filter constants, apply_match_goal, return new MatchProblem |

### Session 13 — Theory compilation chain
**Items: 33, 34** (34 depends on 33)
| # | Function | What to do |
|---|----------|------------|
| 33 | `CompileTheory` | Look up theory name in theory registry (`theory.GetSortTheory`), generate axioms, register on module |
| 34 | `CompileTheories` | Iterate `mod.Interps`, call `CompileTheory` for each sort interpretation |

### Session 14 — Module loading + ReorderProps
**Items: 21, 23** (21 is leaf needed by 20; 23 is leaf needed by 43)
| # | Function | What to do |
|---|----------|------------|
| 21 | `ReorderProps` | Separate spec properties by checking `mod.Attributes`; group by parent name; insert specs before parent; reverse |
| 23 | `ReadModule` etc. | Wire to existing parser: read file, detect `#lang ivy` version, call `parser.New().Parse(s)`; chain IvyLoadFile→ReadModule, IvyFromString→IvyLoadFile |

### Session 15 — Theory-from-string + assert proofs
**Items: 43, 24** (43 depends on 23+33; 24 is standalone)
| # | Function | What to do |
|---|----------|------------|
| 43 | `IvyCompileTheoryFromString` | After `IvyFromString` returns parsed module, call `CompileTheory(decls, sort)` |
| 24 | `ApplyAssertProof/Proofs` | Iterate `mod.Actions`, walk subactions, find AssertAction nodes, look up `mod.Proofs` by ID, attach |

### Session 16 — CheckProperties (largest single function)
**Items: 20** (depends on 21, ProofChecker infrastructure)
| # | Function | What to do |
|---|----------|------------|
| 20 | `CheckProperties` | Call `ReorderProps`; build proof/instance maps; create `ProofChecker` from axioms/definitions/schemata; iterate non-temporal properties calling `admit_proposition`; apply `named_trans` specialization; collect subgoals |

### Session 17 — MC annotation leaves
**Items: 41, 27** (27 needed by 47 and 28)
| # | Function | What to do |
|---|----------|------------|
| 41 | `MatchHandler.Eval` | Look up `cond` symbol in `h.Model` (Z3 model), extract Boolean value |
| 27 | `MatchAnnotation` (mc) | Implement recursive `recur`: RenameAnnotation, Sequence/ComposeAnnotation, IfAction/IteAnnotation, ChoiceAction/unite_annot, CallAction wrapping, base case handler.Handle |

### Session 18 — MC annotation dependents
**Items: 47, 28** (47 depends on 27; 28 depends on 27+41)
| # | Function | What to do |
|---|----------|------------|
| 47 | `MatchAnnotationMC` | Verify/extend to handle ChoiceAction+UniteAnnot and CallAction+Sequence(Ignore,callee,Return) |
| 28 | `AigerWitnessToIvyTrace` | Read AIGER witness file, step simulator, call `MatchAnnotation` with `AigerMatchHandler`, decode latch values, build `IvyMCTrace` with concrete state equalities |

### Session 19 — Complex standalone compiler items
**Items: 19, 18**
| # | Function | What to do |
|---|----------|------------|
| 19 | `InferParameters` | Collect action/mixin declarations; map mixer→mixee; compare formal counts; extend formals; rewrite via `ast.SubstPrefixAtomsAst` |
| 18 | `CompileThunkAction` | All 9 steps: copy sig, compile formals, compile body via sortify, collect fml:/loc: symbols, create sub-sort + $self param, create destructors, build substitution, register run action, build LocalAction |

### Session 20 — DecomposeActionApp (requires Session 7 done)
**Items: 9** (depends on item 5 for SmallModelClauses)
| # | Function | What to do |
|---|----------|------------|
| 9 | `DecomposeActionApp` | After `h.Satisfy(bg)` returns non-nil, unpack (universe, path); for each step create State with Expr, Update, Pred, Universe; requires `History.Satisfy` to return structured data |

### Session 21 — BLOCKED on Z3 Interpolation API
**Items: 7**
| # | Function | Blocker |
|---|----------|---------|
| 7 | `ReverseJoinConcreteClauses` | `solver.BinaryInterpolant` returns error "Z3 interpolation API not available in this build". Options: (a) use unsat-core overapproximation fallback, (b) build Z3 with interpolation (deprecated in Z3 ≥4.8), (c) integrate external interpolation engine |

### Session 22 — SKIP (dead code / low priority)
**Items: 22, 29**
| # | Function | Reason |
|---|----------|--------|
| 22 | `CompileSchemaInstantiation` | Python returns `self` (identity); Go returning input is correct |
| 29 | `GuiArt` | GUI feature not needed for CLI; low priority |

---

## Verification

After each session, run:
```bash
cd /Users/jaten/go/src/github.com/glycerine/goivy && go build ./...
```

For web-integrated tests:
```bash
DYLD_LIBRARY_PATH=/Users/jaten/go/src/github.com/glycerine/goivy/z3ivy/lib:$DYLD_LIBRARY_PATH go test -v ./webui -count=1 -tags web
```

For package-specific tests after each session:
```bash
go test -v ./actions/... -count=1      # after Sessions 3, 8
go test -v ./transrel/... -count=1     # after Session 7
go test -v ./interp/... -count=1       # after Sessions 7, 20
go test -v ./proof/... -count=1        # after Sessions 9-12
go test -v ./compiler/... -count=1     # after Sessions 2, 4-6, 13-16, 19
go test -v ./mc/... -count=1           # after Sessions 1, 17-18
```

## Key Files
- **compiler/phase6.go** — 30 of 55 items (Sessions 2, 4, 5, 6, 13, 14, 15, 16, 19)
- **proof/phase5_matching.go** — 8 items in a strict chain (Sessions 9-12)
- **mc/phase7.go** — 7 items (Sessions 1, 17, 18)
- **interp/phase4.go** — 5 items (Sessions 7, 20, 21)
- **transrel/phase4.go** — 3 items (Sessions 3, 7)
- **actions/phase3.go** — 3 items (Sessions 3, 8)
- **Python source of truth** — `/Users/jaten/pyivy/ivy/ivy/` (ivy_actions.py, ivy_transrel.py, ivy_interp.py, ivy_proof.py, ivy_compiler.py, ivy_mc.py)
