# Fix `proof.GoalVocab` divergence on `rec[index]` — and every sibling `CompiledNode`-unwrap bug

Created: 2026-04-21 08:42 UTC (revised 2026-04-21 09:05 UTC to include all known sibling bugs)

## Context

### Immediate divergence

`make golden-2hr` halts at line 191504 of `log.golden.2hr`:

```
go : XTRACE: proof.GoalVocab EXIT nsorts=0 nsymbols=0 nvariables=1
py : XTRACE: proof.GoalVocab EXIT nsorts=1 nsymbols=3 nvariables=5
```

The ENTER lines match (`label=rec[index]`); only the EXIT counts differ.

### Root cause

Python's `compile_schema_prem` (`ivy_compiler.py:1026-1044`) returns **bare** logic objects for schema premises:
- `TypeDef` → bare `ivy_logic.UninterpretedSort`
- `ConstantDecl` → `ConstantDecl` whose `args[0]` is a bare `ivy_logic.Const`
- `LabeledFormula` → compiled `LabeledFormula` (not wrapped)

Go's `CompileSchemaPremWithSig` (`compiler/phase6.go:1186-1248`) **wraps** the bare logic objects in a Go-only adapter `*ast.CompiledNode` (lines 1200, 1218, 1225). `CompiledNode` is load-bearing elsewhere (many non-schema sites), so we will NOT change the wrap site; we fix every read site to `unwrapCompiledNode(x)` first.

The pattern to mirror is `GoalIsDefn` (`proof/goal.go:306-324`), which correctly unwraps both `p` and `args[0]`. Helper: `unwrapCompiledNode` at `proof/goal.go:331`.

### Why we fix every known site, not just the one golden-2hr surfaced

Latent type-assertion failures on `CompiledNode`-wrapped premises will mask further divergences and waste cycles chasing downstream symptoms. Per project policy, every bug found is fixed now.

## Bug sites (all must be fixed in this change)

### In `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/proof/goal.go`

**(B1)** `GoalDefns` line 349 — `args[0].(*lg.Const)` without unwrap. Fails on `CompiledNode`-wrapped `*lg.Const`.

**(B2)** `GoalDefns` line 355 — `p.(*lg.UninterpretedSort)` without unwrap. Fails on `CompiledNode`-wrapped `*lg.UninterpretedSort`.

**(B3)** `GoalVocab` line 389 — `p.(lg.Sort)` without unwrap. The immediate golden-2hr divergence.

**(B4)** `GoalVocab` line 397 — `args[0].(lg.Expr)` then `.(*lg.Const)` without unwrap.

**(B5)** `GoalVocab` lines 407-417 — the `lf.Formula.(*ast.SchemaBody)` special-case plus `ConcAsExpr(GoalConc(lf))` fallback. Python does neither. Python is simply `fmlas = [x.formula for x in prems if isinstance(x,ia.LabeledFormula)] + [conc]`. Replace with the faithful one-liner: take `lf.Formula.(lg.Expr)` directly.

### In `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/proof/phase5_matching.go`

**(B6)** `extractSymbol` (line 674) — does not unwrap its `ast.Node` argument before `n.(*lg.Const)` at line 678. Called from `ParameterizeSchema` line 570 on `cd.DeclArgs[0]`, which is `CompiledNode`-wrapped. Fix: insert `n = unwrapCompiledNode(n)` at the top of `extractSymbol`; this centralizes the unwrap for all callers.

**(B7)** `ApplyMatchGoalNode` line 1345 — `p.(lg.Sort)` without unwrap.

**(B8)** `ApplyMatchGoalNode` line 1359 — `args[0].(*lg.Const)` without unwrap.

