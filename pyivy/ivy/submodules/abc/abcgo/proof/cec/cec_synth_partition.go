package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

// Gia_ManRegCreatePart extracts the cones of the selected register inputs.
// Other register outputs become primary inputs; selected outputs remain registers.
func Gia_ManRegCreatePart(p *gia.Gia_Man_t, part *vec.Vec_Int_t, countPIs, countRegs *int32, mapBack *[]int32) *gia.Gia_Man_t {
	// The source advances here, then CollectNodesCis advances once more.
	gia.Gia_ManIncrementTravId(p)
	selected := vec.Vec_IntArray(part)
	roots := make([]int32, len(selected))
	for i, reg := range selected {
		roots[i] = gia.Gia_ObjId(p, gia.Gia_ManCo(p, gia.Gia_ManPoNum(p)+reg))
	}
	nodes := vec.Vec_IntArray(gia.Gia_ManCollectNodesCis(p, roots, int32(len(roots))))
	for _, reg := range selected {
		gia.Gia_ObjSetTravIdPrevious(p, gia.Gia_ManCi(p, gia.Gia_ManPiNum(p)+reg))
	}
	var pis, regs int32
	for i := int32(0); i < gia.Gia_ManPiNum(p); i++ {
		pis += gia.Gia_ObjIsTravIdCurrent(p, gia.Gia_ManCi(p, i))
	}
	for i := int32(0); i < gia.Gia_ManRegNum(p); i++ {
		regs += gia.Gia_ObjIsTravIdCurrent(p, gia.Gia_ManCi(p, gia.Gia_ManPiNum(p)+i))
	}
	if countPIs != nil {
		*countPIs = pis
	}
	if countRegs != nil {
		*countRegs = regs
	}
	gia.Gia_ManFillValue(p)
	gia.Gia_ObjSetValue(gia.Gia_ManConst0(p), 0)
	out := gia.Gia_ManStart(int32(len(nodes)))
	for i := int32(0); i < gia.Gia_ManCiNum(p); i++ {
		obj := gia.Gia_ManCi(p, i)
		if gia.Gia_ObjIsTravIdCurrent(p, obj) != 0 {
			gia.Gia_ObjSetValue(obj, gia.Gia_ManAppendCi(out))
		}
	}
	for _, reg := range selected {
		obj := gia.Gia_ManCi(p, gia.Gia_ManPiNum(p)+reg)
		gia.Gia_ObjSetValue(obj, gia.Gia_ManAppendCi(out))
		gia.Gia_ManAppendCo(out, gia.Gia_ObjValue(obj))
		gia.Gia_ObjSetTravIdCurrent(p, obj)
	}
	for _, id := range nodes {
		obj := gia.Gia_ManObj(p, id)
		if gia.Gia_ObjIsAnd(obj) != 0 {
			gia.Gia_ObjSetValue(obj, gia.Gia_ManAppendAnd(out, gia.Gia_ObjFanin0Copy(obj), gia.Gia_ObjFanin1Copy(obj)))
		}
	}
	for _, reg := range selected {
		obj := gia.Gia_ManCo(p, gia.Gia_ManPoNum(p)+reg)
		gia.Gia_ManAppendCo(out, gia.Gia_ObjFanin0Copy(obj))
	}
	gia.Gia_ManSetRegNum(out, int32(len(selected)))
	if mapBack != nil {
		mapping := make([]int32, gia.Gia_ManObjNum(out))
		for i := range mapping {
			mapping[i] = -1
		}
		mapping[0] = 0
		for _, id := range nodes {
			obj := gia.Gia_ManObj(p, id)
			mapping[gia.Gia_ObjValue(obj)>>1] = id
		}
		for _, reg := range selected {
			obj := gia.Gia_ManCi(p, gia.Gia_ManPiNum(p)+reg)
			mapping[gia.Gia_ObjValue(obj)>>1] = gia.Gia_ObjId(p, obj)
		}
		*mapBack = mapping
	}
	return out
}
