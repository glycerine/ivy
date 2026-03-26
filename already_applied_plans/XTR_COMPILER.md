# Plan: Systematic xtracer instrumentation of compiler and actions for golden test conformance

**Created:** 2026-03-26T16:30

## Context

The `lalr_full` parser switchover is complete and showing good conformance. The golden test now diverges at the **compiler** level (line 78259). To systematically advance the golden test, we need matching xtracer traces in every function of `ivy_compiler.py` / `ivy_actions.py` and their Go counterparts. This plan adds ~208 matching trace points (104 per side) across both languages.

**Convention:**
- First line of trace must match exactly between Go and Python (golden test compares these)
- Debug details go after `\n` (not compared)
- Trace strings use Python function names (snake_case) to match source-of-truth
- `ENTER` at function entry, `EXIT` at function exit (EXIT only for infrequent orchestration functions)
- Branch traces: `compiler.ClassName.method branch=X`

## Part 1: Compiler — DomainSetup method ENTER traces (27 new trace points)

All DomainSetup handler methods need an ENTER trace. These are dispatched by `IvyDeclInterp.__call__` (already traced).

**Python file:** `~/pyivy/ivy/ivy/ivy_compiler.py` (inside class `IvyDomainSetup`)
**Go file:** `compiler/decl.go` (methods on `*DomainSetup`)

| # | Trace string | Py method:line | Go method:line |
|---|---|---|---|
| 1 | `compiler.DomainSetup.alias ENTER` | alias:1029 | Alias:1171 |
| 2 | `compiler.DomainSetup.object ENTER` | object:1031 | Object:876 |
| 3 | `compiler.DomainSetup.axiom ENTER` | axiom:1033 | Axiom:524 |
| 4 | `compiler.DomainSetup.property ENTER` | property:1039 | Property:555 |
| 5 | `compiler.DomainSetup.conjecture ENTER` | conjecture:1043 | Conjecture:581 |
| 6 | `compiler.DomainSetup.named ENTER` | named:1045 | Named:1359 |
| 7 | `compiler.DomainSetup.schema ENTER` | schema:1075 | Schema:1235 |
| 8 | `compiler.DomainSetup.theorem ENTER` | theorem:1084 | Theorem:1408 |
| 9 | `compiler.DomainSetup.instantiate ENTER` | instantiate:1092 | Instantiate:1281 |
| 10 | `compiler.DomainSetup.relation ENTER` | relation:1100 | Relation:587 |
| 11 | `compiler.DomainSetup.individual ENTER` | individual:1105 | Individual:618 |
| 12 | `compiler.DomainSetup.parameter ENTER` | parameter:1110 | Parameter:1450 |
| 13 | `compiler.DomainSetup.destructor ENTER` | destructor:1120 | Destructor:1477 |
| 14 | `compiler.DomainSetup.constructor ENTER` | constructor:1127 | Constructor:1495 |
| 15 | `compiler.DomainSetup.derived ENTER` | derived:1157 | Derived:632 |
| 16 | `compiler.DomainSetup.definition ENTER` | definition:1173 | DefinitionDecl:709 |
| 17 | `compiler.DomainSetup.proof ENTER` | proof:1183 | Proof:1311 |
| 18 | `compiler.DomainSetup.progress ENTER` | progress:1199 | Progress:1204 |
| 19 | `compiler.DomainSetup.rely ENTER` | rely:1205 | Rely:1554 |
| 20 | `compiler.DomainSetup.mixord ENTER` | mixord:1208 | Mixord:1570 |
| 21 | `compiler.DomainSetup.concept ENTER` | concept:1210 | Concept:1510 |
| 22 | `compiler.DomainSetup.update ENTER` | update:1216 | Update:1578 |
| 23 | `compiler.DomainSetup.variant ENTER` | variant:1259 | Variant:901 |
| 24 | `compiler.DomainSetup.implementtype ENTER` | implementtype:1265 | Implementtype:1614 |
| 25 | `compiler.DomainSetup.scenario ENTER` | scenario:1347 | Scenario:1593 |
| 26 | `compiler.DomainSetup.mixin ENTER` | mixin:1354 | Mixin:1134 |
| 27 | `compiler.DomainSetup.action ENTER` | action:1359 | Action:787 |

**Python pattern:** `if __debug__: xtracer.trace("compiler.DomainSetup.alias ENTER")`
**Go pattern:** `xtracer.Trace("compiler.DomainSetup.alias ENTER")`

## Part 2: Compiler — ConjSetup + ARGSetup method ENTER traces (13 new trace points)

### ConjSetup (4 new)