**(B9)** `ApplyMatchGoalNode` line 1390 — `args[0].(*lg.Lambda)` without unwrap (lambda-filter guard would let wrapped lambdas through, diverging from Python's `is_lambda`).

**(B10)** `ApplyMatchGoalNodeNonAlt` line 1432 — `p.(lg.Sort)` without unwrap.

**(B11)** `ApplyMatchGoalNodeNonAlt` line 1444 — `args[0].(*lg.Const)` without unwrap.

**(B12)** `ApplyMatchGoalNodeNonAlt` line 1473 — `args[0].(*lg.Lambda)` without unwrap.

### Verified OK (no fix needed)

- `check/ranking.go:224`, `check/ranking_tactic.go:49`, `check/l2s.go:336`, `check/l2s.go:445/465/521`, `check/l2s_auto.go:56/957` — only type-assert on `*ast.LabeledFormula`, which is never wrapped.
- `proof/goal.go:552` (`GoalSubst`), `:561` (`GoalAddPrem`), `:571` (`GoalRemovePrem`), `:595` (`TrivialGoal`) — pass premises opaquely through to `CloneGoal`/`MakeGoal`, or only check `*ast.LabeledFormula`.
- `isolate/helpers.go:761-778` — already correctly unwraps `*ast.CompiledNode` for native-arg resolution.
- `proof/goal.go:306-324` (`GoalIsDefn`) — already correct; used as the template.

### Out of scope

- Changing `CompileSchemaPremWithSig` to stop wrapping (`phase6.go:1200/1218/1225`). That wrapper is load-bearing across many non-schema paths (compiled actions, native defs, condition wrappers); refactoring it out is a separate, cross-cutting port task.
- Python has no `PropertyDecl` branch in `compile_schema_prem`. `phase6.go:1205-1218` has one. That is a separate Python-conformance question to address when a divergence surfaces it; not in scope here.

## Planned edits

### File 1: `proof/goal.go`

**Replace `GoalDefns` (lines 342-360) body loop** with unwrap-first:

```go
for _, p := range GoalPrems(goal) {
    p = unwrapCompiledNode(p)
    if cd, ok := p.(*ast.ConstantDecl); ok {
        args := cd.Args()
        if len(args) > 0 {
            if c, ok := unwrapCompiledNode(args[0]).(*lg.Const); ok {
                res[lg.Key(c)] = c
            }
        }
    }
    if us, ok := p.(*lg.UninterpretedSort); ok {
        res[lg.Key(us)] = us
    }
}
```

**Replace `GoalVocab` per-prem loop (lines 387-418)** with unwrap-first, Python-faithful fmlas collection:

```go
for _, p := range prems {
    p = unwrapCompiledNode(p)
    // Python: sorts = [s for s in prems if isinstance(s, il.UninterpretedSort)]
    if us, ok := p.(*lg.UninterpretedSort); ok {
        sorts = append(sorts, us)
    }
    // Python: symbols = [x.args[0] for x in prems if isinstance(x, ia.ConstantDecl)]
    if cd, ok := p.(*ast.ConstantDecl); ok {
        args := cd.Args()
        if len(args) > 0 {
            if cc, ok := unwrapCompiledNode(args[0]).(*lg.Const); ok {
                symbols = append(symbols, cc)
            }
        }
    }
    // Python: fmlas = [x.formula for x in prems if isinstance(x, ia.LabeledFormula)]
    if lf, ok := p.(*ast.LabeledFormula); ok {
        if f, ok := lf.Formula.(lg.Expr); ok {
            fmlas = append(fmlas, f)
        }
    }
}
```

Keep the `concExpr := ConcAsExpr(conc)` append after the loop (mirrors Python's `+ [conc]`). Keep the `FreeVariables` union loop unchanged.

### File 2: `proof/phase5_matching.go`

**`extractSymbol` at line 674** — add `n = unwrapCompiledNode(n)` as the first statement after the `nil` check. Single-line change; centralizes unwrap for `ParameterizeSchema` and any other caller.

**`ApplyMatchGoalNode` at lines 1342-1379** — unwrap `p` at the top of the loop body; unwrap `args[0]` before `*lg.Const` at line 1359; unwrap `args[0]` before `*lg.Lambda` at line 1390. Write-back paths (Clone with new args, keeping original `p` on non-match) are unchanged.

**`ApplyMatchGoalNodeNonAlt` at lines 1429-1481** — same three unwraps at lines 1432, 1444, 1473.

The unwrap is idempotent and a no-op on non-`CompiledNode` values, so wrapping existing loop bodies with `p = unwrapCompiledNode(p)` is safe even when the premise is a `LabeledFormula` or other unwrapped type.

## Critical files

- `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/proof/goal.go` — `GoalDefns` (342-360) and `GoalVocab` (378-446).
- `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/proof/phase5_matching.go` — `extractSymbol` (674-686), `ApplyMatchGoalNode` (~1336-1418), `ApplyMatchGoalNodeNonAlt` (~1423-1489).

## Reference files (read-only, for verifying Python behavior)

- `/Users/jaten/ivy/pyivy/ivy/ivy/ivy_proof.py` lines 540-547 (`goal_defns`), 719-737 (`goal_vocab`).
- `/Users/jaten/ivy/pyivy/ivy/ivy/ivy_logic_utils.py` lines 523-539 (`variables_ast` / `used_variables_asts`).
- `/Users/jaten/ivy/pyivy/ivy/ivy/ivy_proof.py` — Python `apply_match_goal`, `is_lambda` (for phase5 mirroring).
- `/Users/jaten/ivy/pyivy/ivy/ivy/ivy_compiler.py` lines 1026-1044 (`compile_schema_prem`, shows bare returns).

## Existing utilities to reuse (do not duplicate)

- `unwrapCompiledNode` — `proof/goal.go:331`.
- `ConcAsExpr` / `GoalConc` — already used for conclusion.
- `lu.FreeVariables` — `logicutil/logicutil.go:61`; the union loop following `FreeVariables` mirrors Python's `apply_gen_to_list`.
- `GoalIsDefn` at `proof/goal.go:306-324` — unwrap-pattern template.

## Verification

1. `cd ~/ivy/goivy && make golden-2hr` — should advance past line 191504. Expected new Go-side trace at that point: `nsorts=1 nsymbols=3 nvariables=5`, matching Python.
2. The run should progress further. If a new divergence halts it, capture it; that is a separate follow-up fix.
3. `cd ~/ivy/goivy && make test` — run the unit-test suite to confirm no regressions. Do NOT use `go test ./...` (XTRACE-on default makes it intractable; per project memory).
4. No new unit tests are added — golden-2hr XTRACE conformance is the canonical coverage.

## Risk / rollback

All edits are localized and idempotent (unwraps are no-ops on non-wrapped values), so reverting any single edit cleanly restores prior behavior. The blast radius of an incorrect unwrap is contained to the specific type-assertion branch; adjacent code that writes back (e.g., `cd.Clone([]ast.Node{...})`) is untouched and continues to use whatever value the existing code produced.
