package ivy

import (
	"fmt"
	"sort"
)

// Ivy_ManRewriteAlg performs the algebraic cut rewrite pass. It considers
// only nodes that existed at pass start; newly built nodes are candidates,
// never additional roots in the same sweep.
func Ivy_ManRewriteAlg(p *Ivy_Man_t, updateLevels, useZeroCost int32) int32 {
	if p == nil || p.memoryStopped {
		panic("ivy: rewrite on absent or stopped manager")
	}
	var required []int32
	if updateLevels != 0 {
		required = Ivy_ManRequiredLevels(p)
		defer func() { p.vRequired = nil }()
	}
	front, leaves, cone, solutions := Vec_PtrAlloc(100), Vec_PtrAlloc(100), Vec_PtrAlloc(100), Vec_PtrAlloc(100)
	defer Vec_PtrFree(front)
	defer Vec_PtrFree(leaves)
	defer Vec_PtrFree(cone)
	defer Vec_PtrFree(solutions)

	oldNodeLimit := int32(len(p.vObjs))
	var used, undone int32
	for i := 1; i < len(p.vObjs); i++ {
		node := p.vObjs[i]
		if node == nil {
			continue
		}
		if node.Id >= oldNodeLimit {
			break
		}
		assert(Ivy_ObjIsBuf(node) == 0, "rewrite pass requires buffers to be propagated")
		if Ivy_ObjIsNode(node) == 0 || Ivy_ObjIsExor(node) != 0 || Ivy_ObjIsMuxType(node) != 0 {
			continue
		}
		cut := Ivy_ManFindAlgCut(node, front, leaves, cone)
		if cut == 1 {
			continue
		}
		if cut == -1 {
			Ivy_ObjReplace(p, node, Ivy_ManConst0(p), 1, 0, 1)
			continue
		}
		assert(Vec_PtrSize(leaves) > 2, "Vec_PtrSize(leaves) > 2")
		levelLimit := int32(1 << 30)
		if required != nil {
			levelLimit = required[node.Id]
		}
		result := Ivy_NodeRewriteAlg(p, node, front, leaves, cone, solutions, levelLimit, useZeroCost)
		if result == nil || result == node {
			continue
		}
		assert(Vec_PtrSize(solutions) == 1 || Ivy_IsComplement(result) == 0,
			"a multi-node rewrite result must be regular")
		if Ivy_ObjLevel(result) > levelLimit && Ivy_ObjRefs(result) == 0 {
			Ivy_ObjDelete_rec(p, result, 1)
			undone++
		} else {
			Ivy_ObjReplace(p, node, result, 1, 0, 1)
			used++
		}
	}
	fmt.Printf("Used = %d. Undo = %d.\n", used, undone)
	if removed := Ivy_ManCleanup(p); removed != 0 {
		fmt.Printf("Cleanup after rewriting removed %d dangling nodes.\n", removed)
	}
	if Ivy_ManCheck(p) == 0 {
		fmt.Printf("Ivy_ManRewriteAlg(): The check has failed.\n")
	}
	return 1
}

func ivyVecObj(v *Vec_Ptr_t, i int32) *Ivy_Obj_t {
	obj, ok := Vec_PtrEntry(v, i).(*Ivy_Obj_t)
	if !ok || obj == nil {
		panic("ivy: invalid object vector entry")
	}
	return obj
}

