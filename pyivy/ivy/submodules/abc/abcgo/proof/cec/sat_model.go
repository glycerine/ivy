package cec

import "github.com/glycerine/abcgo/aig/gia"

// Cec_ObjSatVarValue returns the phase-adjusted SAT-variable value, as in
// cecSolve.c. Cec_ObjLogicalValue returns the regular object's logical value.
func Cec_ObjSatVarValue(p *Cec_ManSat_t, obj *Gia_Obj_t) int32 {
	value := Cec_ObjLogicalValue(p, obj)
	if p.pPars != nil && p.proof != nil && p.pPars.fPolarFlip != 0 {
		id := gia.Gia_ObjId(p.pAig, gia.Gia_Regular(obj))
		_, _, _, kind := p.pAig.Gate(id)
		if kind == 2 || kind == 3 || kind == 4 {
			phase, err := p.proof.Model().Value(make([]bool, p.proof.Model().InputCount()), id<<1)
			if err != nil {
				panic(err)
			}
			if phase {
				value ^= 1
			}
		}
	}
	return value
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
