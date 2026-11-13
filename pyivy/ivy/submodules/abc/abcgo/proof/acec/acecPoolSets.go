package acec

// Carry-root grouping from src/proof/acec/acecPool.c.
import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func Acec_ManCountXorTreeInputs_rec(p *Gia_Man_t, node int32, xors, m *Vec_Int_t, isCarry *Vec_Bit_t, carryRoots, xorSet *Vec_Int_t) {
	if node == 0 || gia.Gia_ObjIsTravIdCurrentId(p, node) != 0 {
		return
	}
	gia.Gia_ObjSetTravIdCurrentId(p, node)
	box := vec.Vec_IntEntry(m, int(node))
	if box == -1 {
		if vec.Vec_BitEntry(isCarry, int(node)) != 0 {
			vec.Vec_IntPush(carryRoots, node)
		}
		return
	}
	for k := 1; k < 4; k++ {
		Acec_ManCountXorTreeInputs_rec(p, vec.Vec_IntEntry(xors, int(4*box)+k), xors, m, isCarry, carryRoots, xorSet)
	}
	vec.Vec_IntPush(xorSet, vec.Vec_IntEntry(xors, int(4*box)))
}
func Acec_ManCollectCarryRootSets(p *Gia_Man_t, adds, carryMap, xors, xorRoots, carryRoots *Vec_Int_t) *Vec_Wec_t {
	sets := vec.Vec_WecAlloc(100)
	carrySet, xorSet := vec.Vec_IntAlloc(100), vec.Vec_IntAlloc(100)
	isCarry := vec.Vec_BitStart(int(gia.Gia_ManObjNum(p)))
	xorMap := vec.Vec_IntStartFull(int(gia.Gia_ManObjNum(p)))
	for i := 0; 4*i < vec.Vec_IntSize(xors); i++ {
		vec.Vec_IntWriteEntry(xorMap, int(vec.Vec_IntEntry(xors, 4*i)), int32(i))
	}
	for _, id := range vec.Vec_IntArray(carryRoots) {
		vec.Vec_BitWriteEntry(isCarry, int(id), 1)
	}
	for _, root := range vec.Vec_IntArray(xorRoots) {
		vec.Vec_IntClear(xorSet)
		vec.Vec_IntClear(carrySet)
		gia.Gia_ManIncrementTravId(p)
		Acec_ManCountXorTreeInputs_rec(p, root, xors, xorMap, isCarry, carrySet, xorSet)
		nontrivial := false
		for _, id := range vec.Vec_IntArray(carrySet) {
			if Acec_ManCheckCarryMap(p, id, adds, carryMap) != 0 {
				nontrivial = true
				break
			}
		}
		if !nontrivial {
			continue
		}
		vec.Vec_IntSort(carrySet, 0)
		vec.Vec_IntAppend(vec.Vec_WecPushLevel(sets), xorSet)
		vec.Vec_IntAppend(vec.Vec_WecPushLevel(sets), carrySet)
		for i := 0; i < 3; i++ {
			vec.Vec_WecPushLevel(sets)
		}
		for _, id := range vec.Vec_IntArray(carrySet) {
			if vec.Vec_BitEntry(isCarry, int(id)) == 0 {
				panic("acec: carry root already used")
			}
			vec.Vec_BitWriteEntry(isCarry, int(id), 0)
		}
	}
	for _, id := range vec.Vec_IntArray(carryRoots) {
		if vec.Vec_BitEntry(isCarry, int(id)) == 0 || Acec_ManCheckCarryMap(p, id, adds, carryMap) == 0 {
			continue
		}
		vec.Vec_WecPushLevel(sets)
		vec.Vec_IntPush(vec.Vec_WecPushLevel(sets), id)
		for i := 0; i < 3; i++ {
			vec.Vec_WecPushLevel(sets)
		}
	}
	return sets
}
