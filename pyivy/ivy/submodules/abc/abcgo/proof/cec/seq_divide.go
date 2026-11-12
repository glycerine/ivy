package cec

import "github.com/glycerine/abcgo/aig/gia"

// Gia_ManSeqEquivDivide proves correspondence independently on each half,
// then combines their classes using source-object indices.
func Gia_ManSeqEquivDivide(p *gia.Gia_Man_t, pars *Cec_ParCor_t) {
	parts := make([]*gia.Gia_Man_t, 2)
	for side := int32(0); side < 2; side++ {
		part := gia.Gia_ManDupHalfSeq(p, side)
		parts[side] = part
		Cec_ManLSCorrespondenceClasses(part, pars)
		for id := int32(0); id < gia.Gia_ManObjNum(p); id++ {
			mapped := gia.Gia_ObjValue(gia.Gia_ManObj(p, id))
			if mapped >= 0 {
				gia.Gia_ObjSetValue(gia.Gia_ManObj(part, mapped>>1), id)
			}
		}
	}
	gia.Gia_ManSeqEquivMerge(p, parts)
	for _, part := range parts {
		gia.Gia_ManStop(part)
	}
}
func Gia_ManScorrDivideTest(p *gia.Gia_Man_t, pars *Cec_ParCor_t) *gia.Gia_Man_t {
	p.ClearRepresentatives()
	Gia_ManSeqEquivDivide(p, pars)
	reduced := Gia_ManCorrReduce(p)
	result := gia.Gia_ManSeqCleanup(reduced)
	gia.Gia_ManStop(reduced)
	return result
}
