//go:build wasip1

package main

import "github.com/glycerine/ivy/goivy/smt"

func main() {}

//go:wasmexport smt_solver_sat_true
func smtSolverSatTrue() int32 {
	ctx := smt.NewContext()
	defer ctx.Close()

	solver := ctx.NewSolver()
	solver.Assert(ctx.BoolVal(true))
	return int32(solver.Check())
}

//go:wasmexport smt_solver_unsat_true_and_not_true
func smtSolverUnsatTrueAndNotTrue() int32 {
	ctx := smt.NewContext()
	defer ctx.Close()

	truth := ctx.BoolVal(true)
	solver := ctx.NewSolver()
	solver.Assert(truth)
	solver.Assert(ctx.Not(truth))
	return int32(solver.Check())
}
