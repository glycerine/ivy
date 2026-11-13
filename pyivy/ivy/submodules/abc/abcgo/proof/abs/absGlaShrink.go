package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

// Ga2_ManShrinkAbs rolls back added abstraction objects and SAT mappings.
func Ga2_ManShrinkAbs(p *Ga2_Man_t, nAbs, nValues, nSatVars int32) {
	if nAbs <= 0 || nValues <= 0 || nSatVars <= 0 || int(nAbs) > vec.Vec_IntSize(p.vAbs) || int(nValues) > vec.Vec_IntSize(p.vValues) {
		panic("abs: invalid abstraction rollback")
	}
	for i := 1; i < vec.Vec_IntSize(p.vAbs); i++ {
		obj := gia.Gia_ManObj(p.pGia, vec.Vec_IntEntry(p.vAbs, i))
		if Ga2_ObjCnf0(p, obj) == nil || Ga2_ObjCnf1(p, obj) == nil {
			panic("abs: missing abstract cut covers")
		}
		if i >= int(nAbs) {
			id := Ga2_ObjId(p, obj)
			p.vCnfs[2*id] = nil
			p.vCnfs[2*id+1] = nil
		}
	}
	vec.Vec_IntShrink(p.vAbs, int(nAbs))
	for i := 0; i < vec.Vec_IntSize(p.vValues); i++ {
		obj := gia.Gia_ManObj(p.pGia, vec.Vec_IntEntry(p.vValues, i))
		if Ga2_ObjId(p, obj) < 0 {
			panic("abs: rollback value has no ID")
		}
		if i < int(nValues) {
			continue
		}
		Ga2_ObjSetId(p, obj, -1)
	}
	vec.Vec_IntShrink(p.vValues, int(nValues))
	p.vCnfs = p.vCnfs[:2*int(nValues)]
	if nValues == 1 {
		nValues = 0
	}
	for _, frame := range p.vId2Lit {
		vec.Vec_IntShrink(frame, int(nValues))
		for k := int(p.LimAbs); k < vec.Vec_IntSize(frame); k++ {
			if vec.Vec_IntEntry(frame, k) >= 2*nSatVars {
				vec.Vec_IntWriteEntry(frame, k, -1)
			}
		}
	}
	p.nSatVars = nSatVars
}
