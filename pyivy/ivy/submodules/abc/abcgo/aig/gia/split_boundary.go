package gia

import "github.com/glycerine/abcgo/misc/vec"

func Spl_ManWinFindLeavesRoots(p *Spl_Man_t) {
	g := p.pGia
	vec.Vec_IntClear(p.vLeaves)
	for _, id := range vec.Vec_IntArray(p.vAnds) {
		if vec.Vec_BitEntry(p.vMarksAnd, int(id)) == 0 {
			panic("gia: unmarked window AND")
		}
		obj := Gia_ManObj(g, id)
		for _, fan := range []int32{Gia_ObjFaninId0(obj, id), Gia_ObjFaninId1(obj, id)} {
			if vec.Vec_BitEntry(p.vMarksAnd, int(fan)) != 0 {
				continue
			}
			if Gia_ObjIsLut2(g, fan) == 0 && vec.Vec_BitEntry(p.vMarksCIO, int(fan)) == 0 {
				panic("gia: invalid window boundary")
			}
			vec.Vec_BitWriteEntry(p.vMarksAnd, int(fan), 1)
			vec.Vec_IntPush(p.vLeaves, fan)
		}
	}
	for _, id := range vec.Vec_IntArray(p.vLeaves) {
		vec.Vec_BitWriteEntry(p.vMarksAnd, int(id), 0)
	}
	vec.Vec_IntClear(p.vRoots)
	for _, id := range vec.Vec_IntArray(p.vNodes) {
		for _, fan := range vec.Vec_IntArray(Gia_ObjLutFanins2(g, id)) {
			Gia_ObjLutRefDecId(g, fan)
		}
	}
	for _, id := range vec.Vec_IntArray(p.vAnds) {
		if Gia_ObjLutRefNumId(g, id) != 0 {
			vec.Vec_IntPush(p.vRoots, id)
		}
	}
	for _, id := range vec.Vec_IntArray(p.vNodes) {
		for _, fan := range vec.Vec_IntArray(Gia_ObjLutFanins2(g, id)) {
			Gia_ObjLutRefIncId(g, fan)
		}
	}
}
func Spl_ManLutFanouts_rec(g *Gia_Man_t, id int32, fanouts *Vec_Int_t, marked, cio *Vec_Bit_t) {
	if vec.Vec_BitEntry(marked, int(id)) != 0 || vec.Vec_BitEntry(cio, int(id)) != 0 {
		return
	}
	if Gia_ObjIsLut2(g, id) != 0 {
		vec.Vec_BitWriteEntry(cio, int(id), 1)
		vec.Vec_IntPush(fanouts, id)
		return
	}
	for i := int32(0); i < Gia_ObjFanoutNumId(g, id); i++ {
		Spl_ManLutFanouts_rec(g, Gia_ObjFanoutId(g, id, i), fanouts, marked, cio)
	}
}
func Spl_ManLutFanouts(g *Gia_Man_t, id int32, fanouts *Vec_Int_t, marked, cio *Vec_Bit_t) int32 {
	if Gia_ObjIsLut2(g, id) == 0 {
		panic("gia: expected LUT fanout source")
	}
	vec.Vec_IntClear(fanouts)
	for i := int32(0); i < Gia_ObjFanoutNumId(g, id); i++ {
		Spl_ManLutFanouts_rec(g, Gia_ObjFanoutId(g, id, i), fanouts, marked, cio)
	}
	for _, fan := range vec.Vec_IntArray(fanouts) {
		vec.Vec_BitWriteEntry(cio, int(fan), 0)
	}
	return int32(vec.Vec_IntSize(fanouts))
}
func Spl_ManCountMarkedFanins(g *Gia_Man_t, id int32, marked *Vec_Bit_t) int32 {
	var count int32
	for _, fan := range vec.Vec_IntArray(Gia_ObjLutFanins2(g, id)) {
		if vec.Vec_BitEntry(marked, int(fan)) != 0 {
			count++
		}
	}
	return count
}
func Spl_ManFindGoodCand(p *Spl_Man_t) int32 {
	for _, id := range vec.Vec_IntArray(p.vInputs) {
		vec.Vec_BitWriteEntry(p.vMarksIn, int(id), 1)
	}
	var result int32
	best := int32(-1)
	for _, id := range vec.Vec_IntArray(p.vCands) {
		count := Spl_ManCountMarkedFanins(p.pGia, id, p.vMarksIn)
		if count > best {
			result, best = id, count
		}
	}
	for _, id := range vec.Vec_IntArray(p.vInputs) {
		vec.Vec_BitWriteEntry(p.vMarksIn, int(id), 0)
	}
	return result
}
