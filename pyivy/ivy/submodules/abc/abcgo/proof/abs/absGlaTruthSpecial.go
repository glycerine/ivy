package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

// Ga2_ObjComputeTruthSpecial folds constant leaves, orders surviving SAT
// literals, and expresses the cut truth in that new literal order.
func Ga2_ObjComputeTruthSpecial(p *gia.Gia_Man_t, root *gia.Gia_Obj_t, leaves, lits *vec.Vec_Int_t) uint32 {
	elementary := [...]uint32{0xaaaaaaaa, 0xcccccccc, 0xf0f0f0f0, 0xff00ff00, 0xffff0000}
	n := vec.Vec_IntSize(leaves)
	if gia.Gia_ObjIsAnd(root) == 0 || n > len(elementary) || vec.Vec_IntSize(lits) != n {
		panic("abs: invalid specialized truth arguments")
	}
	for i := 0; i < n; i++ {
		obj := gia.Gia_ManObj(p, vec.Vec_IntEntry(leaves, i))
		lit := vec.Vec_IntEntry(lits, i)
		if lit < 0 {
			panic("abs: missing SAT literal")
		}
		switch lit {
		case 0:
			obj.Value = 0
		case 1:
			obj.Value = ^uint32(0)
		default:
			obj.Value = elementary[i]
		}
	}
	result := Ga2_ObjComputeTruth_rec(p, root, 1)
	if result != 0 && result != ^uint32(0) {
		used := make([]int32, 0, n)
		for i := 0; i < n; i++ {
			if Ga2_ObjTruthDepends(result, int32(i)) != 0 {
				used = append(used, int32(i))
			}
		}
		if len(used) == 0 {
			panic("abs: truth without support")
		}
		vec.Vec_IntSelectSortCost(used, len(used), lits)
		for i := 0; i < n; i++ {
			obj := gia.Gia_ManObj(p, vec.Vec_IntEntry(leaves, i))
			lit := vec.Vec_IntEntry(lits, i)
			switch lit {
			case 0:
				obj.Value = 0
			case 1:
				obj.Value = ^uint32(0)
			default:
				obj.Value = 0xdeadcafe
			}
		}
		for i, index := range used {
			lit := vec.Vec_IntEntry(lits, int(index))
			if lit <= 1 {
				panic("abs: constant is essential")
			}
			value := elementary[i]
			if lit&1 != 0 {
				value = ^value
			}
			gia.Gia_ManObj(p, vec.Vec_IntEntry(leaves, int(index))).Value = value
			used[i] = lit &^ 1
		}
		result = Ga2_ObjComputeTruth_rec(p, root, 1)
		vec.Vec_IntClear(lits)
		for i, lit := range used {
			vec.Vec_IntPush(lits, lit)
			if Ga2_ObjTruthDepends(result, int32(i)) == 0 {
				panic("abs: lost essential variable")
			}
		}
		for i := len(used); i < 5; i++ {
			if Ga2_ObjTruthDepends(result, int32(i)) != 0 {
				panic("abs: unexpected essential variable")
			}
		}
	}
	for i := 0; i < n; i++ {
		gia.Gia_ManObj(p, vec.Vec_IntEntry(leaves, i)).Value = 0
	}
	return result
}
