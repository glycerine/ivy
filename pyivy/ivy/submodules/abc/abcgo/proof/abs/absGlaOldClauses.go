package abs

import (
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

// Gla_ManAddClauses asserts the object's mapped CNF in one timeframe.
func Gla_ManAddClauses(p *Gla_Man_t, id, frame int32, lits *vec.Vec_Int_t) {
	obj := Gla_ManObj(p, id)
	switch {
	case obj.fConst != 0:
		v := Gla_ManGetVar(p, id, frame)
		bsat.Sat_solver2_add_const(p.pSat, v, 1, 0, id)
	case obj.fRo != 0:
		if obj.nFanins != 1 {
			panic("old GLA register must have one fanin")
		}
		v := Gla_ManGetVar(p, id, frame)
		if frame == 0 {
			bsat.Sat_solver2_add_const(p.pSat, v, 1, 0, id)
			return
		}
		fanin := Gla_ManGetVar(p, obj.Fanins[0], frame-1)
		bsat.Sat_solver2_add_buffer(p.pSat, v, fanin, int32(obj.fCompl0), 0, id)
	case obj.fAnd != 0:
		if p.pCnf == nil || int(obj.iGiaObj) >= len(p.pCnf.Obj2Count) {
			panic("missing old GLA cut CNF")
		}
		start, count := p.pCnf.Obj2Clause[obj.iGiaObj], p.pCnf.Obj2Count[obj.iGiaObj]
		if start < 0 || count < 0 || int64(start)+int64(count) > int64(len(p.pCnf.Clauses)) {
			panic("invalid old GLA cut CNF range")
		}
		for _, clause := range p.pCnf.Clauses[start : start+count] {
			vec.Vec_IntClear(lits)
			for _, literal := range clause {
				v := Gla_ManGetVar(p, literal>>1, frame)
				vec.Vec_IntPush(lits, v*2+literal&1)
			}
			bsat.Sat_solver2_addclause(p.pSat, vec.Vec_IntArray(lits), id)
		}
	default:
		// assert( 0 ) in absGlaOld.c, compiled out in release builds: an
		// abstracted PI (an output driven by a PI) gets no clauses
	}
}
