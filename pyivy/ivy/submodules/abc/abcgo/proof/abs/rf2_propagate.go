package abs

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// Rf2_ManPropagate propagates inclusion-minimal group masks through the cone.
func Rf2_ManPropagate(p *Rf2_Man_t, limit int32) *vec.Vec_Int_t {
	g := p.pGia
	zero := gia.Gia_ManConst0(g)
	gia.Gia_ObjSetMark0(zero, 0)
	vec.Vec_IntFill(Rf2_ObjVec(p, zero), 1, 0)
	bit := p.pCex.NRegs
	for frame := int32(0); frame <= p.pCex.IFrame; frame++ {
		for i, id := range vec.Vec_IntArray(p.vMap) {
			obj := gia.Gia_ManObj(g, id)
			if gia.Gia_ObjIsCi(obj) == 0 && gia.Gia_ObjIsAnd(obj) == 0 {
				panic("invalid alternate refinement map")
			}
			value := int32(0)
			if util.Abc_InfoHasBit(p.pCex.PData, bit+int32(i)) {
				value = 1
			}
			gia.Gia_ObjSetMark0(obj, value)
			vec.Vec_IntFill(Rf2_ObjVec(p, obj), 1, 0)
		}
		for i := 0; i < vec.Vec_VecSize(p.vGrp2Ppi); i++ {
			for _, entry := range vec.Vec_IntArray(vec.Vec_VecEntryInt(p.vGrp2Ppi, i)) {
				if i >= 31 {
					panic("too many alternate refinement groups")
				}
				obj := gia.Gia_ManObj(g, vec.Vec_IntEntry(p.vMap, int(entry)))
				v := Rf2_ObjVec(p, obj)
				if vec.Vec_IntSize(v) != 1 {
					panic("invalid alternate refinement mask")
				}
				vec.Vec_IntAddToEntry(v, 0, int32(1)<<i)
			}
		}
		for _, id := range vec.Vec_IntArray(p.vObjs) {
			obj := gia.Gia_ManObj(g, id)
			gia.Gia_ObjSetMark0(obj, 0)
			dst := Rf2_ObjVec(p, obj)
			vec.Vec_IntClear(dst)
			if gia.Gia_ObjIsRo(g, obj) != 0 {
				if frame == 0 {
					vec.Vec_IntPush(dst, 0)
					continue
				}
				prev := gia.Gia_ObjRoToRi(g, obj)
				gia.Gia_ObjSetMark0(obj, gia.Gia_ObjMark0(prev))
				vec.Vec_IntAppend(dst, Rf2_ObjVec(p, prev))
				continue
			}
			left := gia.Gia_ObjFanin0(obj)
			a := gia.Gia_ObjMark0(left) ^ gia.Gia_ObjFaninC0(obj)
			if gia.Gia_ObjIsCo(obj) != 0 {
				gia.Gia_ObjSetMark0(obj, a)
				vec.Vec_IntAppend(dst, Rf2_ObjVec(p, left))
				continue
			}
			if gia.Gia_ObjIsAnd(obj) == 0 {
				panic("invalid alternate refinement cone")
			}
			right := gia.Gia_ObjFanin1(obj)
			b := gia.Gia_ObjMark0(right) ^ gia.Gia_ObjFaninC1(obj)
			gia.Gia_ObjSetMark0(obj, a&b)
			l, r := Rf2_ObjVec(p, left), Rf2_ObjVec(p, right)
			switch {
			case a&b == 1:
				for _, x := range vec.Vec_IntArray(l) {
					for _, y := range vec.Vec_IntArray(r) {
						vec.Vec_IntPush(dst, x|y)
					}
				}
				Rf2_ManProcessVector(dst, limit)
			case a == 0 && b == 0:
				vec.Vec_IntAppend(dst, l)
				vec.Vec_IntAppend(dst, r)
				Rf2_ManProcessVector(dst, limit)
			case a == 0:
				vec.Vec_IntAppend(dst, l)
			default:
				vec.Vec_IntAppend(dst, r)
			}
		}
		bit += p.pCex.NPis
	}
	if bit != p.pCex.NBits {
		panic("counterexample bit count mismatch")
	}
	out := gia.Gia_ManPo(g, 0)
	if gia.Gia_ObjMark0(out) != 1 {
		fmt.Println("Output value is incorrect.")
	}
	return Rf2_ObjVec(p, out)
}
