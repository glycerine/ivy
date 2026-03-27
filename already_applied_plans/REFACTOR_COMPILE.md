# Fix CompileTactic Trace Divergence at Line 88769

**Created:** 2026-03-27 (current session)

## Context

The golden test (`cd ~/goivy && make golden`) shows a trace divergence at line 88769:

```
go : XTRACE: compiler.CompileTactic ENTER
py : XTRACE: compiler.Thing ENTER type=ComposeTactics
```

**Root cause:** Go has a separate `CompileTactic` function (`compiler/compiler.go:1295`) that doesn't exist in Python. In Python, ALL compilation goes through a single `compile()` dispatch:

- Types with explicit compile overrides (LetTactic, IfTactic, etc.): `compile()` calls the override directly, **no `thing()` traces**
- Types without overrides (ComposeTactics, NullTactic, etc.): `compile()` = `thing()` -> `cmpl()` = `other_thing()` -> `clone([a.compile() for a in self.args])`, **producing Thing/CompileNode/OtherThing traces**

Go's `CompileTactic` handles ALL tactic types in one switch with a single "CompileTactic ENTER" trace, which doesn't exist in Python.

## Approach: Refactor CompileTactic to Match Python Dispatch

Restructure `CompileTactic` so its trace output matches Python's `compile()` dispatch:

1. **Types with explicit compile overrides in Python** (bypass `thing()`): handle at top of switch, produce NO Thing traces
2. **Types without compile overrides** (go through `thing()` -> `other_thing()`): handle in default case, produce the full Thing/CompileNode/OtherThing trace chain

No new types, interfaces, or methods needed. Just restructure the existing function.

## Classification of Tactic Types

**Have explicit `compile` overrides in Python (NO thing traces):**
- `SchemaInstantiation` (via TacticWithMatch.compile)
- `LetTactic` (compile_let_tactic -> return self)
- `WitnessTactic` (compile_witness_tactic -> return self)
- `UnfoldTactic` (compile_unfold_tactic -> return self)
- `ForgetTactic` (compile_forget_tactic -> return self)
- `FunctionTactic` (compile_function_tactic -> return self)
- `IfTactic` (compile_if_tactic -> compile cond + branches)
- `PropertyTactic` (compile_property_tactic -> compile name args + proof)
- `TacticTactic` (compile_tactic_tactic -> clone(self.args))
- `ProofTactic` (compile_proof_tactic -> clone([label, proof.compile()]))

**No compile override (thing -> other_thing, PRODUCE traces):**
- `ComposeTactics`
- `AssumeTactic`
- `AssumeGlobalTactic`
- `ShowGoalsTactic`
- `DeferGoalTactic`
- `NullTactic`
- `SpoilTactic`
- `Tactic` (base)

## Files to Modify

### 1. `compiler/compiler.go` (lines ~1295-1428)

Refactor `CompileTactic`:

