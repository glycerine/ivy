package bmc

import (
	"fmt"

	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
)

func bmcSolverForCNF(data *cnf.CNF) (*bsat.Solver, error) {
	if err := data.Validate(); err != nil {
		return nil, err
	}
	solver := bsat.NewSolver()
	for solver.NumVars() < int32(data.Variables) {
		solver.NewVariable()
	}
	for clauseIndex, clause := range data.Clauses {
		literals := make([]bsat.Literal, len(clause))
		for i, literal := range clause {
			variable := literal >> 1
			if literal&1 != 0 {
				literals[i] = bsat.Negative(variable)
			} else {
				literals[i] = bsat.Positive(variable)
			}
		}
		if !solver.AddClause(literals...) {
			return nil, fmt.Errorf("bmc: inconsistent base CNF at clause %d", clauseIndex)
		}
	}
	return solver, nil
}

func bmcStatusName(status bsat.Status) string {
	switch status {
	case bsat.Satisfiable:
		return "SAT"
	case bsat.Unsatisfiable:
		return "UNSAT"
	default:
		return "UNKNOWN"
	}
}
