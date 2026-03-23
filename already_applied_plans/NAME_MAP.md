# NAME_MAP.md — Python-to-Go Name Correspondence

Generated: 2026-03-17

This documents every Python→Go name mapping in the Ivy port, flags
problematic renames, and notes what should be renamed to conform.

## RULES FOR FUTURE PORTING

```
1. Python class name -> Go struct name: SAME NAME. App stays App.
2. Python function name -> Go function name: SAME NAME, capital first letter, PascalCase.
   substitute_ast -> SubstituteAst.
3. Python field name -> Go field name: SAME NAME, capital first letter.
   self.rep -> Rep. self.args -> Args.
4. One Python file -> one Go file with the same base name.
5. Do not merge Python classes.
6. Do not omit functions.
7. Do not add abstractions that don't exist in Python.
8. When in doubt, translate literally.
```

---

## 1. AST Layer: ivy_ast.py -> ast/

| Python | Go | RENAME NEEDED? |
|---|---|---|
| `AST` (base) | `ast.Node` (interface) + `ast.Base` | OK — Go doesn't have class inheritance |
| `Formula` (base) | (no separate type) | Consider adding `Formula` interface |
| `Term` (base) | (no separate type) | Consider adding `Term` interface |
| `Atom` | `ast.Atom` | OK |
| `App` | `ast.App` | OK |
| `Old` | `ast.Old` | OK |
| `Symbol` | `ast.Symbol` | OK |
| `Variable` | `ast.Variable` | OK |
| `KeyArg` | `ast.KeyArg` | OK |
| `Definition` | `ast.Definition` | OK |
| `Some` | `ast.Some` | OK |
| `And` | `ast.And` | OK |
| `Or` | `ast.Or` | OK |
| `Not` | `ast.Not` | OK |
| `Implies` | `ast.Implies` | OK |
| `Iff` | `ast.Iff` | OK |
| `Ite` | `ast.Ite` | OK |
| `ForAll` | `ast.Forall` | **RENAME to `ast.ForAll`** (inconsistent with `logic.ForAll`) |
| `Exists` | `ast.Exists` | OK |
| `Lambda` | (not in ast/) | Consider adding if Python has it in AST layer |
| `NamedBinder` | `ast.NamedBinder` | OK |
| `Cond` | (not in ast/) | Consider adding |
| `WhenOperator` | `ast.WhenOperator` | OK |
| `Globally` | `ast.Globally` | OK |
| `Eventually` | `ast.Eventually` | OK |

### AST Fields

| Python field | Go field | RENAME NEEDED? |
|---|---|---|
| `self.rep` | `.Rep` | OK |
| `self.args` | `.Terms` / `.Args()` | **Should be `.Args` consistently** — Python uses `self.args` everywhere |
| `self.sort` | `.ASort` | OK (AST-level sort annotation) |

---

## 2. Logic Layer: logic.py -> logic/, ivy_logic.py -> ivylogic/

### Types

| Python | Go | RENAME NEEDED? |
|---|---|---|
| `Variable` | `logic.Var` | **RENAME to `logic.Variable`** to match Python |
| `Symbol` | `logic.Const` | **RENAME to `logic.Symbol`** to match Python |
| `Apply` | `logic.Apply` | OK |
| `Eq` | `logic.Eq` | OK |
| `Not` | `logic.Not` | OK |
| `And` | `logic.And` | OK |
| `Or` | `logic.Or` | OK |
| `Implies` | `logic.Implies` | OK |
| `Iff` | `logic.Iff` | OK |
| `ForAll` | `logic.ForAll` | OK |
| `Exists` | `logic.Exists` | OK |
| `Lambda` | `logic.Lambda` | OK |
| `NamedBinder` | `logic.NamedBinder` | OK |
| `Ite` | `logic.Ite` | OK |
| `Cond` | `logic.Cond` | OK |
| `WhenOperator` | `logic.WhenOperator` | OK |
| `Globally` | `logic.Globally` | OK |
| `Eventually` | `logic.Eventually` | OK |
| `Definition` | `logic.Definition` | OK |
| `DefinitionSchema` | `logic.DefinitionSchema` | OK |
| `Some` | `ivylogic.Some` | OK |
| `Literal` | `ivylogic.Literal` | OK |
| `Sig` | `ivylogic.Sig` | OK |

### Logic Fields — THE WORST DEVIATION

