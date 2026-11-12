package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
)

func Gia_ManSpeedupObj_rec(p *Gia_Man_t, obj *Gia_Obj_t, nodes *Vec_Int_t) int32 {
	if Gia_ObjIsTravIdCurrent(p, obj) != 0 {
		return 1
	}
	Gia_ObjSetTravIdCurrent(p, obj)
	if Gia_ObjIsCi(obj) != 0 {
		return 0
	}
	if Gia_ObjIsAnd(obj) == 0 {
		panic("gia: invalid speedup cone")
	}
	if Gia_ManSpeedupObj_rec(p, Gia_ObjFanin0(obj), nodes) == 0 || Gia_ManSpeedupObj_rec(p, Gia_ObjFanin1(obj), nodes) == 0 {
		return 0
	}
	vec.Vec_IntPush(nodes, obj.id)
	return 1
}
func Gia_ManSpeedupObj(dst, p *Gia_Man_t, obj *Gia_Obj_t, leaves, times *Vec_Int_t) {
	if vec.Vec_IntSize(times) > 5 {
		panic("gia: too many critical cofactors")
	}
	Gia_ManIncrementTravId(p)
	Gia_ObjSetTravIdCurrent(p, Gia_ManConst0(p))
	for _, id := range vec.Vec_IntArray(leaves) {
		Gia_ObjSetTravIdCurrentId(p, id)
	}
	nodes := vec.Vec_IntAlloc(100)
	if Gia_ManSpeedupObj_rec(p, obj, nodes) == 0 {
		fmt.Println("Bad node!!!")
		return
	}
	if vec.Vec_IntSize(nodes) == 0 {
		return
	}
	count := 1 << uint(vec.Vec_IntSize(times))
	cofs := make([]int32, count)
	for mint := 0; mint < count; mint++ {
		for _, id := range vec.Vec_IntArray(leaves) {
			Gia_ManObj(p, id).Value = uint32(2 * id)
		}
		for k, id := range vec.Vec_IntArray(times) {
			Gia_ManObj(p, id).Value = uint32((mint >> uint(k)) & 1)
		}
		for _, id := range vec.Vec_IntArray(nodes) {
			node := Gia_ManObj(p, id)
			node.Value = uint32(Gia_ManHashAnd(dst, Gia_ObjFanin0Copy(node), Gia_ObjFanin1Copy(node)))
		}
		cofs[mint] = int32(obj.Value)
	}
	for k, id := range vec.Vec_IntArray(times) {
		skip := 1 << uint(k)
		for i := 0; i < count; i += 2 * skip {
			cofs[i] = Gia_ManHashMux(dst, 2*id, cofs[i+skip], cofs[i])
		}
	}
	result := cofs[0] >> 1
	if result <= obj.id {
		return
	}
	Gia_ObjSetRepr(dst, result, obj.id)
	Gia_ObjSetNext(dst, result, Gia_ObjNext(dst, obj.id))
	Gia_ObjSetNext(dst, obj.id, result)
}
