# Fix IfAction Not-wrapper loss in substitute_constants_action divergence

**Created:** 2026-05-02 ~UTC

## Context

The 2hr golden test (`Test2hrOrdLive`) diverges at trace line 33778931 (~line 228528 of `log.golden.2hr`). During `substitute_constants_action` traversal of a callee action's body inside `CallAction.IntUpdate` → `applyActuals`, an inner IfAction's condition differs:

- **Go:** `type=Apply nargs=2`
- **Python:** `type=Not nargs=1`

The `Not` wrapper is missing from Go's `actions.IfAction.Cond` field.

### Tree structure at divergence

```
33778924: IfAction nargs=2          ← outer IfAction
33778925:   SomeMin nargs=3         ← outer's condition (AstCond)
33778926:     And nargs=2
33778927:       Not nargs=1         ← both agree here
33778928:         Apply nargs=2
33778929:       Apply nargs=2
33778930:   IfAction nargs=2        ← inner IfAction (ThenBody of outer)
33778931:     DIVERGE: Go=Apply(2) vs Python=Not(1) ← inner's CONDITION
```

## Exhaustive code review findings

Every code path that constructs, clones, or substitutes IfActions **correctly preserves Not wrappers**:

| Code path | File | Verdict |
|-----------|------|---------|
| `compileNot` | `compiler/compiler.go:629` | Correct: `&lg.Not{Body: args[0]}` |
| `lg.Not.Clone` | `logic/ast_compat.go:76` | Correct: `&Not{Body: args[0].(Expr)}` |
| `IfAction.Clone` default | `actions/action_expr.go:354` | Correct: `cond = args[0].(lg.Expr)` |
| `SubstituteConstantsAST` | `module/astutil.go:108` | Correct: recurse + clone |
| `ExprContext.Extract` | `compiler/compiler.go:76` | Correct: returns single item directly |
| `SortInfer` | `compiler/compiler.go:1188` | Sort inference only, no structural change |
| `someCondFromAST` | `actions/action.go:473` | Only handles Some/SomeMin/SomeMax |
| No post-compilation code strips Not | `actions/`, `isolate/`, `module/` | Confirmed: no Not stripping found |

### Conclusion

Since **all** runtime code paths preserve Not, the inner IfAction's `Cond` was set to `Apply(...)` at **construction time** — either during compilation or action composition. The Not was never present on this specific IfAction.

## Most likely root causes (ordered by probability)

1. **Parser bug:** The Go AST parser produced `ast.IfAction{Cond: <not-a-Not>}` for this specific condition, while Python's parser produced `IfAction(Not(Apply(...)), ...)`.

2. **Action composition (mixin/after/before):** The callee action's body was assembled from multiple definitions. During composition, a new IfAction was created with `Cond = Apply(...)` instead of `Cond = Not(Apply(...))`.

3. **Double-compilation:** The condition node was already a compiled `lg.Apply` (not `ast.Not`) when it reached `CompileIf`, so `CompileNode` routed it through `OtherThing` → `compileGeneric` instead of `compileNot`, producing a re-wrapped Apply without the Not.

## Plan

### Phase 1: Add targeted diagnostic assertions

Add **panic-on-detection** assertions that fire at the exact moment the Not wrapper is found missing, producing a stack trace. These use `panic()` so they don't affect XTRACE golden comparison.

#### 1a. In `CompileIf` — catch compilation-time loss

**File:** `compiler/action.go:1255` (after `SortifyWithInference` returns)

```go
// DIAGNOSTIC: detect Not lost during compilation
if _, isAstNot := condNode.(*ast.Not); isAstNot {
    if _, isLgNot := cond.(*lg.Not); !isLgNot {
        panic(fmt.Sprintf("CompileIf: ast.Not compiled to non-lg.Not: condNode=%T, cond=%T(%v)", condNode, cond, cond))
    }
}
```

If this fires → the compiler's `SortifyWithInference` lost the Not.

#### 1b. In `IfAction.Clone` — catch clone-time loss

**File:** `actions/action_expr.go:356` (in the `default` case)

```go
// DIAGNOSTIC: detect Not-stripping in Clone
if _, origIsNot := a.Cond.(*lg.Not); origIsNot {
    if _, newIsNot := args[0].(*lg.Not); !newIsNot {
        panic(fmt.Sprintf("IfAction.Clone: Not lost! orig=%T, args[0]=%T(%v)", a.Cond, args[0], args[0]))
    }
}
```

If this fires → `SubstituteConstantsAST` returned a non-Not for args[0].

#### 1c. In `CallAction.IntUpdate` — identify the callee

**File:** `actions/update.go:1883` (after calleeName is computed)

```go
fmt.Fprintf(os.Stderr, "DIAG CallAction.IntUpdate callee=%s\n", calleeName)
```

This identifies which action definition is involved (stderr doesn't affect golden comparison).

#### 1d. In `SubstituteConstantsAST` — catch Not.Clone loss

**File:** `module/astutil.go:131` (after `node.Clone(newArgs)`)

```go
// DIAGNOSTIC: detect Not.Clone producing non-Not
if _, isNot := node.(*lg.Not); isNot {
    if _, resultIsNot := node.Clone(newArgs).(*lg.Not); !resultIsNot {
        panic(fmt.Sprintf("SubstituteConstantsAST: Not.Clone lost Not: %T", node.Clone(newArgs)))
    }
}
```

### Phase 2: Run test once (4 hours)

```bash
cd ~/ivy/goivy && make golden-2hr 2>diag.stderr
```

Analyze results:
- **If 1a fires:** Fix is in `SortifyWithInference` or `CompileNode` — the sort inference or re-compilation path loses Not.
- **If 1b fires:** Fix is in the clone/substitution path — `SubstituteConstantsAST` returned a body-of-Not instead of Not.
- **If 1d fires:** Fix is in `lg.Not.Clone()`.
- **If NONE fires and test still diverges:** The AST condition was NEVER `ast.Not`. This is a **parser bug** or **action composition bug**. Use the callee name from 1c to trace backward to the Ivy source and the compilation path.

### Phase 3: Fix the root cause

Based on Phase 2 results:

| Trigger | Root cause | Fix location |
|---------|-----------|--------------|
| 1a fires | `SortifyWithInference` or `SortInfer` strips Not | `compiler/compiler.go` or `compiler/phase6.go` |
| 1b fires | Clone receives unwrapped arg | Trace caller via stack trace |
| 1d fires | `lg.Not.Clone` implementation bug | `logic/ast_compat.go:76` |
| None fire | Parser or composition doesn't produce Not | Parser grammar rules or mixin composition in `isolate/` |

For the "none fire" case (most likely): use callee name from stderr to find the Ivy source, then check the Go parser's AST output for that specific if-condition. Compare with Python parser's output.

### Phase 4: Verify

1. Remove all diagnostic assertions
2. Run `cd ~/ivy/goivy && make test` — verify no regressions
3. Run `cd ~/ivy/goivy && make golden-2hr` — verify test advances past line 33778931

## Critical files

- `actions/action_expr.go:334-374` — IfAction.Args(), Clone()
- `actions/update.go:1882-2058` — CallAction.IntUpdate, applyActuals
- `module/astutil.go:108-132` — SubstituteConstantsAST
- `compiler/action.go:1228-1287` — CompileIf (non-Some path)
- `compiler/action.go:1289-1373` — compileIfSome
- `compiler/compiler.go:629-636` — compileNot
- `logic/ast_compat.go:75-78` — lg.Not.Args(), Clone()
- `actions/action.go:371-385` — IfAction struct, NewIfAction
