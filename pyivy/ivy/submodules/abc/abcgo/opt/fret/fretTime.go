package fret

import (
	"fmt"
	"os"

	abc "github.com/glycerine/abcgo/base/abc"
	vec "github.com/glycerine/abcgo/misc/vec"
)

// trace2 prints a node and follows the first fanin whose level can lie on a
// longest path to it. This is the debug helper from fretTime.c.
func trace2(obj *abc.Abc_Obj_t) {
	print_node(obj)
	for _, edge := range obj.Fanins() {
		next := edge.Object
		if next.Level >= obj.Level-1 {
			trace2(next)
			break
		}
	}
}

// Timing constraints are stored per object ID. Each edge points to a node on
// the critical path from the constrained object to a latch boundary.
func Abc_FlowRetime_InitTiming(ntk *abc.Abc_Ntk_t) {
	pManMR.nConservConstraints = 0
	pManMR.nExactConstraints = 0
	pManMR.vExactNodes = vec.Vec_PtrAlloc(1000)
	pManMR.vTimeEdges = make([]vec.Vec_Ptr_t, int(abc.Abc_NtkObjNumMax(ntk)+1))
}

func Abc_FlowRetime_ConstrainConserv(ntk *abc.Abc_Ntk_t) {
	clearExactConstraints()
	if pManMR.fIsForward != 0 {
		Abc_FlowRetime_ConstrainConserv_forw(ntk)
	} else {
		Abc_FlowRetime_ConstrainConserv_back(ntk)
	}
	for _, obj := range ntk.Objects() {
		assert(vec.Vec_PtrSize(FTIMEEDGES(obj)) == 0, "no exact edges after conservative constraint pass")
	}
}

func clearExactConstraints() {
	pManMR.nExactConstraints = 0
	for vec.Vec_PtrSize(pManMR.vExactNodes) > 0 {
		obj, ok := vec.Vec_PtrPop(pManMR.vExactNodes).(*abc.Abc_Obj_t)
		assert(ok && obj != nil, "exact-node list contains an object")
		vec.Vec_PtrClear(FTIMEEDGES(obj))
	}
}

func Abc_FlowRetime_ConstrainConserv_forw(ntk *abc.Abc_Ntk_t) {
	assert(vec.Vec_PtrSize(pManMR.vNodes) == 0, "scratch node vector is empty")
	pManMR.nConservConstraints = 0

	// First measure the combinational depth from primary inputs.
	abc.Abc_NtkIncrementTravId(ntk)
	for _, pi := range ntk.ObjectsOfKind(abc.ABC_OBJ_PI) {
		Abc_FlowRetime_Dfs_forw(pi, pManMR.vNodes)
	}
	for _, obj := range reverseVecObjects(pManMR.vNodes) {
		level := maxFaninLevel(obj, true)
		setObjectLevel(obj, level)
		if abc.Abc_ObjIsBi(obj) != 0 {
			obj.MarkA = true
		}
		assert(level <= pManMR.maxDelay, "PI arrival depth is within max delay")
	}

	// Seed latch outputs from the arrival time at each BI.
	vec.Vec_PtrClear(pManMR.vNodes)
	abc.Abc_NtkIncrementTravId(ntk)
	for _, latch := range ntk.ObjectsOfKind(abc.ABC_OBJ_LATCH) {
		bo, bi := abc.Abc_ObjFanout0(latch), abc.Abc_ObjFanin0(latch)
		abc.Abc_NodeSetTravIdCurrent(latch)
		Abc_FlowRetime_Dfs_forw(bo, pManMR.vNodes)
		if bi.MarkA {
			bi.MarkA = false
			setObjectLevel(latch, objectLevel(bi))
		} else {
			setObjectLevel(latch, 0)
		}
		assert(objectLevel(latch) <= pManMR.maxDelay, "latch arrival depth is within max delay")
	}
	checkForwardOrder(pManMR.vNodes)

	// Mark nodes that violate the hard delay limit.
	for _, obj := range reverseVecObjects(pManMR.vNodes) {
		level := maxFaninLevel(obj, true)
		setObjectLevel(obj, level)
		if level > pManMR.maxDelay {
			FSET(obj, BLOCK)
		}
	}

	// Start from T=0 at each latch, then propagate the conservative timing
	// envelope through the latch cones.
	for _, latch := range ntk.ObjectsOfKind(abc.ABC_OBJ_LATCH) {
		setObjectLevel(latch, 0)
	}
	for _, obj := range reverseVecObjects(pManMR.vNodes) {
		level := maxFaninLevel(obj, true)
		setObjectLevel(obj, level)
		if abc.Abc_ObjIsBi(obj) != 0 {
			obj.MarkA = true
		}
	}
	for _, latch := range ntk.ObjectsOfKind(abc.ABC_OBJ_LATCH) {
		bi := abc.Abc_ObjFanin0(latch)
		if bi.MarkA {
			bi.MarkA = false
			setObjectLevel(latch, objectLevel(bi))
		} else {
			setObjectLevel(latch, 0)
		}
		assert(objectLevel(latch) <= pManMR.maxDelay, "latch output depth is within max delay")
	}
	for _, obj := range reverseVecObjects(pManMR.vNodes) {
		level := maxFaninLevel(obj, true)
		setObjectLevel(obj, level)
		if level > pManMR.maxDelay {
			FSET(obj, CONSERVATIVE)
			pManMR.nConservConstraints++
		} else {
			FUNSET(obj, CONSERVATIVE)
		}
	}
	vec.Vec_PtrClear(pManMR.vNodes)
}

