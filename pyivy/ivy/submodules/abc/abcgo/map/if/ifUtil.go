package if_

import "fmt"

// If_ManCleanNodeCopy clears temporary node-copy references.
func If_ManCleanNodeCopy(p *If_Man_t) {
	for _, entry := range Vec_PtrArray(p.vObjs) {
		If_ObjSetCopy(entry.(*If_Obj_t), nil)
	}
}

// If_ManCleanCutData clears data attached to each best cut.
func If_ManCleanCutData(p *If_Man_t) {
	for _, entry := range Vec_PtrArray(p.vObjs) {
		If_CutSetData(If_ObjCutBest(entry.(*If_Obj_t)), nil)
	}
}

// If_ManCleanMarkV clears temporary visit marks.
func If_ManCleanMarkV(p *If_Man_t) {
	for _, entry := range Vec_PtrArray(p.vObjs) {
		entry.(*If_Obj_t).fVisit = 0
	}
}

// If_ManResetOriginalRefs recomputes fanout counts from graph edges.
func If_ManResetOriginalRefs(p *If_Man_t) {
	for _, entry := range Vec_PtrArray(p.vObjs) {
		entry.(*If_Obj_t).nRefs = 0
	}
	for _, entry := range Vec_PtrArray(p.vObjs) {
		obj := entry.(*If_Obj_t)
		switch obj.Type {
		case IF_AND:
			obj.pFanin0.nRefs++
			obj.pFanin1.nRefs++
		case IF_CO:
			obj.pFanin0.nRefs++
		}
	}
}

// If_ManCrossCut computes the maximum number of simultaneously live nodes.
// Visit counts are restored before returning, including when the invariant fails.
func If_ManCrossCut(p *If_Man_t) int32 {
	var size, maximum int32
	objs := Vec_PtrArray(p.vObjs)
	defer func() {
		for _, entry := range objs {
			entry.(*If_Obj_t).nVisits = entry.(*If_Obj_t).nVisitsCopy
		}
	}()
	for _, entry := range objs {
		obj := entry.(*If_Obj_t)
		if obj.Type != IF_AND {
			continue
		}
		size++
		if size > maximum {
			maximum = size
		}
		if obj.nVisits == 0 {
			size--
		}
		for _, fanin := range []*If_Obj_t{obj.pFanin0, obj.pFanin1} {
			if fanin.Type != IF_CI {
				fanin.nVisits--
				if fanin.nVisits == 0 {
					size--
				}
			}
		}
		if obj.fRepr != 0 {
			for choice := obj; choice != nil; choice = choice.pEquiv {
				if choice.Type != IF_CI {
					choice.nVisits--
					if choice.nVisits == 0 {
						size--
					}
				}
			}
		}
	}
	for _, entry := range objs {
		obj := entry.(*If_Obj_t)
		assert(obj.Type == IF_CI || obj.fVisit == 0, "obj.Type == IF_CI || obj.fVisit == 0")
	}
	assert(size == 0, "cross-cut sweep returned to zero")
	return maximum
}

// If_ManReverseOrder groups nodes by level, descending, preserving the C
// implementation's reverse insertion order within each level.
func If_ManReverseOrder(p *If_Man_t) *Vec_Ptr_t {
	if p.nLevelMax < 0 {
		panic("if: negative maximum node level")
	}
	byLevel := make([][]*If_Obj_t, int(p.nLevelMax)+1)
	for _, entry := range Vec_PtrArray(p.vObjs) {
		obj := entry.(*If_Obj_t)
		assert(obj.Level <= uint32(p.nLevelMax), "obj.Level <= p.nLevelMax")
		level := int(obj.Level)
		byLevel[level] = append(byLevel[level], obj)
	}
	order := Vec_PtrAlloc(int(If_ManObjNum(p)))
	for level := len(byLevel) - 1; level >= 0; level-- {
		bucket := byLevel[level]
		for i := len(bucket) - 1; i >= 0; i-- {
			Vec_PtrPush(order, bucket[i])
		}
	}
	return order
}

// If_ManMarkMapping_rec marks the transitive fanin of the chosen cuts and
// returns the area contribution of newly referenced nodes.
func If_ManMarkMapping_rec(p *If_Man_t, obj *If_Obj_t) float32 {
	alreadyReferenced := obj.nRefs != 0
	obj.nRefs++
	if alreadyReferenced || obj.Type == IF_CI || obj.Type == IF_CONST1 {
		return 0
	}
	assert(obj.Type == IF_AND, "mapped internal node is an AND")
	cut := If_ObjCutBest(obj)
	p.nNets += int32(cut.nLeaves)
	area := If_CutLutArea(p, cut)
	for _, leafID := range cut.pLeaves {
		if p.pPars.fLiftLeaves != 0 {
			leafID >>= 8
		}
		leaf := If_ManObj(p, leafID)
		if p.vSwitching != nil && int(leaf.Id) < Vec_IntSize(p.vSwitching) {
			p.dPower += float32(Vec_IntEntry(p.vSwitching, int(leaf.Id)))
		}
		area += If_ManMarkMapping_rec(p, leaf)
	}
	return area
}

// If_ManMarkMapping computes references, area, and power for the selected map.
func If_ManMarkMapping(p *If_Man_t) {
	for _, entry := range Vec_PtrArray(p.vObjs) {
		obj := entry.(*If_Obj_t)
		obj.Required = IF_FLOAT_LARGE
		obj.nVisits = obj.nVisitsCopy
		obj.nRefs = 0
	}
	p.nNets = 0
	p.dPower = 0
	p.AreaGlo = 0
	for _, entry := range Vec_PtrArray(p.vCos) {
		p.AreaGlo += If_ManMarkMapping_rec(p, If_ObjFanin0(entry.(*If_Obj_t)))
	}
}

