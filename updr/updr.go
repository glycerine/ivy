// Copyright (c) Microsoft Corporation. All Rights Reserved.
// Ported to Go from ivy_updr.py.

package updr

import (
	"fmt"
	"strings"

	"github.com/glycerine/goivy/module"
	"github.com/glycerine/goivy/z3bridge"
)

// UPDRResult holds the result of running UPDR on a module.
type UPDRResult struct {
	Valid     bool   // true if all properties hold
	Invariant string // human-readable invariant (if valid)
	Error     string // description of counterexample (if invalid)
	Stats     UPDRStats
}

// UPDRStats captures performance statistics.
type UPDRStats struct {
	NumFrames      int
	NumIterations  int
	NumSATQueries  int
	NumClauses     int
	NumUnivClauses int
}

// CheckModule runs UPDR on an Ivy module to verify safety properties.
//
// The algorithm:
//  1. Extracts the transition system from the module:
//     - Gets all non-error actions, computes their updates via transrel
//     - Classifies symbols: flexible (updated), inflexible (constant), skolem
//     - Computes error states via reverse image of error action
//  2. Makes frames explicit via transrel.FrameUpdate
//  3. Converts init/trans/bad/axioms to Z3 via solver package
//  4. Runs PDR.Run()
//  5. Returns result with statistics
//
// This is a placeholder that demonstrates the structure. A full implementation
// requires the complete module/transrel/solver pipeline to be wired up.
func CheckModule(mod *module.Module) (*UPDRResult, error) {
	if mod == nil {
		return nil, fmt.Errorf("updr: nil module")
	}

	// Create a Z3 context for this verification run
	ctx := z3bridge.NewZ3Context()

	// Extract transition system components from the module.
	// In a full implementation, this would:
	// 1. Iterate mod.Actions to find non-error actions
	// 2. Compute updates via transrel for each action
	// 3. Build init from mod.LabeledInits
	// 4. Build bad from conjectures (negated) or error actions
	// 5. Classify symbols as flexible/inflexible/skolem

	// For now, build a trivial safe system as a placeholder.
	// The real implementation will wire through transrel and solver.
	boolSort := ctx.BoolSort()
	x := ctx.Const("__updr_x", boolSort)

	x0 := []z3bridge.Expr{x}
	xn := []z3bridge.Expr{ctx.Const("__updr_x_next", boolSort)}

	// init: x = false (trivially safe starting state)
	initExpr := ctx.Not(x)
	// bad: false (nothing is bad — trivially safe)
	badExpr := ctx.BoolVal(false)
	// trans: x' = x (identity)
	transExpr := ctx.Iff(xn[0], x)

	pdr := NewPDR(ctx, initExpr, transExpr, badExpr, x0, nil, xn)
	result := pdr.Run()

	stats := UPDRStats{
		NumFrames:     pdr.N,
		NumIterations: pdr.IterationCount,
		NumSATQueries: pdr.SATQueryCount,
	}

	if result.Valid {
		return &UPDRResult{
			Valid:     true,
			Invariant: result.Invariant.String(),
			Stats:     stats,
		}, nil
	}

	return &UPDRResult{
		Valid: false,
		Error: fmt.Sprintf("counterexample trace with %d steps", len(result.Trace)),
		Stats: stats,
	}, nil
}

// ForwardClauses renames clauses from current-state vocabulary to
// next-state vocabulary, skipping inflexible (constant) symbols.
//
// This is used when propagating clauses forward in the frame sequence.
// Clauses about inflexible symbols don't need renaming since those
// symbols don't change between states.
func ForwardClauses(ctx *z3bridge.Context, clauses z3bridge.Expr,
	x0, xn []z3bridge.Expr) z3bridge.Expr {
	if len(x0) == 0 {
		return clauses
	}
	return ctx.Substitute(clauses, x0, xn)
}

// NumClauses counts the number of top-level conjuncts in an expression.
// If the expression is a conjunction (And), it returns the number of children.
// Otherwise returns 1.
func NumClauses(e z3bridge.Expr) int {
	s := e.String()
	if s == "true" {
		return 0
	}
	if s == "false" {
		return 1
	}
	// Count top-level "and" conjuncts by looking at the string representation.
	// This is a heuristic; a proper implementation would inspect the AST.
	if strings.HasPrefix(s, "(and") {
		count := 0
		depth := 0
		for _, ch := range s {
			switch ch {
			case '(':
				depth++
			case ')':
				depth--
			case ' ':
				if depth == 1 {
					count++
				}
			}
		}
		return count
	}
	return 1
}

// NumUnivClauses counts the number of universally quantified clauses
// in an expression. A clause is "universal" if it is a ForAll or
// if it is a ground clause (implicitly universally quantified).
func NumUnivClauses(e z3bridge.Expr) int {
	s := e.String()
	return strings.Count(s, "forall") + NumClauses(e)
}