func Abc_FlowRetime_ConstrainConserv_back(ntk *abc.Abc_Ntk_t) {
	assert(vec.Vec_PtrSize(pManMR.vNodes) == 0, "scratch node vector is empty")
	pManMR.nConservConstraints = 0

	// First measure combinational depth from primary outputs, traversing toward
	// their drivers.
	abc.Abc_NtkIncrementTravId(ntk)
	for _, po := range ntk.ObjectsOfKind(abc.ABC_OBJ_PO) {
		Abc_FlowRetime_Dfs_back(po, pManMR.vNodes)
	}
	for _, obj := range reverseVecObjects(pManMR.vNodes) {
		level := maxFanoutLevel(obj, true)
		setObjectLevel(obj, level)
		if abc.Abc_ObjIsBo(obj) != 0 {
			obj.MarkA = true
		}
		assert(level <= pManMR.maxDelay, "PO arrival depth is within max delay")
	}

	vec.Vec_PtrClear(pManMR.vNodes)
	abc.Abc_NtkIncrementTravId(ntk)
	for _, latch := range ntk.ObjectsOfKind(abc.ABC_OBJ_LATCH) {
		bo, bi := abc.Abc_ObjFanout0(latch), abc.Abc_ObjFanin0(latch)
		abc.Abc_NodeSetTravIdCurrent(latch)
		Abc_FlowRetime_Dfs_back(bi, pManMR.vNodes)
		if bo.MarkA {
			bo.MarkA = false
			setObjectLevel(latch, objectLevel(bo))
		} else {
			setObjectLevel(latch, 0)
		}
		assert(objectLevel(latch) <= pManMR.maxDelay, "latch output depth is within max delay")
	}
	checkBackwardOrder(pManMR.vNodes)
	for _, obj := range reverseVecObjects(pManMR.vNodes) {
		level := maxFanoutLevel(obj, true)
		setObjectLevel(obj, level)
		if level+boolInt(abc.Abc_ObjIsNode(obj) != 0) > pManMR.maxDelay {
			FSET(obj, BLOCK)
		}
	}

	for _, latch := range ntk.ObjectsOfKind(abc.ABC_OBJ_LATCH) {
		setObjectLevel(latch, 0)
	}
	for _, obj := range reverseVecObjects(pManMR.vNodes) {
		level := maxFanoutLevel(obj, true)
		setObjectLevel(obj, level)
		if abc.Abc_ObjIsBo(obj) != 0 {
			obj.MarkA = true
		}
	}
	for _, latch := range ntk.ObjectsOfKind(abc.ABC_OBJ_LATCH) {
		bo := abc.Abc_ObjFanout0(latch)
		if bo.MarkA {
			bo.MarkA = false
			setObjectLevel(latch, objectLevel(bo))
		} else {
			setObjectLevel(latch, 0)
		}
	}
	for _, obj := range reverseVecObjects(pManMR.vNodes) {
		level := maxFanoutLevel(obj, true)
		setObjectLevel(obj, level)
		if level > pManMR.maxDelay {
			FSET(obj, CONSERVATIVE)
			pManMR.nConservConstraints++
		} else {
			FUNSET(obj, CONSERVATIVE)
		}
	}
	vec.Vec_PtrClear(pManMR.vNodes)
}