```go
func (c *Compiler) CompileTactic(node ast.Node) (ast.Node, error) {
    if node == nil { return nil, nil }

    // --- Types with EXPLICIT compile overrides in Python ---
    // These bypass thing(), producing no Thing/CompileNode/OtherThing traces.
    switch n := node.(type) {
    case *ast.SchemaInstantiation:
        return n, nil
    case *ast.LetTactic:
        return n, nil
    case *ast.WitnessTactic:
        return n, nil
    case *ast.UnfoldTactic:
        return n, nil
    case *ast.ForgetTactic:
        return n, nil
    case *ast.FunctionTactic:
        return n, nil
    case *ast.IfTactic:
        // Python compile_if_tactic: sortify cond, compile branches
        cond, err := c.SortifyWithInference(n.Cond)
        if err != nil {
            cond, err = c.Thing(n.Cond)
            if err != nil { return n, nil }
        }
        thenBranch, err := c.CompileTactic(n.Then)
        if err != nil { return n, nil }
        elseBranch, err := c.CompileTactic(n.Else)
        if err != nil { return n, nil }
        condWrapper := &ast.CompiledNode{Node: cond}
        return n.Clone([]ast.Node{condWrapper, thenBranch, elseBranch}), nil
    case *ast.PropertyTactic:
        // Python compile_property_tactic
        prop := n.Prop
        name := n.PName
        if _, isNone := name.(*ast.NoneAST); !isNone && name != nil {
            nameArgs := name.Args()
            compiledArgs := make([]ast.Node, len(nameArgs))
            for i, arg := range nameArgs {
                compiled, err := c.Thing(arg)
                if err != nil { compiledArgs[i] = arg; continue }
                compiledArgs[i] = &ast.CompiledNode{Node: compiled}
            }
            name = name.Clone(compiledArgs)
        }
        proof, err := c.CompileTactic(n.Proof)
        if err != nil { return nil, err }
        return node.Clone([]ast.Node{prop, name, proof}), nil
    case *ast.TacticTactic:
        return n.Clone(n.Args()), nil
    case *ast.ProofTactic:
        proof, err := c.CompileTactic(n.Proof)
        if err != nil { proof = n.Proof }
        return &ast.ProofTactic{Base: n.Base, TLabel: n.TLabel, Proof: proof}, nil

    // --- Types WITHOUT compile overrides ---
    // These go through thing() -> other_thing() in Python.
    // Produces: Thing ENTER, CompileNode ENTER, CompileNode return default,
    //           OtherThing ENTER, [child compilation], OtherThing return, Thing return
    default:
        tn := typeName(node)
        xtracer.Trace(fmt.Sprintf("compiler.Thing ENTER type=%s", tn))
        xtracer.Trace(fmt.Sprintf("compiler.CompileNode ENTER type=%s", tn))
        xtracer.Trace(fmt.Sprintf("compiler.CompileNode return case=default type=%s", tn))
        xtracer.Trace(fmt.Sprintf("compiler.OtherThing ENTER type=%s", tn))
        // Python: self.clone([a.compile() for a in self.args])
        args := node.Args()
        compiled := make([]ast.Node, len(args))
        for i, a := range args {
            ca, err := c.CompileTactic(a)
            if err != nil { return nil, err }
            compiled[i] = ca
        }
        result := node.Clone(compiled)
        xtracer.Trace(fmt.Sprintf("compiler.OtherThing return type=%s sort_infer_root=false", tn))
        xtracer.Trace(fmt.Sprintf("compiler.Thing return type=%s", tn))
        return result, nil
    }
}
```

Key changes:
- **Remove** the `xtracer.Trace("compiler.CompileTactic ENTER")` at the top
- **Remove** individual cases for AssumeTactic, AssumeGlobalTactic, ShowGoalsTactic, DeferGoalTactic, NullTactic, SpoilTactic, Tactic — these all go to default
- **Keep** explicit-override types at top of switch
- **Default case** produces the correct Thing/CompileNode/OtherThing traces and does generic `clone(compile children)`

### 2. `compiler/phase6.go` — Check for duplicate functions

`CompileIfTactic` (line 1390), `CompilePropertyTactic` (line 1425), `CompileProofTactic` (line 1466) in phase6.go appear to be duplicates of the logic in CompileTactic's switch cases. Verify whether they are called from anywhere outside of CompileTactic. If they are only called from CompileTactic, they can be removed or left as-is (they're not causing the trace issue).

### 3. No caller changes needed

All callers (`decl.go:1385,1409`, `action.go:265`, `ivy_compile.go:502`) call `CompileTactic` which now produces the correct traces internally.

## Verification

1. Run `cd ~/goivy && make golden` and verify line 88769 no longer diverges
2. Check subsequent trace lines also match (the OtherThing return and Thing return traces for ComposeTactics)
3. Run `go build ./...` to ensure compilation passes
4. Run `go test ./compiler/...` to check for regressions
