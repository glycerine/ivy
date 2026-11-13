package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

// Ga2_ManAbsDerive collects marked flops and ANDs already in the abstraction.
func Ga2_ManAbsDerive(p *gia.Gia_Man_t) *vec.Vec_Int_t {
	result := vec.Vec_IntAlloc(1000)
	classes := gia.Gia_ManGateClasses(p)
	for i := int32(0); i < gia.Gia_ManRegNum(p); i++ {
		o := gia.Gia_ManRo(p, i)
		id := gia.Gia_ObjId(p, o)
		if gia.Gia_ObjPhase(o) != 0 && vec.Vec_IntEntry(classes, int(id)) != 0 {
			vec.Vec_IntPush(result, id)
		}
	}
	for id := int32(0); id < gia.Gia_ManObjNum(p); id++ {
		o := gia.Gia_ManObj(p, id)
		if gia.Gia_ObjIsAnd(o) != 0 && gia.Gia_ObjPhase(o) != 0 && vec.Vec_IntEntry(classes, int(id)) != 0 {
			vec.Vec_IntPush(result, id)
		}
	}
	return result
}
