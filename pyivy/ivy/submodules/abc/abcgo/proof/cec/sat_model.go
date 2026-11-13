package cec

import "github.com/glycerine/abcgo/aig/gia"

// Cec_ObjSatVarValue returns the phase-adjusted SAT-variable value, as in
// cecSolve.c. Cec_ObjLogicalValue returns the regular object's logical value.
func Cec_ObjSatVarValue(p *Cec_ManSat_t, obj *Gia_Obj_t) int32 {
	id := gia.Gia_ObjId(p.pAig, gia.Gia_Regular(obj))
	if p.source == nil {
		value := Cec_ObjLogicalValue(p, obj)
		if p.pPars != nil && p.pPars.fPolarFlip != 0 && gia.Gia_ObjPhase(gia.Gia_Regular(obj)) != 0 {
			value ^= 1
		}
		return value
	}
	if id < 0 || int(id) >= len(p.source.ids) {
		panic("cec: invalid SAT variable")
	}
	value, ok := p.source.sat.Value(p.source.ids[id])
	if !ok {
		panic("cec: SAT variable value requested without a valid witness")
	}
	if value {
		return 1
	}
	return 0
}

func Cec_ObjLogicalValue(p *Cec_ManSat_t, obj *Gia_Obj_t) int32 {
	id := gia.Gia_ObjId(p.pAig, gia.Gia_Regular(obj))
	if id < 0 || int(id) >= len(p.values) {
		panic("cec: SAT variable value requested without a valid witness")
	}
	if p.values[id] {
		return 1
	}
	return 0
}
