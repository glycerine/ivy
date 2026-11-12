// Algebraic AIG balancing, ported from ABC hopBalance.c (Alan Mishchenko).
package hop

import (
	"fmt"
	"sort"
)

// Reusable supergate storage indexed by recursion depth.
type Vec_Vec_t struct {
	nSize  int32
	pArray []*Vec_Ptr_t
}

func Vec_VecAlloc(n int32) *Vec_Vec_t { return &Vec_Vec_t{pArray: make([]*Vec_Ptr_t, 0, n)} }
func Vec_VecSize(v *Vec_Vec_t) int32  { return v.nSize }
func Vec_VecExpand(v *Vec_Vec_t, level int32) {
	for v.nSize <= level {
		v.pArray = append(v.pArray, Vec_PtrAlloc(16))
		v.nSize++
	}
}
func Vec_VecEntry(v *Vec_Vec_t, level int32) *Vec_Ptr_t { return v.pArray[level] }
func Vec_VecPush(v *Vec_Vec_t, level int32, value any) {
	Vec_VecExpand(v, level)
	Vec_PtrPush(v.pArray[level], value)
}
func Vec_VecFree(v *Vec_Vec_t) {
	if v != nil {
		for _, x := range v.pArray {
			Vec_PtrFree(x)
		}
		v.pArray = nil
		v.nSize = 0
	}
}

