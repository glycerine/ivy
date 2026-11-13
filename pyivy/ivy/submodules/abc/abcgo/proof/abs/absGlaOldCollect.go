package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func Gla_ManCollectInternal_rec(g *gia.Gia_Man_t, object *gia.Gia_Obj_t, nodes *vec.Vec_Int_t) {
	if gia.Gia_ObjIsTravIdCurrent(g, object) != 0 {
		return
	}
	gia.Gia_ObjSetTravIdCurrent(g, object)
	if gia.Gia_ObjIsAnd(object) == 0 {
		panic("GLA internal cone reached an unmarked input")
	}
	Gla_ManCollectInternal_rec(g, gia.Gia_ObjFanin0(object), nodes)
	Gla_ManCollectInternal_rec(g, gia.Gia_ObjFanin1(object), nodes)
	vec.Vec_IntPush(nodes, gia.Gia_ObjId(g, object))
}

// Gla_ManCollect collects the abstraction's input boundaries, latch outputs,
// and internal nodes in topological order for ternary refinement.
func Gla_ManCollect(p *Gla_Man_t, pis, ppis, cos, nodes *vec.Vec_Int_t) {
	g := p.pGia
	vec.Vec_IntPush(cos, gia.Gia_ObjId(g, gia.Gia_ManPo(g, 0)))
	for _, id := range vec.Vec_IntArray(p.vAbs) {
		object := Gla_ManObj(p, id)
		if object.fConst == 0 && object.fRo == 0 && object.fAnd == 0 {
			panic("invalid GLA abstraction object")
		}
		if object.fRo != 0 {
			ro := gia.Gia_ManObj(g, object.iGiaObj)
			vec.Vec_IntPush(cos, gia.Gia_ObjId(g, gia.Gia_ObjRoToRi(g, ro)))
		}
		for k := uint32(0); k < object.nFanins; k++ {
			fanin := Gla_ManObj(p, object.Fanins[k])
			if fanin.fAbs != 0 {
				continue
			}
			if fanin.fPi != 0 {
				vec.Vec_IntPush(pis, fanin.iGiaObj)
			} else {
				vec.Vec_IntPush(ppis, fanin.iGiaObj)
			}
		}
	}
	vec.Vec_IntUniqify(pis)
	vec.Vec_IntUniqify(ppis)
	vec.Vec_IntSort(cos, 0)
	gia.Gia_ManIncrementTravId(g)
	gia.Gia_ObjSetTravIdCurrent(g, gia.Gia_ManConst0(g))
	for _, id := range vec.Vec_IntArray(pis) {
		gia.Gia_ObjSetTravIdCurrent(g, gia.Gia_ManObj(g, id))
	}
	for _, id := range vec.Vec_IntArray(ppis) {
		gia.Gia_ObjSetTravIdCurrent(g, gia.Gia_ManObj(g, id))
	}
	for i, id := range vec.Vec_IntArray(cos) {
		if i == 0 {
			continue
		}
		ro := gia.Gia_ObjRiToRo(g, gia.Gia_ManObj(g, id))
		gia.Gia_ObjSetTravIdCurrent(g, ro)
		vec.Vec_IntPush(nodes, gia.Gia_ObjId(g, ro))
	}
	for _, id := range vec.Vec_IntArray(cos) {
		Gla_ManCollectInternal_rec(g, gia.Gia_ObjFanin0(gia.Gia_ManObj(g, id)), nodes)
	}
}
