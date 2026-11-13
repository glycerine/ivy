package acec

// Common structurally hashed AIG construction from acecCore.c.
import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func Acec_ManCecSetDefaultParams(p *Acec_ParCec_t) { *p = Acec_ParCec_t{nBTLimit: 1000, iOutFail: -1} }
func Acec_CommonStart(base, add *Gia_Man_t) *Gia_Man_t {
	gia.Gia_ManFillValue(add)
	gia.Gia_ManConst0(add).Value = 0
	if base == nil {
		base = gia.Gia_ManStart(gia.Gia_ManObjNum(add))
		base.CopyIdentityFrom(add)
		for i := int32(0); i < gia.Gia_ManCiNum(add); i++ {
			gia.Gia_ManCi(add, i).Value = uint32(gia.Gia_ManAppendCi(base))
		}
		gia.Gia_ManHashAlloc(base)
	} else {
		if gia.Gia_ManCiNum(base) != gia.Gia_ManCiNum(add) {
			panic("acec: different input counts")
		}
		for i := int32(0); i < gia.Gia_ManCiNum(add); i++ {
			gia.Gia_ManCi(add, i).Value = uint32(gia.Gia_Obj2Lit(base, gia.Gia_ManCi(base, i)))
		}
	}
	for id := int32(1); id < gia.Gia_ManObjNum(add); id++ {
		o := gia.Gia_ManObj(add, id)
		if gia.Gia_ObjIsAnd(o) != 0 {
			a := gia.Gia_ObjFanin0Copy(o)
			if gia.Gia_ObjIsBuf(o) != 0 {
				o.Value = uint32(a)
			} else if gia.Gia_ObjIsMux(add, o) != 0 {
				o.Value = uint32(gia.Gia_ManHashMux(base, gia.Gia_ObjFanin2Copy(add, o), gia.Gia_ObjFanin1Copy(o), a))
			} else if gia.Gia_ObjIsXor(o) != 0 {
				o.Value = uint32(gia.Gia_ManHashXor(base, a, gia.Gia_ObjFanin1Copy(o)))
			} else {
				o.Value = uint32(gia.Gia_ManHashAnd(base, a, gia.Gia_ObjFanin1Copy(o)))
			}
		}
	}
	return base
}
func Acec_CommonFinish(base *Gia_Man_t) {
	gia.Gia_ManCreateRefs(base)
	for id := int32(1); id < gia.Gia_ManObjNum(base); id++ {
		o := gia.Gia_ManObj(base, id)
		if gia.Gia_ObjIsAnd(o) != 0 && gia.Gia_ObjRefNumId(base, id) == 0 {
			gia.Gia_ManAppendCo(base, id<<1)
		}
	}
}
func Acec_CountRemap(add, base *Gia_Man_t) *Vec_Int_t {
	mapping := vec.Vec_IntStartFull(int(gia.Gia_ManObjNum(add)))
	gia.Gia_ManSetPhase(add)
	vec.Vec_IntWriteEntry(mapping, 0, 0)
	for id := int32(1); id < gia.Gia_ManObjNum(add); id++ {
		o := gia.Gia_ManObj(add, id)
		if gia.Gia_ObjIsCand(o) == 0 {
			continue
		}
		repr := gia.Gia_ManObj(base, int32(o.Value>>1))
		vec.Vec_IntWriteEntry(mapping, int(id), int32(repr.Value>>1)<<1|gia.Gia_ObjPhase(o))
	}
	return mapping
}
