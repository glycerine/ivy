package acec

// Booth-pattern recognition ported from src/proof/acec/acecBo.c.
import (
	"fmt"

	"github.com/glycerine/abcgo/aig/gia"
)

// Acec_DetectBoothXorMux checks whether mux and xor share the same data inputs.
// ins receives the two data input IDs followed by the mux control ID.
func Acec_DetectBoothXorMux(p *gia.Gia_Man_t, mux, xor *gia.Gia_Obj_t, ins []int32) int32 {
	if gia.Gia_ObjIsMuxType(mux) == 0 || gia.Gia_ObjIsMuxType(xor) == 0 {
		return 0
	}
	var fan0, fan1, dat0, dat1 *gia.Gia_Obj_t
	if gia.Gia_ObjRecognizeExor(xor, &fan0, &fan1) == 0 {
		return 0
	}
	fan0, fan1 = gia.Gia_Regular(fan0), gia.Gia_Regular(fan1)
	if gia.Gia_ObjId(p, fan0) > gia.Gia_ObjId(p, fan1) {
		fan0, fan1 = fan1, fan0
	}
	ctrl := gia.Gia_ObjRecognizeMux(mux, &dat0, &dat1)
	if ctrl == nil {
		return 0
	}
	dat0, dat1, ctrl = gia.Gia_Regular(dat0), gia.Gia_Regular(dat1), gia.Gia_Regular(ctrl)
	if gia.Gia_ObjIsAnd(dat0) == 0 || gia.Gia_ObjIsAnd(dat1) == 0 ||
		gia.Gia_ObjFaninId0p(p, dat0) != gia.Gia_ObjFaninId0p(p, dat1) ||
		gia.Gia_ObjFaninId1p(p, dat0) != gia.Gia_ObjFaninId1p(p, dat1) ||
		gia.Gia_ObjFaninId0p(p, dat0) != gia.Gia_ObjId(p, fan0) ||
		gia.Gia_ObjFaninId1p(p, dat0) != gia.Gia_ObjId(p, fan1) {
		return 0
	}
	ins[0], ins[1], ins[2] = gia.Gia_ObjId(p, fan0), gia.Gia_ObjId(p, fan1), gia.Gia_ObjId(p, ctrl)
	return 1
}

func Acec_DetectBoothXorFanin(p *gia.Gia_Man_t, obj *gia.Gia_Obj_t, ins []int32) int32 {
	if gia.Gia_ObjIsAnd(obj) == 0 || gia.Gia_ObjFaninC0(obj) == 0 || gia.Gia_ObjFaninC1(obj) == 0 {
		return 0
	}
	fan0, fan1 := gia.Gia_ObjFanin0(obj), gia.Gia_ObjFanin1(obj)
	if gia.Gia_ObjIsAnd(fan0) == 0 || gia.Gia_ObjIsAnd(fan1) == 0 {
		return 0
	}
	left := [2]*gia.Gia_Obj_t{gia.Gia_ObjFanin0(fan0), gia.Gia_ObjFanin1(fan0)}
	right := [2]*gia.Gia_Obj_t{gia.Gia_ObjFanin0(fan1), gia.Gia_ObjFanin1(fan1)}
	for i := range left {
		for j := range right {
			if Acec_DetectBoothXorMux(p, left[i], right[j], ins) != 0 {
				ins[3], ins[4] = gia.Gia_ObjId(p, left[1-i]), gia.Gia_ObjId(p, right[1-j])
				return 1
			}
		}
	}
	return 0
}
func Acec_DetectBoothOne(p *gia.Gia_Man_t, obj *gia.Gia_Obj_t, ins []int32) int32 {
	var fan0, fan1 *gia.Gia_Obj_t
	if gia.Gia_ObjRecognizeExor(obj, &fan0, &fan1) == 0 {
		return 0
	}
	fan0, fan1 = gia.Gia_Regular(fan0), gia.Gia_Regular(fan1)
	if Acec_DetectBoothXorFanin(p, fan0, ins) != 0 && ins[2] == gia.Gia_ObjId(p, fan1) {
		return 1
	}
	if Acec_DetectBoothXorFanin(p, fan1, ins) != 0 && ins[2] == gia.Gia_ObjId(p, fan0) {
		return 1
	}
	return 0
}
func Acec_DetectBoothTwoXor(p *gia.Gia_Man_t, obj *gia.Gia_Obj_t, ins []int32) int32 {
	if gia.Gia_ObjIsAnd(obj) == 0 {
		return 0
	}
	children := [2]*gia.Gia_Obj_t{gia.Gia_ObjFanin0(obj), gia.Gia_ObjFanin1(obj)}
	for i, child := range children {
		var fan0, fan1 *gia.Gia_Obj_t
		if gia.Gia_ObjRecognizeExor(child, &fan0, &fan1) != 0 {
			ins[0], ins[1] = gia.Gia_ObjId(p, gia.Gia_Regular(fan0)), gia.Gia_ObjId(p, gia.Gia_Regular(fan1))
			ins[2], ins[3], ins[4] = -1, 0, gia.Gia_ObjId(p, children[1-i])
			return 1
		}
	}
	return 0
}
func Acec_DetectBoothTwo(p *gia.Gia_Man_t, obj *gia.Gia_Obj_t, ins []int32) int32 {
	var fan0, fan1 *gia.Gia_Obj_t
	if gia.Gia_ObjRecognizeExor(obj, &fan0, &fan1) == 0 {
		return 0
	}
	fan0, fan1 = gia.Gia_Regular(fan0), gia.Gia_Regular(fan1)
	if Acec_DetectBoothTwoXor(p, fan0, ins) != 0 {
		ins[2] = gia.Gia_ObjId(p, fan1)
		return 1
	}
	if Acec_DetectBoothTwoXor(p, fan1, ins) != 0 {
		ins[2] = gia.Gia_ObjId(p, fan0)
		return 1
	}
	return 0
}
func Acec_DetectBoothTest(p *gia.Gia_Man_t) {
	var ins [5]int32
	for id := int32(1); id < gia.Gia_ManObjNum(p); id++ {
		obj := gia.Gia_ManObj(p, id)
		if gia.Gia_ObjIsAnd(obj) == 0 || (Acec_DetectBoothOne(p, obj, ins[:]) == 0 && Acec_DetectBoothTwo(p, obj, ins[:]) == 0) {
			continue
		}
		fmt.Printf("obj = %4d  :  b0 = %4d  b1 = %4d  b2 = %4d    a0 = %4d  a1 = %4d\n", id, ins[0], ins[1], ins[2], ins[3], ins[4])
	}
}
