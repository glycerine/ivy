// Package fracomb holds combinational FRA verification primitives shared by
// sequential FRA and CEC without depending on either higher-level driver.
package fracomb

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
	"time"
)

type fraCecSolver interface {
	AddClause(...bsat.Literal) bool
	NewVariable() int32
	NumVars() int32
	Simplify() bool
	Solve(...bsat.Literal) bsat.Status
	SetLimits(int64, int64)
	Value(int32) (bool, bool)
}

// Fra_FraigSat checks the outputs of a combinational miter using ABC's SAT
// solver. It returns 1 for UNSAT, 0 for SAT and -1 for an exhausted limit.
// A SAT model is recorded on graph.Data in primary-input order.
func Fra_FraigSat(graph *aig.Graph, conflictLimit, propagationLimit int64, learnedStart, learnedDelta, learnedPercent, flipBits, andOutputs, newSolver, verbose int32) int32 {
	if graph == nil || aig.Aig_ManRegNum(graph) != 0 {
		panic("fra: SAT miter must be combinational")
	}
	graph.Data = nil
	started := time.Now()
	formula := cnf.Cnf_Derive(graph, int32(graph.OutputCount()))
	if formula == nil {
		return -1
	}
	if flipBits != 0 {
		cnf.Cnf_DataTranformPolarity(formula, 0)
	}
	var solver fraCecSolver
	if newSolver != 0 {
		solver = bsat.NewSolver2()
	} else {
		primary := bsat.NewSolver()
		primary.SetLearntReduction(learnedStart, learnedDelta, learnedPercent)
		solver = primary
	}
	for int(solver.NumVars()) < formula.Variables {
		solver.NewVariable()
	}
	for _, row := range formula.Clauses {
		if newSolver != 0 {
			// Cnf_DataWriteIntoSolver2 passes ID zero, rejected by the
			// assertion in ABC's sat_solver2_addclause.
			panic("fra: source solver2 rejects clause ID zero")
		}
		if !solver.AddClause(row...) {
			return 1
		}
	}
	outputs := make([]bsat.Literal, 0, graph.OutputCount())
	for _, edge := range graph.Outputs() {
		variable := formula.VarNums[edge.NodeID()]
		if variable < 0 {
			return -1
		}
		outputs = append(outputs, bsat.Positive(variable))
	}
	if andOutputs != 0 {
		for _, output := range outputs {
			if !solver.AddClause(output) {
				return 1
			}
		}
	} else if !solver.AddClause(outputs...) {
		return 1
	}
	if conflictLimit < 0 || propagationLimit < 0 {
		panic("fra: negative SAT limit")
	}
	if !solver.Simplify() {
		return 1
	}
	solver.SetLimits(conflictLimit, propagationLimit)
	status := solver.Solve()
	if verbose != 0 {
		fmt.Printf("CNF stats: Vars = %6d. Clauses = %7d. Time = %9.2f sec\n", formula.Variables, len(formula.Clauses), time.Since(started).Seconds())
	}
	switch status {
	case bsat.Satisfiable:
		model := make([]int32, graph.InputCount())
		for i, input := range graph.Inputs() {
			value, ok := solver.Value(formula.VarNums[input.NodeID()])
			if !ok {
				// An input absent from all clauses is unconstrained.
				continue
			}
			// ABC returns the solver's raw PI model after CNF phase
			// transformation; it does not undo a nonzero PI phase.
			if value {
				model[i] = 1
			}
		}
		graph.Data = model
		return 0
	case bsat.Unsatisfiable:
		return 1
	default:
		return -1
	}
}
