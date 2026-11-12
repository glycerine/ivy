package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"time"
)

// Sbl_ManComputeCuts enumerates window cuts and finds current mapped cuts.
func Sbl_ManComputeCuts(p *Sbl_Man_t) int32 {
	started := time.Now()
	leaves := vec.Vec_IntArray(p.vLeaves)
	ands := vec.Vec_IntArray(p.vAnds)
	if len(leaves) > 128 || len(ands) > int(p.nVars) {
		panic("gia: SAT window exceeds cut capacity")
	}
	for _, v := range []*Vec_Int_t{p.vCutsStart, p.vCutsObj, p.vCutsNum} {
		vec.Vec_IntClear(v)
	}
	for _, v := range []*Vec_Wrd_t{p.vCutsI1, p.vCutsI2, p.vCutsN1, p.vCutsN2} {
		vec.Vec_WrdClear(v)
	}
	defer func() {
		for _, id := range leaves {
			Gia_ManObj(p.pGia, id).Value = ^uint32(0)
		}
		for _, id := range ands {
			Gia_ManObj(p.pGia, id).Value = ^uint32(0)
		}
		p.timeCut += time.Since(started)
	}()
	for i, id := range leaves {
		vec.Vec_IntPush(p.vCutsStart, int32(vec.Vec_WrdSize(p.vCutsI1)))
		vec.Vec_IntPush(p.vCutsObj, -1)
		vec.Vec_IntPush(p.vCutsNum, 1)
		var a, b word
		if i < 64 {
			a = word(1) << uint(i)
		} else {
			b = word(1) << uint(i-64)
		}
		vec.Vec_WrdPush(p.vCutsI1, a)
		vec.Vec_WrdPush(p.vCutsI2, b)
		vec.Vec_WrdPush(p.vCutsN1, 0)
		vec.Vec_WrdPush(p.vCutsN2, 0)
		Gia_ManObj(p.pGia, id).Value = uint32(i)
	}
	for i, id := range ands {
		obj := Gia_ManObj(p.pGia, id)
		a, b := Gia_ObjFanin0(obj).Value, Gia_ObjFanin1(obj).Value
		if Gia_ObjIsAnd(obj) == 0 || a == ^uint32(0) || b == ^uint32(0) {
			panic("gia: invalid SAT window order")
		}
		Sbl_ManComputeCutsOne(p, int32(a), int32(b), int32(i))
		obj.Value = uint32(len(leaves) + i)
	}
	vec.Vec_IntClear(p.vRootVars)
	for _, id := range vec.Vec_IntArray(p.vRoots) {
		obj := Gia_ManObj(p.pGia, id)
		if Gia_ObjIsCi(obj) != 0 {
			continue
		}
		if Gia_ObjIsLut2(p.pGia, id) == 0 || obj.Value == ^uint32(0) {
			panic("gia: invalid SAT window root")
		}
		vec.Vec_IntPush(p.vRootVars, int32(obj.Value)-int32(len(leaves)))
	}
	vec.Vec_IntClear(p.vPolar)
	vec.Vec_IntClear(p.vSolInit)
	for i, id := range ands {
		if Gia_ObjIsLut2(p.pGia, id) == 0 {
			continue
		}
		vec.Vec_IntPush(p.vPolar, int32(i))
		vec.Vec_IntPush(p.vSolInit, int32(i))
		var masks [4]word
		for _, fan := range vec.Vec_IntArray(Gia_ObjLutFanins2(p.pGia, id)) {
			value := Gia_ManObj(p.pGia, fan).Value
			if value == ^uint32(0) {
				continue
			}
			if int(value) < len(leaves) {
				masks[value/64] |= word(1) << (value % 64)
			} else {
				n := int(value) - len(leaves)
				if n >= 128 {
					panic("gia: invalid SAT window fanin")
				}
				masks[2+n/64] |= word(1) << uint(n%64)
			}
		}
		index := Sbl_ManFindCut(p, int32(len(leaves)+i), masks[0], masks[1], masks[2], masks[3])
		if index >= 0 {
			vec.Vec_IntPush(p.vPolar, p.FirstVar+index)
		}
	}
	return int32(vec.Vec_WrdSize(p.vCutsI1))
}
