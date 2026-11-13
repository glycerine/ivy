package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

// Ga2_ManBreakTree_rec marks boundaries so each resulting tree has fewer than N leaves.
func Ga2_ManBreakTree_rec(p *gia.Gia_Man_t, obj *gia.Gia_Obj_t, first, n int32) int32 {
	if gia.Gia_ObjPhase(obj) != 0 && first == 0 {
		return 1
	}
	a, b := gia.Gia_ObjFanin0(obj), gia.Gia_ObjFanin1(obj)
	va, vb := Ga2_ManBreakTree_rec(p, a, 0, n), Ga2_ManBreakTree_rec(p, b, 0, n)
	if va+vb < n {
		return va + vb
	}
	if va+vb == n {
		gia.Gia_ObjSetPhaseValue(obj, 1)
		return 1
	}
	if va >= n || vb >= n {
		panic("abs: invalid tree leaf count")
	}
	if va >= vb {
		gia.Gia_ObjSetPhaseValue(a, 1)
		va = 1
	} else {
		gia.Gia_ObjSetPhaseValue(b, 1)
		vb = 1
	}
	if va+vb < n {
		return va + vb
	}
	if va+vb == n {
		gia.Gia_ObjSetPhaseValue(obj, 1)
		return 1
	}
	panic("abs: unable to break tree")
}
func Ga2_ManCheckNodesAnd(p *gia.Gia_Man_t, nodes *vec.Vec_Int_t) int32 {
	for i := 0; i < vec.Vec_IntSize(nodes); i++ {
		obj := gia.Gia_ManObj(p, vec.Vec_IntEntry(nodes, i))
		if gia.Gia_ObjPhase(gia.Gia_ObjFanin0(obj)) == 0 && gia.Gia_ObjFaninC0(obj) != 0 || gia.Gia_ObjPhase(gia.Gia_ObjFanin1(obj)) == 0 && gia.Gia_ObjFaninC1(obj) != 0 {
			return 0
		}
	}
	return 1
}
func Ga2_ManCollectNodes_rec(p *gia.Gia_Man_t, obj *gia.Gia_Obj_t, nodes *vec.Vec_Int_t, first int32) {
	if gia.Gia_ObjPhase(obj) != 0 && first == 0 {
		return
	}
	if gia.Gia_ObjIsAnd(obj) == 0 {
		panic("abs: expected AND node")
	}
	Ga2_ManCollectNodes_rec(p, gia.Gia_ObjFanin0(obj), nodes, 0)
	Ga2_ManCollectNodes_rec(p, gia.Gia_ObjFanin1(obj), nodes, 0)
	vec.Vec_IntPush(nodes, gia.Gia_ObjId(p, obj))
}
func Ga2_ManCollectLeaves_rec(p *gia.Gia_Man_t, obj *gia.Gia_Obj_t, leaves *vec.Vec_Int_t, first int32) {
	if gia.Gia_ObjPhase(obj) != 0 && first == 0 {
		vec.Vec_IntPushUnique(leaves, gia.Gia_ObjId(p, obj))
		return
	}
	if gia.Gia_ObjIsAnd(obj) == 0 {
		panic("abs: expected AND node")
	}
	Ga2_ManCollectLeaves_rec(p, gia.Gia_ObjFanin0(obj), leaves, 0)
	Ga2_ManCollectLeaves_rec(p, gia.Gia_ObjFanin1(obj), leaves, 0)
}

// Ga2_ObjComputeTruth_rec evaluates a marked cone using its leaves' truth words.
func Ga2_ObjComputeTruth_rec(p *gia.Gia_Man_t, obj *gia.Gia_Obj_t, first int32) uint32 {
	if gia.Gia_ObjPhase(obj) != 0 && first == 0 {
		return obj.Value
	}
	if gia.Gia_ObjIsAnd(obj) == 0 {
		panic("abs: expected AND node")
	}
	a := Ga2_ObjComputeTruth_rec(p, gia.Gia_ObjFanin0(obj), 0)
	b := Ga2_ObjComputeTruth_rec(p, gia.Gia_ObjFanin1(obj), 0)
	if gia.Gia_ObjFaninC0(obj) != 0 {
		a = ^a
	}
	if gia.Gia_ObjFaninC1(obj) != 0 {
		b = ^b
	}
	return a & b
}
func Ga2_ManComputeTruth(p *gia.Gia_Man_t, root *gia.Gia_Obj_t, leaves *vec.Vec_Int_t) uint32 {
	truth := [...]uint32{0xaaaaaaaa, 0xcccccccc, 0xf0f0f0f0, 0xff00ff00, 0xffff0000}
	if vec.Vec_IntSize(leaves) > len(truth) {
		panic("abs: too many truth-table leaves")
	}
	for i := 0; i < vec.Vec_IntSize(leaves); i++ {
		gia.Gia_ManObj(p, vec.Vec_IntEntry(leaves, i)).Value = truth[i]
	}
	result := Ga2_ObjComputeTruth_rec(p, root, 1)
	for i := 0; i < vec.Vec_IntSize(leaves); i++ {
		gia.Gia_ManObj(p, vec.Vec_IntEntry(leaves, i)).Value = 0
	}
	return result
}