**Go file:** `compiler/ivy_compile.go` (inside `ConjSetup.ProcessDecls` switch)

| # | Trace string | Py method:line | Go location |
|---|---|---|---|
| 28 | `compiler.ConjSetup.property ENTER` | property:1397 | ivy_compile.go switch |
| 29 | `compiler.ConjSetup.definition ENTER` | definition:1399 | ivy_compile.go switch |
| 30 | `compiler.ConjSetup.theorem ENTER` | theorem:1401 | ivy_compile.go switch |
| 31 | `compiler.ConjSetup.proof ENTER` | proof:1403 | ivy_compile.go switch |

### ARGSetup (9 new, 4 already traced)

**Go file:** `compiler/ivy_compile.go` (inside `ARGSetup.ProcessDecls` switch)

| # | Trace string | Py method:line | Status |
|---|---|---|---|
| 32 | `compiler.ARGSetup.state ENTER` | state:1434 | NEEDS TRACE |
| 33 | `compiler.ARGSetup.mixin ENTER` | mixin:1436 | NEEDS TRACE |
| 34 | `compiler.ARGSetup._assert ENTER` | _assert:1440 | NEEDS TRACE |
| 35 | `compiler.ARGSetup.import_ ENTER` | import_:1456 | NEEDS TRACE |
| 36 | `compiler.ARGSetup.private ENTER` | private:1459 | NEEDS TRACE |
| 37 | `compiler.ARGSetup.delegate ENTER` | delegate:1461 | NEEDS TRACE |
| 38 | `compiler.ARGSetup.native ENTER` | native:1463 | NEEDS TRACE |
| 39 | `compiler.ARGSetup.attribute ENTER` | attribute:1465 | NEEDS TRACE |
| 40 | `compiler.ARGSetup.scenario ENTER` | scenario:1480 | NEEDS TRACE |

## Part 3: Compiler — Orchestration function ENTER/EXIT traces (14 new pairs = 28 trace points)

These are called once per `ivy_compile` invocation and are structurally critical.

**Go files:** `compiler/ivy_compile.go`, `compiler/phase6.go`
**Python file:** `ivy_compiler.py`

| # | ENTER trace | EXIT trace | Py:line | Go file:line |
|---|---|---|---|---|
| 41-42 | `compiler.CheckProperties ENTER` | `compiler.CheckProperties EXIT` | 2005 | phase6.go:2054 |
| 43-44 | `compiler.CreateSortOrder ENTER` | `compiler.CreateSortOrder EXIT` | 1652 | ivy_compile.go:1069 |
| 45-46 | `compiler.AttachProofs ENTER` | `compiler.AttachProofs EXIT` | 1692 | ivy_compile.go:1285 |
| 47-48 | `compiler.FixConstructors ENTER` | `compiler.FixConstructors EXIT` | 1937 | ivy_compile.go:1002 |
| 49-50 | `compiler.CreateConstructorSchemata ENTER` | `compiler.CreateConstructorSchemata EXIT` | 1892 | ivy_compile.go:1117 |
| 51-52 | `compiler.CreateConjActions ENTER` | `compiler.CreateConjActions EXIT` | 2122 | ivy_compile.go:1600 |
| 53-54 | `compiler.HandleTemporals ENTER` | `compiler.HandleTemporals EXIT` | 2182 | ivy_compile.go:1672 |
| 55-56 | `compiler.InferParameters ENTER` | `compiler.InferParameters EXIT` | 1596 | phase6.go:1406 |
| 57-58 | `compiler.CompileTheories ENTER` | `compiler.CompileTheories EXIT` | 2109 | phase6.go:2314 |
| 59-60 | `compiler.CompileTheory ENTER sortname=%s theoryname=%s` | `compiler.CompileTheory EXIT` | 2104 | phase6.go:2292 |
| 61-62 | `compiler.IvyCompileTheory ENTER` | `compiler.IvyCompileTheory EXIT` | 2093 | phase6.go:2282 |
| 63-64 | `compiler.IvyCompileTheoryFromString ENTER sortname=%s` | `compiler.IvyCompileTheoryFromString EXIT` | 2096 | phase6.go:2583 |
| 65-66 | `compiler.ReorderProps ENTER` | `compiler.ReorderProps EXIT` | 1866 | phase6.go:1687 |
| 67-68 | `compiler.BalancedChoice ENTER` | (no EXIT — recursive) | 1551 | phase6.go:1769 |

## Part 4: Compiler — AST compilation function ENTER traces (16 new trace points)

These fire per AST node during compilation. ENTER only (EXIT too noisy).

