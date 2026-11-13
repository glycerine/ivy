package acec

// Full-adder cut truth computation from src/proof/acec/acecFadds.c.
import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func Dtc_ManCutMergeOne(a, b, out []int32) int32 { return Ree_ManCutMergeOne(a, b, out) }
func Dtc_ManCutCheckEqual(cuts *Vec_Int_t, cut []int32) int32 {
	list := vec.Vec_IntArray(cuts)
	off := 1
	for i := int32(0); i < list[0]; i++ {
		n := list[off]
		if n == cut[0] {
			equal := true
			for j := int32(1); j <= n; j++ {
				if list[off+int(j)] != cut[j] {
					equal = false
					break
				}
			}
			if equal {
				return 1
			}
		}
		off += int(n) + 1
	}
	return 0
}
func Dtc_ObjComputeTruth_rec(o *Gia_Obj_t) uint32 {
	if o.Value != 0 {
		return o.Value
	}
	if gia.Gia_ObjIsAnd(o) == 0 {
		panic("acec: unmarked truth input")
	}
	a, b := Dtc_ObjComputeTruth_rec(gia.Gia_ObjFanin0(o)), Dtc_ObjComputeTruth_rec(gia.Gia_ObjFanin1(o))
	if gia.Gia_ObjFaninC0(o) != 0 {
		a = ^a
	}
	if gia.Gia_ObjFaninC1(o) != 0 {
		b = ^b
	}
	if gia.Gia_ObjIsXor(o) != 0 {
		o.Value = a ^ b
	} else {
		o.Value = a & b
	}
	return o.Value
}
func Dtc_ObjCleanTruth_rec(o *Gia_Obj_t) {
	// A zero-valued gate may have nonzero fanins. Visit every node in the
	// cone rather than using Value as a visited flag.
	seen := make(map[*Gia_Obj_t]bool)
	var clear func(*Gia_Obj_t)
	clear = func(node *Gia_Obj_t) {
		if seen[node] {
			return
		}
		seen[node] = true
		node.Value = 0
		if gia.Gia_ObjIsAnd(node) == 0 {
			return
		}
		clear(gia.Gia_ObjFanin0(node))
		clear(gia.Gia_ObjFanin1(node))
	}
	clear(o)
}
func Dtc_ObjComputeTruth(p *Gia_Man_t, id int32, cut []int32, truth *int32) int32 {
	base := [3]uint32{0xaa, 0xcc, 0xf0}
	for i := int32(1); i <= cut[0]; i++ {
		gia.Gia_ManObj(p, cut[i]).Value = base[i-1]
	}
	o := gia.Gia_ManObj(p, id)
	t := int32(Dtc_ObjComputeTruth_rec(o) & 0xff)
	Dtc_ObjCleanTruth_rec(o)
	if truth != nil {
		*truth = t
	}
	if t == 0x66 || t == 0x99 {
		return 3
	}
	if t == 0x96 || t == 0x69 {
		return 1
	}
	if t == 0xe8 || t == 0xd4 || t == 0xb2 || t == 0x71 || t == 0x17 || t == 0x2b || t == 0x4d || t == 0x8e {
		return 2
	}
	return 0
}