func Ivy_NodeRewriteAlg(p *Ivy_Man_t, root *Ivy_Obj_t, front, leaves, cone, solutions *Vec_Ptr_t, levelLimit, useZeroCost int32) *Ivy_Obj_t {
	if Ivy_ObjIsExor(root) != 0 {
		var nodes []*Ivy_Obj_t
		Ivy_ManCollectCone(root, vectorObjects(front), &nodes)
		setVectorObjects(cone, nodes)
	} else {
		var nodes []*Ivy_Obj_t
		Ivy_ManCollectCone(root, vectorObjects(leaves), &nodes)
		setVectorObjects(cone, nodes)
	}
	coneNodes := vectorObjects(cone)
	for _, node := range coneNodes {
		Ivy_ObjRefsDec(Ivy_ObjFanin0(node))
		if Ivy_ObjIsNode(node) != 0 {
			Ivy_ObjRefsDec(Ivy_ObjFanin1(node))
		}
		node.fMarkB = 1
	}
	frontNodes := vectorObjects(front)
	for _, node := range frontNodes {
		Ivy_Regular(node).fMarkA = 1
	}
	mffc := Ivy_NodeCountMffc(root)
	for _, node := range frontNodes {
		Ivy_Regular(node).fMarkA = 0
	}

	if Ivy_MultiPlus(p, leaves, cone, Ivy_ObjType(root), mffc+useZeroCost, solutions) == 0 {
		for _, node := range coneNodes {
			Ivy_ObjRefsInc(Ivy_ObjFanin0(node))
			if Ivy_ObjIsNode(node) != 0 {
				Ivy_ObjRefsInc(Ivy_ObjFanin1(node))
			}
			node.fMarkA, node.fMarkB = 0, 0
		}
		return nil
	}
	for _, node := range coneNodes {
		Ivy_ObjRefsInc(Ivy_ObjFanin0(node))
		if Ivy_ObjIsNode(node) != 0 {
			Ivy_ObjRefsInc(Ivy_ObjFanin1(node))
		}
		node.fMarkA, node.fMarkB = 0, 0
	}
	if Vec_PtrSize(solutions) == 1 {
		return ivyVecObj(solutions, 0)
	}
	args := vectorObjects(solutions)
	return Ivy_NodeBalanceBuildSuper(p, &args, Ivy_ObjType(root), 1)
}

func vectorObjects(v *Vec_Ptr_t) []*Ivy_Obj_t {
	objs := make([]*Ivy_Obj_t, int(Vec_PtrSize(v)))
	for i := range objs {
		objs[i] = ivyVecObj(v, int32(i))
	}
	return objs
}

func setVectorObjects(v *Vec_Ptr_t, objs []*Ivy_Obj_t) {
	Vec_PtrClear(v)
	for _, obj := range objs {
		Vec_PtrPush(v, obj)
	}
}

func Ivy_NodeCountMffcRec(node *Ivy_Obj_t) int32 {
	if Ivy_ObjRefs(node) > 0 || Ivy_ObjIsCi(node) != 0 || node.fMarkA != 0 {
		return 0
	}
	assert(node.fMarkB != 0, "node.fMarkB != 0")
	node.fMarkA = 1
	if Ivy_ObjIsBuf(node) != 0 {
		return Ivy_NodeCountMffcRec(Ivy_ObjFanin0(node))
	}
	return 1 + Ivy_NodeCountMffcRec(Ivy_ObjFanin0(node)) + Ivy_NodeCountMffcRec(Ivy_ObjFanin1(node))
}

func Ivy_NodeCountMffc(node *Ivy_Obj_t) int32 {
	assert(node.fMarkB != 0, "node.fMarkB != 0")
	return 1 + Ivy_NodeCountMffcRec(Ivy_ObjFanin0(node)) + Ivy_NodeCountMffcRec(Ivy_ObjFanin1(node))
}

func Ivy_ManFindAlgCutCompare(a, b *Ivy_Obj_t) int {
	if Ivy_ObjId(a) < Ivy_ObjId(b) {
		return -1
	}
	if Ivy_ObjId(a) > Ivy_ObjId(b) {
		return 1
	}
	if Ivy_IsComplement(a) < Ivy_IsComplement(b) {
		return -1
	}
	if Ivy_IsComplement(a) > Ivy_IsComplement(b) {
		return 1
	}
	return 0
}

