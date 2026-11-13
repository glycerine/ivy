package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func Rf2_ManCollect_rec(g *gia.Gia_Man_t, obj *gia.Gia_Obj_t, objects *vec.Vec_Int_t) {
	if gia.Gia_ObjIsTravIdCurrent(g, obj) != 0 {
		return
	}
	gia.Gia_ObjSetTravIdCurrent(g, obj)
	switch {
	case gia.Gia_ObjIsCo(obj) != 0:
		Rf2_ManCollect_rec(g, gia.Gia_ObjFanin0(obj), objects)
	case gia.Gia_ObjIsAnd(obj) != 0:
		Rf2_ManCollect_rec(g, gia.Gia_ObjFanin0(obj), objects)
		Rf2_ManCollect_rec(g, gia.Gia_ObjFanin1(obj), objects)
	case gia.Gia_ObjIsRo(g, obj) == 0:
		panic("refinement cone reached an unmapped primary input")
	}
	vec.Vec_IntPush(objects, gia.Gia_ObjId(g, obj))
}

func Rf2_ManCollect(p *Rf2_Man_t) {
	gia.Gia_ManIncrementTravId(p.pGia)
	gia.Gia_ObjSetTravIdCurrent(p.pGia, gia.Gia_ManConst0(p.pGia))
	for _, id := range vec.Vec_IntArray(p.vMap) {
		obj := gia.Gia_ManObj(p.pGia, id)
		if gia.Gia_ObjIsCi(obj) == 0 && gia.Gia_ObjIsAnd(obj) == 0 {
			panic("invalid refinement map object")
		}
		gia.Gia_ObjSetTravIdCurrent(p.pGia, obj)
	}
	vec.Vec_IntClear(p.vObjs)
	Rf2_ManCollect_rec(p.pGia, gia.Gia_ManPo(p.pGia, 0), p.vObjs)
	for i := 0; i < vec.Vec_IntSize(p.vObjs); i++ {
		obj := gia.Gia_ManObj(p.pGia, vec.Vec_IntEntry(p.vObjs, i))
		if gia.Gia_ObjIsRo(p.pGia, obj) != 0 {
			Rf2_ManCollect_rec(p.pGia, gia.Gia_ObjRoToRi(p.pGia, obj), p.vObjs)
		}
	}
	if gia.Gia_ObjIsCo(gia.Gia_ManObj(p.pGia, vec.Vec_IntEntryLast(p.vObjs))) == 0 {
		panic("invalid refinement cone order")
	}
}
