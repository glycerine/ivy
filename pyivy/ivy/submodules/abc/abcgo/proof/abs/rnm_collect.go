package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

// Rnm_ManCollect_rec collects the cone in topological order until it reaches
// a previously visited object. Values store one-based indices after the map.
func Rnm_ManCollect_rec(p *gia.Gia_Man_t, obj *gia.Gia_Obj_t, objects *vec.Vec_Int_t, offset int32) {
	if gia.Gia_ObjIsTravIdCurrent(p, obj) != 0 {
		return
	}
	gia.Gia_ObjSetTravIdCurrent(p, obj)
	switch {
	case gia.Gia_ObjIsCo(obj) != 0:
		Rnm_ManCollect_rec(p, gia.Gia_ObjFanin0(obj), objects, offset)
	case gia.Gia_ObjIsAnd(obj) != 0:
		Rnm_ManCollect_rec(p, gia.Gia_ObjFanin0(obj), objects, offset)
		Rnm_ManCollect_rec(p, gia.Gia_ObjFanin1(obj), objects, offset)
	case gia.Gia_ObjIsRo(p, obj) == 0:
		panic("refinement cone reached an unmapped primary input")
	}
	obj.Value = uint32(vec.Vec_IntSize(objects)) + uint32(offset)
	vec.Vec_IntPush(objects, gia.Gia_ObjId(p, obj))
}

func Rnm_ManCollect(p *Rnm_Man_t) {
	gia.Gia_ManIncrementTravId(p.pGia)
	gia.Gia_ObjSetTravIdCurrent(p.pGia, gia.Gia_ManConst0(p.pGia))
	gia.Gia_ManConst0(p.pGia).Value = 0
	for i, id := range vec.Vec_IntArray(p.vMap) {
		obj := gia.Gia_ManObj(p.pGia, id)
		if gia.Gia_ObjIsCi(obj) == 0 && gia.Gia_ObjIsAnd(obj) == 0 {
			panic("invalid refinement map object")
		}
		gia.Gia_ObjSetTravIdCurrent(p.pGia, obj)
		obj.Value = uint32(i + 1)
	}
	vec.Vec_IntClear(p.vObjs)
	offset := int32(1 + vec.Vec_IntSize(p.vMap))
	Rnm_ManCollect_rec(p.pGia, gia.Gia_ManPo(p.pGia, 0), p.vObjs, offset)
	// The traversal may append more register outputs while it runs.
	for i := 0; i < vec.Vec_IntSize(p.vObjs); i++ {
		obj := gia.Gia_ManObj(p.pGia, vec.Vec_IntEntry(p.vObjs, i))
		if gia.Gia_ObjIsRo(p.pGia, obj) != 0 {
			Rnm_ManCollect_rec(p.pGia, gia.Gia_ObjRoToRi(p.pGia, obj), p.vObjs, offset)
		}
	}
	last := gia.Gia_ManObj(p.pGia, vec.Vec_IntEntryLast(p.vObjs))
	if gia.Gia_ObjIsCo(last) == 0 || int(last.Value) != vec.Vec_IntSize(p.vMap)+vec.Vec_IntSize(p.vObjs) {
		panic("invalid refinement cone order")
	}
}

func Rnm_ManCleanValues(p *Rnm_Man_t) {
	for _, id := range vec.Vec_IntArray(p.vMap) {
		gia.Gia_ManObj(p.pGia, id).Value = 0
	}
	for _, id := range vec.Vec_IntArray(p.vObjs) {
		gia.Gia_ManObj(p.pGia, id).Value = 0
	}
}
