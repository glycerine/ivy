package abs

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func Rf2_ManObj(p *Rf2_Man_t, obj *gia.Gia_Obj_t, frame int32) *Rf2_Obj_t {
	if (gia.Gia_ObjIsConst0(obj) == 0 && obj.Value == 0) || int32(obj.Value) >= p.nObjsFrame || frame < 0 || frame > p.pCex.IFrame {
		panic("invalid alternate refinement object")
	}
	return &p.pObjs[frame*p.nObjsFrame+int32(obj.Value)]
}
func Rf2_ObjVec(p *Rf2_Man_t, obj *gia.Gia_Obj_t) *vec.Vec_Int_t {
	return &p.pvVecs[gia.Gia_ObjId(p.pGia, obj)]
}
func Rf2_ObjA(p *Rf2_Man_t, obj *gia.Gia_Obj_t) []int32 {
	return vec.Vec_IntArray(Rf2_ObjVec(p, obj))[:p.nMapWords]
}
func Rf2_ObjN(p *Rf2_Man_t, obj *gia.Gia_Obj_t) []int32 {
	return vec.Vec_IntArray(Rf2_ObjVec(p, obj))[p.nMapWords:]
}
func Rf2_ObjClear(p *Rf2_Man_t, obj *gia.Gia_Obj_t) {
	vec.Vec_IntFill(Rf2_ObjVec(p, obj), int(2*p.nMapWords), 0)
}
func Rf2_ObjStart(p *Rf2_Man_t, obj *gia.Gia_Obj_t, bit int32) {
	Rf2_ObjClear(p, obj)
	a, n := Rf2_ObjA(p, obj), Rf2_ObjN(p, obj)
	for i := range n {
		n[i] = -1
	}
	mask := int32(uint32(1) << uint(bit&31))
	a[bit>>5] |= mask
	n[bit>>5] ^= mask
}
func Rf2_ObjCopy(p *Rf2_Man_t, obj, fanin *gia.Gia_Obj_t) {
	dest, src := vec.Vec_IntArray(Rf2_ObjVec(p, obj)), vec.Vec_IntArray(Rf2_ObjVec(p, fanin))
	if len(dest) != int(2*p.nMapWords) || len(src) != len(dest) {
		panic("invalid alternate refinement vector")
	}
	copy(dest, src)
}
func Rf2_ObjDeriveAnd(p *Rf2_Man_t, obj *gia.Gia_Obj_t, one int32) {
	left, right := gia.Gia_ObjFanin0(obj), gia.Gia_ObjFanin1(obj)
	if gia.Gia_ObjIsAnd(obj) == 0 || one != gia.Gia_ObjMark0(obj) || one != gia.Gia_ObjMark0(left)^gia.Gia_ObjFaninC0(obj) || one != gia.Gia_ObjMark0(right)^gia.Gia_ObjFaninC1(obj) {
		panic("invalid alternate refinement AND")
	}
	a, a0, a1 := Rf2_ObjA(p, obj), Rf2_ObjA(p, left), Rf2_ObjA(p, right)
	n, n0, n1 := Rf2_ObjN(p, obj), Rf2_ObjN(p, left), Rf2_ObjN(p, right)
	for i := range a {
		if one != 0 {
			a[i] = a0[i] & a1[i]
			n[i] = n0[i] | n1[i]
		} else {
			a[i] = a0[i] | a1[i]
			n[i] = n0[i] & n1[i]
		}
	}
}
func Rf2_ObjPrint(p *Rf2_Man_t, root *gia.Gia_Obj_t) {
	for side, words := range [][]int32{Rf2_ObjA(p, root), Rf2_ObjN(p, root)} {
		for i, id := range vec.Vec_IntArray(p.vMap) {
			if gia.Gia_ObjIsPi(p.pGia, gia.Gia_ManObj(p.pGia, id)) != 0 {
				continue
			}
			bit := (uint32(words[i>>5]) >> uint(i&31)) & 1
			if side == 1 {
				bit ^= 1
			}
			fmt.Print(bit)
		}
		fmt.Println()
	}
}
