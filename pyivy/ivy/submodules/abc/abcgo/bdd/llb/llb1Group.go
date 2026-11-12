package llb

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

func Llb_ManGroupAlloc(manager *Llb_Man_t) *Llb_Grp_t {
	if manager.vGroups == nil {
		manager.vGroups = vec.Vec_PtrAlloc(8)
	}
	group := &Llb_Grp_t_{
		pMan:  manager,
		vIns:  vec.Vec_PtrAlloc(8),
		vOuts: vec.Vec_PtrAlloc(8),
		Id:    int32(vec.Vec_PtrSize(manager.vGroups)),
	}
	vec.Vec_PtrPush(manager.vGroups, group)
	return group
}

func Llb_ManGroupStop(group *Llb_Grp_t) {
	if group == nil {
		return
	}
	if group.pMan != nil && group.pMan.vGroups != nil && int(group.Id) < vec.Vec_PtrSize(group.pMan.vGroups) {
		vec.Vec_PtrWriteEntry(group.pMan.vGroups, int(group.Id), nil)
	}
	vec.Vec_PtrFree(group.vIns)
	vec.Vec_PtrFree(group.vOuts)
	vec.Vec_PtrFree(group.vNodes)
	group.vIns, group.vOuts, group.vNodes = nil, nil, nil
}

func llbGroupEdge(v *Vec_Ptr_t, index int) aig.Edge {
	value := vec.Vec_PtrEntry(v, index)
	switch edge := value.(type) {
	case aig.Edge:
		return edge
	case *aig.Edge:
		return *edge
	default:
		panic(fmt.Sprintf("llb: group entry has type %T, want aig.Edge", value))
	}
}

func llbGroupPush(v *Vec_Ptr_t, edge aig.Edge) { vec.Vec_PtrPush(v, edge) }

func Llb_ManGroupCollect_rec(graph *Aig_Man_t, edge aig.Edge, nodes *Vec_Ptr_t) {
	if aig.Aig_ObjIsTravIdCurrent(graph, edge) {
		return
	}
	aig.Aig_ObjSetTravIdCurrent(graph, edge)
	node := graph.Node(aig.Aig_Regular(edge))
	if node.Kind == aig.Constant {
		return
	}
	if node.Kind == aig.Output {
		Llb_ManGroupCollect_rec(graph, node.Fanins[0], nodes)
		return
	}
	if node.Kind == aig.Input {
		return
	}
	if node.Kind != aig.And && node.Kind != aig.ExclusiveOr && node.Kind != aig.Buffer {
		panic("llb: unsupported node in group cone")
	}
	Llb_ManGroupCollect_rec(graph, node.Fanins[0], nodes)
	if node.Kind != aig.Buffer {
		Llb_ManGroupCollect_rec(graph, node.Fanins[1], nodes)
	}
	llbGroupPush(nodes, aig.Aig_Regular(edge))
}

// Llb_ManGroupCollect returns internal nodes between the group's boundary sets.
func Llb_ManGroupCollect(group *Llb_Grp_t) *Vec_Ptr_t {
	graph := group.pMan.pAig
	nodes := vec.Vec_PtrAlloc(100)
	aig.Aig_ManIncrementTravId(graph)
	for i := 0; i < vec.Vec_PtrSize(group.vIns); i++ {
		aig.Aig_ObjSetTravIdCurrent(graph, llbGroupEdge(group.vIns, i))
	}
	for i := 0; i < vec.Vec_PtrSize(group.vOuts); i++ {
		aig.Aig_ObjSetTravIdPrevious(graph, llbGroupEdge(group.vOuts, i))
	}
	for i := 0; i < vec.Vec_PtrSize(group.vOuts); i++ {
		Llb_ManGroupCollect_rec(graph, llbGroupEdge(group.vOuts, i), nodes)
	}
	return nodes
}

func Llb_ManGroupCreate_rec(graph *Aig_Man_t, edge aig.Edge, support *Vec_Ptr_t) {
	if aig.Aig_ObjIsTravIdCurrent(graph, edge) {
		return
	}
	aig.Aig_ObjSetTravIdCurrent(graph, edge)
	node := graph.Node(aig.Aig_Regular(edge))
	if node.Kind == aig.Constant {
		return
	}
	if aig.Aig_ObjIsMarkA(graph, edge) || node.Kind == aig.Input {
		llbGroupPush(support, aig.Aig_Regular(edge))
		return
	}
	if node.Kind == aig.Output {
		Llb_ManGroupCreate_rec(graph, node.Fanins[0], support)
		return
	}
	if node.Kind != aig.And && node.Kind != aig.ExclusiveOr && node.Kind != aig.Buffer {
		panic("llb: unsupported node while deriving group support")
	}
	Llb_ManGroupCreate_rec(graph, node.Fanins[0], support)
	if node.Kind != aig.Buffer {
		Llb_ManGroupCreate_rec(graph, node.Fanins[1], support)
	}
}

