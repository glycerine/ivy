package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func Ga2_ManAbsTranslate_rec(p *gia.Gia_Man_t, obj *gia.Gia_Obj_t, classes *vec.Vec_Int_t, first int32) {
	if gia.Gia_ObjPhase(obj) != 0 && first == 0 {
		return
	}
	if gia.Gia_ObjIsAnd(obj) == 0 {
		panic("abs: expected AND in cut")
	}
	Ga2_ManAbsTranslate_rec(p, gia.Gia_ObjFanin0(obj), classes, 0)
	Ga2_ManAbsTranslate_rec(p, gia.Gia_ObjFanin1(obj), classes, 0)
	vec.Vec_IntWriteEntry(classes, int(gia.Gia_ObjId(p, obj)), 1)
}
func Ga2_ManAbsTranslate(p *Ga2_Man_t) *vec.Vec_Int_t {
	classes := vec.Vec_IntStart(int(gia.Gia_ManObjNum(p.pGia)))
	vec.Vec_IntWriteEntry(classes, 0, 1)
	for i := 0; i < vec.Vec_IntSize(p.vAbs); i++ {
		obj := gia.Gia_ManObj(p.pGia, vec.Vec_IntEntry(p.vAbs, i))
		if gia.Gia_ObjIsAnd(obj) != 0 {
			Ga2_ManAbsTranslate_rec(p.pGia, obj, classes, 1)
		} else if gia.Gia_ObjIsRo(p.pGia, obj) != 0 {
			vec.Vec_IntWriteEntry(classes, int(gia.Gia_ObjId(p.pGia, obj)), 1)
		} else if gia.Gia_ObjIsConst0(obj) == 0 {
			panic("abs: invalid abstraction object")
		}
	}
	return classes
}
