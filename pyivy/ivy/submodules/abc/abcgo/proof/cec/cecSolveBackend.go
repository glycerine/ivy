package cec

import (
	"github.com/glycerine/abcgo/proof/cec/core"
	"github.com/glycerine/abcgo/sat/bsat"
)

// cecClassicSolver uses ABC's bsat package for the default SolverType -1.
type cecClassicSolver struct {
	graph *Gia_Man_t
	pars  *Cec_ParSat_t
	sat   *bsat.Solver
	vars  []int32
	since int32
}

func newCecClassicSolver(g *Gia_Man_t, pars *Cec_ParSat_t) *cecClassicSolver {
	s := &cecClassicSolver{graph: g, pars: pars}
	s.reset()
	return s
}
func (s *cecClassicSolver) reset() {
	s.sat = bsat.NewSolver()
	s.vars = make([]int32, s.graph.ObjectCount())
	for i := range s.vars {
		s.vars[i] = -1
	}
	s.vars[0] = s.sat.NewVariable() << 1
	if !s.sat.AddClause(s.vars[0] ^ 1) {
		panic("cec: constant-zero SAT clause failed")
	}
	s.since = 0
}
func (s *cecClassicSolver) binary(a, b int32, xor bool) int32 {
	if a == b {
		if xor {
			return s.vars[0]
		}
		return a
	}
	if a == b^1 {
		if xor {
			return s.vars[0] ^ 1
		}
		return s.vars[0]
	}
	z := s.sat.NewVariable() << 1
	clauses := [][]int32{{a ^ 1, b ^ 1, z}, {a, z ^ 1}, {b, z ^ 1}}
	if xor {
		clauses = [][]int32{{a, b, z ^ 1}, {a ^ 1, b ^ 1, z ^ 1}, {a, b ^ 1, z}, {a ^ 1, b, z}}
	}
	for _, clause := range clauses {
		if !s.sat.AddClause(clause...) {
			panic("cec: inconsistent circuit clauses")
		}
	}
	return z
}
func (s *cecClassicSolver) encode(lit int32) int32 {
	id := lit >> 1
	if id < 0 || int(id) >= len(s.vars) {
		panic("cec: invalid graph literal")
	}
	if s.vars[id] >= 0 {
		return s.vars[id] ^ (lit & 1)
	}
	a, b, c, kind := s.graph.Gate(id)
	var value int32
	switch kind {
	case core.Input:
		value = s.sat.NewVariable() << 1
	case core.Buffer, core.Output:
		value = s.encode(a)
	case core.And, core.Xor:
		value = s.binary(s.encode(a), s.encode(b), kind == core.Xor)
	case core.Mux:
		cond, t, e := s.encode(a), s.encode(b), s.encode(c)
		x := s.binary(cond, t, false)
		y := s.binary(cond^1, e, false)
		value = s.binary(x^1, y^1, false) ^ 1
	default:
		panic("cec: unsupported graph object")
	}
	s.vars[id] = value
	return value ^ (lit & 1)
}
func (s *cecClassicSolver) check(lit int32) int32 {
	if lit == 0 {
		return 1
	}
	if lit == 1 {
		return 0
	}
	if s.pars.nBTLimit < 0 {
		return -1
	}
	s.since++
	if s.pars.nSatVarMax > 0 && s.sat.NumVars() > s.pars.nSatVarMax && s.since > s.pars.nCallsRecycle {
		s.reset()
	}
	encoded := s.encode(lit)
	s.sat.SetLimits(int64(s.pars.nBTLimit), 0)
	switch s.sat.Solve(encoded) {
	case bsat.Unsatisfiable:
		if !s.sat.AddClause(encoded ^ 1) {
			panic("cec: failed to retain proved clause")
		}
		return 1
	case bsat.Satisfiable:
		return 0
	default:
		return -1
	}
}
func (s *cecClassicSolver) inputs() []bool {
	inputs := s.graph.InputLiterals()
	model := make([]bool, len(inputs))
	for i, lit := range inputs {
		cached := s.vars[lit>>1]
		if cached < 0 {
			continue
		}
		value, ok := s.sat.Value(cached >> 1)
		if ok {
			model[i] = value != (cached&1 != 0)
		}
	}
	return model
}
