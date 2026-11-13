// Incremental PDR clause reduction, ported from ABC pdrIncr.c.
package pdr

import (
	"fmt"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

// IPdr_ManCheckCubeReduce keeps the assumptions in the final UNSAT core.
// The caller uses a dedicated last-frame solver; this permanently adds the
// property and guarded clauses to that solver, as in the original routine.
func IPdr_ManCheckCubeReduce(p *Pdr_Man_t, clauses *vec.Vec_Ptr_t, cube *Pdr_Set_t, limit int32) int32 {
	started := time.Now()
	solver := Pdr_ManSolver(p, 1)
	if cube != nil {
		fmt.Printf("SAT at the last frame.    Time = %v\n", time.Since(started))
		return 0
	}
	property := Pdr_ObjSatVar(p, 1, 2, aig.Aig_ManCo(p.pAig, p.iOutCur))
	if !solver.AddClause(bsat.Positive(property)) {
		panic("pdr: inconsistent property in reduction solver")
	}
	n := vec.Vec_PtrSize(clauses)
	first := Pdr_ManFreeVar(p, 1)
	for solver.NumVars() <= first {
		solver.NewVariable()
	}
	assumptions := make([]bsat.Literal, n)
	for i := 1; i < n; i++ {
		variable := Pdr_ManFreeVar(p, 1)
		if variable != first+int32(i) {
			panic("pdr: noncontiguous activation variables")
		}
		for solver.NumVars() <= variable {
			solver.NewVariable()
		}
	}
	for i := 0; i < n; i++ {
		variable := first + int32(i)
		clause, ok := vec.Vec_PtrEntry(clauses, i).(*Pdr_Set_t)
		if !ok || clause == nil {
			panic("pdr: invalid reduction clause")
		}
		lits := append([]bsat.Literal(nil), pdrLiterals(Pdr_ManCubeToLits(p, 1, clause, 1, 0))...)
		lits = append(lits, bsat.Negative(variable))
		if !solver.AddClause(lits...) {
			panic("pdr: inconsistent guarded reduction clause")
		}
		assumptions[i] = bsat.Positive(variable)
	}
	solver.Compress()
	oldConflicts, oldPropagations := solver.Limits()
	solver.SetLimits(int64(limit), 0)
	result := solver.Solve(assumptions...)
	solver.SetLimits(oldConflicts, oldPropagations)
	if result == bsat.Unknown {
		return -1
	}
	if result == bsat.Satisfiable {
		fmt.Printf("SAT at the last frame.    Time = %v\n", time.Since(started))
		return 0
	}
	core := solver.UnsatCore()
	keep := make(map[int32]bool, len(core))
	for _, lit := range core {
		keep[int32(lit)>>1] = true
	}
	old := make([]*Pdr_Set_t, n)
	for i := range old {
		old[i] = vec.Vec_PtrEntry(clauses, i).(*Pdr_Set_t)
	}
	vec.Vec_PtrClear(clauses)
	for i, clause := range old {
		if keep[first+int32(i)] {
			vec.Vec_PtrPush(clauses, clause)
		} else {
			Pdr_SetDeref(clause)
		}
	}
	fmt.Printf("UNSAT at the last frame. nCores = %d (out of %d).    Time = %v\n", len(core), n, time.Since(started))
	return 1
}

// Abc_NtkDarIPdr converts the network, runs incremental PDR and transfers
// the resulting witnesses back to the caller's network.
func Abc_NtkDarIPdr(network *abc.Network, pars *Pdr_Par_t) int32 {
	started := time.Now()
	graph := abc.Abc_NtkToDar(network, 0, 1)
	result := IPdr_ManSolve(graph, pars)
	switch result {
	case 1:
		fmt.Print("Property proved.  ")
	case 0:
		cex := graph.SequentialCounterexample()
		if cex == nil {
			fmt.Println("Counter-example is not available.")
		} else {
			fmt.Printf("Output %d of miter %q was asserted in frame %d.  ", cex.IPo, network.Name, cex.IFrame)
			if saig.Saig_ManVerifyCex(graph, cex) == 0 {
				fmt.Println("Counter-example verification has FAILED.")
			}
		}
	case -1:
		fmt.Print("Property UNDECIDED.  ")
	default:
		panic("pdr: unexpected incremental solver result")
	}
	fmt.Printf("Time = %v\n", time.Since(started))
	network.SequenceModel = graph.TakeSequentialCounterexample()
	network.SetSequentialCounterexamples(graph.TakeSequentialCounterexamples())
	return result
}