**Go files:** `compiler/compiler.go`, `compiler/phase6.go`, `compiler/helpers.go`
**Python file:** `ivy_compiler.py` (monkey-patched methods)

| # | Trace string | Py function:line | Go method | Go file |
|---|---|---|---|---|
| 69 | `compiler.thing ENTER` | thing:52 | Thing | phase6.go:51 |
| 70 | `compiler.other_thing ENTER` | other_thing:61 | OtherThing | phase6.go:59 |
| 71 | `compiler.compile_root_args ENTER` | compile_root_args:58 | CompileRootArgs | phase6.go:166 |
| 72 | `compiler.compile_args ENTER` | compile_args:86 | compileArgs | compiler.go:316 |
| 73 | `compiler.compile_app ENTER old=%s` | compile_app:301 | CompileApp | compiler.go:419 |
| 74 | `compiler.compile_method_call ENTER` | compile_method_call:326 | compileMethodCall | compiler.go:604 |
| 75 | `compiler.compile_variable ENTER` | compile_variable:380 | CompileVariable | compiler.go:552 |
| 76 | `compiler.compile_quantifier ENTER` | compile_quantifier:398 | CompileQuantifier | compiler.go:826 |
| 77 | `compiler.compile_field_reference ENTER name=%s` | compile_field_reference:205 | CompileFieldReference | helpers.go:80 |
| 78 | `compiler.compile_field_reference_rec ENTER name=%s` | compile_field_reference_rec:164 | compileFieldReferenceRec | helpers.go:100 |
| 79 | `compiler.compile_inline_call ENTER` | compile_inline_call:234 | CompileInlineCall | helpers.go:228 |
| 80 | `compiler.sortify_with_inference ENTER` | sortify_with_inference:454 | SortifyWithInference | compiler.go:911 |
| 81 | `compiler.sortify ENTER` | sortify:450 | Sortify | phase6.go:467 |
| 82 | `compiler.sort_infer_covariant ENTER` | sort_infer_covariant:216 | SortInferCovariant | phase6.go:203 |
| 83 | `compiler.sort_infer_contravariant ENTER` | sort_infer_contravariant:225 | SortInferContravariant | phase6.go:239 |
| 84 | `compiler.compile_defn ENTER` | compile_defn:832 | CompileDefn | compiler.go:1038 |

## Part 5: Compiler — Action compilation ENTER traces (11 new trace points)

**Go file:** `compiler/action.go`, `compiler/phase6.go`
**Python file:** `ivy_compiler.py`

| # | Trace string | Py function:line | Go method | Go file |
|---|---|---|---|---|
| 85 | `compiler.compile_local ENTER` | compile_local:473 | CompileLocal | action.go:763 |
| 86 | `compiler.compile_assign ENTER` | compile_assign:530 | CompileAssign | action.go:444 |
| 87 | `compiler.compile_call ENTER` | compile_call:576 | CompileCall | action.go:610 |
| 88 | `compiler.compile_if_action ENTER` | compile_if_action:613 | CompileIf | action.go:913 |
| 89 | `compiler.compile_while_action ENTER` | compile_while_action:638 | CompileWhile | action.go:1045 |
| 90 | `compiler.compile_assert_action ENTER` | compile_assert_action:656 | CompileAssertFormula | action.go:1159 |
| 91 | `compiler.compile_crash_action ENTER` | compile_crash_action:675 | CompileCrashAction | phase6.go:497 |
| 92 | `compiler.compile_thunk_action ENTER` | compile_thunk_action:685 | CompileThunkAction | phase6.go:535 |
| 93 | `compiler.compile_debug_action ENTER` | compile_debug_action:741 | CompileDebugAction | phase6.go:728 |
| 94 | `compiler.compile_native_action ENTER` | compile_native_action:779 | CompileNativeAction | phase6.go:915 |
| 95 | `compiler.compile_action_def ENTER` | compile_action_def:798 | (in DomainSetup.Action) | decl.go:787 |

## Part 6: Actions — Runtime dispatch + action_update/int_update traces (72 new trace points)

**Python file:** `~/pyivy/ivy/ivy/ivy_actions.py`
**Go file:** `actions/update.go`, `actions/extra_actions.go`, `actions/action.go`

### 6a. Central dispatch (4 traces)

| # | Trace string | Py location | Go location |
|---|---|---|---|
| 96 | `actions.IntUpdate ENTER type=%s` | Action.int_update:201 | update.go IntUpdate() |
| 97 | `actions.GetUpdate ENTER type=%s` | Action.update:218 | update.go GetUpdate() |
| 98 | `actions.IntUpdate EXIT type=%s` | (at return) | (at return) |
| 99 | `actions.GetUpdate EXIT type=%s` | (at return) | (at return) |

