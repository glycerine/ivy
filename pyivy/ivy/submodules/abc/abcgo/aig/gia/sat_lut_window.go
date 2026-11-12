package gia

import "github.com/glycerine/abcgo/misc/vec"

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
