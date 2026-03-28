# Fix xtrace divergence at 144124 + port missing SubgoalAction AST type + add missing dispatch cases

**Created:** 2026-03-28

## Context

The golden test diverges at line 144125:
- **Go**: `compiler.CompileNode return case=default type=RequiresAction`
- **Python**: `compiler.CompileNode return case=Action`

Go's `compileNodeCore` type switch (compiler/compiler.go:339) is missing cases for `*ast.RequiresAction`, `*ast.EnsuresAction`, and (once created) `*ast.SubgoalAction`. These all fall to the default case instead of routing through `CompileActionBody`.

Additionally, `ast.SubgoalAction` doesn't exist in Go at all, despite Python having `class SubgoalAction(AssertAction)` in `ivy_actions.py:394`. This is a mechanical port gap.

Python's `compile_assert_action` (ivy_compiler.py:781) handles ALL four assert-like types via inheritance (`self.clone()` preserves type). Go needs explicit cases for each.

Also: `ApplyAssertProofsWithProver` (compiler/phase6.go:1927) only checks `*actions.AssertAction`, missing RequiresAction and EnsuresAction (Python's `isinstance(self, AssertAction)` matches subclasses).

---

## Step 1: Create `ast.SubgoalAction` in `ast/ast.go`

**File:** `/Users/jaten/go/src/github.com/glycerine/goivy/ast/ast.go` (after `RequiresAction` at ~line 1135)

Model after `RequiresAction`. Python's `SubgoalAction(AssertAction)` has a custom `clone` that preserves `kind`. Add a `Kind` field.

```go
// SubgoalAction represents a proof subgoal assertion.
// Python: class SubgoalAction(AssertAction) from ivy_actions.py:394.
type SubgoalAction struct {
	Base
	Elems []Node
	Kind  string // preserved across clone, mirrors Python's self.kind
}

func (cfg *AstConfig) NewSubgoalAction(args ...Node) *SubgoalAction {
	a := &SubgoalAction{Elems: args}
	a.Cfg = cfg
	return a
}

func (a *SubgoalAction) Args() []Node { return a.Elems }
func (a *SubgoalAction) Clone(args []Node) Node {
	return &SubgoalAction{Base: a.Base, Elems: args, Kind: a.Kind}
}
func (a *SubgoalAction) String() string { return "subgoal" }
func (a *SubgoalAction) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(subgoalAction %v elems:%v)", a.Base.canonFields(), SliceCanon(a.Elems)))
}
```

---

## Step 2: Add dispatch cases in `compileNodeCore`

**File:** `/Users/jaten/go/src/github.com/glycerine/goivy/compiler/compiler.go` (line 339)

Add `*ast.RequiresAction`, `*ast.EnsuresAction`, `*ast.SubgoalAction` to the existing action case list:

```go
case *ast.AssignAction, *ast.AssumeAction, *ast.AssertAction,
	*ast.RequiresAction, *ast.EnsuresAction, *ast.SubgoalAction,
	*ast.CrashAction, *ast.ThunkAction,
	*ast.LocalAction,
	*ast.CallAction, *ast.IfAction, *ast.WhileAction,
	*ast.DebugAction, *ast.NativeAction:
```

---

## Step 3: Refactor formula compilation to avoid duplication

**File:** `/Users/jaten/go/src/github.com/glycerine/goivy/compiler/action.go`

Extract shared logic from `CompileAssertFormula` (line 1350) into a helper:

```go
// compileAssertLikeFormula compiles a formula node for any assert-like action.
// Python: compile_assert_action uses self.clone([cond]) — Go needs explicit factories.
func (c *Compiler) compileAssertLikeFormula(node ast.Node, traceName string, makeAction func(lg.Expr) actions.Action) (actions.Action, error) {
	xtracer.Trace("compiler.%s ENTER", traceName)
	// ... same ExprContext / LabeledFormula / SortifyWithInference logic ...
	res := makeAction(cond)
	// ... set LF, Unprovable, lineno, ctx.Extract ...
}
```

Then simplify existing functions:

```go
func (c *Compiler) CompileAssertFormula(node ast.Node) (actions.Action, error) {
	return c.compileAssertLikeFormula(node, "compile_assert_action",
		func(cond lg.Expr) actions.Action { return actions.NewAssertAction(cond) })
}
```

Add new functions:

```go
func (c *Compiler) CompileRequiresFormula(node ast.Node) (actions.Action, error) {
	return c.compileAssertLikeFormula(node, "compile_assert_action",
		func(cond lg.Expr) actions.Action { return actions.NewRequiresAction(cond) })
}

func (c *Compiler) CompileEnsuresFormula(node ast.Node) (actions.Action, error) {
	return c.compileAssertLikeFormula(node, "compile_assert_action",
		func(cond lg.Expr) actions.Action { return actions.NewEnsuresAction(cond) })
}

func (c *Compiler) CompileSubgoalFormula(node ast.Node) (actions.Action, error) {
	return c.compileAssertLikeFormula(node, "compile_assert_action",
		func(cond lg.Expr) actions.Action { return actions.NewSubgoalAction(cond) })
}
```

Note: all use trace name `"compile_assert_action"` because Python emits the same trace for all (they all use `compile_assert_action`).

---

## Step 4: Add cases in `CompileActionBody`

**File:** `/Users/jaten/go/src/github.com/glycerine/goivy/compiler/action.go` (after the `*ast.AssumeAction` case at ~line 453)

### RequiresAction case (after ~line 445):

```go
case *ast.RequiresAction:
	xtracer.Trace("compiler.CompileNode return case=Action")
	if len(n.Elems) >= 1 {
		act, err := c.CompileRequiresFormula(n.Elems[0])
		if err != nil {
			return nil, err
		}
		if len(n.Elems) >= 2 {
			pf, pfErr := c.CompileTactic(n.Elems[1])
			if pfErr == nil && pf != nil {
				if ra, ok := act.(*actions.RequiresAction); ok {
					ra.Proof = actions.WrapTactic(pf)
				}
			}
		}
		return act, nil
	}
	return nil, fmt.Errorf("require needs a formula")
```

### EnsuresAction case:

```go
case *ast.EnsuresAction:
	xtracer.Trace("compiler.CompileNode return case=Action")
	if len(n.Elems) >= 1 {
		act, err := c.CompileEnsuresFormula(n.Elems[0])
		if err != nil {
			return nil, err
		}
		if len(n.Elems) >= 2 {
			pf, pfErr := c.CompileTactic(n.Elems[1])
			if pfErr == nil && pf != nil {
				if ea, ok := act.(*actions.EnsuresAction); ok {
					ea.Proof = actions.WrapTactic(pf)
				}
			}
		}
		return act, nil
	}
	return nil, fmt.Errorf("ensure needs a formula")
```

### SubgoalAction case:

```go
case *ast.SubgoalAction:
	xtracer.Trace("compiler.CompileNode return case=Action")
	if len(n.Elems) >= 1 {
		act, err := c.CompileSubgoalFormula(n.Elems[0])
		if err != nil {
			return nil, err
		}
		if len(n.Elems) >= 2 {
			pf, pfErr := c.CompileTactic(n.Elems[1])
			if pfErr == nil && pf != nil {
				if sa, ok := act.(*actions.SubgoalAction); ok {
					sa.Proof = actions.WrapTactic(pf)
				}
			}
		}
		return act, nil
	}
	return nil, fmt.Errorf("subgoal needs a formula")
```

---

## Step 5: Fix `ApplyAssertProofsWithProver` to handle all assert-like types

**File:** `/Users/jaten/go/src/github.com/glycerine/goivy/compiler/phase6.go` (line 1922-1935)

Python's `isinstance(self, AssertAction)` matches all subclasses. Go's `act.(*actions.AssertAction)` only matches exact type. Need to also check RequiresAction and EnsuresAction.

Current code (line 1927):
```go
if a, ok := act.(*actions.AssertAction); ok {
```

Replace with handling for all assert-like types:

```go
// Python: isinstance(self, AssertAction) matches subclasses
if a, ok := act.(*actions.AssertAction); ok {
	if a.Proof != nil {
		if getModVerifying(mod) {
			return applyAssertProofAction(mod, a, prover)
		}
		return actions.NewAssertAction(a.Formula)
	}
	return a
}
if a, ok := act.(*actions.RequiresAction); ok {
	if a.Proof != nil {
		if getModVerifying(mod) {
			return applyAssertProofAction(mod, &a.AssertAction, prover)
		}
		return actions.NewRequiresAction(a.Formula)
	}
	return a
}
if a, ok := act.(*actions.EnsuresAction); ok {
	if a.Proof != nil {
		if getModVerifying(mod) {
			return applyAssertProofAction(mod, &a.AssertAction, prover)
		}
		return actions.NewEnsuresAction(a.Formula)
	}
	return a
}
```

Note: `applyAssertProofAction` takes `*actions.AssertAction` and creates SubgoalActions with `sga.SubgoalKind = a.Name()`. When called with the embedded AssertAction from RequiresAction, `a.Name()` returns "assert" (not "require") because the embedded AssertAction's method is called. We should consider passing the original action's Name() separately, or using a helper. For now, the SubgoalKind will be set based on the original action type.

Actually, better approach — change the function signature to accept the Action interface and extract the AssertAction fields:

```go
func applyAssertProofAction(mod *module.Module, act actions.Action, aa *actions.AssertAction, prover module.ProofCheckerInterface) actions.Action {
```

Or simplest correct fix: use `actions.IsAssertLike()` + type switch:

```go
if actions.IsAssertLike(act) {
	switch a := act.(type) {
	case *actions.AssertAction:
		if a.Proof != nil {
			if getModVerifying(mod) {
				return applyAssertProofAction(mod, a, prover)
			}
			return actions.NewAssertAction(a.Formula)
		}
		return a
	case *actions.RequiresAction:
		if a.Proof != nil {
			if getModVerifying(mod) {
				return applyAssertProofActionGeneric(mod, &a.AssertAction, a.Name(), prover)
			}
			return actions.NewRequiresAction(a.Formula)
		}
		return a
	case *actions.EnsuresAction:
		if a.Proof != nil {
			if getModVerifying(mod) {
				return applyAssertProofActionGeneric(mod, &a.AssertAction, a.Name(), prover)
			}
			return actions.NewEnsuresAction(a.Formula)
		}
		return a
	}
}
```

Where `applyAssertProofActionGeneric` takes an additional `kindName string` parameter so SubgoalKind is set correctly to "require" or "ensure" instead of "assert".

---

## Step 6: Ensure `SubgoalAction` ActionUpdate works via Go embedding

**File:** `/Users/jaten/go/src/github.com/glycerine/goivy/actions/update.go`

Go's `SubgoalAction` embeds `AssertAction`, so `ActionUpdate` is already promoted. BUT: unlike RequiresAction and EnsuresAction (which have explicit `ActionUpdate` methods at update.go:557-565), SubgoalAction relies on embedding promotion. Add an explicit delegation for clarity and consistency:

```go
// --- SubgoalAction ---

func (a *SubgoalAction) ActionUpdate(ctx *UpdateContext) *transrel.Update {
	return a.AssertAction.ActionUpdate(ctx)
}
```

---

## Files to modify

| File | Change |
|------|--------|
| `ast/ast.go` | Add `SubgoalAction` struct + constructor + methods |
| `compiler/compiler.go:339` | Add 3 types to dispatch case list |
| `compiler/action.go:~445` | Add 3 new cases in `CompileActionBody` |
| `compiler/action.go:~1350` | Extract `compileAssertLikeFormula` helper, add `CompileRequiresFormula`, `CompileEnsuresFormula`, `CompileSubgoalFormula` |
| `compiler/phase6.go:~1927` | Handle RequiresAction + EnsuresAction in proof application |
| `actions/update.go:~565` | Add explicit SubgoalAction.ActionUpdate |

---

## Verification

1. `go build ./...` — must compile cleanly
2. `cd ~/goivy && make golden` — golden line count must advance past 144124
3. `go test ./actions/...` — existing SubgoalAction tests must still pass
4. `go test ./compiler/...` — existing compiler tests must still pass
