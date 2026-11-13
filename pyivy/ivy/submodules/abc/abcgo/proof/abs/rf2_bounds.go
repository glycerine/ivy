package abs

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func Rf2_ManBounds(p *Rf2_Man_t) {
	g := p.pGia
	zero := gia.Gia_ManConst0(g)
	gia.Gia_ObjSetMark0(zero, 0)
	Rf2_ObjStart(p, zero, int32(vec.Vec_IntSize(p.vMap)+vec.Vec_IntSize(p.vObjs)))
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
			Rf2_ObjStart(p, obj, int32(i))
		}
		for i, id := range vec.Vec_IntArray(p.vObjs) {
			obj := gia.Gia_ManObj(g, id)
			gia.Gia_ObjSetMark0(obj, 0)
			Rf2_ObjClear(p, obj)
			if gia.Gia_ObjIsRo(g, obj) != 0 {
				if frame == 0 {
					Rf2_ObjStart(p, obj, int32(vec.Vec_IntSize(p.vMap)+i))
					continue
				}
				prev := gia.Gia_ObjRoToRi(g, obj)
				gia.Gia_ObjSetMark0(obj, gia.Gia_ObjMark0(prev))
				Rf2_ObjCopy(p, obj, prev)
				continue
			}
			left := gia.Gia_ObjFanin0(obj)
			a := gia.Gia_ObjMark0(left) ^ gia.Gia_ObjFaninC0(obj)
			if gia.Gia_ObjIsCo(obj) != 0 {
				gia.Gia_ObjSetMark0(obj, a)
				Rf2_ObjCopy(p, obj, left)
				continue
			}
			if gia.Gia_ObjIsAnd(obj) == 0 {
				panic("invalid alternate refinement cone")
			}
			right := gia.Gia_ObjFanin1(obj)
			b := gia.Gia_ObjMark0(right) ^ gia.Gia_ObjFaninC1(obj)
			gia.Gia_ObjSetMark0(obj, a&b)
			switch {
			case a&b == 1:
				Rf2_ObjDeriveAnd(p, obj, 1)
			case a == 0 && b == 0:
				Rf2_ObjDeriveAnd(p, obj, 0)
			case a == 0:
				Rf2_ObjCopy(p, obj, left)
			default:
				Rf2_ObjCopy(p, obj, right)
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
	fmt.Println("Bounds:")
	Rf2_ObjPrint(p, out)
}