func objectLevel(obj *abc.Abc_Obj_t) int32           { return int32(obj.Level) }
func setObjectLevel(obj *abc.Abc_Obj_t, level int32) { obj.Level = int(level) }

func maxFaninLevel(obj *abc.Abc_Obj_t, onlyCurrent bool) int32 {
	var level int32
	for _, edge := range obj.Fanins() {
		fanin := edge.Object
		if onlyCurrent && abc.Abc_NodeIsTravIdCurrent(fanin) == 0 {
			continue
		}
		if objectLevel(fanin) > level {
			level = objectLevel(fanin)
		}
	}
	if abc.Abc_ObjIsNode(obj) != 0 {
		level++
	}
	return level
}

func maxFanoutLevel(obj *abc.Abc_Obj_t, onlyCurrent bool) int32 {
	var level int32
	for _, fanout := range obj.Fanouts() {
		if onlyCurrent && abc.Abc_NodeIsTravIdCurrent(fanout) == 0 {
			continue
		}
		candidate := objectLevel(fanout)
		if abc.Abc_ObjIsNode(obj) != 0 {
			candidate++
		}
		if candidate > level {
			level = candidate
		}
	}
	return level
}

func reverseVecObjects(nodes *vec.Vec_Ptr_t) []*abc.Abc_Obj_t {
	result := make([]*abc.Abc_Obj_t, 0, vec.Vec_PtrSize(nodes))
	for i := vec.Vec_PtrSize(nodes) - 1; i >= 0; i-- {
		obj, ok := vec.Vec_PtrEntry(nodes, i).(*abc.Abc_Obj_t)
		assert(ok && obj != nil, "DFS node vector contains an object")
		result = append(result, obj)
	}
	return result
}

func checkForwardOrder(nodes *vec.Vec_Ptr_t) {
	ordered := reverseVecObjects(nodes)
	for _, obj := range ordered {
		obj.MarkB = true
	}
	for _, obj := range ordered {
		for _, edge := range obj.Fanins() {
			fanin := edge.Object
			if abc.Abc_NodeIsTravIdCurrent(fanin) != 0 && abc.Abc_ObjIsLatch(fanin) == 0 {
				assert(fanin.MarkB, "forward DFS places each fanin before its consumer")
			}
		}
	}
	for _, obj := range ordered {
		obj.MarkB = false
	}
}

func checkBackwardOrder(nodes *vec.Vec_Ptr_t) {
	ordered := reverseVecObjects(nodes)
	for _, obj := range ordered {
		obj.MarkB = true
	}
	for _, obj := range ordered {
		for _, fanout := range obj.Fanouts() {
			if abc.Abc_NodeIsTravIdCurrent(fanout) != 0 && abc.Abc_ObjIsLatch(fanout) == 0 {
				assert(fanout.MarkB, "backward DFS places each fanout before its consumer")
			}
		}
	}
	for _, obj := range ordered {
		obj.MarkB = false
	}
}

func Abc_FlowRetime_ConstrainExact(obj *abc.Abc_Obj_t) {
	if FTEST(obj, CONSERVATIVE) != 0 {
		pManMR.nConservConstraints--
		FUNSET(obj, CONSERVATIVE)
	}
	if pManMR.fIsForward != 0 {
		constrainExact(obj, true)
	} else {
		constrainExact(obj, false)
	}
}

