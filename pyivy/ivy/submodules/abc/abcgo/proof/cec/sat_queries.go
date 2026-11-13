package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/proof/cec/core"
)

func Cec_ManSatSolverRecycle(p *Cec_ManSat_t) { p.restart() }
func (p *Cec_ManSat_t) query(a int32, b *int32) int32 {
	p.values = nil
	if p.vCex != nil {
		vec.Vec_IntClear(p.vCex)
	}
	if p.proof == nil {
		return -1
	}
	if p.Error = p.proof.SetConflictLimit(int64(p.pPars.nBTLimit)); p.Error != nil {
		return -1
	}
	if b == nil {
		p.result, p.Error = p.proof.CheckLiteral(a)
	} else {
		p.result, p.Error = p.proof.CheckConjunction(a, *b)
	}
	if p.Error != nil {
		return -1
	}
	if p.result.Status == core.Different {
		p.values, p.Error = p.proof.Model().Values(p.result.Counterexample.Inputs)
		if p.Error != nil {
			return -1
		}
	}
	return int32(p.result.Status)
}
func Cec_ManSatCheckNode(p *Cec_ManSat_t, obj *Gia_Obj_t) int32 {
	lit := cecObjectLiteral(p, obj)
	if lit == 1 {
		panic("cec: constant-one node query")
	}
	return p.query(lit, nil)
}
func Cec_ManSatCheckNodeTwo(p *Cec_ManSat_t, a, b *Gia_Obj_t) int32 {
	lit := cecObjectLiteral(p, b)
	return p.query(cecObjectLiteral(p, a), &lit)
}

func cecObjectLiteral(p *Cec_ManSat_t, obj *Gia_Obj_t) int32 {
	return gia.Gia_ObjId(p.pAig, gia.Gia_Regular(obj))<<1 | gia.Gia_IsComplement(obj)
}
