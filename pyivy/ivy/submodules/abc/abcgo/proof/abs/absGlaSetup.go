package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

// Ga2_ManSetupNode assigns an abstraction ID and, when requested, its cut CNF.
func Ga2_ManSetupNode(p *Ga2_Man_t, obj *gia.Gia_Obj_t, abstract int32) {
	if gia.Gia_ObjPhase(obj) == 0 || len(p.vCnfs) != 2*vec.Vec_IntSize(p.vValues) {
		panic("abs: inconsistent GLA node setup")
	}
	if Ga2_ObjId(p, obj) == -1 {
		Ga2_ObjSetId(p, obj, int32(vec.Vec_IntSize(p.vValues)))
		vec.Vec_IntPush(p.vValues, gia.Gia_ObjId(p.pGia, obj))
		p.vCnfs = append(p.vCnfs, nil, nil)
	}
	if Ga2_ObjCnf0(p, obj) != nil {
		panic("abs: node already abstracted")
	}
	if abstract == 0 {
		return
	}
	if gia.Gia_ObjIsAnd(obj) == 0 && gia.Gia_ObjIsRo(p.pGia, obj) == 0 {
		panic("abs: invalid abstracted node")
	}
	vec.Vec_IntPush(p.vAbs, gia.Gia_ObjId(p.pGia, obj))
	n := Ga2_ObjLeaveNum(p.pGia, obj)
	truth := Ga2_ObjTruth(p.pGia, obj)
	id := Ga2_ObjId(p, obj)
	p.vCnfs[2*id] = Ga2_ManCnfCompute(truth, n, p.vIsopMem)
	p.vCnfs[2*id+1] = Ga2_ManCnfCompute(^truth, n, p.vIsopMem)
}
