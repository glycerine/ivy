//go:build !tinygo && wasip1

package main

import "github.com/glycerine/ivy/goivy/smt"

func probeSMTSolverSatTrue() int32 {
	ctx := smt.NewContext()
	defer ctx.Close()

	solver := ctx.NewSolver()
	solver.Assert(ctx.BoolVal(true))
	return int32(solver.Check())
}

func probeSMTSolverUnsatTrueAndNotTrue() int32 {
	ctx := smt.NewContext()
	defer ctx.Close()

	truth := ctx.BoolVal(true)
	solver := ctx.NewSolver()
	solver.Assert(truth)
	solver.Assert(ctx.Not(truth))
	return int32(solver.Check())
}