// If_ManCollectMappingDirect returns mapped AND nodes in graph order.
func If_ManCollectMappingDirect(p *If_Man_t) *Vec_Ptr_t {
	If_ManMarkMapping(p)
	order := Vec_PtrAlloc(int(If_ManObjNum(p)))
	for _, entry := range Vec_PtrArray(p.vObjs) {
		obj := entry.(*If_Obj_t)
		if obj.Type == IF_AND && obj.nRefs != 0 {
			Vec_PtrPush(order, obj)
		}
	}
	return order
}

// If_ManCollectMappingInt encodes each mapped cut as leaf count, leaf IDs,
// and root ID, in graph order.
func If_ManCollectMappingInt(p *If_Man_t) *Vec_Int_t {
	If_ManMarkMapping(p)
	order := Vec_IntAlloc(int(If_ManObjNum(p)))
	for _, entry := range Vec_PtrArray(p.vObjs) {
		obj := entry.(*If_Obj_t)
		if obj.Type != IF_AND || obj.nRefs == 0 {
			continue
		}
		cut := If_ObjCutBest(obj)
		Vec_IntPush(order, int32(cut.nLeaves))
		for _, leafID := range cut.pLeaves {
			Vec_IntPush(order, leafID)
		}
		Vec_IntPush(order, obj.Id)
	}
	return order
}

// If_ManCountSpecialPos counts complemented outputs whose regular driver also
// feeds at least one non-complemented output.
func If_ManCountSpecialPos(p *If_Man_t) int32 {
	start := int(p.pPars.nLatchesCoBox)
	stop := If_ManCoNum(p) - p.pPars.nLatchesCo
	if start < 0 {
		start = 0
	}
	if stop > int32(Vec_PtrSize(p.vCos)) {
		stop = int32(Vec_PtrSize(p.vCos))
	}
	if start > int(stop) {
		return 0
	}
	outputs := Vec_PtrArray(p.vCos)
	for i := start; i < int(stop); i++ {
		If_ObjFanin0(outputs[i].(*If_Obj_t)).fMark = 0
	}
	for i := start; i < int(stop); i++ {
		co := outputs[i].(*If_Obj_t)
		if co.fCompl0 == 0 {
			If_ObjFanin0(co).fMark = 1
		}
	}
	var count int32
	for i := start; i < int(stop); i++ {
		co := outputs[i].(*If_Obj_t)
		if co.fCompl0 != 0 && If_ObjFanin0(co).fMark != 0 {
			count++
		}
	}
	for i := start; i < int(stop); i++ {
		If_ObjFanin0(outputs[i].(*If_Obj_t)).fMark = 0
	}
	return count
}

// If_CutTraverse_rec appends each node once in fanin-first order.
func If_CutTraverse_rec(node *If_Obj_t, nodes *Vec_Ptr_t) {
	if node.fMark != 0 {
		return
	}
	node.fMark = 1
	if node.Type == IF_AND {
		If_CutTraverse_rec(node.pFanin0, nodes)
		If_CutTraverse_rec(node.pFanin1, nodes)
	}
	Vec_PtrPush(nodes, node)
}

// If_CutTraverse returns the cut leaves and internal nodes below root, then
// clears all temporary marks it set.
func If_CutTraverse(p *If_Man_t, root *If_Obj_t, cut *If_Cut_t, nodes *Vec_Ptr_t) {
	Vec_PtrClear(nodes)
	for _, leafID := range cut.pLeaves[:cut.nLeaves] {
		leaf := If_ManObj(p, leafID)
		assert(leaf.fMark == 0, "leaf mark is clear")
		leaf.fMark = 1
		Vec_PtrPush(nodes, leaf)
	}
	If_CutTraverse_rec(root, nodes)
	for _, entry := range Vec_PtrArray(nodes) {
		entry.(*If_Obj_t).fMark = 0
	}
}

// If_CutTraverseTest runs a traversal for diagnostics.
func If_CutTraverseTest(p *If_Man_t, root *If_Obj_t, cut *If_Cut_t) {
	nodes := Vec_PtrAlloc(1000)
	If_CutTraverse(p, root, cut, nodes)
	Vec_PtrFree(nodes)
}

// If_ObjPrint prints a compact summary of an object.
func If_ObjPrint(obj *If_Obj_t) {
	if obj == nil {
		fmt.Print("Object is NULL.")
		return
	}
	fmt.Printf("Obj %4d : ", If_ObjId(obj))
	switch obj.Type {
	case IF_CONST1:
		fmt.Print("constant 1")
	case IF_CI:
		fmt.Print("PI")
	case IF_CO:
		fmt.Printf("PO( %4d%s )", If_ObjId(obj.pFanin0), complementSuffix(obj.fCompl0))
	default:
		fmt.Printf("AND( %4d%s, %4d%s )", If_ObjId(obj.pFanin0), complementSuffix(obj.fCompl0), If_ObjId(obj.pFanin1), complementSuffix(obj.fCompl1))
	}
	fmt.Printf(" (refs = %3d)\n", obj.nVisitsCopy)
}

func complementSuffix(complement uint32) string {
	if complement != 0 {
		return "'"
	}
	return " "
}
