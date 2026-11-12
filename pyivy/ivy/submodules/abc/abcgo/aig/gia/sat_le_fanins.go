package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
)

func Sle_ManPrintCut(cut []int32) {
	fmt.Print("{")
	for _, id := range Sle_CutLeaves(cut) {
		fmt.Printf(" %d", id)
	}
	fmt.Println(" }")
}
func Sle_ManPrintCuts(_ *Gia_Man_t, cuts *Vec_Int_t, id int32) {
	fmt.Printf("Obj %3d\n", id)
	for _, cut := range sleList(cuts, id, false) {
		Sle_ManPrintCut(cut)
	}
	fmt.Println()
}
func Sle_ManPrintCutsAll(p *Gia_Man_t, cuts *Vec_Int_t) {
	for id := int32(1); id < p.nObjs; id++ {
		if Gia_ObjIsAnd(Gia_ManObj(p, id)) != 0 {
			Sle_ManPrintCuts(p, cuts, id)
		}
	}
}
func Sle_ManComputeCutsTest(p *Gia_Man_t) { Sle_ManComputeCuts(p, 4, 1) }
func Sle_ManInternalNodeMask(p *Gia_Man_t) *Vec_Bit_t {
	mask := vec.Vec_BitStart(int(p.nObjs))
	for id := int32(1); id < p.nObjs; id++ {
		if Gia_ObjIsAnd(Gia_ManObj(p, id)) != 0 {
			vec.Vec_BitWriteEntry(mask, int(id), 1)
		}
	}
	return mask
}
func Sle_ManCutHasPisOnly(cut []int32, mask *Vec_Bit_t) int32 {
	for _, id := range Sle_CutLeaves(cut) {
		if vec.Vec_BitEntry(mask, int(id)) != 0 {
			return 0
		}
	}
	return 1
}
func Sle_ManCollectCutFaninsOne(_ *Gia_Man_t, id int32, cuts *Vec_Int_t, mask *Vec_Bit_t, fanins *Vec_Int_t, marked *Vec_Bit_t) {
	for _, cut := range sleList(cuts, id, false) {
		if Sle_CutSize(cut) <= 1 {
			panic("gia: invalid non-unit LUT cut")
		}
		for _, leaf := range Sle_CutLeaves(cut) {
			if vec.Vec_BitEntry(mask, int(leaf)) != 0 && vec.Vec_BitEntry(marked, int(leaf)) == 0 {
				vec.Vec_BitWriteEntry(marked, int(leaf), 1)
				vec.Vec_IntPush(fanins, leaf)
			}
		}
	}
	for _, leaf := range vec.Vec_IntArray(fanins) {
		vec.Vec_BitWriteEntry(marked, int(leaf), 0)
	}
}
func Sle_ManCollectCutFanins(p *Gia_Man_t, cuts *Vec_Int_t, mask *Vec_Bit_t) *Vec_Wec_t {
	marked := vec.Vec_BitStart(int(p.nObjs))
	result := vec.Vec_WecStart(int(p.nObjs))
	for id := int32(1); id < p.nObjs; id++ {
		if Gia_ObjIsAnd(Gia_ManObj(p, id)) != 0 {
			Sle_ManCollectCutFaninsOne(p, id, cuts, mask, vec.Vec_WecEntry(result, int(id)), marked)
		}
	}
	return result
}
