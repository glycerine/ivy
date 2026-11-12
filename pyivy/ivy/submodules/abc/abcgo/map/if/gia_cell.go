package if_

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

// Gia_ManFromIfLogicFindCell is disabled in ABC by an unconditional assert.
// Its only caller in the original reconstruction routine is also commented out.
func Gia_ManFromIfLogicFindCell(p *If_Man_t, dst, temp *Gia_Man_t, cut *If_Cut_t, cell *Ifn_Ntk_t, lutMax int32, leaves, lits, cover, mapping, records, configs *Vec_Int_t) int32 {
	panic("if: programmable-cell reconstruction is disabled in ABC")
}

// giaManFromIfLogicFindCell implements the body retained after the assertion
// in ABC, for use when programmable-cell reconstruction is enabled upstream.
func giaManFromIfLogicFindCell(p *If_Man_t, dst, temp *Gia_Man_t, cut *If_Cut_t, cell *Ifn_Ntk_t, lutMax int32, leaves, cover, mapping, records *Vec_Int_t) int32 {
	if Vec_IntSize(leaves) <= int(lutMax) {
		if If_CutTruthW(p, cut) == nil {
			panic("if: missing cut truth")
		}
		return Gia_ManFromIfLogicCreateLut(dst, p.puTempW[:p.nTruth6Words[cut.nLeaves]], leaves, cover, mapping, records)
	}
	config := If_DsdManGetFuncConfig(p.pIfDsdMan, If_CutDsdLit(p, cut))
	if len(config) == 0 {
		panic("if: missing cell configuration")
	}
	lit := If_ManSatDeriveGiaFromBits(temp, cell, config[1:], leaves, cover)
	gia.Gia_ManCleanCopyArray(temp)
	gia.Gia_ObjSetCopyArray(temp, 0, 0)
	for i, leaf := range Vec_IntArray(leaves) {
		gia.Gia_ObjSetCopyArray(temp, gia.Gia_ManCiIdToId(temp, int32(i)), leaf)
	}
	gia.Gia_ManIncrementTravId(temp)
	id := lit >> 1
	gia.Gia_ManCollectAnds(temp, []int32{id}, 1, cover, nil)
	vec.Vec_IntPrint(cover)
	for _, node := range Vec_IntArray(cover) {
		gia.Gia_ObjPrint(temp, gia.Gia_ManObj(temp, node))
	}
	for _, node := range Vec_IntArray(cover) {
		obj := gia.Gia_ManObj(temp, node)
		copy := gia.Gia_ManAppendAnd(dst, gia.Gia_ObjFanin0CopyArray(temp, obj), gia.Gia_ObjFanin1CopyArray(temp, obj))
		gia.Gia_ObjSetCopyArray(temp, node, copy)
	}
	return gia.Gia_ObjCopyArray(temp, id) ^ (lit & 1)
}
