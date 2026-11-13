// Package integration contains small adapters that exercise interactions
// between independently maintained ABC Go packages.
package integration

import (
	"errors"

	"github.com/glycerine/abcgo/aig/miniaig"
	"github.com/glycerine/abcgo/sat/satoko"
)

// MiniAIGSolver builds a SAT solver whose variables and clauses encode p.
// AIG literal l maps to SAT literal l, so callers can use graph literals
// directly as assumptions. Node zero is constrained to false.
func MiniAIGSolver(p *miniaig.Mini_Aig_t) (*satoko.Solver, error) {
	if p == nil {
		return nil, errors.New("integration: nil MiniAIG")
	}

	s := satoko.New()
	s.SetVariables(miniaig.Mini_AigNodeNum(p))
	s.AddClause([]int{1})
	for id := 1; id < miniaig.Mini_AigNodeNum(p); id++ {
		if !miniaig.Mini_AigNodeIsAnd(p, id) {
			continue
		}
		a, b, z := miniaig.Mini_AigNodeFanin0(p, id), miniaig.Mini_AigNodeFanin1(p, id), id*2
		s.AddClause([]int{z ^ 1, a})
		s.AddClause([]int{z ^ 1, b})
		s.AddClause([]int{z, a ^ 1, b ^ 1})
	}
	return s, nil
}