### 6b. Leaf action traces (28 traces)

**AssumeAction** (3):
`actions.AssumeAction.action_update ENTER`, `branch=unprovable_skip`, `EXIT`

**AssertAction** (4):
`actions.AssertAction.action_update ENTER`, `branch=unprovable_mismatch`, `branch=checked_assert_skip`, `EXIT`

**AssignAction** (10 — most complex):
`actions.AssignAction.action_update ENTER`, `branch=hierarchy`, `branch=xtra_negative`, `branch=xtra_positive`, `branch=destructor`, `branch=variant`, `branch=standard`, `branch=individual_to_bool`, `branch=multiply_assigned`, `EXIT`

**HavocAction** (4):
`actions.HavocAction.action_update ENTER`, `branch=is_atom`, `branch=is_individual`, `EXIT`

**SetAction** (2):
`actions.SetAction.action_update ENTER`, `EXIT`

**CrashAction** (2):
`actions.CrashAction.action_update ENTER`, `EXIT`

### 6c. Compound action traces (37 traces)

**Sequence** (3): `ENTER`, `child=%d`, `EXIT`
**ChoiceAction** (4): `ENTER`, `branch=determinize`, `branch=join`, `EXIT`
**EnvAction** (4): `ENTER`, `branch=determinize`, `branch=join`, `EXIT`
**IfAction** (6): `ENTER`, `branch=simple_boolean`, `branch=some_condition`, `subactions ENTER`, `branch=some_min_max`, `EXIT`
**WhileAction** (5): `ENTER`, `branch=unroll`, `branch=expand`, `expand ENTER`, `EXIT`
**LocalAction** (3): `__init__ uniqueID=%d` (Py already has), `int_update ENTER`, `EXIT`
**LetAction** (2): `ENTER`, `EXIT`
**CallAction** (5): `ENTER`, `branch=action`, `branch=state`, `apply_actuals ENTER`, `EXIT`
**BindOldsAction** (2): `ENTER`, `EXIT`
**InstantiateAction** (3): `ENTER`, `branch=macro`, `branch=schema`

### 6d. Module-level function traces (3 traces)

| # | Trace string | Py function:line | Go function |
|---|---|---|---|
| 100 | `actions.apply_mixin ENTER` | apply_mixin:1341 | ApplyMixin |
| 101 | `actions.env_action ENTER` | env_action:1670 | BuildEnvAction |
| 102 | `actions.match_annotation ENTER` | match_annotation:1551 | MatchAnnotation |

## Summary

| Category | New trace points (per side) |
|---|---|
| Part 1: DomainSetup methods | 27 |
| Part 2: ConjSetup + ARGSetup | 13 |
| Part 3: Orchestration ENTER/EXIT | 28 |
| Part 4: AST compilation | 16 |
| Part 5: Action compilation | 11 |
| Part 6: Runtime actions | 72 |
| **Total** | **167 per side (334 total lines)** |

## Files to Modify

| File | Changes |
|---|---|
| `~/pyivy/ivy/ivy/ivy_compiler.py` | ~95 new xtracer.trace() calls |
| `~/pyivy/ivy/ivy/ivy_actions.py` | ~72 new xtracer.trace() calls |
| `compiler/decl.go` | 27 DomainSetup ENTER traces |
| `compiler/ivy_compile.go` | 13 ConjSetup/ARGSetup + ~10 orchestration traces |
| `compiler/phase6.go` | ~20 orchestration + compilation function traces |
| `compiler/compiler.go` | ~8 compilation function traces |
| `compiler/helpers.go` | 3 compilation function traces |
| `compiler/action.go` | 11 action compilation traces |
| `actions/update.go` | ~50 action runtime traces + xtracer import |
| `actions/extra_actions.go` | ~5 InstantiateAction + BuildEnvAction traces |
| `actions/action.go` | ~3 constructor traces (LocalAction, CallAction, ChoiceAction) |

## Implementation Order

1. **Parts 1-2 first** (DomainSetup/ConjSetup/ARGSetup) — these are the current golden test divergence area
2. **Part 3 next** (orchestration) — structural backbone
3. **Parts 4-5** (compilation functions) — per-node traces
4. **Part 6 last** (runtime actions) — these fire later in the pipeline

## Verification

```bash
# Build
go build ./...

# Tests green
XTRACE_OFF=1 go test ./... -count=1 -timeout 120s

# Golden test advancement
cd ~/goivy && make golden
```

Golden test should advance significantly past line 78259 as trace coverage increases.