func Llb_ManGroupCreate(manager *Llb_Man_t, edge aig.Edge) *Llb_Grp_t {
	assert(aig.Aig_ObjIsMarkA(manager.pAig, edge), "group root is marked")
	group := Llb_ManGroupAlloc(manager)
	llbGroupPush(group.vOuts, aig.Aig_Regular(edge))
	aig.Aig_ManIncrementTravId(manager.pAig)
	node := manager.pAig.Node(edge)
	if node.Kind == aig.Output {
		Llb_ManGroupCreate_rec(manager.pAig, node.Fanins[0], group.vIns)
	} else {
		Llb_ManGroupCreate_rec(manager.pAig, node.Fanins[0], group.vIns)
		if node.Kind != aig.Buffer {
			Llb_ManGroupCreate_rec(manager.pAig, node.Fanins[1], group.vIns)
		}
	}
	group.vNodes = Llb_ManGroupCollect(group)
	return group
}

func Llb_ManGroupCreateFirst(manager *Llb_Man_t) *Llb_Grp_t {
	group := Llb_ManGroupAlloc(manager)
	inputs := manager.pAig.Inputs()
	for i := len(inputs) - int(manager.pAig.RegisterCount()); i < len(inputs); i++ {
		if i >= 0 {
			llbGroupPush(group.vOuts, inputs[i])
		}
	}
	return group
}

func Llb_ManGroupCreateLast(manager *Llb_Man_t) *Llb_Grp_t {
	group := Llb_ManGroupAlloc(manager)
	outputs := manager.pAig.Outputs()
	for i := len(outputs) - int(manager.pAig.RegisterCount()); i < len(outputs); i++ {
		if i >= 0 {
			llbGroupPush(group.vIns, outputs[i])
		}
	}
	return group
}

func Llb_ManGroupsCombine(first, second *Llb_Grp_t) *Llb_Grp_t {
	group := Llb_ManGroupAlloc(first.pMan)
	appendEntries := func(dst, src *Vec_Ptr_t, unique bool) {
		for i := 0; i < vec.Vec_PtrSize(src); i++ {
			edge := llbGroupEdge(src, i)
			if unique {
				found := false
				for j := 0; j < vec.Vec_PtrSize(dst); j++ {
					if llbGroupEdge(dst, j) == edge {
						found = true
						break
					}
				}
				if found {
					continue
				}
			}
			llbGroupPush(dst, edge)
		}
	}
	appendEntries(group.vIns, first.vIns, false)
	appendEntries(group.vIns, second.vIns, true)
	appendEntries(group.vOuts, first.vOuts, false)
	appendEntries(group.vOuts, second.vOuts, true)
	group.vNodes = Llb_ManGroupCollect(group)
	return group
}

func Llb_ManGroupMarkNodes_rec(graph *Aig_Man_t, edge aig.Edge) {
	if aig.Aig_ObjIsTravIdCurrent(graph, edge) {
		return
	}
	if aig.Aig_ObjIsTravIdPrevious(graph, edge) {
		aig.Aig_ObjSetTravIdCurrent(graph, edge)
		return
	}
	aig.Aig_ObjSetTravIdCurrent(graph, edge)
	node := graph.Node(aig.Aig_Regular(edge))
	if node.Kind == aig.Output || node.Kind == aig.Buffer {
		Llb_ManGroupMarkNodes_rec(graph, node.Fanins[0])
		return
	}
	if node.Kind != aig.And && node.Kind != aig.ExclusiveOr {
		panic("llb: expected internal node in cut cone")
	}
	Llb_ManGroupMarkNodes_rec(graph, node.Fanins[0])
	Llb_ManGroupMarkNodes_rec(graph, node.Fanins[1])
}

