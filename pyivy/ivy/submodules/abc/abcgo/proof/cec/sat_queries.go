package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/proof/cec/core"
	"github.com/glycerine/abcgo/sat/bsat"
	"time"
)

func Cec_ManSatSolverRecycle(p *Cec_ManSat_t) {
	p.source.reset()
	p.values = nil
}
func (p *Cec_ManSat_t) query(a int32, b *int32) int32 {
	p.values = nil
	if p.proof == nil {
		return -1
	}
	if p.Error = p.proof.SetConflictLimit(int64(p.pPars.NBTLimit)); p.Error != nil {
		return -1
	}
	if a == 0 && b == nil {
		return 1
	}
	if a == 1 && b == nil {
		panic("cec: constant-one node query")
	}
	if !p.pPars.deadline.IsZero() && !time.Now().Before(p.pPars.deadline) {
		return -1
	}
	s := p.source
	s.since++
	s.calls++
	if s.cnf.Solver() == nil || p.pPars.NSatVarMax > 0 && s.cnf.Solver().NumVars() > p.pPars.NSatVarMax && s.since > p.pPars.NCallsRecycle {
		s.reset()
	}
	s.cnf.Solver().SetDeadline(p.pPars.deadline)
	s.addNode(a >> 1)
	lits := []int32{s.literal(a)}
	if b != nil {
		s.addNode(*b >> 1)
		lits = append(lits, s.literal(*b))
	}
	result := s.cnf.Solver().SolveWithResourceLimits(lits, int64(p.pPars.NBTLimit), 0, 0, 0)
	switch result {
	case bsat.Unsatisfiable:
		s.unsatCount++
		for i := range lits {
			lits[i] ^= 1
		}
		s.addClause(lits...)
		p.result = core.Result{Status: core.Equivalent}
	case bsat.Satisfiable:
		s.satCount++
		inputs := make([]bool, len(p.pAig.InputLiterals()))
		for i, lit := range p.pAig.InputLiterals() {
			id := s.cnf.IDs()[lit>>1]
			if id != 0 {
				inputs[i], _ = s.cnf.Solver().Value(id)
			}
		}
		p.values, p.Error = p.proof.Model().Values(inputs)
		p.result = core.Result{Status: core.Different, Counterexample: &core.Counterexample{Inputs: inputs}}
	default:
		s.undecCount++
		p.result = core.Result{Status: core.Undecided}
	}
	return int32(p.result.Status)
}
func Cec_ManSatCheckNode(p *Cec_ManSat_t, obj *Gia_Obj_t) int32 {
	lit := cecObjectLiteral(p, obj)
	if lit == 0 {
		return 1
	}
	if lit == 1 {
		panic("cec: constant-one node query")
	}
	return p.query(lit, nil)
}
func Cec_ManSatCheckNodeTwo(p *Cec_ManSat_t, a, b *Gia_Obj_t) int32 {
	first, second := cecObjectLiteral(p, a), cecObjectLiteral(p, b)
	// These trivial cases precede SAT manager allocation and counters in C.
	if first == 0 || second == 0 || first == second^1 {
		return 1
	}
	if first == 1 && second == 1 {
		panic("cec: constant-one conjunction query")
	}
	return p.query(first, &second)
}

func cecObjectLiteral(p *Cec_ManSat_t, obj *Gia_Obj_t) int32 {
	return gia.Gia_ObjId(p.pAig, gia.Gia_Regular(obj))<<1 | gia.Gia_IsComplement(obj)
}