type timingPathNode struct {
	obj          *abc.Abc_Obj_t
	throughLatch bool
}

func constrainExact(root *abc.Abc_Obj_t, forward bool) {
	assert(vec.Vec_PtrSize(pManMR.vNodes) == 0, "scratch node vector is empty")
	assert(abc.Abc_ObjIsLatch(root) == 0, "exact timing root is not a latch")
	assert(vec.Vec_PtrSize(FTIMEEDGES(root)) == 0, "exact timing edges were cleared")
	vec.Vec_PtrPush(pManMR.vExactNodes, root)
	paths := collectTimingPaths(root, forward)

	// DFS appends after visiting neighbors. Walk it backward so each object's
	// level is propagated to its fanins or fanouts before those are processed.
	for i := len(paths) - 1; i >= 0; i-- {
		path := paths[i]
		if path.throughLatch {
			continue
		}
		propagateExactLevel(path.obj, forward, false)
		assert(objectLevel(path.obj) <= pManMR.maxDelay, "unregistered path is within max delay")
		path.obj.Level = 0
		path.obj.MarkA, path.obj.MarkB = false, false
	}
	for i := len(paths) - 1; i >= 0; i-- {
		path := paths[i]
		if !path.throughLatch {
			continue
		}
		propagateExactLevel(path.obj, forward, true)
		if objectLevel(path.obj) == pManMR.maxDelay {
			vec.Vec_PtrPush(FTIMEEDGES(root), path.obj)
			pManMR.nExactConstraints++
		}
		path.obj.Level = 0
		path.obj.MarkA, path.obj.MarkB = false, false
	}
}

func propagateExactLevel(obj *abc.Abc_Obj_t, forward, skipLatchNeighbors bool) {
	level := objectLevel(obj)
	if abc.Abc_ObjIsNode(obj) != 0 {
		level++
	}
	for _, neighbor := range timingNeighbors(obj, forward) {
		if skipLatchNeighbors && abc.Abc_ObjIsLatch(neighbor.obj) != 0 {
			continue
		}
		if level > objectLevel(neighbor.obj) {
			setObjectLevel(neighbor.obj, level)
		}
	}
}

func collectTimingPaths(root *abc.Abc_Obj_t, forward bool) []timingPathNode {
	type visitKey struct {
		obj          *abc.Abc_Obj_t
		throughLatch bool
	}
	visited := make(map[visitKey]bool)
	var ordered []timingPathNode
	var visit func(*abc.Abc_Obj_t, bool)
	visit = func(obj *abc.Abc_Obj_t, crossed bool) {
		if abc.Abc_ObjIsLatch(obj) != 0 {
			if crossed {
				return
			}
			crossed = true
		}
		key := visitKey{obj: obj, throughLatch: crossed}
		if visited[key] {
			return
		}
		visited[key] = true
		obj.Level = 0
		obj.MarkA = !crossed
		obj.MarkB = crossed
		for _, next := range timingNeighbors(obj, forward) {
			visit(next.obj, crossed)
		}
		ordered = append(ordered, timingPathNode{obj: obj, throughLatch: crossed})
	}
	visit(root, false)
	return ordered
}

func timingNeighbors(obj *abc.Abc_Obj_t, forward bool) []timingPathNode {
	var neighbors []*abc.Abc_Obj_t
	if forward {
		neighbors = faninObjects(obj)
	} else {
		neighbors = obj.Fanouts()
	}
	result := make([]timingPathNode, 0, len(neighbors))
	for _, next := range neighbors {
		result = append(result, timingPathNode{obj: next})
	}
	return result
}

func Abc_FlowRetime_ConstrainExact_forw(obj *abc.Abc_Obj_t) { constrainExact(obj, true) }
func Abc_FlowRetime_ConstrainExact_back(obj *abc.Abc_Obj_t) { constrainExact(obj, false) }

