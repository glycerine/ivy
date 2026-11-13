package fra

import (
	"errors"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
)

// clausInductivePass translates one Fra_ClausInductiveClauses refinement
// round. Clauses are encoded at frame zero; costs of rejected candidates are
// set to -1. Proven clauses and currently enabled candidates constrain all
// preceding frames. Recorded SAT models accelerate later candidate checks.
func clausInductivePass(data *cnf.Cnf_Dat_t, frames int32, proven, candidates [][]int32, costs []int32, limit int64) (int, error) {
	if data == nil || frames <= 0 || len(candidates) != len(costs) {
		return 0, errors.New("invalid clause induction input")
	}
	solver := cnf.Cnf_DataWriteIntoSolver(data, frames+1, 0)
	if solver == nil {
		return 0, errors.New("induction transition relation is inconsistent")
	}
	addFrame := func(c []int32, frame int32) bool {
		lifted := make([]int32, len(c))
		offset := 2 * int32(data.Variables) * frame
		for i, lit := range c {
			lifted[i] = lit + offset
		}
		return solver.AddClause(lifted...)
	}
	for frame := int32(0); frame < frames; frame++ {
		for _, clause := range proven {
			if !addFrame(clause, frame) {
				return 0, errors.New("proven clauses are inconsistent")
			}
		}
		for i, clause := range candidates {
			if costs[i] != -1 && !addFrame(clause, frame) {
				return 0, errors.New("candidate clauses are inconsistent")
			}
		}
	}
	return clausCheckInduction(solver, data.Variables, frames, candidates, costs, limit)
}

// Reuse the same incremental solver as the source after its property query.
func clausCheckInduction(solver *bsat.Solver, variables int, frames int32, candidates [][]int32, costs []int32, limit int64) (int, error) {
	models := newClausModels(variables)
	offset := frames * int32(variables)
	removed := 0
	for i, clause := range candidates {
		if costs[i] == -1 {
			continue
		}
		lifted := make([]int32, len(clause))
		for j, lit := range clause {
			lifted[j] = lit + 2*offset
		}
		if models.disproves(lifted, offset) {
			costs[i] = -1
			removed++
			continue
		}
		assumptions := make([]int32, len(lifted))
		for j, lit := range lifted {
			assumptions[j] = lit ^ 1
		}
		status := solver.SolveWithSourceLimits(assumptions, limit, 0, 0, 0)
		if status == bsat.Unsatisfiable {
			continue
		}
		if status == bsat.Satisfiable {
			model := make([]bool, variables)
			for j := range model {
				value, ok := solver.Value(offset + int32(j))
				if !ok {
					return removed, errors.New("incomplete induction model")
				}
				model[j] = value
			}
			models.record(model)
		}
		// Unknown cannot establish induction and must never become a proof.
		costs[i] = -1
		removed++
	}
	return removed, nil
}