| Python field | Go field | On which types | RENAME NEEDED? |
|---|---|---|---|
| `self.name` | `.Name` | Var, Const, NamedBinder | OK |
| `self.sort` | `.VSort` | Var | **RENAME to `.Sort`** |
| `self.sort` | `.CSort` | Const | **RENAME to `.Sort`** |
| `self.sort` | `.ISort` | Ite | **RENAME to `.Sort`** |
| `self.sort` | `.WSort` | WhenOperator | **RENAME to `.Sort`** |
| `self.sort` | `.CSort` | Cond | **RENAME to `.Sort`** |
| `self.func` | `.Func` | Apply | OK |
| `self.terms` | `.Terms` | Apply | OK |
| `self.variables` | `.Variables` | ForAll, Exists, Lambda, NamedBinder | OK |
| `self.body` | `.Body` | ForAll, Exists, Lambda, NamedBinder, Not, Globally, Eventually | OK |
| `self.environ` | `.Environ` | Globally, Eventually, NamedBinder | OK |
| `self.t1` / `self.t2` | `.T1` / `.T2` | Eq, Implies, Iff, Cond, WhenOperator | OK |
| `self.cond` | `.Cond` | Ite | OK |
| `self.t_then` | `.Then` | Ite | OK (minor) |
| `self.t_else` | `.Else` | Ite | OK (minor) |

**The `.Sort` rename is the single most impactful fix.** Python uses `x.sort` on every
logic node. Go uses 5 different prefixed names (`VSort`, `CSort`, `ISort`, `WSort`, `CSort`).
This makes grep impossible and violates the "same name" rule.

The reason for the prefixed names was to avoid Go's compile-time ambiguity with the
`Sort` interface. Fix: use a different field name that doesn't collide, like `NodeSort_`
or use a method `Sort() Sort` on each type (which already exists as `NodeSort()`).

Actually the simplest conformance fix: **rename `NodeSort()` to `Sort()`** on all types,
and rename the fields to something that doesn't collide with the method name. Or:
keep the fields as `VSort`/`CSort` but add a `Sort()` method that Python code would use.
The `NodeSort()` method already exists and does this — but it should be named `Sort()`.

However, Go's `Sort` interface name collides with a `Sort()` method. This is a
fundamental Go naming issue. Best approach: rename the interface from `Sort` to
`IvySort` and use `Sort()` as the method name on nodes.

---

## 3. Logic Utils: ivy_logic_utils.py -> clauseops/, logicutil/

| Python function | Go function | Package | RENAME NEEDED? |
|---|---|---|---|
| `free_variables` | `FreeVariables` | logicutil | OK |
| `used_variables` | `UsedVariables` | logicutil | OK |
| `bound_variables` | `BoundVariables` | logicutil | OK |
| `used_constants` | `UsedConstants` | logicutil | OK |
| `substitute` | `Substitute` | logicutil | OK |
| `substitute_ast` | (missing) | — | **PORT IT** |
| `substitute_constants_ast` | `SubstituteConstantsAST` | clauseops | OK |
| `rename_ast` | `RenameAST` | clauseops | OK |
| `symbols_ast` | `SymbolsAST` | clauseops | OK |
| `used_symbols_ast` | `UsedSymbolsAST` | clauseops | OK |
| `variables_ast` | `VariablesAST` | clauseops | OK |
| `sorts_ast` | `SortsAst` | logicutil | **RENAME to `SortsAST`** (inconsistent caps) |
| `resort_sort` | `ResortSort` | logicutil | OK |
| `resort_ast` | `ResortAst` | logicutil | **RENAME to `ResortAST`** |
| `close_epr` | `CloseEPR` | logicutil | OK |
| `normalize_quantifiers` | `NormalizeQuantifiers` | logicutil | OK |
| `and_clauses` | `AndClauses` / `AndClausesTyped` | clauseops | OK |
| `or_clauses` | `OrClauses` / `OrClausesTyped` | clauseops | OK |
| `ite_clauses` | `IteClauses` | clauseops | OK |
| `negate_clauses` | `NegateClauses` | clauseops | OK |
| `equal_mod_alpha` | `EqualModAlpha` | logicutil | OK |
| `is_tautology_equality` | `IsTautologyEquality` | logicutil | OK |

---

## 4. Transrel: ivy_transrel.py -> transrel/

