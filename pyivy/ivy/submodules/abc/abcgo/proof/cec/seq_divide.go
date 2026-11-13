package cec

import "github.com/glycerine/abcgo/aig/gia"

// Gia_ManSeqEquivDivide proves correspondence independently on each half,
// then combines their classes using source-object indices.
func Gia_ManSeqEquivDivide(p *gia.Gia_Man_t, pars *Cec_ParCor_t) {
	giaManSeqEquivDivide(p, pars)
}
func giaManSeqEquivDivide(p *gia.Gia_Man_t, pars *Cec_ParCor_t) bool {
	parts := make([]*gia.Gia_Man_t, 2)
	defer func() {
		for _, part := range parts {
			if part != nil {
				gia.Gia_ManStop(part)
			}
		}
	}()
	for side := int32(0); side < 2; side++ {
		part := gia.Gia_ManDupHalfSeq(p, side)
		parts[side] = part
		if Cec_ManLSCorrespondenceClasses(part, pars) < 0 || cecCorrDeadlineExpired(pars) {
			return false
		}
		for id := int32(0); id < gia.Gia_ManObjNum(p); id++ {
			mapped := gia.Gia_ObjValue(gia.Gia_ManObj(p, id))
			if mapped >= 0 {
				gia.Gia_ObjSetValue(gia.Gia_ManObj(part, mapped>>1), id)
			}
		}
	}
	gia.Gia_ManSeqEquivMerge(p, parts)
	return true
}
func Gia_ManScorrDivideTest(p *gia.Gia_Man_t, pars *Cec_ParCor_t) *gia.Gia_Man_t {
	if cecCorrDeadlineExpired(pars) {
		return nil
	}
	p.ClearRepresentatives()
	if !giaManSeqEquivDivide(p, pars) {
		return nil
	}
	reduced := Gia_ManCorrReduce(p)
	result := gia.Gia_ManSeqCleanup(reduced)
	gia.Gia_ManStop(reduced)
	return result
}
