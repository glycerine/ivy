package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

// Facts about a cofactored Boolean function. Bit zero/one records whether that
// output value was checked and whether an assignment producing it exists.
// These facts belong to pNew's literal namespace and expire with that graph.
type giaAreCofTruth struct{ known, possible uint8 }

func giaAreCofactorPossible(p *Gia_ManAre_t, co, lit int32, value uint8) bsat.Status {
	if lit < 2 {
		if lit == int32(value) {
			return bsat.Satisfiable
		}
		return bsat.Unsatisfiable
	}
	if Gia_ObjIsCi(Gia_ManObj(p.pNew, lit>>1)) != 0 {
		return bsat.Satisfiable
	}
	bit := uint8(1) << value
	fact := p.cofTruth[lit]
	if fact.known&bit != 0 {
		if fact.possible&bit != 0 {
			return bsat.Satisfiable
		}
		return bsat.Unsatisfiable
	}
	if p.pSat == nil {
		Gia_ManAreDeriveCexSatStart(p)
	}
	// The static source CNF is constrained by the current cube and recursive
	// cofactor choices. Every unresolved CI remains free, as in pNew.
	vec.Vec_IntClear(p.vAssumps)
	for i := int32(0); i < Gia_ManCiNum(p.pAig); i++ {
		v := Gia_ObjValue(Gia_ManCi(p.pAig, i))
		if v < 2 {
			variable := vec.Vec_IntEntry(p.vSatNumCis, int(i))
			literal := bsat.Negative(variable)
			if v != 0 {
				literal = bsat.Positive(variable)
			}
			vec.Vec_IntPush(p.vAssumps, literal)
		}
	}
	variable := vec.Vec_IntEntry(p.vSatNumCos, int(co))
	literal := bsat.Negative(variable)
	if value != 0 {
		literal = bsat.Positive(variable)
	}
	vec.Vec_IntPush(p.vAssumps, literal)
	status := p.pSat.SolveWithResourceLimits(vec.Vec_IntArray(p.vAssumps), 1000000, 0, 0, 0)
	if status == bsat.Unknown {
		p.fStopped = 1
		return status
	}
	fact.known |= bit
	if status == bsat.Satisfiable {
		fact.possible |= bit
	}
	if p.cofTruth == nil {
		p.cofTruth = make(map[int32]giaAreCofTruth)
	}
	p.cofTruth[lit] = fact
	return status
}

// Structural nonconstancy is only a candidate: reconvergent logic can be
// identically false. Check POs in source order and stop only on a real model.
func giaAreOutputAsserted(p *Gia_ManAre_t) bool {
	for i := int32(0); i < Gia_ManPoNum(p.pAig); i++ {
		lit := Gia_ObjValue(Gia_ManPo(p.pAig, i))
		switch giaAreCofactorPossible(p, i, lit, 1) {
		case bsat.Satisfiable:
			p.iOutFail = i
			return true
		case bsat.Unknown:
			return false
		}
	}
	return false
}

// A structurally nonconstant RI is a don't-care only when both polarities
// are feasible. Otherwise preserve the proven constant in the successor.
func giaAreRefineSuccessor(p *Gia_ManAre_t, next *Gia_StaAre_t) bool {
	for i := int32(0); i < p.pAig.nRegs; i++ {
		lit := Gia_ObjValue(Gia_ManRi(p.pAig, i))
		if lit < 2 {
			continue
		}
		co := Gia_ManPoNum(p.pAig) + i
		if status := giaAreCofactorPossible(p, co, lit, 1); status == bsat.Unknown {
			return false
		} else if status == bsat.Unsatisfiable {
			Gia_StaSetValue0(next, i)
			continue
		}
		if status := giaAreCofactorPossible(p, co, lit, 0); status == bsat.Unknown {
			return false
		} else if status == bsat.Unsatisfiable {
			Gia_StaSetValue1(next, i)
		}
	}
	return true
}
