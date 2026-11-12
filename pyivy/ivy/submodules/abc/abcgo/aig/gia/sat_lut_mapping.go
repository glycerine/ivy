package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"math/bits"
)

func sblCutFanins(p *Sbl_Man_t, cut int32, result *Vec_Int_t) {
	vec.Vec_IntClear(result)
	masks := [4]word{vec.Vec_WrdEntry(p.vCutsI1, int(cut)), vec.Vec_WrdEntry(p.vCutsI2, int(cut)), vec.Vec_WrdEntry(p.vCutsN1, int(cut)), vec.Vec_WrdEntry(p.vCutsN2, int(cut))}
	for part, mask := range masks {
		source := p.vLeaves
		if part >= 2 {
			source = p.vAnds
		}
		for mask != 0 {
			bit := bits.TrailingZeros64(mask)
			vec.Vec_IntPush(result, vec.Vec_IntEntry(source, (part%2)*64+bit))
			mask &= mask - 1
		}
	}
}
func Sbl_ManGetCurrentMapping(p *Sbl_Man_t) {
	vec.Vec_WecClear(p.vWindow)
	vec.Vec_WecInit(p.vWindow, vec.Vec_IntSize(p.vAnds))
	if vec.Vec_IntSize(p.vSolCur) == 0 {
		panic("gia: empty SAT window mapping")
	}
	for _, cut := range vec.Vec_IntArray(p.vSolCur) {
		obj := vec.Vec_IntEntry(p.vCutsObj, int(cut))
		if obj < 0 {
			panic("gia: invalid SAT window cut")
		}
		sblCutFanins(p, cut, vec.Vec_WecEntry(p.vWindow, int(obj)))
	}
}
func Sbl_ManUpdateMapping(p *Sbl_Man_t) {
	if vec.Vec_IntSize(p.vSolBest) >= vec.Vec_IntSize(p.vSolInit) {
		panic("gia: SAT window mapping did not improve")
	}
	for _, id := range vec.Vec_IntArray(p.vAnds) {
		fanins := vec.Vec_WecEntry(p.pGia.vMapping2, int(id))
		for _, fan := range vec.Vec_IntArray(fanins) {
			Gia_ObjLutRefDecId(p.pGia, fan)
		}
		vec.Vec_IntClear(fanins)
	}
	for _, cut := range vec.Vec_IntArray(p.vSolBest) {
		obj := vec.Vec_IntEntry(p.vCutsObj, int(cut))
		id := vec.Vec_IntEntry(p.vAnds, int(obj))
		fanins := vec.Vec_WecEntry(p.pGia.vMapping2, int(id))
		sblCutFanins(p, cut, fanins)
		for _, fan := range vec.Vec_IntArray(fanins) {
			Gia_ObjLutRefIncId(p.pGia, fan)
		}
	}
}
