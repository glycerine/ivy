package abs

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func Rf2_ManGatherFanins_rec(p *Rf2_Man_t, obj *gia.Gia_Obj_t, fanins *vec.Vec_Int_t, depth, rootID, first int32) {
	if gia.Gia_ObjIsTravIdCurrent(p.pGia, obj) != 0 {
		return
	}
	gia.Gia_ObjSetTravIdCurrent(p.pGia, obj)
	phase := gia.Gia_ObjPhase(obj)
	if phase != 0 && first == 0 {
		v := Rf2_ObjVec(p, obj)
		if vec.Vec_IntSize(v) == 0 {
			vec.Vec_IntPush(fanins, gia.Gia_ObjId(p.pGia, obj))
		}
		vec.Vec_IntPushUnique(v, rootID)
		if depth == 0 {
			return
		}
	}
	if gia.Gia_ObjIsPi(p.pGia, obj) != 0 || gia.Gia_ObjIsConst0(obj) != 0 {
		return
	}
	if gia.Gia_ObjIsRo(p.pGia, obj) != 0 {
		if phase == 0 {
			panic("unmarked register output")
		}
		obj = gia.Gia_ObjRoToRi(p.pGia, obj)
		Rf2_ManGatherFanins_rec(p, gia.Gia_ObjFanin0(obj), fanins, depth-1, rootID, 0)
	} else if gia.Gia_ObjIsAnd(obj) != 0 {
		Rf2_ManGatherFanins_rec(p, gia.Gia_ObjFanin0(obj), fanins, depth-phase, rootID, 0)
		Rf2_ManGatherFanins_rec(p, gia.Gia_ObjFanin1(obj), fanins, depth-phase, rootID, 0)
	} else {
		panic("invalid alternate refinement fanin")
	}
}
func Rf2_ManGatherFanins(p *Rf2_Man_t, depth int32) {
	for _, id := range vec.Vec_IntArray(p.vMap) {
		v := Rf2_ObjVec(p, gia.Gia_ManObj(p.pGia, id))
		if vec.Vec_IntSize(v) != 0 {
			panic("nonempty alternate refinement vector")
		}
		vec.Vec_IntPush(v, 0)
	}
	vec.Vec_IntClear(p.vFanins)
	for i, id := range vec.Vec_IntArray(p.vMap) {
		obj := gia.Gia_ManObj(p.pGia, id)
		if gia.Gia_ObjIsPi(p.pGia, obj) != 0 {
			continue
		}
		gia.Gia_ManIncrementTravId(p.pGia)
		Rf2_ManGatherFanins_rec(p, obj, p.vFanins, depth, int32(i), 1)
	}
	used := vec.Vec_IntStart(vec.Vec_IntSize(p.vMap))
	fmt.Printf("\nMap (%d): ", vec.Vec_IntSize(p.vMap))
	for i, id := range vec.Vec_IntArray(p.vMap) {
		v := Rf2_ObjVec(p, gia.Gia_ManObj(p.pGia, id))
		entries := vec.Vec_IntArray(v)
		if len(entries) > 1 {
			fmt.Printf("%d=%d ", i, len(entries)-1)
		}
		for _, entry := range entries[1:] {
			vec.Vec_IntAddToEntry(used, int(entry), 1)
		}
		vec.Vec_IntClear(v)
	}
	fmt.Printf("\nInt (%d): ", vec.Vec_IntSize(p.vFanins))
	for i, id := range vec.Vec_IntArray(p.vFanins) {
		v := Rf2_ObjVec(p, gia.Gia_ManObj(p.pGia, id))
		entries := vec.Vec_IntArray(v)
		if len(entries) > 1 {
			fmt.Printf("%d=%d ", i, len(entries))
			for _, entry := range entries {
				vec.Vec_IntAddToEntry(used, int(entry), 1)
			}
		}
		vec.Vec_IntClear(v)
	}
	fmt.Println()
	for _, entry := range vec.Vec_IntArray(used) {
		fmt.Printf("%d ", entry)
	}
	fmt.Println()
}
