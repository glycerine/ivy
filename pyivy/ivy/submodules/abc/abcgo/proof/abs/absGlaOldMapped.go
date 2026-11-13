package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func Gia_ManDupMapped_rec(p *gia.Gia_Man_t, obj *gia.Gia_Obj_t, dst *gia.Gia_Man_t) {
	if gia.Gia_ObjIsTravIdCurrent(p, obj) != 0 {
		return
	}
	gia.Gia_ObjSetTravIdCurrent(p, obj)
	if gia.Gia_ObjIsAnd(obj) == 0 {
		panic("mapped cone reaches an unmarked non-AND")
	}
	Gia_ManDupMapped_rec(p, gia.Gia_ObjFanin0(obj), dst)
	Gia_ManDupMapped_rec(p, gia.Gia_ObjFanin1(obj), dst)
	obj.Value = uint32(gia.Gia_ManAppendAnd(dst, gia.Gia_ObjFanin0Copy(obj), gia.Gia_ObjFanin1Copy(obj)))
	vec.Vec_IntPush(gia.Gia_ManLutConfigs(dst), gia.Gia_ObjId(p, obj))
}

// Gia_ManDupMapped duplicates the logic of every mapped LUT separately.
func Gia_ManDupMapped(p *gia.Gia_Man_t, mapping *vec.Vec_Int_t) *gia.Gia_Man_t {
	dst := gia.Gia_ManStart(gia.Gia_ManObjNum(p))
	gia.Gia_ManCopyNameSpec(dst, p)
	gia.Gia_ManFillValue(p)
	copies := make([]int32, gia.Gia_ManObjNum(p))
	for i := range copies {
		copies[i] = -1
	}
	copies[0] = 0
	configs := vec.Vec_IntAlloc(int(gia.Gia_ManObjNum(p)) * 4 / 3)
	vec.Vec_IntPush(configs, 0)
	gia.Gia_ManSetLutConfigs(dst, configs)
	for i := int32(1); i < gia.Gia_ManObjNum(p); i++ {
		obj := gia.Gia_ManObj(p, i)
		switch {
		case gia.Gia_ObjIsAnd(obj) != 0:
			offset := vec.Vec_IntEntry(mapping, int(i))
			if offset == 0 {
				continue
			}
			gia.Gia_ManIncrementTravId(p)
			for k := int32(1); k <= 4; k++ {
				id := vec.Vec_IntEntry(mapping, int(offset+k))
				if id == -1 {
					continue
				}
				fanin := gia.Gia_ManObj(p, id)
				gia.Gia_ObjSetTravIdCurrent(p, fanin)
				if copies[id] < 0 {
					panic("mapped LUT fanin was not created")
				}
				fanin.Value = uint32(copies[id])
			}
			if gia.Gia_ObjIsTravIdCurrent(p, obj) != 0 || obj.Value != ^uint32(0) {
				panic("invalid mapped LUT root")
			}
			Gia_ManDupMapped_rec(p, obj, dst)
			copies[i] = int32(obj.Value)
		case gia.Gia_ObjIsCi(obj) != 0:
			copies[i] = gia.Gia_ManAppendCi(dst)
			vec.Vec_IntPush(configs, i)
		case gia.Gia_ObjIsCo(obj) != 0:
			id := gia.Gia_ObjId(p, gia.Gia_ObjFanin0(obj))
			if copies[id] < 0 {
				panic("mapped output fanin was not created")
			}
			gia.Gia_ObjFanin0(obj).Value = uint32(copies[id])
			copies[i] = gia.Gia_ManAppendCo(dst, gia.Gia_ObjFanin0Copy(obj))
			vec.Vec_IntPush(configs, i)
		}
	}
	if vec.Vec_IntSize(configs) != int(gia.Gia_ManObjNum(dst)) {
		panic("invalid LUT provenance size")
	}
	gia.Gia_ManSetRegNum(dst, gia.Gia_ManRegNum(p))
	for i, value := range copies {
		gia.Gia_ManObj(p, int32(i)).Value = uint32(value)
	}
	return dst
}