func Llb_ManGroupCreateFromCuts(manager *Llb_Man_t, cut1, cut2 *Vec_Int_t) *Llb_Grp_t {
	group := Llb_ManGroupAlloc(manager)
	firstCut, reached := map[int]bool{}, map[int]bool{}
	for _, id := range vec.Vec_IntArray(cut1) {
		firstCut[int(id)] = true
	}
	graph := manager.pAig
	for _, id := range vec.Vec_IntArray(cut2) {
		edge := aig.Aig_ManObj(graph, id)
		if !firstCut[edge.NodeID()] {
			llbGroupPush(group.vOuts, edge)
		}
	}
	var visit func(aig.Edge)
	visit = func(edge aig.Edge) {
		id := edge.NodeID()
		if firstCut[id] {
			reached[id] = true
			return
		}
		if reached[id] {
			return
		}
		reached[id] = true
		node := graph.Node(aig.Aig_Regular(edge))
		switch node.Kind {
		case aig.Constant, aig.Input:
			return
		case aig.Output, aig.Buffer:
			visit(node.Fanins[0])
		default:
			visit(node.Fanins[0])
			visit(node.Fanins[1])
		}
	}
	for _, id := range vec.Vec_IntArray(cut2) {
		visit(aig.Aig_ManObj(graph, id))
	}
	for _, id := range vec.Vec_IntArray(cut1) {
		if reached[int(id)] {
			llbGroupPush(group.vIns, aig.Aig_ManObj(graph, id))
		}
	}
	group.vNodes = Llb_ManGroupCollect(group)
	return group
}

func Llb_ManPrepareGroups(manager *Llb_Man_t) {
	assert(manager.vGroups == nil, "groups are uninitialized")
	manager.vGroups = vec.Vec_PtrAlloc(1000)
	Llb_ManGroupCreateFirst(manager)
	for _, edge := range manager.pAig.LiveEdges() {
		if manager.pAig.Node(edge).Kind == aig.And && aig.Aig_ObjIsMarkA(manager.pAig, edge) {
			Llb_ManGroupCreate(manager, edge)
		}
	}
	outputs := manager.pAig.Outputs()
	for i := len(outputs) - int(manager.pAig.RegisterCount()); i < len(outputs); i++ {
		if i >= 0 && aig.Aig_ObjIsMarkA(manager.pAig, outputs[i]) {
			Llb_ManGroupCreate(manager, outputs[i])
		}
	}
	Llb_ManGroupCreateLast(manager)
}

func Llb_ManPrintSpan(manager *Llb_Man_t) {
	var span, maximum int32
	for i := 0; i < vec.Vec_PtrSize(manager.vGroups); i++ {
		group, _ := vec.Vec_PtrEntry(manager.vGroups, i).(*Llb_Grp_t)
		if group == nil {
			continue
		}
		for _, entries := range []*Vec_Ptr_t{group.vIns, group.vOuts} {
			for k := 0; k < vec.Vec_PtrSize(entries); k++ {
				edge := llbGroupEdge(entries, k)
				if vec.Vec_IntEntry(manager.vVarBegs, int(aig.Aig_ObjId(edge))) == int32(i) {
					span++
				}
			}
		}
		maximum = max(maximum, span)
		fmt.Printf("%d ", span)
		for _, entries := range []*Vec_Ptr_t{group.vIns, group.vOuts} {
			for k := 0; k < vec.Vec_PtrSize(entries); k++ {
				edge := llbGroupEdge(entries, k)
				if vec.Vec_IntEntry(manager.vVarEnds, int(aig.Aig_ObjId(edge))) == int32(i) {
					span--
				}
			}
		}
	}
	fmt.Printf("\nMax = %d\n", maximum)
}

func Llb_ManGroupHasVar(manager *Llb_Man_t, groupID, variable int32) int32 {
	if manager.vGroups == nil || groupID < 0 || int(groupID) >= vec.Vec_PtrSize(manager.vGroups) {
		return 0
	}
	group, _ := vec.Vec_PtrEntry(manager.vGroups, int(groupID)).(*Llb_Grp_t)
	if group == nil {
		return 0
	}
	for _, entries := range []*Vec_Ptr_t{group.vIns, group.vOuts} {
		for i := 0; i < vec.Vec_PtrSize(entries); i++ {
			if aig.Aig_ObjId(llbGroupEdge(entries, i)) == variable {
				return 1
			}
		}
	}
	return 0
}

func Llb_ManPrintHisto(manager *Llb_Man_t) {
	for _, edge := range manager.pAig.LiveEdges() {
		id := aig.Aig_ObjId(edge)
		if vec.Vec_IntEntry(manager.vObj2Var, int(id)) < 0 {
			continue
		}
		fmt.Printf("%3d :", id)
		for group := int32(0); group < vec.Vec_IntEntry(manager.vVarBegs, int(id)); group++ {
			fmt.Print(" ")
		}
		for group := vec.Vec_IntEntry(manager.vVarBegs, int(id)); group <= vec.Vec_IntEntry(manager.vVarEnds, int(id)); group++ {
			if Llb_ManGroupHasVar(manager, group, id) != 0 {
				fmt.Print("*")
			} else {
				fmt.Print("-")
			}
		}
		fmt.Println()
	}
}
