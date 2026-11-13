package bmc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

// Bmc_EcoMiter builds onset and offset mismatch outputs with a free patch root,
// followed by the requested old-circuit fanin signals.
func Bmc_EcoMiter(gold, old *gia.Gia_Man_t, fans *vec.Vec_Int_t) *gia.Gia_Man_t {
	if gold == nil || old == nil || fans == nil || gia.Gia_ManCiNum(gold) != gia.Gia_ManCiNum(old) || gia.Gia_ManCoNum(old) != gia.Gia_ManCoNum(gold)+1 || gia.Gia_ManPoNum(old) == 0 {
		return nil
	}
	last := gia.Gia_ManPo(old, gia.Gia_ManPoNum(old)-1)
	root := gia.Gia_ObjFaninId0(last, gia.Gia_ObjId(old, last))
	if gia.Gia_ObjIsAnd(gia.Gia_ManObj(old, root)) == 0 {
		return nil
	}
	q := gia.Gia_ManStart(max(1, 3*gia.Gia_ManObjNum(gold)))
	gia.Gia_ManHashStart(q)
	inputs := make([]int32, gia.Gia_ManCiNum(gold))
	for i := range inputs {
		inputs[i] = gia.Gia_ManAppendCi(q)
	}
	patch := gia.Gia_ManAppendCi(q)
	copyGraph := func(src *gia.Gia_Man_t, replacement int32, useRoot bool) []int32 {
		mapping := make([]int32, gia.Gia_ManObjNum(src))
		for i, lit := range inputs {
			mapping[gia.Gia_ObjId(src, gia.Gia_ManCi(src, int32(i)))] = lit
		}
		read := func(lit int32) int32 { return mapping[lit>>1] ^ (lit & 1) }
		for id := int32(1); id < gia.Gia_ManObjNum(src); id++ {
			if useRoot && id == root {
				mapping[id] = replacement
				continue
			}
			a, b, c, kind := src.Gate(id)
			switch kind {
			case 2:
				mapping[id] = gia.Gia_ManHashAnd(q, read(a), read(b))
			case 3:
				mapping[id] = gia.Gia_ManHashXor(q, read(a), read(b))
			case 4:
				mapping[id] = gia.Gia_ManHashMux(q, read(a), read(b), read(c))
			case 5, 6:
				mapping[id] = read(a)
			}
		}
		return mapping
	}
	gm, on := copyGraph(gold, 0, false), copyGraph(old, patch, true)
	for i := int32(0); i < gia.Gia_ManCoNum(gold); i++ {
		x := gm[gia.Gia_ObjId(gold, gia.Gia_ManCo(gold, i))]
		y := on[gia.Gia_ObjId(old, gia.Gia_ManCo(old, i))]
		gia.Gia_ManAppendCo(q, gia.Gia_ManHashXor(q, x, y))
	}
	off := copyGraph(old, patch^1, true)
	miter := int32(0)
	for i := int32(0); i < gia.Gia_ManCoNum(gold); i++ {
		x := gm[gia.Gia_ObjId(gold, gia.Gia_ManCo(gold, i))]
		y := off[gia.Gia_ObjId(old, gia.Gia_ManCo(old, i))]
		miter = gia.Gia_ManHashOr(q, miter, gia.Gia_ManHashXor(q, x, y))
	}
	gia.Gia_ManAppendCo(q, miter)
	for _, id := range vec.Vec_IntArray(fans) {
		if id < 0 || id >= gia.Gia_ManObjNum(old) {
			return nil
		}
		gia.Gia_ManAppendCo(q, off[id])
	}
	gia.Gia_ManHashStop(q)
	return gia.Gia_ManCleanup(q)
}
