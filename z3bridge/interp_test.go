package z3bridge

import (
	"testing"
)

// TestBinaryInterpolant verifies that the Z3 interpolation API works.
// Computes binary_interpolant(x < 0, x > 2) and verifies the result
// is an interpolant (i.e., implied by x < 0 and inconsistent with x > 2).
// Matches the Python z3 docstring example.
func TestBinaryInterpolant(t *testing.T) {
	// Create an interpolation-capable context
	ctx := NewInterpolationContext()

	// Create integer variable x
	intSort := ctx.IntSort()
	x := ctx.Const("x", intSort)

	// a: x < 0
	a := ctx.Lt(x, ctx.IntVal(0))
	// b: x > 2
	b := ctx.Gt(x, ctx.IntVal(2))

	// Build interpolation pattern: And(Interpolant(a), b)
	marked := ctx.MkInterpolant(a)
	pattern := ctx.And(marked, b)

	// Compute interpolant
	interps, err := ctx.ComputeInterpolant(pattern)
	if err != nil {
		t.Fatalf("ComputeInterpolant failed: %v", err)
	}
	if len(interps) == 0 {
		t.Fatal("ComputeInterpolant returned no interpolants")
	}

	itp := interps[0]
	t.Logf("interpolant: %s", itp.String())

	// Verify interpolant properties:
	// 1) a implies itp
	solver1 := ctx.NewSolver()
	solver1.Assert(a)
	solver1.Assert(ctx.Not(itp))
	if solver1.Check() != Unsat {
		t.Error("interpolant is not implied by a (x < 0)")
	}

	// 2) itp AND b is unsat
	solver2 := ctx.NewSolver()
	solver2.Assert(itp)
	solver2.Assert(b)
	if solver2.Check() != Unsat {
		t.Error("interpolant AND b (x > 2) is satisfiable")
	}
}

// TestComputeInterpolantSat verifies that ComputeInterpolant returns
// an error when the formula is satisfiable (no interpolant exists).
func TestComputeInterpolantSat(t *testing.T) {
	ctx := NewInterpolationContext()
	intSort := ctx.IntSort()
	x := ctx.Const("x", intSort)

	// a: x > 0, b: x < 10 — these are satisfiable together
	a := ctx.Gt(x, ctx.IntVal(0))
	b := ctx.Lt(x, ctx.IntVal(10))

	marked := ctx.MkInterpolant(a)
	pattern := ctx.And(marked, b)

	_, err := ctx.ComputeInterpolant(pattern)
	if err == nil {
		t.Error("expected error for satisfiable formula, got nil")
	}
}