func Hop_ManBalance(p *Hop_Man_t, update int32) *Hop_Man_t {
	result := Hop_ManStart()
	result.fRefCount = 0
	Hop_ManCleanData(p)
	p.pConst1.pData = result.pConst1
	for i := int32(0); i < Vec_PtrSize(p.vPis); i++ {
		Vec_PtrEntry(p.vPis, i).(*Hop_Obj_t).pData = Hop_ObjCreatePi(result)
	}
	store := Vec_VecAlloc(50)
	defer Vec_VecFree(store)
	for i := int32(0); i < Vec_PtrSize(p.vPos); i++ {
		po := Vec_PtrEntry(p.vPos, i).(*Hop_Obj_t)
		driver := Hop_NodeBalance_rec(result, Hop_ObjFanin0(po), store, 0, update)
		Hop_ObjCreatePo(result, Hop_NotCond(driver, Hop_ObjFaninC0(po)))
	}
	if Hop_ManCheck(result) == 0 {
		fmt.Println("Hop_ManBalance(): The check has failed.")
	}
	return result
}
func Hop_NodeBalance_rec(p *Hop_Man_t, old *Hop_Obj_t, store *Vec_Vec_t, level, update int32) *Hop_Obj_t {
	assert(Hop_IsComplement(old) == 0, "balancing requires a regular node")
	if old.pData != nil {
		return old.pData.(*Hop_Obj_t)
	}
	assert(Hop_ObjIsNode(old) != 0, "balancing requires a node or mapped input")
	cone := Hop_NodeBalanceCone(old, store, level)
	if cone.nSize == 0 {
		old.pData = Hop_ManConst0(p)
		return old.pData.(*Hop_Obj_t)
	}
	for i := int32(0); i < cone.nSize; i++ {
		edge := cone.pArray[i].(*Hop_Obj_t)
		node := Hop_NodeBalance_rec(p, Hop_Regular(edge), store, level+1, update)
		cone.pArray[i] = Hop_NotCond(node, Hop_IsComplement(edge))
	}
	result := Hop_NodeBalanceBuildSuper(p, cone, Hop_ObjType(old), update)
	assert(old.pData == nil, "node already mapped")
	old.pData = result
	return result
}
func Hop_NodeBalanceCone_rec(root, node *Hop_Obj_t, cone *Vec_Ptr_t) int32 {
	regular := Hop_Regular(node)
	if regular.fMarkB != 0 {
		for i := int32(0); i < cone.nSize; i++ {
			if cone.pArray[i].(*Hop_Obj_t) == node {
				return 1
			}
		}
		for i := int32(0); i < cone.nSize; i++ {
			if cone.pArray[i].(*Hop_Obj_t) == Hop_Not(node) {
				return -1
			}
		}
		panic("marked supergate leaf not found")
	}
	if node != root && (Hop_IsComplement(node) != 0 || Hop_ObjType(node) != Hop_ObjType(root) || Hop_ObjRefs(node) > 1 || cone.nSize > 10000) {
		Vec_PtrPush(cone, node)
		regular.fMarkB = 1
		return 0
	}
	assert(Hop_IsComplement(node) == 0 && Hop_ObjIsNode(node) != 0, "invalid supergate internal node")
	a := Hop_NodeBalanceCone_rec(root, Hop_ObjChild0(node), cone)
	b := Hop_NodeBalanceCone_rec(root, Hop_ObjChild1(node), cone)
	if a == -1 || b == -1 {
		return -1
	}
	if a != 0 || b != 0 {
		return 1
	}
	return 0
}
func Hop_NodeBalanceCone(node *Hop_Obj_t, store *Vec_Vec_t, level int32) *Vec_Ptr_t {
	assert(Hop_IsComplement(node) == 0, "regular cone root")
	Vec_VecExpand(store, level)
	cone := Vec_VecEntry(store, level)
	cone.nSize = 0
	result := Hop_NodeBalanceCone_rec(node, node, cone)
	for i := int32(0); i < cone.nSize; i++ {
		Hop_Regular(cone.pArray[i].(*Hop_Obj_t)).fMarkB = 0
	}
	if result == -1 {
		cone.nSize = 0
	}
	return cone
}
func Hop_NodeCompareLevelsDecrease(a, b *Hop_Obj_t) int32 {
	a, b = Hop_Regular(a), Hop_Regular(b)
	if Hop_ObjLevel(a) > Hop_ObjLevel(b) {
		return -1
	}
	if Hop_ObjLevel(a) < Hop_ObjLevel(b) {
		return 1
	}
	if a.Id > b.Id {
		return -1
	}
	if a.Id < b.Id {
		return 1
	}
	return 0
}
func Hop_NodeBalanceBuildSuper(p *Hop_Man_t, cone *Vec_Ptr_t, typ Hop_Type_t, update int32) *Hop_Obj_t {
	assert(cone.nSize > 0, "empty supergate")
	sort.Slice(cone.pArray[:cone.nSize], func(i, j int) bool {
		return Hop_NodeCompareLevelsDecrease(cone.pArray[i].(*Hop_Obj_t), cone.pArray[j].(*Hop_Obj_t)) < 0
	})
	for cone.nSize > 1 {
		var left int32
		if update != 0 {
			left = Hop_NodeBalanceFindLeft(cone)
		}
		Hop_NodeBalancePermute(p, cone, left, hopBool(typ == AIG_EXOR))
		a := cone.pArray[cone.nSize-1].(*Hop_Obj_t)
		b := cone.pArray[cone.nSize-2].(*Hop_Obj_t)
		cone.nSize -= 2
		Hop_NodeBalancePushUniqueOrderByLevel(cone, Hop_Oper(p, a, b, typ))
	}
	return cone.pArray[0].(*Hop_Obj_t)
}
func Hop_NodeBalanceFindLeft(cone *Vec_Ptr_t) int32 {
	if cone.nSize < 3 {
		return 0
	}
	current := cone.nSize - 2
	level := Hop_ObjLevel(Hop_Regular(cone.pArray[current].(*Hop_Obj_t)))
	for current > 0 && Hop_ObjLevel(Hop_Regular(cone.pArray[current-1].(*Hop_Obj_t))) == level {
		current--
	}
	return current
}
func Hop_NodeBalancePermute(p *Hop_Man_t, cone *Vec_Ptr_t, left, exor int32) {
	right := cone.nSize - 2
	assert(left <= right, "invalid balancing bound")
	if left == right {
		return
	}
	a, b := cone.pArray[right+1].(*Hop_Obj_t), cone.pArray[right].(*Hop_Obj_t)
	if Hop_Regular(a) == p.pConst1 || Hop_Regular(b) == p.pConst1 {
		return
	}
	typ := AIG_AND
	if exor != 0 {
		typ = AIG_EXOR
	}
	for i := right; i >= left; i-- {
		candidate := cone.pArray[i].(*Hop_Obj_t)
		if Hop_Regular(candidate) == p.pConst1 || Hop_TableLookup(p, Hop_ObjCreateGhost(p, a, candidate, typ)) != nil {
			if candidate != b {
				cone.pArray[i], cone.pArray[right] = b, candidate
			}
			return
		}
	}
}
func Hop_NodeBalancePushUniqueOrderByLevel(cone *Vec_Ptr_t, node *Hop_Obj_t) {
	for i := int32(0); i < cone.nSize; i++ {
		if cone.pArray[i].(*Hop_Obj_t) == node {
			return
		}
	}
	Vec_PtrPush(cone, node)
	for i := cone.nSize - 1; i > 0; i-- {
		a, b := cone.pArray[i].(*Hop_Obj_t), cone.pArray[i-1].(*Hop_Obj_t)
		if Hop_ObjLevel(Hop_Regular(a)) <= Hop_ObjLevel(Hop_Regular(b)) {
			break
		}
		cone.pArray[i], cone.pArray[i-1] = b, a
	}
}
