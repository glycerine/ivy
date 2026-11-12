package gia

import "github.com/glycerine/abcgo/misc/vec"

// Gia_Min2ManGatherSupp_rec gathers CIs and AND nodes in fanin-first order.
// A node visited by the previous traversal signals overlap between the cones.
func Gia_Min2ManGatherSupp_rec(p *Gia_Man_t, id int32, cis, objs *Vec_Int_t) int32 {
	if Gia_ObjIsTravIdPreviousId(p, id) != 0 {
		return 1
	}
	if Gia_ObjIsTravIdCurrentId(p, id) != 0 {
		return 0
	}
	Gia_ObjSetTravIdCurrentId(p, id)
	obj := Gia_ManObj(p, id)
	if Gia_ObjIsCi(obj) != 0 {
		vec.Vec_IntPush(cis, id)
		return 0
	}
	if Gia_ObjIsAnd(obj) == 0 {
		panic("gia: support cone contains a non-AND node")
	}
	a := Gia_Min2ManGatherSupp_rec(p, Gia_ObjFaninId0(obj, id), cis, objs)
	b := Gia_Min2ManGatherSupp_rec(p, Gia_ObjFaninId1(obj, id), cis, objs)
	vec.Vec_IntPush(objs, id)
	return a | b
}
