package if_

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
)

// Gia_ManNodeIfToGia rebuilds the local cone above the selected cut leaves.
// Choice alternatives are searched when the representative depends on nodes
// outside the cut. A failed search returns -1.
func Gia_ManNodeIfToGia(dst *Gia_Man_t, p *If_Man_t, root *If_Obj_t, leaves *Vec_Int_t, hash int32) int32 {
	cut := If_ObjCutBest(root)
	assert(cut.nLeaves > 1 && Vec_IntSize(leaves) == int(cut.nLeaves), "cut leaves match GIA literals")
	copies := make(map[*If_Obj_t]int32)
	for i := int32(0); i < int32(cut.nLeaves); i++ {
		copies[If_CutLeaf(p, cut, i)] = Vec_IntEntry(leaves, int(i))
	}
	var visit func(*If_Obj_t) int32
	visit = func(obj *If_Obj_t) int32 {
		if obj == nil {
			return -1
		}
		if lit, ok := copies[obj]; ok {
			return lit
		}
		copies[obj] = -1
		if If_ObjIsCi(obj) != 0 || If_ObjIsConst1(obj) != 0 {
			return -1
		}
		for alternative := obj; alternative != nil; alternative = alternative.pEquiv {
			a := visit(alternative.pFanin0)
			if a < 0 {
				continue
			}
			b := visit(alternative.pFanin1)
			if b < 0 {
				continue
			}
			a ^= int32(alternative.fCompl0)
			b ^= int32(alternative.fCompl1)
			var lit int32
			if hash != 0 {
				lit = Gia_ManHashAnd(dst, a, b)
			} else {
				lit = gia.Gia_ManAppendAnd2(dst, a, b)
			}
			if alternative.fPhase != obj.fPhase {
				lit ^= 1
			}
			copies[obj] = lit
			return lit
		}
		return -1
	}
	result := visit(root)
	if result < 0 {
		fmt.Println("Gia_ManNodeIfToGia(): Computing local AIG has failed.")
	}
	return result
}