func Abc_FlowRetime_ConstrainExactAll(ntk *abc.Abc_Ntk_t) {
	for _, obj := range ntk.Objects() {
		vec.Vec_PtrClear(FTIMEEDGES(obj))
	}
	pManMR.nExactConstraints = 0
	for _, obj := range ntk.Objects() {
		if abc.Abc_ObjIsLatch(obj) == 0 && FTEST(obj, CONSERVATIVE) != 0 && FTEST(obj, BLOCK) == 0 && vec.Vec_PtrSize(FTIMEEDGES(obj)) == 0 {
			Abc_FlowRetime_ConstrainExact(obj)
		}
	}
}

func Abc_FlowRetime_FreeTiming(ntk *abc.Abc_Ntk_t) {
	clearExactConstraints()
	pManMR.vExactNodes = nil
	pManMR.vTimeEdges = nil
	_ = ntk
}

func Abc_FlowRetime_Dfs_forw(obj *abc.Abc_Obj_t, nodes *vec.Vec_Ptr_t) {
	if abc.Abc_ObjIsLatch(obj) != 0 || abc.Abc_NodeIsTravIdCurrent(obj) != 0 {
		return
	}
	abc.Abc_NodeSetTravIdCurrent(obj)
	for _, next := range obj.Fanouts() {
		Abc_FlowRetime_Dfs_forw(next, nodes)
	}
	vec.Vec_PtrPush(nodes, obj)
}

func Abc_FlowRetime_Dfs_back(obj *abc.Abc_Obj_t, nodes *vec.Vec_Ptr_t) {
	if abc.Abc_ObjIsLatch(obj) != 0 || abc.Abc_NodeIsTravIdCurrent(obj) != 0 {
		return
	}
	abc.Abc_NodeSetTravIdCurrent(obj)
	for _, edge := range obj.Fanins() {
		Abc_FlowRetime_Dfs_back(edge.Object, nodes)
	}
	vec.Vec_PtrPush(nodes, obj)
}

func Abc_FlowRetime_RefineConstraints() int32 {
	ntk := pManMR.pNtk
	var count int32
	pManMR.constraintMask = BLOCK | CONSERVATIVE
	fretPrintf("\t\trefinement: over ")
	flow := Abc_FlowRetime_PushFlows(ntk, 0)
	fretPrintf("= %d ", flow)
	for _, obj := range ntk.Objects() {
		if (pManMR.fIsForward != 0 && FTEST(obj, VISITED_R) == 0) || (pManMR.fIsForward == 0 && FTEST(obj, VISITED_E) == 0) {
			obj.MarkC = true
		}
	}
	if pManMR.fConservTimingOnly != 0 {
		fretPrintf(" done\n")
		return 0
	}
	pManMR.constraintMask = BLOCK
	Abc_FlowRetime_ClearFlows(0)
	fretPrintf("under = ")
	flow = Abc_FlowRetime_PushFlows(ntk, 0)
	fretPrintf("%d refined nodes = ", flow)
	for _, obj := range ntk.Objects() {
		visited := FTEST(obj, VISITED_E)
		if pManMR.fIsForward != 0 {
			visited = FTEST(obj, VISITED_R)
		}
		if obj.MarkC && visited != 0 && FTEST(obj, CONSERVATIVE) != 0 && count < 99999 {
			count++
			Abc_FlowRetime_ConstrainExact(obj)
		}
		obj.MarkC = false
	}
	fretPrintf("%d\n", count)
	if count > 0 {
		return 1
	}
	return 0
}

func fretPrintf(format string, args ...any) {
	if pManMR != nil && pManMR.fVerbose != 0 {
		fmt.Fprintf(os.Stdout, format, args...)
	}
}

func faninObjects(obj *abc.Abc_Obj_t) []*abc.Abc_Obj_t {
	result := make([]*abc.Abc_Obj_t, 0, abc.Abc_ObjFaninNum(obj))
	for _, edge := range obj.Fanins() {
		result = append(result, edge.Object)
	}
	return result
}