| Python | Go | RENAME NEEDED? |
|---|---|---|
| `new(sym)` | `New(name)` | OK |
| `old(sym)` | `Old(name)` | OK |
| `is_new(sym)` | `IsNew(name)` | OK |
| `new_of(sym)` | `NewOf(name)` | OK |
| `is_skolem(name)` | `IsSkolem(name)` | OK |
| `Update` (triple) | `Update` struct | OK |
| `frame` | `Frame` | OK |
| `frame_def` | `FrameDef` | OK |
| `diff_frame` | `DiffFrame` / `DiffFrameConst` | Minor — Go has Const variant |
| `updated_join` | `UpdatedJoin` / `UpdatedJoinConst` | Minor |
| `compose_updates` | `ComposeUpdates` | OK |
| `forward_image_map` | `ForwardImageMap` | OK |
| `forward_image` | `ForwardImage` | OK |
| `reverse_image` | `ReverseImage` | OK |

---

## 5. Actions: ivy_actions.py -> actions/

| Python | Go | RENAME NEEDED? |
|---|---|---|
| `Action` (base) | `actions.Action` (interface) | OK |
| `Sequence` | `actions.Sequence` | OK |
| `ChoiceAction` | `actions.ChoiceAction` | OK |
| `IfAction` | `actions.IfAction` | OK |
| `WhileAction` | `actions.WhileAction` | OK |
| `AssignAction` | `actions.AssignAction` | OK |
| `AssertAction` | `actions.AssertAction` | OK |
| `AssumeAction` | `actions.AssumeAction` | OK |
| `CallAction` | `actions.CallAction` | OK |
| `HavocAction` | `actions.HavocAction` | OK |
| `LocalAction` | `actions.LocalAction` | OK |
| `mk_assign_clauses` | `mkAssignClauses` (unexported) | **EXPORT as `MkAssignClauses`** |
| `int_update` | `IntUpdate` | OK |

---

## 6. Solver: ivy_solver.py -> solver/

| Python | Go | RENAME NEEDED? |
|---|---|---|
| `solver_name(sym)` | `Solver.SolverName(sym)` | OK (method on Solver) |
| `lookup_native(sym)` | `Solver.LookupNative(sym, isRelation)` | OK |
| `formula_to_z3(fmla)` | `Solver.FormulaToZ3(fmla)` | OK |
| `clauses_imply(c1, c2)` | `Solver.ClausesImply(c1, c2)` | OK |
| `clauses_sat(c)` | `Solver.ClausesSat(c)` | OK |

---

## 7. Compiler: ivy_compiler.py -> compiler/

| Python | Go | RENAME NEEDED? |
|---|---|---|
| `IvyDomainSetup` | `DeclInterp` | **RENAME to `DomainSetup`** |
| `IvyConjectureSetup` | `ConjSetup` | **RENAME to `ConjectureSetup`** |
| `IvyARGSetup` | `ARGSetup` | OK |
| `collect_actions` | `CollectActions` | OK |
| `compile_action_def` | `CompileAction` | **RENAME to `CompileActionDef`** |
| `ivy_compile` | `IvyCompile` | OK |

---

## 8. Module: ivy_module.py -> module/

All field names are straightforward `snake_case` -> `CamelCase`. No problematic renames.

---

## 9. Art: ivy_art.py -> art/

| Python | Go | RENAME NEEDED? |
|---|---|---|
| `AnalysisGraph` | `AnalysisGraph` | OK |
| `State` | `State` | OK |

---

## Priority Renames (for future pass)

### P0 — Critical for grep cross-referencing

1. **`logic.Const` -> `logic.Symbol`**: Matches Python's `ivy_logic.Symbol`. This is the
   most confusing rename — every reference to Python's "Symbol" maps to Go's "Const".
   Affects ~500+ references across the codebase.

2. **`logic.Var` -> `logic.Variable`**: Matches Python's `logic.Variable`. Less impactful
   since `Var` is a common abbreviation, but still a deviation.

3. **Sort field names**: `VSort`/`CSort`/`ISort`/`WSort` should ideally all be accessible
   via a single `.Sort` accessor matching Python's `self.sort`. The `NodeSort()` method
   exists but has a non-matching name. Rename to match Python.

### P1 — Important for conformance

4. **`ast.Forall` -> `ast.ForAll`**: Inconsistent with `logic.ForAll`.

5. **`DeclInterp` -> `DomainSetup`**: Match Python's `IvyDomainSetup`.

6. **`ConjSetup` -> `ConjectureSetup`**: Match Python's `IvyConjectureSetup`.

7. **`SortsAst` -> `SortsAST`**: Inconsistent acronym capitalization.

8. **`ResortAst` -> `ResortAST`**: Same.

### P2 — Nice to have

9. **`CompileAction` -> `CompileActionDef`**: Match Python's `compile_action_def`.

10. **`mkAssignClauses` -> `MkAssignClauses`**: Export so other packages can call it.