func Ivy_ManFindAlgCutRec(edge *Ivy_Obj_t, typ Ivy_Type_t, front, cone *Vec_Ptr_t) int32 {
	node := Ivy_Regular(edge)
	assert(Ivy_ObjIsBuf(node) == 0, "algebraic cut cannot contain buffers")
	assert(typ != IVY_EXOR || Ivy_IsComplement(edge) == 0, "EXOR cut root edges must be regular")
	if Ivy_IsComplement(edge) != 0 {
		if node.fMarkA != 0 {
			return -1
		}
		node.fMarkB = 1
	} else {
		if node.fMarkB != 0 {
			return -1
		}
		node.fMarkA = 1
	}
	Vec_PtrPush(cone, node)
	if Ivy_IsComplement(edge) != 0 || Ivy_ObjType(edge) != typ {
		if Ivy_ObjRefs(node) == 1 {
			return 1
		}
		assert(Ivy_ObjRefs(node) > 1, "boundary node has external references")
		Vec_PtrPush(front, edge)
		return 0
	}
	assert(Ivy_IsComplement(edge) == 0 && Ivy_ObjIsNode(edge) != 0, "internal node must be regular logic")
	r0 := Ivy_ManFindAlgCutRec(Ivy_ObjReal(Ivy_ObjChild0(edge)), typ, front, cone)
	r1 := Ivy_ManFindAlgCutRec(Ivy_ObjReal(Ivy_ObjChild1(edge)), typ, front, cone)
	if r0 == -1 || r1 == -1 {
		return -1
	}
	if r0 != 0 && r1 != 0 {
		if Ivy_ObjRefs(node) == 1 {
			return 1
		}
		assert(Ivy_ObjRefs(node) > 1, "internal node has external references")
		Vec_PtrPush(front, edge)
		return 0
	}
	if r0 != 0 {
		Vec_PtrPush(front, Ivy_ObjReal(Ivy_ObjChild0(edge)))
	}
	if r1 != 0 {
		Vec_PtrPush(front, Ivy_ObjReal(Ivy_ObjChild1(edge)))
	}
	return 0
}

func Ivy_ManFindAlgCut(root *Ivy_Obj_t, front, leaves, cone *Vec_Ptr_t) int32 {
	assert(Ivy_IsComplement(root) == 0 && Ivy_ObjIsNode(root) != 0, "invalid algebraic cut root")
	Vec_PtrClear(front)
	Vec_PtrClear(leaves)
	Vec_PtrClear(cone)
	result := Ivy_ManFindAlgCutRec(root, Ivy_ObjType(root), front, cone)
	for i := int32(0); i < Vec_PtrSize(cone); i++ {
		node := ivyVecObj(cone, i)
		node.fMarkA, node.fMarkB = 0, 0
	}
	if result != 0 {
		return result
	}
	if Vec_PtrSize(front) <= 2 {
		return 1
	}
	frontRaw := Vec_PtrArray(front)
	sort.Slice(frontRaw, func(i, j int) bool {
		return Ivy_ManFindAlgCutCompare(frontRaw[i].(*Ivy_Obj_t), frontRaw[j].(*Ivy_Obj_t)) < 0
	})
	var prev *Ivy_Obj_t
	for i := 0; i < len(frontRaw); i++ {
		obj := frontRaw[i].(*Ivy_Obj_t)
		if prev != nil && obj == prev {
			if Ivy_ObjIsExor(root) != 0 {
				if Vec_PtrSize(leaves) == 0 {
					prev = nil
				} else {
					Vec_PtrPop(leaves)
					if Vec_PtrSize(leaves) == 0 {
						prev = nil
					} else {
						prev = ivyVecObj(leaves, Vec_PtrSize(leaves)-1)
					}
				}
			}
			continue
		}
		if prev != nil && obj == Ivy_Not(prev) {
			assert(Ivy_ObjIsAnd(root) != 0, "opposite leaves only cancel in AND cut")
			return -1
		}
		prev = obj
		Vec_PtrPush(leaves, obj)
	}
	if Vec_PtrSize(leaves) == 0 {
		return -1
	}
	if Vec_PtrSize(leaves) <= 2 {
		return 1
	}
	return 0
}
