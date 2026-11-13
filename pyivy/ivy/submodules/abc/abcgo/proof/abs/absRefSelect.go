package abs

// Go port of absRefSelect.c by Alan Mishchenko, UC Berkeley (2005).

import (
	"fmt"
	"sync/atomic"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

// Rnm_ManPrintSelected prints which non-primary inputs were selected.
func Rnm_ManPrintSelected(p *Rnm_Man_t, selected *vec.Vec_Int_t) {
	count := 0
	for _, id := range vec.Vec_IntArray(p.vMap) {
		obj := gia.Gia_ManObj(p.pGia, id)
		if gia.Gia_ObjIsPi(p.pGia, obj) != 0 {
			fmt.Print("-")
		} else if vec.Vec_IntFind(selected, id) >= 0 {
			fmt.Print("1")
			count++
		} else {
			fmt.Print("0")
		}
	}
	fmt.Printf(" %3d\n", count)
}

// Ga2_StructAnalize prints the relationship between selected cut leaves.
// As in the C implementation it uses the GIA objects' two temporary marks.
func Ga2_StructAnalize(p *gia.Gia_Man_t, front, internal, selected *vec.Vec_Int_t) {
	for i := int32(0); i < gia.Gia_ManObjNum(p); i++ {
		gia.Gia_ObjTerSimSetC(gia.Gia_ManObj(p, i))
	}
	for _, id := range vec.Vec_IntArray(front) {
		gia.Gia_ObjTerSimSet0(gia.Gia_ManObj(p, id))
	}
	markOne := func(obj *gia.Gia_Obj_t) {
		if gia.Gia_ObjTerSimGet0(obj) != 0 || gia.Gia_ObjTerSimGetX(obj) != 0 {
			gia.Gia_ObjTerSimSetX(obj)
		} else {
			gia.Gia_ObjTerSimSet1(obj)
		}
	}
	for _, id := range vec.Vec_IntArray(internal) {
		markOne(gia.Gia_ManObj(p, id))
	}
	for _, id := range vec.Vec_IntArray(selected) {
		markOne(gia.Gia_ManObj(p, id))
	}
	for i, id := range vec.Vec_IntArray(selected) {
		obj := gia.Gia_ManObj(p, id)
		fmt.Printf("Selected PPI %3d : %6d \n", i+1, id)
		for _, fanin := range vec.Vec_IntArray(Ga2_ObjLeaves(p, obj)) {
			f := gia.Gia_ManObj(p, fanin)
			kind := "free variable"
			switch {
			case gia.Gia_ObjTerSimGetX(f) != 0:
				kind = "selected PPI"
			case gia.Gia_ObjTerSimGet0(f) != 0:
				kind = "frontier (original PI or PPI)"
			case gia.Gia_ObjTerSimGet1(f) != 0:
				kind = "abstracted node"
			}
			fmt.Printf("    %6d %s\n", fanin, kind)
		}
	}
}

// Rnm_ManFilterSelected retains flop outputs and PPIs with justified or shared
// cut leaves. Its temporary fanin counts are restored before returning.
func Rnm_ManFilterSelected(p *Rnm_Man_t, old *vec.Vec_Int_t) *vec.Vec_Int_t {
	vec.Vec_IntClear(p.vFanins)
	for _, id := range vec.Vec_IntArray(old) {
		obj := gia.Gia_ManObj(p.pGia, id)
		for _, fanin := range vec.Vec_IntArray(Ga2_ObjLeaves(p.pGia, obj)) {
			f := gia.Gia_ManObj(p.pGia, fanin)
			if Rnm_ObjAddToCount(p, f) == 0 {
				vec.Vec_IntPush(p.vFanins, fanin)
			}
		}
	}
	result := vec.Vec_IntAlloc(100)
	for _, id := range vec.Vec_IntArray(old) {
		obj := gia.Gia_ManObj(p.pGia, id)
		if gia.Gia_ObjIsRo(p.pGia, obj) != 0 {
			vec.Vec_IntPush(result, id)
			continue
		}
		for _, fanin := range vec.Vec_IntArray(Ga2_ObjLeaves(p.pGia, obj)) {
			f := gia.Gia_ManObj(p.pGia, fanin)
			if Rnm_ObjIsJust(p, f) != 0 || Rnm_ObjCount(p, f) > 1 {
				vec.Vec_IntPush(result, id)
				break
			}
		}
	}
	if vec.Vec_IntUniqify(result) != 0 {
		panic("refinement selection contains duplicate PPIs")
	}
	for _, id := range vec.Vec_IntArray(p.vFanins) {
		Rnm_ObjSetCount(p, gia.Gia_ManObj(p.pGia, id), 0)
	}
	return result
}

var refinementSelectionCounter atomic.Uint64

// Rnm_ManFilterSelectedNew periodically keeps the complete original set, as
// the active branch of the original refinement heuristic does.
func Rnm_ManFilterSelectedNew(p *Rnm_Man_t, old *vec.Vec_Int_t) *vec.Vec_Int_t {
	if refinementSelectionCounter.Add(1)%9 == 0 {
		return vec.Vec_IntDup(old)
	}
	return Rnm_ManFilterSelected(p, old)
}
