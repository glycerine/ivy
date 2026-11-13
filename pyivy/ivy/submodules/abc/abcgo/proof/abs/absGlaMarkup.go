package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

// Ga2_ManMarkup partitions a GIA into marked cuts and records their truths.
func Ga2_ManMarkup(p *gia.Gia_Man_t, n, simple int32) int32 {
	if n < 2 || n > 5 {
		panic("abs: invalid cut size")
	}
	leaves := vec.Vec_IntAlloc(100)
	num := gia.Gia_ManObjNum(p)
	if simple != 0 {
		for id := int32(0); id < num; id++ {
			o := gia.Gia_ManObj(p, id)
			if gia.Gia_ObjIsCo(o) == 0 {
				gia.Gia_ObjSetPhaseValue(o, 1)
			} else {
				gia.Gia_ObjSetPhaseValue(o, 0)
			}
		}
	} else {
		for id := int32(0); id < num; id++ {
			o := gia.Gia_ManObj(p, id)
			o.Value = 0
			if gia.Gia_ObjIsAnd(o) == 0 {
				continue
			}
			a, b := gia.Gia_ObjFanin0(o), gia.Gia_ObjFanin1(o)
			a.Value++
			b.Value++
			if gia.Gia_ObjIsMuxType(o) != 0 {
				gia.Gia_ObjFanin0(a).Value++
				gia.Gia_ObjFanin1(a).Value++
				gia.Gia_ObjFanin0(b).Value++
				gia.Gia_ObjFanin1(b).Value++
			}
		}
		for id := int32(0); id < num; id++ {
			o := gia.Gia_ManObj(p, id)
			gia.Gia_ObjSetPhaseValue(o, 0)
			if gia.Gia_ObjIsAnd(o) != 0 {
				if o.Value > 1 {
					gia.Gia_ObjSetPhaseValue(o, 1)
				}
			} else if gia.Gia_ObjIsCo(o) != 0 {
				gia.Gia_ObjSetPhaseValue(gia.Gia_ObjFanin0(o), 1)
			} else {
				gia.Gia_ObjSetPhaseValue(o, 1)
			}
		}
		for id := int32(0); id < num; id++ {
			o := gia.Gia_ManObj(p, id)
			if gia.Gia_ObjIsAnd(o) == 0 || gia.Gia_ObjPhase(o) == 0 {
				continue
			}
			vec.Vec_IntClear(leaves)
			Ga2_ManCollectLeaves_rec(p, o, leaves, 1)
			if vec.Vec_IntSize(leaves) > int(n) {
				Ga2_ManBreakTree_rec(p, o, 1, n)
			}
		}
	}
	mapping := vec.Vec_IntStart(int(num))
	gia.Gia_ManSetMapping(p, mapping)
	for i := int32(0); i < gia.Gia_ManRegNum(p); i++ {
		o := gia.Gia_ManRo(p, i)
		ri := gia.Gia_ObjRoToRi(p, o)
		if gia.Gia_ObjPhase(o) == 0 || gia.Gia_ObjPhase(gia.Gia_ObjFanin0(ri)) == 0 {
			panic("abs: unmarked register cut")
		}
		vec.Vec_IntWriteEntry(mapping, int(gia.Gia_ObjId(p, o)), int32(vec.Vec_IntSize(mapping)))
		vec.Vec_IntPush(mapping, 1)
		vec.Vec_IntPush(mapping, gia.Gia_ObjFaninId0p(p, ri))
		truth := uint32(0xaaaaaaaa)
		if gia.Gia_ObjFaninC0(ri) != 0 {
			truth = 0x55555555
		}
		vec.Vec_IntPush(mapping, int32(truth))
		vec.Vec_IntPush(mapping, -1)
	}
	count := gia.Gia_ManRegNum(p)
	for id := int32(0); id < num; id++ {
		o := gia.Gia_ManObj(p, id)
		if gia.Gia_ObjIsAnd(o) == 0 || gia.Gia_ObjPhase(o) == 0 {
			continue
		}
		vec.Vec_IntClear(leaves)
		Ga2_ManCollectLeaves_rec(p, o, leaves, 1)
		if vec.Vec_IntSize(leaves) > int(n) {
			panic("abs: cut exceeds limit")
		}
		vec.Vec_IntWriteEntry(mapping, int(id), int32(vec.Vec_IntSize(mapping)))
		vec.Vec_IntPush(mapping, int32(vec.Vec_IntSize(leaves)))
		for j := 0; j < vec.Vec_IntSize(leaves); j++ {
			leaf := vec.Vec_IntEntry(leaves, j)
			if gia.Gia_ObjPhase(gia.Gia_ManObj(p, leaf)) == 0 {
				panic("abs: unmarked cut leaf")
			}
			vec.Vec_IntPush(mapping, leaf)
		}
		vec.Vec_IntPush(mapping, int32(Ga2_ManComputeTruth(p, o, leaves)))
		vec.Vec_IntPush(mapping, -1)
		count++
	}
	gia.Gia_ManCleanValue(p)
	return count
}
