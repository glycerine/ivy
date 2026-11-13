package abs

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// Rf2_ManSensitize propagates counterexample values and refinement priorities.
func Rf2_ManSensitize(p *Rf2_Man_t) int32 {
	bit := p.pCex.NRegs
	for frame := int32(0); frame <= p.pCex.IFrame; frame++ {
		for i, id := range vec.Vec_IntArray(p.vMap) {
			obj := gia.Gia_ManObj(p.pGia, id)
			if gia.Gia_ObjIsCi(obj) == 0 && gia.Gia_ObjIsAnd(obj) == 0 {
				panic("invalid refinement map object")
			}
			r := Rf2_ManObj(p, obj, frame)
			if util.Abc_InfoHasBit(p.pCex.PData, bit+int32(i)) {
				r.Value = 1
			} else {
				r.Value = 0
			}
			if gia.Gia_ObjIsPi(p.pGia, obj) == 0 {
				if obj.Value == 0 {
					panic("unindexed pseudo-input")
				}
				r.Prio, r.fPPi = obj.Value, 1
			}
		}
		for _, id := range vec.Vec_IntArray(p.vObjs) {
			obj := gia.Gia_ManObj(p.pGia, id)
			r := Rf2_ManObj(p, obj, frame)
			if r.fPPi != 0 {
				panic("pseudo-input in refinement cone")
			}
			if gia.Gia_ObjIsRo(p.pGia, obj) != 0 {
				if frame > 0 {
					previous := Rf2_ManObj(p, gia.Gia_ObjRoToRi(p.pGia, obj), frame-1)
					r.Value, r.Prio = previous.Value, previous.Prio
				}
				continue
			}
			left := Rf2_ManObj(p, gia.Gia_ObjFanin0(obj), frame)
			a := left.Value ^ uint32(gia.Gia_ObjFaninC0(obj))
			if gia.Gia_ObjIsCo(obj) != 0 {
				r.Value, r.Prio = a, left.Prio
				continue
			}
			if gia.Gia_ObjIsAnd(obj) == 0 {
				panic("invalid refinement cone object")
			}
			right := Rf2_ManObj(p, gia.Gia_ObjFanin1(obj), frame)
			b := right.Value ^ uint32(gia.Gia_ObjFaninC1(obj))
			r.Value = a & b
			switch {
			case r.Value == 1:
				r.Prio = max(left.Prio, right.Prio)
			case a == 0 && b == 0:
				r.Prio = min(left.Prio, right.Prio)
			case a == 0:
				r.Prio = left.Prio
			default:
				r.Prio = right.Prio
			}
		}
		bit += p.pCex.NPis
	}
	if bit != p.pCex.NBits {
		panic("counterexample bit count mismatch")
	}
	out := Rf2_ManObj(p, gia.Gia_ManPo(p.pGia, 0), p.pCex.IFrame)
	if out.Value != 1 {
		fmt.Println("Output value is incorrect.")
	}
	return int32(out.Prio)
}

// Rf2_ManVerifyUsingTerSim uses the same ternary simulation as Rnm refinement.
func Rf2_ManVerifyUsingTerSim(g *gia.Gia_Man_t, cex *util.Abc_Cex_t, mapping, objects, selected *vec.Vec_Int_t) {
	Rnm_ManVerifyUsingTerSim(g, cex, mapping, objects, selected)
}
