package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"time"
)

// Sbl_ManWindow uses the entire graph as the SAT window.
func Sbl_ManWindow(p *Sbl_Man_t) {
	vec.Vec_IntClear(p.vLeaves)
	vec.Vec_IntClear(p.vAnds)
	vec.Vec_IntClear(p.vRoots)
	for i := int32(0); i < Gia_ManCiNum(p.pGia); i++ {
		vec.Vec_IntPush(p.vLeaves, Gia_ObjId(p.pGia, Gia_ManCi(p.pGia, i)))
	}
	for id := int32(1); id < p.pGia.nObjs; id++ {
		if Gia_ObjIsAnd(Gia_ManObj(p.pGia, id)) != 0 {
			vec.Vec_IntPush(p.vAnds, id)
		}
	}
	for i := int32(0); i < Gia_ManCoNum(p.pGia); i++ {
		vec.Vec_IntPush(p.vRoots, Gia_ObjFaninId0p(p.pGia, Gia_ManCo(p.pGia, i)))
	}
}

func Sbl_ManWindow2(p *Sbl_Man_t, pivot int32) int32 {
	started := time.Now()
	var roots, nodes, leaves, ands *Vec_Int_t
	count := Gia_ManComputeOneWin(p.pGia, pivot, &roots, &nodes, &leaves, &ands)
	p.timeWin += time.Since(started)
	if count == 0 {
		return 0
	}
	for i, dst := range []*Vec_Int_t{p.vRoots, p.vNodes, p.vLeaves, p.vAnds} {
		src := []*Vec_Int_t{roots, nodes, leaves, ands}[i]
		vec.Vec_IntClear(dst)
		for _, id := range vec.Vec_IntArray(src) {
			vec.Vec_IntPush(dst, id)
		}
	}
	return count
}
