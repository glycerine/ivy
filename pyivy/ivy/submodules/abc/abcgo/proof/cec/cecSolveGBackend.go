package cec

import (
	"fmt"
	"github.com/glycerine/abcgo/proof/cec/core"
	"github.com/glycerine/abcgo/sat/glucose2"
)

// cecGeneralSolver follows cecSolveG.c's Glucose2 modes: mode 0 uses CNF,
// mode 1 uses CNF plus justification, and mode 2 uses circuit justification.
type cecGeneralSolver struct {
	graph                                             *Gia_Man_t
	pars                                              *Cec_ParSat_t
	sat                                               *glucose2.SimpSolver
	vars                                              []int32 // object id -> SAT literal, or -1 when not encoded
	phases                                            []bool
	since                                             int32
	recycles, calls, satCount, unsatCount, undecCount int32
	totalConflicts                                    int64
}

func newCecGeneralSolver(graph *Gia_Man_t, pars *Cec_ParSat_t) *cecGeneralSolver {
	if pars.SolverType < 0 || pars.SolverType > 2 {
		panic("cec: invalid general solver type")
	}
	s := &cecGeneralSolver{graph: graph, pars: pars, phases: cecSolverPhases(graph, pars.SolverType == 0 && pars.fPolarFlip != 0)}
	s.reset()
	return s
}
func (s *cecGeneralSolver) reset() {
	if s.sat != nil {
		s.totalConflicts += s.sat.Conflicts
	}
	s.sat = glucose2.NewABCSolver()
	s.sat.JustificationMode = int(s.pars.SolverType)
	s.vars = make([]int32, s.graph.ObjectCount())
	for i := range s.vars {
		s.vars[i] = -1
	}
	v := s.sat.NewVar(true, true)
	s.vars[0] = int32(v) << 1
	if !s.sat.AddClause([]glucose2.Lit{glucose2.Lit(s.vars[0] ^ 1)}) {
		panic("cec: constant-zero clause failed")
	}
	s.since = 0
	s.recycles++
}
func (s *cecGeneralSolver) binary(a, b int32, xor, phase bool) int32 {
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
	v := s.sat.NewVar(true, true)
	z := int32(v) << 1
	if phase {
		z ^= 1
	}
	if s.pars.SolverType != 0 {
		left, right := a, b
		if (left > right) != xor {
			left, right = right, left
		}
		s.sat.SetVarFaninLits(v, glucose2.Lit(left), glucose2.Lit(right))
	}
	if s.pars.SolverType < 2 {
		clauses := [][]int32{{a ^ 1, b ^ 1, z}, {a, z ^ 1}, {b, z ^ 1}}
		if xor {
			clauses = [][]int32{{a, b, z ^ 1}, {a ^ 1, b ^ 1, z ^ 1}, {a, b ^ 1, z}, {a ^ 1, b, z}}
		}
		for _, clause := range clauses {
			lits := make([]glucose2.Lit, len(clause))
			for i, l := range clause {
				lits[i] = glucose2.Lit(l)
			}
			if !s.sat.AddClause(lits) {
				panic("cec: inconsistent general solver CNF")
			}
		}
	}
	return z
}
func (s *cecGeneralSolver) encode(lit int32) int32 {
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
		value = int32(s.sat.NewVar(true, true)) << 1
	case core.Buffer, core.Output:
		value = s.encode(a)
	case core.And, core.Xor:
		value = s.binary(s.encode(a), s.encode(b), kind == core.Xor, s.phases[id])
	case core.Mux:
		cond, t, e := s.encode(a), s.encode(b), s.encode(c)
		x := s.binary(cond, t, false, false)
		y := s.binary(cond^1, e, false, false)
		value = s.binary(x^1, y^1, false, false) ^ 1
	default:
		panic(fmt.Sprintf("cec: unsupported graph object %d", id))
	}
	s.vars[id] = value
	return value ^ (lit & 1)
}
func (s *cecGeneralSolver) check(lit int32) int32 {
	if lit == 0 {
		return 1
	}
	if lit == 1 {
		return 0
	}
	s.since++
	s.calls++
	if s.pars.nSatVarMax > 0 && s.sat.NVars() > int(s.pars.nSatVarMax) && s.since > s.pars.nCallsRecycle {
		s.reset()
	}
	encoded := s.encode(lit)
	if s.pars.SolverType != 0 {
		s.sat.SetNewRound()
		s.sat.MarkCone(glucose2.Lit(encoded).Var())
	}
	if s.pars.nBTLimit > 0 {
		s.sat.SetConfBudget(int64(s.pars.nBTLimit))
	} else {
		s.sat.BudgetOff()
	}
	switch s.sat.SolveLimited([]glucose2.Lit{glucose2.Lit(encoded)}, false) {
	case glucose2.LFalse:
		s.unsatCount++
		if !s.sat.AddClause([]glucose2.Lit{glucose2.Lit(encoded ^ 1)}) {
			panic("cec: failed to retain proved clause")
		}
		return 1
	case glucose2.LTrue:
		s.satCount++
		return 0
	default:
		s.undecCount++
		return -1
	}
}

func (s *cecGeneralSolver) printStats() {
	fmt.Printf("CEC Glucose2: %d queries, %d SAT, %d UNSAT, %d undecided, %d conflicts, %d recycles\n", s.calls, s.satCount, s.unsatCount, s.undecCount, s.totalConflicts+s.sat.Conflicts, s.recycles)
}
