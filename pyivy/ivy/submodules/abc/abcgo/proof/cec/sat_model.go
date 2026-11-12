package cec

import "github.com/glycerine/abcgo/aig/gia"

// Cec_ObjSatVarValue returns the regular object's value in the most recent
// independently simulated SAT witness. No witness exists after an UNSAT query.
func Cec_ObjSatVarValue(p *Cec_ManSat_t, obj *Gia_Obj_t) int32 {
	id := gia.Gia_ObjId(p.pAig, gia.Gia_Regular(obj))
	if id < 0 || int(id) >= len(p.values) {
		panic("cec: SAT variable value requested without a valid witness")
	}
	if p.values[id] {
		return 1
	}
	return 0
}
